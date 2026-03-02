package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kowon/naver-news-search-skills/internal/dotenv"
	"github.com/kowon/naver-news-search-skills/internal/exa"
	"github.com/kowon/naver-news-search-skills/internal/naver"
	"github.com/kowon/naver-news-search-skills/internal/notion"
)

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
	fmt.Println("  naver-news search --query <검색어> [--display <1-100>] [--sort <sim|date>] [--fetch] [--highlights] [--highlight-query <검색어>] [--highlight-chars <N>]")
	fmt.Println("  naver-news fetch  --url <URL>")
	fmt.Println("  naver-news notion --parent-id <페이지ID> --title <제목>   (stdin으로 Markdown 입력)")
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "검색어 (필수)")
	display := fs.Int("display", 10, "결과 개수 (1-100, 기본 10)")
	sort := fs.String("sort", "sim", "정렬: sim(정확도순), date(날짜순)")
	fetchContent := fs.Bool("fetch", false, "각 기사 본문을 Exa로 함께 가져오기")
	highlights := fs.Bool("highlights", false, "각 기사 핵심 문장을 Exa highlights로 가져오기")
	highlightQuery := fs.String("highlight-query", "", "highlights 추출 방향 지정 (빈 문자열이면 자동)")
	highlightChars := fs.Int("highlight-chars", 500, "URL당 최대 문자 수")
	fs.Parse(args)

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Error: --query is required")
		os.Exit(1)
	}
	if *display < 1 || *display > 100 {
		fmt.Fprintln(os.Stderr, "Error: --display must be between 1 and 100")
		os.Exit(1)
	}
	if *sort != "sim" && *sort != "date" {
		fmt.Fprintln(os.Stderr, "Error: --sort must be 'sim' or 'date'")
		os.Exit(1)
	}

	items, err := naver.Search(*query, *display, *sort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# 네이버 뉴스 검색 결과: \"%s\"\n\n", *query)
	fmt.Printf("총 %d개 기사\n\n", len(items))

	for i, item := range items {
		fmt.Printf("## %d. %s\n\n", i+1, item.Title)
		fmt.Printf("- **날짜**: %s\n", item.PubDate)
		fmt.Printf("- **원문 링크**: %s\n", item.OriginalLink)
		fmt.Printf("- **네이버 링크**: %s\n", item.Link)
		fmt.Printf("\n%s\n\n", item.Description)

		targetURL := item.OriginalLink
		if targetURL == "" {
			targetURL = item.Link
		}

		if *highlights {
			fmt.Printf("**핵심 내용:**\n")
			snippets, err := exa.FetchHighlights(targetURL, *highlightQuery, *highlightChars)
			if err != nil {
				fmt.Printf("_핵심 내용을 가져오지 못했습니다: %v_\n\n", err)
			} else if len(snippets) == 0 {
				fmt.Printf("_핵심 내용 없음_\n\n")
			} else {
				for j, s := range snippets {
					fmt.Printf("%d. %s\n", j+1, s)
				}
				fmt.Println()
			}
		} else if *fetchContent {
			fmt.Printf("### 기사 본문\n\n")
			content, err := exa.FetchContent(targetURL)
			if err != nil {
				fmt.Printf("_본문을 가져오지 못했습니다: %v_\n\n", err)
			} else {
				fmt.Printf("%s\n\n", content)
			}
		}

		fmt.Println("---")
		fmt.Println()
	}
}

func runFetch(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	rawURL := fs.String("url", "", "기사 URL (필수)")
	fs.Parse(args)

	if *rawURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		os.Exit(1)
	}

	content, err := exa.FetchContent(*rawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# 기사 본문\n\n")
	fmt.Printf("**URL**: %s\n\n", *rawURL)
	fmt.Println("---")
	fmt.Println()
	fmt.Println(content)
}

func runNotion(args []string) {
	fs := flag.NewFlagSet("notion", flag.ExitOnError)
	parentID := fs.String("parent-id", "", "부모 페이지 ID (필수)")
	title := fs.String("title", "", "새 페이지 제목 (필수)")
	fs.Parse(args)

	if *parentID == "" {
		fmt.Fprintln(os.Stderr, "Error: --parent-id is required")
		os.Exit(1)
	}
	if *title == "" {
		fmt.Fprintln(os.Stderr, "Error: --title is required")
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

	blocks := notion.ParseMarkdownToBlocks(string(content))
	pageURL, err := notion.CreatePage(*parentID, *title, blocks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("노션 페이지 생성 완료: %s\n", pageURL)
}
