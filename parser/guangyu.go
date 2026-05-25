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

var httpClient = &http.Client{Timeout: 15 * time.Second}

type GuangyuParser struct{}

func (p *GuangyuParser) SearchBooks(ctx context.Context, keyword string) ([]model.Book, error) {
	return tryAllHosts(ctx, func(baseURL string) ([]model.Book, error) {
		reqURL := baseURL + "/search?" + url.Values{
			"title":           {keyword},
			"tab":             {"小说"},
			"source":          {"番茄"},
			"page":            {"1"},
			"disabled_sources": {"0"},
		}.Encode()

		req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Data []mapItem `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return mapBookList(result.Data), nil
	})
}

func (p *GuangyuParser) GetChapters(ctx context.Context, bookID, innerSource, innerTab string) ([]model.Chapter, error) {
	return tryAllHosts(ctx, func(baseURL string) ([]model.Chapter, error) {
		reqURL := baseURL + "/catalog?" + url.Values{
			"book_id": {bookID},
			"source":  {innerSource},
			"tab":     {innerTab},
		}.Encode()

		req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Data []mapItem `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		chapters := make([]model.Chapter, 0, len(result.Data))
		for _, item := range result.Data {
			chapters = append(chapters, model.Chapter{
				Title:  str(item["title"]),
				ItemID: str(item["item_id"]),
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

		req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/content", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := httpClient.Do(req)
		if err != nil {
			return model.ChapterContent{}, err
		}
		defer resp.Body.Close()

		raw, _ := io.ReadAll(resp.Body)
		var result struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
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

// --- helpers ---

type mapItem map[string]any

func str(v any) string {
	s, _ := v.(string)
	return s
}

func tryAllHosts[T any](ctx context.Context, fn func(string) (T, error)) (T, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val T
		err error
	}
	ch := make(chan result, len(hosts))

	for _, host := range hosts {
		go func(h string) {
			val, err := fn(h)
			select {
			case ch <- result{val, err}:
			case <-ctx.Done():
			}
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

func mapBookList(items []mapItem) []model.Book {
	books := make([]model.Book, 0, len(items))
	for _, item := range items {
		books = append(books, model.Book{
			Title:       cleanBookName(str(item["book_name"])),
			Author:      str(item["author"]),
			Cover:       str(item["thumb_url"]),
			Intro:       str(item["abstract"]),
			Kind:        joinNonEmpty([]string{str(item["status"]), str(item["score"]), str(item["tags"]), str(item["last_chapter_update_time"])}, " / "),
			LastChapter: strings.TrimSpace(str(item["source"]) + " " + str(item["last_chapter_title"])),
			WordCount:   str(item["word_number"]),
			BookID:      str(item["book_id"]),
			SourceKey:   "guangyu",
			Source:      str(item["source"]),
			Tab:         str(item["tab"]),
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

var adPatterns = []*regexp.Regexp{
	regexp.MustCompile(`打赏`),
	regexp.MustCompile(`非\s*[Vv][Ii][Pp]\s*用户`),
	regexp.MustCompile(`[Vv][Ii][Pp]\s*服务器`),
	regexp.MustCompile(`开通\s*[Vv][Ii][Pp]`),
	regexp.MustCompile(`封禁`),
	regexp.MustCompile(`(?i)电报群|t\.me`),
	regexp.MustCompile(`(?i)telegram`),
	regexp.MustCompile(`联系作者`),
	regexp.MustCompile(`后台页面`),
	regexp.MustCompile(`(?i)gmai?l\.com`),
	regexp.MustCompile(`限时折扣`),
	regexp.MustCompile(`恢复原价`),
	regexp.MustCompile(`删除普通账户`),
	regexp.MustCompile(`服务器压力`),
	regexp.MustCompile(`纯净`),
	regexp.MustCompile(`未登录.*访问`),
	regexp.MustCompile(`已访问.*次`),
	regexp.MustCompile(`缓存操作`),
}

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
		skip := false
		for _, p := range adPatterns {
			if p.MatchString(line) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		out = append(out, "<p>"+line+"</p>")
	}
	return strings.Join(out, "\n")
}
