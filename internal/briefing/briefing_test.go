package briefing

import (
	"strings"
	"testing"
)

func TestSplitQueriesTrimsDedupesAndSkipsEmpty(t *testing.T) {
	got := SplitQueries(" AI, 경제,AI, , 반도체 ")
	want := []string{"AI", "경제", "반도체"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDedupeItemsIgnoresQueryStrings(t *testing.T) {
	items := []Item{
		{Title: "A", URL: "https://example.com/news/1?utm_source=a"},
		{Title: "A again", URL: "https://example.com/news/1?utm_source=b"},
		{Title: "B", NaverURL: "https://n.news.naver.com/article/1"},
	}
	got := DedupeItems(items)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %#v", len(got), got)
	}
	if got[0].Title != "A" || got[1].Title != "B" {
		t.Fatalf("unexpected order/content: %#v", got)
	}
}

func TestMarkdownAndPlainTextContainExpectedBriefingFields(t *testing.T) {
	out := Output{
		GeneratedAt: "2026-05-28T09:00:00+09:00",
		Queries:     []string{"AI"},
		Source:      "google",
		Items: []Item{{
			Query:       "AI",
			Source:      "google",
			Title:       "테스트 기사",
			Description: "요약",
			URL:         "https://example.com/a",
			PubDate:     "Thu, 28 May 2026 09:00:00 +0900",
		}},
	}

	md := Markdown(out)
	for _, needle := range []string{"# AI", "[테스트 기사](https://example.com/a)", "- 검색어: AI", "요약"} {
		if !strings.Contains(md, needle) {
			t.Fatalf("markdown missing %q:\n%s", needle, md)
		}
	}

	text := PlainText(out, 1)
	for _, needle := range []string{"뉴스 브리핑: AI", "테스트 기사", "https://example.com/a"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("plain text missing %q:\n%s", needle, text)
		}
	}
}

func TestParseOutputAcceptsLegacySearchJSON(t *testing.T) {
	data := []byte(`{"query":"AI","source":"naver","items":[{"title":"기사","description":"요약","url":"https://example.com"}]}`)
	out, ok := ParseOutput(data)
	if !ok {
		t.Fatal("ParseOutput returned false")
	}
	if len(out.Queries) != 1 || out.Queries[0] != "AI" {
		t.Fatalf("queries = %#v, want [AI]", out.Queries)
	}
	if len(out.Items) != 1 || out.Items[0].Query != "AI" || out.Items[0].Source != "naver" {
		t.Fatalf("items not normalized: %#v", out.Items)
	}
}
