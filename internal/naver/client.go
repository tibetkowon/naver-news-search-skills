package naver

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

// NewsItem represents a single news article returned by Search.
type NewsItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	NaverURL    string `json:"naver_url"`
	PubDate     string `json:"pub_date"`
	Content     string `json:"content,omitempty"`
}

// apiItem matches the raw Naver API response field names.
type apiItem struct {
	Title        string `json:"title"`
	OriginalLink string `json:"originallink"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	PubDate      string `json:"pubDate"`
}

type searchResponse struct {
	Items []apiItem `json:"items"`
}

var boldTagRe = regexp.MustCompile(`</?b>`)

func stripBoldTags(s string) string {
	return boldTagRe.ReplaceAllString(s, "")
}

// Search queries the Naver News Search API and returns a list of news items.
// start is 1-based (Naver API default). Reads NAVER_CLIENT_ID and NAVER_CLIENT_SECRET from env.
func Search(query string, display int, sort string, start int) ([]NewsItem, error) {
	clientID := os.Getenv("NAVER_CLIENT_ID")
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("NAVER_CLIENT_ID and NAVER_CLIENT_SECRET environment variables are required")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("display", strconv.Itoa(display))
	params.Set("sort", sort)
	params.Set("start", strconv.Itoa(start))

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

	items := make([]NewsItem, len(result.Items))
	for i, a := range result.Items {
		items[i] = NewsItem{
			Title:       html.UnescapeString(stripBoldTags(a.Title)),
			Description: html.UnescapeString(stripBoldTags(a.Description)),
			URL:         a.OriginalLink,
			NaverURL:    a.Link,
			PubDate:     a.PubDate,
		}
	}
	return items, nil
}
