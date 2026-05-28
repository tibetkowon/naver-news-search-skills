package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kowon/naver-news-search-skills/internal/dotenv"
	"github.com/kowon/naver-news-search-skills/internal/exa"
	"github.com/kowon/naver-news-search-skills/internal/google"
	"github.com/kowon/naver-news-search-skills/internal/naver"
	"github.com/kowon/naver-news-search-skills/internal/notion"
)

// urlList accepts multiple --url flags.
type urlList []string

func (u *urlList) String() string     { return strings.Join(*u, ", ") }
func (u *urlList) Set(v string) error { *u = append(*u, v); return nil }

// outputItem is the common JSON structure for search results from any source.
type outputItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	NaverURL    string `json:"naver_url"`
	PubDate     string `json:"pub_date"`
	Content     string `json:"content,omitempty"`
}

type searchOutput struct {
	Query  string       `json:"query"`
	Source string       `json:"source"`
	Items  []outputItem `json:"items"`
}

func main() {
	if err := dotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read .env: %v\n", err)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "search":
		runSearch(os.Args[2:])
	case "fetch":
		runFetch(os.Args[2:])
	case "notion":
		runNotion(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  naver-news search --query <검색어> [--display N] [--sort sim|date] [--start N] [--source naver|google] [--fetch] [--format json|markdown]")
	fmt.Println("  naver-news fetch  --url <URL> [--url <URL> ...] [--format json|markdown]")
	fmt.Println("  naver-news notion --parent-id <ID> --title <제목>   (stdin: JSON 또는 Markdown)")
	fmt.Println("  naver-news notion --page-id <ID>                    (기존 페이지에 append)")
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "검색어 (필수)")
	display := fs.Int("display", 10, "결과 개수 (1-100)")
	sort := fs.String("sort", "sim", "정렬: sim(정확도순), date(날짜순) — naver 전용")
	start := fs.Int("start", 1, "시작 위치 1-based — naver 전용")
	source := fs.String("source", "naver", "검색 소스: naver|google")
	fetchContent := fs.Bool("fetch", false, "각 기사 전문을 Exa로 가져오기")
	format := fs.String("format", "json", "출력 형식: json|markdown")
	fs.Parse(args)

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Error: --query is required")
		os.Exit(1)
	}
	if *display < 1 || *display > 100 {
		fmt.Fprintln(os.Stderr, "Error: --display must be between 1 and 100")
		os.Exit(1)
	}
	if *format != "json" && *format != "markdown" {
		fmt.Fprintln(os.Stderr, "Error: --format must be 'json' or 'markdown'")
		os.Exit(1)
	}
	if *source != "naver" && *source != "google" {
		fmt.Fprintln(os.Stderr, "Error: --source must be 'naver' or 'google'")
		os.Exit(1)
	}

	var items []outputItem

	switch *source {
	case "naver":
		if *sort != "sim" && *sort != "date" {
			fmt.Fprintln(os.Stderr, "Error: --sort must be 'sim' or 'date'")
			os.Exit(1)
		}
		naverItems, err := naver.Search(*query, *display, *sort, *start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, ni := range naverItems {
			items = append(items, outputItem{
				Title:       ni.Title,
				Description: ni.Description,
				URL:         ni.URL,
				NaverURL:    ni.NaverURL,
				PubDate:     ni.PubDate,
			})
		}

	case "google":
		googleItems, err := google.Search(*query, *display)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, gi := range googleItems {
			items = append(items, outputItem{
				Title:       gi.Title,
				Description: gi.Description,
				URL:         gi.URL,
				PubDate:     gi.PubDate,
			})
		}
	}

	if *fetchContent && len(items) > 0 {
		urls := make([]string, len(items))
		for i, item := range items {
			if item.URL != "" {
				urls[i] = item.URL
			} else {
				urls[i] = item.NaverURL
			}
		}
		results, err := exa.FetchContents(urls)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch content: %v\n", err)
		} else {
			contentMap := make(map[string]string, len(results))
			for _, r := range results {
				contentMap[r.URL] = r.Content
			}
			for i := range items {
				u := items[i].URL
				if u == "" {
					u = items[i].NaverURL
				}
				if c, ok := contentMap[u]; ok {
					items[i].Content = c
				}
			}
		}
	}

	if *format == "json" {
		out := searchOutput{Query: *query, Source: *source, Items: items}
		if out.Items == nil {
			out.Items = []outputItem{}
		}
		json.NewEncoder(os.Stdout).Encode(out)
		return
	}

	// Markdown output
	fmt.Printf("# 네이버 뉴스 검색 결과: \"%s\"\n\n", *query)
	fmt.Printf("총 %d개 기사\n\n", len(items))
	for i, item := range items {
		fmt.Printf("## %d. %s\n\n", i+1, item.Title)
		fmt.Printf("- **날짜**: %s\n", item.PubDate)
		fmt.Printf("- **원문 링크**: %s\n", item.URL)
		if item.NaverURL != "" {
			fmt.Printf("- **네이버 링크**: %s\n", item.NaverURL)
		}
		fmt.Printf("\n%s\n\n", item.Description)
		if item.Content != "" {
			fmt.Printf("### 기사 본문\n\n%s\n\n", item.Content)
		}
		fmt.Println("---")
		fmt.Println()
	}
}

func runFetch(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	format := fs.String("format", "json", "출력 형식: json|markdown")
	var urls urlList
	fs.Var(&urls, "url", "기사 URL (여러 번 사용 가능)")
	fs.Parse(args)

	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		os.Exit(1)
	}
	if *format != "json" && *format != "markdown" {
		fmt.Fprintln(os.Stderr, "Error: --format must be 'json' or 'markdown'")
		os.Exit(1)
	}

	results, err := exa.FetchContents(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *format == "json" {
		if len(urls) == 1 {
			single := struct {
				URL     string `json:"url"`
				Content string `json:"content"`
			}{URL: urls[0]}
			if len(results) > 0 {
				single.Content = results[0].Content
			}
			json.NewEncoder(os.Stdout).Encode(single)
		} else {
			type multiResult struct {
				Results []exa.ContentResult `json:"results"`
			}
			json.NewEncoder(os.Stdout).Encode(multiResult{Results: results})
		}
		return
	}

	// Markdown output
	for _, r := range results {
		fmt.Printf("# 기사 본문\n\n")
		fmt.Printf("**URL**: %s\n\n", r.URL)
		fmt.Println("---")
		fmt.Println()
		fmt.Println(r.Content)
		fmt.Println()
	}
}

func runNotion(args []string) {
	fs := flag.NewFlagSet("notion", flag.ExitOnError)
	parentID := fs.String("parent-id", "", "부모 페이지 ID (새 페이지 생성 시 필수)")
	title := fs.String("title", "", "새 페이지 제목 (새 페이지 생성 시 필수)")
	pageID := fs.String("page-id", "", "기존 페이지 ID (append 모드)")
	fs.Parse(args)

	if *pageID == "" && (*parentID == "" || *title == "") {
		fmt.Fprintln(os.Stderr, "Error: --page-id 또는 (--parent-id + --title) 중 하나가 필요합니다")
		os.Exit(1)
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
	if len(content) == 0 {
		fmt.Fprintln(os.Stderr, "Error: stdin is empty — pipe search output into this command")
		os.Exit(1)
	}

	var blocks []notion.Block
	trimmed := strings.TrimSpace(string(content))
	if len(trimmed) > 0 && trimmed[0] == '{' {
		blocks, err = notion.ParseSearchJSONToBlocks(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		blocks = notion.ParseMarkdownToBlocks(string(content))
	}

	var pageURL string
	if *pageID != "" {
		pageURL, err = notion.AppendBlocks(*pageID, blocks)
	} else {
		pageURL, err = notion.CreatePage(*parentID, *title, blocks)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *pageID != "" {
		fmt.Printf("노션 페이지 업데이트 완료: %s\n", pageURL)
	} else {
		fmt.Printf("노션 페이지 생성 완료: %s\n", pageURL)
	}
}
