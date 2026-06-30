package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"yueduqi-desktop/cache"
	"yueduqi-desktop/model"
)

var hosts = []string{
	"https://v1.gyks.cf",
	"https://v2.gyks.cf",
	"https://v3.gyks.cf",
	"https://v4.gyks.cf",
	"https://v5.gyks.cf",
	"https://v6.gyks.cf",
	"https://v7.gyks.cf",
}

// Per-operation deadlines, shorter than the client-wide fallback.
const (
	searchTimeout  = 8 * time.Second
	contentTimeout = 15 * time.Second
)

// Shared transport with connection pooling limits to avoid overwhelming
// upstream hosts and to keep idle connections alive across requests.
// httpClient uses 0 Timeout because each operation applies its own
// context.WithTimeout (searchTimeout=8s, contentTimeout=20s). The global
// timeout must be >= the longest per-operation deadline to avoid the
// client-level deadline cutting short a request the context would allow.
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
		MaxConnsPerHost: 5,
	},
}

type GuangyuParser struct{}

func init() {
	Register("guangyu", &GuangyuParser{})
}

func (p *GuangyuParser) SearchBooks(ctx context.Context, keyword string) ([]model.Book, error) {
	// Cache keyed by keyword only — source/tab are hardcoded to 番茄/小说 in this parser.
	if books, ok := cache.Search.Get(keyword); ok {
		return books, nil
	}
	books, err := tryAllHosts(ctx, func(baseURL string) ([]model.Book, error) {
		reqURL := baseURL + "/search?" + url.Values{
			"title":            {keyword},
			"tab":              {"小说"},
			"source":           {"番茄"},
			"page":             {"1"},
			"disabled_sources": {"0"},
		}.Encode()

		// Per-operation deadline so search cannot hang past this window.
		reqCtx, cancel := context.WithTimeout(ctx, searchTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating search request: %w", err)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result searchResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return mapBookList(result.Data), nil
	})
	if err == nil {
		cache.Search.Set(keyword, books)
	}
	return books, err
}

func (p *GuangyuParser) GetChapters(ctx context.Context, bookID, innerSource, innerTab string) ([]model.Chapter, error) {
	return tryAllHosts(ctx, func(baseURL string) ([]model.Chapter, error) {
		reqURL := baseURL + "/catalog?" + url.Values{
			"book_id": {bookID},
			"source":  {innerSource},
			"tab":     {innerTab},
		}.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating catalog request: %w", err)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result catalogResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		chapters := make([]model.Chapter, 0, len(result.Data))
		for _, item := range result.Data {
			chapters = append(chapters, model.Chapter{
				Title:  item.Title,
				ItemID: item.ItemID,
			})
		}
		return chapters, nil
	})
}

func (p *GuangyuParser) GetChapterContent(ctx context.Context, bookID, itemID, innerSource, innerTab string) (model.ChapterContent, error) {
	return tryAllHosts(ctx, func(baseURL string) (model.ChapterContent, error) {
		body := fmt.Sprintf(`html=&item_id=%s&source=%s&tab=%s&tone_id=4&variable=&version=4.11.5.1`,
			url.QueryEscape(itemID),
			url.QueryEscape(innerSource),
			url.QueryEscape(innerTab),
		)

		// Per-operation deadline so content fetch cannot hang past this window.
		reqCtx, cancel := context.WithTimeout(ctx, contentTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, "POST", baseURL+"/content", strings.NewReader(body))
		if err != nil {
			return model.ChapterContent{}, fmt.Errorf("creating content request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := httpClient.Do(req)
		if err != nil {
			return model.ChapterContent{}, err
		}
		defer resp.Body.Close()

		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return model.ChapterContent{}, fmt.Errorf("reading content body: %w", err)
		}
		var result contentResponse
		if err := json.Unmarshal(raw, &result); err != nil {
			return model.ChapterContent{}, err
		}

		if strings.Contains(result.Content, "免登录访问次数已达上限") {
			return model.ChapterContent{}, fmt.Errorf("今日免费阅读次数已用完（每日3次），请明天再试")
		}

		return model.ChapterContent{
			Title:   result.Title,
			Content: cleanContent(result.Content),
		}, nil
	})
}

// --- API response types ---

// searchResponse decodes the /search endpoint JSON. Also shared by
// hot.go's /get_discover because both return the same book-item shape.
type searchResponse struct {
	Data []searchBookItem `json:"data"`
}

type searchBookItem struct {
	BookName              string `json:"book_name"`
	Author                string `json:"author"`
	ThumbURL              string `json:"thumb_url"`
	Abstract              string `json:"abstract"`
	Status                string `json:"status"`
	Score                 string `json:"score"`
	Tags                  string `json:"tags"`
	LastChapterUpdateTime string `json:"last_chapter_update_time"`
	Source                string `json:"source"`
	LastChapterTitle      string `json:"last_chapter_title"`
	WordNumber            string `json:"word_number"`
	BookID                string `json:"book_id"`
	Tab                   string `json:"tab"`
}

type catalogResponse struct {
	Data []catalogItem `json:"data"`
}

type catalogItem struct {
	Title  string `json:"title"`
	ItemID string `json:"item_id"`
}

type contentResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// --- helpers ---

func tryAllHosts[T any](ctx context.Context, fn func(string) (T, error)) (T, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val T
		err error
	}
	// Buffered to capacity so every goroutine can send without blocking,
	// even after we return early on the first success. This avoids the
	// deadlock where cancel() causes remaining goroutines to skip their
	// send, leaving the drain loop hung on a channel that will never fill.
	ch := make(chan result, len(hosts))

	for _, host := range hosts {
		go func(h string) {
			val, err := fn(h)
			ch <- result{val, err}
		}(host)
	}

	var lastErr error
	for range hosts {
		res := <-ch
		if res.err == nil {
			cancel()
			return res.val, nil
		}
		lastErr = res.err
	}
	var zero T
	return zero, lastErr
}

var nameCleanRe = regexp.MustCompile(`[（(]别名[：:].*?[）)]`)

func cleanBookName(name string) string {
	return strings.TrimSpace(nameCleanRe.ReplaceAllString(name, ""))
}

func mapBookList(items []searchBookItem) []model.Book {
	books := make([]model.Book, 0, len(items))
	for _, item := range items {
		books = append(books, model.Book{
			Title:       cleanBookName(item.BookName),
			Author:      item.Author,
			Cover:       item.ThumbURL,
			Intro:       item.Abstract,
			Kind:        joinNonEmpty([]string{item.Status, item.Score, item.Tags, item.LastChapterUpdateTime}, " / "),
			LastChapter: strings.TrimSpace(item.Source + " " + item.LastChapterTitle),
			WordCount:   item.WordNumber,
			BookID:      item.BookID,
			SourceKey:   "guangyu",
			Source:      item.Source,
			Tab:         item.Tab,
		})
	}
	return books
}

func joinNonEmpty(parts []string, sep string) string {
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, sep)
}

// adRe combines the former adPatterns slice into a single alternation.
// Aho-Corasick would be faster for large line counts, but the per-line
// overhead of 18 separate regexp.MatchString calls already dominates;
// a single combined regex trades some readability for a 1-call-per-line
// match and is sufficient for the typical chapter length.
var adRe = regexp.MustCompile(
	`打赏|非\s*[Vv][Ii][Pp]\s*用户|[Vv][Ii][Pp]\s*服务器|开通\s*[Vv][Ii][Pp]|封禁|(?i)电报群|t\.me|(?i)telegram|联系作者|后台页面|(?i)gmai?l\.com|限时折扣|恢复原价|删除普通账户|服务器压力|纯净|未登录.*访问|已访问.*次|缓存操作`,
)

var identRe = regexp.MustCompile(`\s*ident="[^"]*"`)

func cleanContent(content string) string {
	content = identRe.ReplaceAllString(content, "")
	lines := strings.Split(content, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if adRe.MatchString(line) {
			continue
		}
		out = append(out, "<p>"+line+"</p>")
	}
	return strings.Join(out, "\n")
}
