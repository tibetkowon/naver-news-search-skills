package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	notionVersion = "2022-06-28"
	notionAPIBase = "https://api.notion.com/v1"
)

// ---- Notion block types ----

type TextLink struct {
	URL string `json:"url"`
}

type TextContent struct {
	Content string    `json:"content"`
	Link    *TextLink `json:"link,omitempty"`
}

type Annotation struct {
	Bold bool `json:"bold,omitempty"`
}

type RichText struct {
	Type        string      `json:"type"`
	Text        TextContent `json:"text"`
	Annotations *Annotation `json:"annotations,omitempty"`
}

type TextBlock struct {
	RichText []RichText `json:"rich_text"`
}

type Icon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type CalloutBlock struct {
	RichText []RichText `json:"rich_text"`
	Icon     Icon       `json:"icon"`
}

type Block struct {
	Object           string        `json:"object"`
	Type             string        `json:"type"`
	Heading1         *TextBlock    `json:"heading_1,omitempty"`
	Heading2         *TextBlock    `json:"heading_2,omitempty"`
	Paragraph        *TextBlock    `json:"paragraph,omitempty"`
	BulletedListItem *TextBlock    `json:"bulleted_list_item,omitempty"`
	NumberedListItem *TextBlock    `json:"numbered_list_item,omitempty"`
	Divider          *struct{}     `json:"divider,omitempty"`
	Callout          *CalloutBlock `json:"callout,omitempty"`
}

// ---- Category emoji mapping ----

var categoryEmojis = []struct{ keyword, emoji string }{
	{"인공지능", "🤖"}, {"AI", "🤖"},
	{"경제", "💰"}, {"주식", "📈"}, {"금융", "💳"},
	{"정치", "🏛"}, {"국회", "🏛"},
	{"스포츠", "⚽"}, {"축구", "⚽"},
	{"기술", "💻"}, {"과학", "🔬"},
	{"사회", "👥"}, {"문화", "🎭"}, {"연예", "🎬"},
	{"국제", "🌐"}, {"날씨", "🌤"},
}

func categoryEmoji(keyword string) string {
	for _, m := range categoryEmojis {
		if strings.Contains(keyword, m.keyword) {
			return m.emoji
		}
	}
	return "📰"
}

// ---- Rich text helpers ----

func plainRichText(content string) []RichText {
	return []RichText{{Type: "text", Text: TextContent{Content: content}}}
}

func linkedRichText(content, url string) []RichText {
	return []RichText{{Type: "text", Text: TextContent{Content: content, Link: &TextLink{URL: url}}}}
}

// parseRichText splits "**bold** normal **bold2**" into RichText slices.
var boldRe = regexp.MustCompile(`\*\*(.+?)\*\*`)

func parseRichText(text string) []RichText {
	var result []RichText
	last := 0
	matches := boldRe.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		// text before bold
		if m[0] > last {
			result = append(result, RichText{
				Type: "text",
				Text: TextContent{Content: text[last:m[0]]},
			})
		}
		// bold text
		result = append(result, RichText{
			Type:        "text",
			Text:        TextContent{Content: text[m[2]:m[3]]},
			Annotations: &Annotation{Bold: true},
		})
		last = m[1]
	}
	if last < len(text) {
		result = append(result, RichText{
			Type: "text",
			Text: TextContent{Content: text[last:]},
		})
	}
	if len(result) == 0 {
		result = plainRichText(text)
	}
	return result
}

func newBlock(blockType string) Block {
	return Block{Object: "block", Type: blockType}
}

// ---- Markdown patterns ----

var (
	// # 네이버 뉴스 검색 결과: "query"  (search 커맨드 출력)
	h1SearchRe = regexp.MustCompile(`^# 네이버 뉴스 검색 결과: "(.+)"$`)
	// # 일반 제목  (에이전트가 직접 쓴 heading)
	h1Re = regexp.MustCompile(`^# (.+)$`)
	// ## [제목](url)  (에이전트가 링크 포함해서 쓴 heading)
	h2LinkRe = regexp.MustCompile(`^## \[(.+?)\]\((.+?)\)$`)
	// ## N. 제목  or  ## 제목  (search 출력 또는 에이전트가 쓴 heading)
	h2Re = regexp.MustCompile(`^## (?:\d+\. )?(.+)$`)
	// - **원문 링크**: URL
	originalLinkRe = regexp.MustCompile(`^- \*\*원문 링크\*\*: (.+)$`)
	// - **네이버 링크**: URL
	naverLinkRe = regexp.MustCompile(`^- \*\*네이버 링크\*\*: (.+)$`)
	// - **날짜**: ...
	dateLinkRe = regexp.MustCompile(`^- \*\*날짜\*\*:`)
	// N. text (numbered list)
	numberedRe = regexp.MustCompile(`^(\d+)\. (.+)$`)
	// 총 N개 기사
	totalRe = regexp.MustCompile(`^총 \d+개 기사$`)
)

// ParseMarkdownToBlocks converts the Markdown output of the search command into
// a slice of Notion blocks. It is stateful: article titles and URLs are buffered
// until a blank line triggers block emission with the hyperlink applied.
func ParseMarkdownToBlocks(markdown string) []Block {
	lines := strings.Split(markdown, "\n")
	var blocks []Block

	pendingTitle := ""
	pendingURL := ""

	flushPending := func() {
		if pendingTitle == "" {
			return
		}
		b := newBlock("heading_2")
		if pendingURL != "" {
			b.Heading2 = &TextBlock{RichText: linkedRichText(pendingTitle, pendingURL)}
		} else {
			b.Heading2 = &TextBlock{RichText: plainRichText(pendingTitle)}
		}
		blocks = append(blocks, b)
		pendingTitle = ""
		pendingURL = ""
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t")

		// blank line: skip entirely — do NOT flush here.
		// Titles are buffered until actual content arrives, so that the URL
		// lines that follow ## headings are captured before heading_2 is emitted.
		if line == "" {
			continue
		}

		// skip "총 N개 기사"
		if totalRe.MatchString(line) {
			continue
		}

		// # 네이버 뉴스 검색 결과: "query" → heading_1 with category emoji mapping
		if m := h1SearchRe.FindStringSubmatch(line); m != nil {
			flushPending()
			keyword := m[1]
			emoji := categoryEmoji(keyword)
			b := newBlock("heading_1")
			b.Heading1 = &TextBlock{RichText: plainRichText(emoji + " " + keyword)}
			blocks = append(blocks, b)
			continue
		}

		// # 일반 heading → heading_1 as-is (에이전트가 직접 쓴 경우, 이모지 포함)
		if m := h1Re.FindStringSubmatch(line); m != nil {
			flushPending()
			b := newBlock("heading_1")
			b.Heading1 = &TextBlock{RichText: plainRichText(m[1])}
			blocks = append(blocks, b)
			continue
		}

		// ## [제목](url) → heading_2 with link (에이전트가 링크 포함해서 쓴 경우)
		if m := h2LinkRe.FindStringSubmatch(line); m != nil {
			flushPending()
			pendingTitle = m[1]
			pendingURL = m[2]
			continue
		}

		// ## N. 제목  or  ## 제목 → buffer title; reset URL
		if m := h2Re.FindStringSubmatch(line); m != nil {
			flushPending()
			pendingTitle = m[1]
			pendingURL = ""
			continue
		}

		// skip date lines
		if dateLinkRe.MatchString(line) {
			continue
		}

		// 원문 링크 → store URL
		if m := originalLinkRe.FindStringSubmatch(line); m != nil {
			pendingURL = strings.TrimSpace(m[1])
			continue
		}

		// 네이버 링크 → store only if no original link yet
		if m := naverLinkRe.FindStringSubmatch(line); m != nil {
			if pendingURL == "" {
				pendingURL = strings.TrimSpace(m[1])
			}
			continue
		}

		// Any other non-empty, non-metadata line: flush pending title first
		flushPending()

		// divider
		if line == "---" {
			b := newBlock("divider")
			b.Divider = &struct{}{}
			blocks = append(blocks, b)
			continue
		}

		// numbered list item: N. text
		if m := numberedRe.FindStringSubmatch(line); m != nil {
			b := newBlock("numbered_list_item")
			b.NumberedListItem = &TextBlock{RichText: parseRichText(m[2])}
			blocks = append(blocks, b)
			continue
		}

		// default: paragraph
		b := newBlock("paragraph")
		b.Paragraph = &TextBlock{RichText: parseRichText(line)}
		blocks = append(blocks, b)
	}

	flushPending()
	return blocks
}

// ---- Notion HTTP helper ----

func notionRequest(apiKey, method, url string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Notion-Version", notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Notion API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Notion API returned status %d: %s", resp.StatusCode, string(respBytes))
	}
	return respBytes, nil
}

// ---- Page creation ----

type pageParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type pageTitleProp struct {
	Title []RichText `json:"title"`
}

type createPageRequest struct {
	Parent     pageParent               `json:"parent"`
	Properties map[string]pageTitleProp `json:"properties"`
	Children   []Block                  `json:"children,omitempty"`
}

type appendBlocksRequest struct {
	Children []Block `json:"children"`
}

// CreatePage creates a new Notion page under parentPageID with the given title
// and appends all blocks. Blocks are sent in batches of 100 (Notion API limit).
// It reads NOTION_API_KEY from the environment.
func CreatePage(parentPageID, title string, blocks []Block) (string, error) {
	apiKey := os.Getenv("NOTION_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NOTION_API_KEY environment variable is required")
	}

	allBlocks := blocks

	const batchSize = 100
	firstBatch := allBlocks
	if len(firstBatch) > batchSize {
		firstBatch = allBlocks[:batchSize]
	}

	payload := createPageRequest{
		Parent: pageParent{Type: "page_id", PageID: parentPageID},
		Properties: map[string]pageTitleProp{
			"title": {Title: plainRichText(title)},
		},
		Children: firstBatch,
	}

	respBytes, err := notionRequest(apiKey, http.MethodPost, notionAPIBase+"/pages", payload)
	if err != nil {
		return "", err
	}

	var pageResp struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBytes, &pageResp); err != nil {
		return "", fmt.Errorf("parsing page response: %w", err)
	}

	// Append remaining blocks in batches
	for i := batchSize; i < len(allBlocks); i += batchSize {
		end := i + batchSize
		if end > len(allBlocks) {
			end = len(allBlocks)
		}
		batch := appendBlocksRequest{Children: allBlocks[i:end]}
		patchURL := notionAPIBase + "/blocks/" + pageResp.ID + "/children"
		if _, err := notionRequest(apiKey, http.MethodPatch, patchURL, batch); err != nil {
			return "", fmt.Errorf("appending blocks (batch starting at %s): %w", strconv.Itoa(i), err)
		}
	}

	return pageResp.URL, nil
}
