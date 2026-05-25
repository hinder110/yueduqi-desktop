package parser

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"yueduqi-desktop/model"
)

const qixingeBase = "http://www.qixinge.net"

type QixingeParser struct{}

func (p *QixingeParser) SearchBooks(ctx context.Context, keyword string) ([]model.Book, error) {
	reqURL := qixingeBase + "/search.php?q=" + url.QueryEscape(keyword) + "&p=1"
	req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	req.Header.Set("Referer", qixingeBase+"/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var books []model.Book
	doc.Find(".col-md-6 dl").Each(func(_ int, el *goquery.Selection) {
		coverImg, _ := el.Find("dt img").Attr("src")
		nameLink := el.Find("h3 a")
		name := nameLink.Text()
		name = strings.ReplaceAll(name, "免费阅读小说", "")
		name = stripBrackets(name)
		name = strings.TrimSpace(name)

		href, _ := nameLink.Attr("href")
		if name == "" || href == "" {
			return
		}

		bookOthers := el.Find(".book_other")
		author := strings.TrimSpace(bookOthers.Eq(0).Find("span").First().Text())

		kind := ""
		bookOthers.Eq(1).Each(func(_ int, s *goquery.Selection) {
			kind = strings.TrimSpace(strings.Replace(s.Text(), ".*：", "", 1))
		})

		lastChapter := ""
		bookOthers.Eq(3).Find("a").Each(func(_ int, s *goquery.Selection) {
			lastChapter = strings.TrimSpace(s.Text())
		})

		books = append(books, model.Book{
			Title:       name,
			Author:      author,
			Cover:       toAbsURL(coverImg, qixingeBase),
			Kind:        kind,
			LastChapter: lastChapter,
			BookID:      toAbsURL(href, qixingeBase),
			SourceKey:   "qixinge",
			Source:      "qixinge",
		})
	})
	return books, nil
}

func (p *QixingeParser) GetChapters(ctx context.Context, bookID, _, _ string) ([]model.Chapter, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", bookID, nil)
	req.Header.Set("Referer", qixingeBase+"/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var chapters []model.Chapter
	doc.Find(".book_list2 li a").Each(func(_ int, el *goquery.Selection) {
		href, _ := el.Attr("href")
		title := strings.TrimSpace(el.Text())
		if href == "" || title == "" {
			return
		}
		chapters = append(chapters, model.Chapter{
			Title:  title,
			ItemID: toAbsURL(href, qixingeBase),
		})
	})
	return chapters, nil
}

func (p *QixingeParser) GetChapterContent(ctx context.Context, _, itemID, _, _ string) (model.ChapterContent, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", itemID, nil)
	req.Header.Set("Referer", qixingeBase+"/")
	resp, err := httpClient.Do(req)
	if err != nil {
		return model.ChapterContent{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return model.ChapterContent{}, err
	}

	rawTitle := strings.TrimSpace(doc.Find("h1").First().Text())
	title := strings.TrimSpace(strings.Replace(rawTitle, "-《.*》", "", 1))

	article := doc.Find("article.font_max")
	article.Find("script, style, div, a").Remove()
	article.Find("br").ReplaceWithHtml("\n")
	text := article.Text()

	text = strings.ReplaceAll(text, "本章未完", "")
	text = strings.TrimSpace(text)

	return model.ChapterContent{
		Title:   title,
		Content: cleanContent(text),
	}, nil
}

func stripBrackets(s string) string {
	// remove leading [xxx] prefix
	for strings.HasPrefix(s, "[") {
		end := strings.Index(s, "]")
		if end < 0 {
			break
		}
		s = s[end+1:]
	}
	return s
}
