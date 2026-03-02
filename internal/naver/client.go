package naver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

// NewsItem represents a single news article from the Naver News API.
type NewsItem struct {
	Title        string `json:"title"`
	OriginalLink string `json:"originallink"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	PubDate      string `json:"pubDate"`
}

type searchResponse struct {
	Items []NewsItem `json:"items"`
}

var boldTagRe = regexp.MustCompile(`</?b>`)

func stripBoldTags(s string) string {
	return boldTagRe.ReplaceAllString(s, "")
}

// Search queries the Naver News Search API and returns a list of news items.
// It reads NAVER_CLIENT_ID and NAVER_CLIENT_SECRET from environment variables.
func Search(query string, display int, sort string) ([]NewsItem, error) {
	clientID := os.Getenv("NAVER_CLIENT_ID")
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("NAVER_CLIENT_ID and NAVER_CLIENT_SECRET environment variables are required")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("display", strconv.Itoa(display))
	params.Set("sort", sort)

	reqURL := "https://openapi.naver.com/v1/search/news.json?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Naver-Client-Id", clientID)
	req.Header.Set("X-Naver-Client-Secret", clientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Naver API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Naver API returned status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	for i := range result.Items {
		result.Items[i].Title = stripBoldTags(result.Items[i].Title)
		result.Items[i].Description = stripBoldTags(result.Items[i].Description)
	}

	return result.Items, nil
}
