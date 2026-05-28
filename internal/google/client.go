package google

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
)

// NewsItem represents a single news article from Google News RSS.
// Fields match naver.NewsItem for consistent JSON output.
type NewsItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	NaverURL    string `json:"naver_url"` // always empty for Google source
	PubDate     string `json:"pub_date"`
}

// rssItem is the internal XML structure for a Google News RSS <item>.
// Google News RSS 2.0 uses a plain text <link> element.
// Using a custom type to handle the link element which sits between two tags.
type rssLink struct {
	Value string `xml:",chardata"`
}

type rssItem struct {
	Title       string `xml:"title"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	// Link in RSS 2.0 is a text node; captured via chardata on a wrapper type.
	// Fallback: source url attribute if link is empty.
	Link   rssLink   `xml:"link"`
	Source rssSource `xml:"source"`
}

type rssSource struct {
	URL string `xml:"url,attr"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

// Search queries Google News RSS for Korean news and returns up to display items.
// No API key required.
func Search(query string, display int) ([]NewsItem, error) {
	feedURL := "https://news.google.com/rss/search?q=" +
		url.QueryEscape(query) + "&hl=ko&gl=KR&ceid=KR:ko"

	req, err := http.NewRequest(http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; naver-news-cli/2.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Google News RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google News RSS returned status %d", resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decoding RSS: %w", err)
	}

	raw := feed.Channel.Items
	if display < len(raw) {
		raw = raw[:display]
	}

	items := make([]NewsItem, len(raw))
	for i, r := range raw {
		link := r.Link.Value
		if link == "" {
			link = r.Source.URL
		}
		items[i] = NewsItem{
			Title:       html.UnescapeString(r.Title),
			Description: html.UnescapeString(r.Description),
			URL:         link,
			PubDate:     r.PubDate,
		}
	}
	return items, nil
}
