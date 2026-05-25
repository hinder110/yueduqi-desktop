package parser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"yueduqi-desktop/model"
)

func GetHotBooks(ctx context.Context) ([]model.Book, error) {
	return tryAllHosts(ctx, func(baseURL string) ([]model.Book, error) {
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

		var result struct {
			Data []mapItem `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		data := result.Data
		if len(data) > 12 {
			data = data[:12]
		}
		return mapBookList(data), nil
	})
}
