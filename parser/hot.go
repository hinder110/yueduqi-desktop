package parser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"yueduqi-desktop/cache"
	"yueduqi-desktop/model"
)

// maxHotBooks caps the discover/hot list to avoid overwhelming the UI.
const maxHotBooks = 12

func GetHotBooks(ctx context.Context) ([]model.Book, error) {
	// Single-key cache: the hot list has no parameters, so one entry covers all callers.
	if books, ok := cache.HotBooks.Get("hot"); ok {
		return books, nil
	}
	books, err := tryAllHosts(ctx, func(baseURL string) ([]model.Book, error) {
		reqURL := baseURL + "/get_discover?" + url.Values{
			"source":     {"番茄"},
			"tab":        {"小说"},
			"bdtype":     {"热搜榜"},
			"gender":     {"1"},
			"is_ranking": {"1"},
			"page":       {"1"},
		}.Encode()

		req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result searchResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		data := result.Data
		if len(data) > maxHotBooks {
			data = data[:maxHotBooks]
		}
		return mapBookList(data), nil
	})
	if err == nil {
		cache.HotBooks.Set("hot", books)
	}
	return books, err
}
