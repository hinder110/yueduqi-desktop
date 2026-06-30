package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"yueduqi-desktop/model"
)

const biqugeBase = "http://m.biquge900.com"

// gbkEnc is cached at package init so charset.Lookup("gbk") is not
// repeated on every gbkEncode call.
var gbkEnc encoding.Encoding

func init() {
	gbkEnc, _ = charset.Lookup("gbk")
	Register("biquge900", &BiqugeParser{})
}

type BiqugeParser struct{}

func (p *BiqugeParser) SearchBooks(ctx context.Context, keyword string) ([]model.Book, error) {
	body := fmt.Sprintf("searchkey=%s&t=1", gbkEncode(keyword))

	req, _ := http.NewRequestWithContext(ctx, "POST", biqugeBase+"/modules/article/search.php", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", biqugeBase+"/")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(utf8Reader(resp.Body))
	if err != nil {
		return nil, err
	}

	var books []model.Book
	doc.Find(".hot_sale").Each(func(_ int, el *goquery.Selection) {
		a := el.Find("a").First()
		href, _ := a.Attr("href")
		if href == "" {
			return
		}
		name := strings.TrimSpace(a.Find(".title").Text())
		if name == "" {
			name = strings.TrimSpace(a.Find("p").First().Text())
		}
		if name == "" {
			return
		}
		books = append(books, model.Book{
			Title:     name,
			Author:    strings.TrimSpace(a.Find(".author").Text()),
			Kind:      strings.TrimSpace(a.Find(".review").Text()),
			BookID:    toAbsURL(href, biqugeBase),
			SourceKey: "biquge900",
			Source:    "biquge900",
		})
	})
	return books, nil
}

func (p *BiqugeParser) GetChapters(ctx context.Context, bookID, _, _ string) ([]model.Chapter, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", bookID, nil)
	req.Header.Set("Referer", biqugeBase+"/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(utf8Reader(resp.Body))
	if err != nil {
		return nil, err
	}

	var chapters []model.Chapter
	doc.Find(".directoryArea p").Each(func(_ int, el *goquery.Selection) {
		a := el.Find("a")
		href, _ := a.Attr("href")
		title := strings.TrimSpace(a.Text())
		if href == "" || title == "" {
			return
		}
		chapters = append(chapters, model.Chapter{
			Title:  title,
			ItemID: toAbsURL(href, biqugeBase),
		})
	})
	return chapters, nil
}

func (p *BiqugeParser) GetChapterContent(ctx context.Context, _, itemID, _, _ string) (model.ChapterContent, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", itemID, nil)
	req.Header.Set("Referer", biqugeBase+"/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return model.ChapterContent{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(utf8Reader(resp.Body))
	if err != nil {
		return model.ChapterContent{}, err
	}

	title := strings.TrimSpace(doc.Find(".title").First().Text())

	chapterDiv := doc.Find("#chaptercontent")
	chapterDiv.Find("script, style, div, a").Remove()
	chapterDiv.Find("br").ReplaceWithHtml("\n")
	raw := chapterDiv.Text()

	raw = strings.NewReplacer(
		"笔趣阁最新域名：", "",
		"，请牢记本域名并相互转告！", "",
	).Replace(raw)
	raw = strings.TrimSpace(raw)

	return model.ChapterContent{
		Title:   title,
		Content: cleanContent(raw),
	}, nil
}

// --- helpers ---

func gbkEncode(s string) string {
	if gbkEnc == nil {
		return s
	}
	var buf strings.Builder
	w := transform.NewWriter(&buf, gbkEnc.NewEncoder())
	io.WriteString(w, s)
	w.Close()
	return buf.String()
}

// utf8Reader wraps r in a streaming charset → UTF-8 decoder.
// charset.NewReader peeks only the first ~1 KB to sniff the encoding;
// the returned reader streams the decoded output without buffering the
// entire body in memory.
func utf8Reader(r io.Reader) io.Reader {
	reader, err := charset.NewReader(r, "")
	if err != nil {
		return r
	}
	return reader
}

func toAbsURL(path, base string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http") {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}
