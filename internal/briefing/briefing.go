package briefing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Item struct {
	Query       string `json:"query"`
	Source      string `json:"source"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	NaverURL    string `json:"naver_url,omitempty"`
	PubDate     string `json:"pub_date"`
}

type Output struct {
	GeneratedAt string   `json:"generated_at"`
	Queries     []string `json:"queries"`
	Source      string   `json:"source"`
	Items       []Item   `json:"items"`
}

func NewOutput(queries []string, source string, items []Item) Output {
	return Output{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Queries:     queries,
		Source:      source,
		Items:       items,
	}
}

func SplitQueries(raw string) []string {
	seen := map[string]bool{}
	var queries []string
	for _, part := range strings.Split(raw, ",") {
		q := strings.TrimSpace(part)
		if q == "" || seen[q] {
			continue
		}
		seen[q] = true
		queries = append(queries, q)
	}
	return queries
}

func DedupeItems(items []Item) []Item {
	seen := map[string]bool{}
	result := make([]Item, 0, len(items))
	for _, item := range items {
		key := itemKey(item)
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(item.Title))
		}
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func SortByQueryOrder(items []Item, queries []string) {
	order := map[string]int{}
	for i, q := range queries {
		order[q] = i
	}
	sort.SliceStable(items, func(i, j int) bool {
		return order[items[i].Query] < order[items[j].Query]
	})
}

func WriteJSON(w io.Writer, out Output) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func Markdown(out Output) string {
	var b strings.Builder
	title := "뉴스 브리핑"
	if len(out.Queries) > 0 {
		title = strings.Join(out.Queries, ", ")
	}
	fmt.Fprintf(&b, "# %s\n\n", title)
	if out.GeneratedAt != "" {
		fmt.Fprintf(&b, "- 생성 시각: %s\n", out.GeneratedAt)
	}
	if out.Source != "" {
		fmt.Fprintf(&b, "- 소스: %s\n", out.Source)
	}
	fmt.Fprintf(&b, "- 기사 수: %d\n\n", len(out.Items))

	for i, item := range out.Items {
		link := item.URL
		if link == "" {
			link = item.NaverURL
		}
		if link != "" {
			fmt.Fprintf(&b, "## %d. [%s](%s)\n\n", i+1, item.Title, link)
		} else {
			fmt.Fprintf(&b, "## %d. %s\n\n", i+1, item.Title)
		}
		if item.Query != "" {
			fmt.Fprintf(&b, "- 검색어: %s\n", item.Query)
		}
		if item.Source != "" {
			fmt.Fprintf(&b, "- 출처: %s\n", item.Source)
		}
		if item.PubDate != "" {
			fmt.Fprintf(&b, "- 날짜: %s\n", item.PubDate)
		}
		if item.NaverURL != "" && item.NaverURL != link {
			fmt.Fprintf(&b, "- 네이버 링크: %s\n", item.NaverURL)
		}
		if item.Description != "" {
			fmt.Fprintf(&b, "\n%s\n\n", item.Description)
		} else {
			b.WriteString("\n")
		}
		b.WriteString("---\n\n")
	}
	return b.String()
}

func PlainText(out Output, maxItems int) string {
	items := out.Items
	if maxItems > 0 && maxItems < len(items) {
		items = items[:maxItems]
	}
	var b strings.Builder
	title := "뉴스 브리핑"
	if len(out.Queries) > 0 {
		title += ": " + strings.Join(out.Queries, ", ")
	}
	fmt.Fprintf(&b, "%s\n", title)
	for i, item := range items {
		link := item.URL
		if link == "" {
			link = item.NaverURL
		}
		fmt.Fprintf(&b, "%d. %s", i+1, item.Title)
		if item.Query != "" {
			fmt.Fprintf(&b, " [%s]", item.Query)
		}
		if link != "" {
			fmt.Fprintf(&b, "\n%s", link)
		}
		if item.Description != "" {
			fmt.Fprintf(&b, "\n%s", item.Description)
		}
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func ParseOutput(data []byte) (Output, bool) {
	var raw struct {
		GeneratedAt string   `json:"generated_at"`
		Queries     []string `json:"queries"`
		Query       string   `json:"query"`
		Source      string   `json:"source"`
		Items       []Item   `json:"items"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Output{}, false
	}
	if raw.GeneratedAt == "" && len(raw.Queries) == 0 && raw.Query == "" {
		return Output{}, false
	}
	queries := raw.Queries
	if len(queries) == 0 && raw.Query != "" {
		queries = []string{raw.Query}
	}
	items := raw.Items
	for i := range items {
		if items[i].Query == "" && raw.Query != "" {
			items[i].Query = raw.Query
		}
		if items[i].Source == "" && raw.Source != "" {
			items[i].Source = raw.Source
		}
	}
	return Output{
		GeneratedAt: raw.GeneratedAt,
		Queries:     queries,
		Source:      raw.Source,
		Items:       items,
	}, true
}

func itemKey(item Item) string {
	for _, candidate := range []string{item.URL, item.NaverURL} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		u, err := url.Parse(candidate)
		if err != nil {
			return strings.ToLower(candidate)
		}
		u.RawQuery = ""
		u.Fragment = ""
		return strings.ToLower(u.String())
	}
	return ""
}
