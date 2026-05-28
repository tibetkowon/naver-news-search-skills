package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kowon/naver-news-search-skills/internal/briefing"
	"github.com/kowon/naver-news-search-skills/internal/dotenv"
	"github.com/kowon/naver-news-search-skills/internal/google"
	"github.com/kowon/naver-news-search-skills/internal/naver"
	"github.com/kowon/naver-news-search-skills/internal/notion"
	"github.com/kowon/naver-news-search-skills/internal/webhook"
)

// outputItem is the common JSON structure for search results from any source.
type outputItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	NaverURL    string `json:"naver_url"`
	PubDate     string `json:"pub_date"`
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
	case "brief":
		runBrief(os.Args[2:])
	case "notion":
		runNotion(os.Args[2:])
	case "publish":
		runPublish(os.Args[2:])
	case "run":
		runRun(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  naver-news search --query <검색어> [--display N] [--sort sim|date] [--start N] [--source naver|google] [--format json|markdown]")
	fmt.Println("  naver-news brief --queries <검색어,검색어> [--source auto|naver|google] [--format json|markdown]")
	fmt.Println("  naver-news publish --target slack|discord|webhook|notion [target options]   (stdin: JSON 또는 Markdown)")
	fmt.Println("  naver-news run --queries <검색어,검색어> --target stdout|slack|discord|webhook|notion [--interval 30m]")
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
		fmt.Println("---")
		fmt.Println()
	}
}

func runBrief(args []string) {
	fs := flag.NewFlagSet("brief", flag.ExitOnError)
	queriesRaw := fs.String("queries", "", "쉼표로 구분한 검색어 목록 (필수)")
	display := fs.Int("display", 5, "검색어별 결과 개수 (1-100)")
	source := fs.String("source", "auto", "검색 소스: auto|naver|google")
	sort := fs.String("sort", "date", "정렬: sim(정확도순), date(날짜순) — naver 전용")
	format := fs.String("format", "json", "출력 형식: json|markdown")
	concurrency := fs.Int("concurrency", 4, "동시 검색 개수")
	dedupe := fs.Bool("dedupe", true, "URL 기준 중복 기사 제거")
	fs.Parse(args)

	queries := briefing.SplitQueries(*queriesRaw)
	if len(queries) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --queries is required")
		os.Exit(1)
	}
	if err := validateBriefFlags(*display, *source, *sort, *format, *concurrency); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out, err := collectBriefing(queries, *source, *display, *sort, *concurrency, *dedupe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	writeBriefingOutput(out, *format)
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

func runPublish(args []string) {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	target := fs.String("target", "", "발행 대상: slack|discord|webhook|notion")
	webhookURL := fs.String("webhook-url", "", "Slack/Discord/Generic Webhook URL")
	parentID := fs.String("parent-id", "", "Notion 부모 페이지 ID (새 페이지 생성)")
	title := fs.String("title", "뉴스 브리핑", "Notion 새 페이지 제목")
	pageID := fs.String("page-id", "", "Notion 기존 페이지 ID (append)")
	maxItems := fs.Int("max-items", 20, "채널 메시지에 포함할 최대 기사 수")
	fs.Parse(args)

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		fmt.Fprintln(os.Stderr, "Error: stdin is empty")
		os.Exit(1)
	}

	if err := publishContent(*target, content, *webhookURL, *parentID, *title, *pageID, *maxItems); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	queriesRaw := fs.String("queries", "", "쉼표로 구분한 검색어 목록 (필수)")
	display := fs.Int("display", 5, "검색어별 결과 개수 (1-100)")
	source := fs.String("source", "auto", "검색 소스: auto|naver|google")
	sort := fs.String("sort", "date", "정렬: sim(정확도순), date(날짜순) — naver 전용")
	format := fs.String("format", "markdown", "stdout 출력 형식: json|markdown")
	concurrency := fs.Int("concurrency", 4, "동시 검색 개수")
	dedupe := fs.Bool("dedupe", true, "URL 기준 중복 기사 제거")
	target := fs.String("target", "stdout", "발행 대상: stdout|slack|discord|webhook|notion")
	webhookURL := fs.String("webhook-url", "", "Slack/Discord/Generic Webhook URL")
	parentID := fs.String("parent-id", "", "Notion 부모 페이지 ID (새 페이지 생성)")
	title := fs.String("title", "뉴스 브리핑", "Notion 새 페이지 제목")
	pageID := fs.String("page-id", "", "Notion 기존 페이지 ID (append)")
	intervalRaw := fs.String("interval", "", "반복 주기 예: 30m, 2h. 비우면 한 번만 실행")
	fs.Parse(args)

	queries := briefing.SplitQueries(*queriesRaw)
	if len(queries) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --queries is required")
		os.Exit(1)
	}
	if err := validateBriefFlags(*display, *source, *sort, *format, *concurrency); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var interval time.Duration
	if *intervalRaw != "" {
		var err error
		interval, err = time.ParseDuration(*intervalRaw)
		if err != nil || interval <= 0 {
			fmt.Fprintln(os.Stderr, "Error: --interval must be a positive duration such as 30m or 2h")
			os.Exit(1)
		}
	}

	for {
		out, err := collectBriefing(queries, *source, *display, *sort, *concurrency, *dedupe)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if interval == 0 {
				os.Exit(1)
			}
		} else if *target == "stdout" {
			writeBriefingOutput(out, *format)
		} else {
			data, err := json.Marshal(out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling briefing: %v\n", err)
				os.Exit(1)
			}
			if err := publishContent(*target, data, *webhookURL, *parentID, *title, *pageID, 20); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				if interval == 0 {
					os.Exit(1)
				}
			}
		}
		if interval == 0 {
			return
		}
		time.Sleep(interval)
	}
}

func validateBriefFlags(display int, source, sort, format string, concurrency int) error {
	if display < 1 || display > 100 {
		return fmt.Errorf("--display must be between 1 and 100")
	}
	if source != "auto" && source != "naver" && source != "google" {
		return fmt.Errorf("--source must be 'auto', 'naver', or 'google'")
	}
	if sort != "sim" && sort != "date" {
		return fmt.Errorf("--sort must be 'sim' or 'date'")
	}
	if format != "json" && format != "markdown" {
		return fmt.Errorf("--format must be 'json' or 'markdown'")
	}
	if concurrency < 1 || concurrency > 20 {
		return fmt.Errorf("--concurrency must be between 1 and 20")
	}
	return nil
}

func collectBriefing(queries []string, source string, display int, sort string, concurrency int, dedupe bool) (briefing.Output, error) {
	type result struct {
		items []briefing.Item
		err   error
	}

	sem := make(chan struct{}, concurrency)
	results := make(chan result, len(queries))
	var wg sync.WaitGroup

	for _, query := range queries {
		query := query
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			items, err := searchBriefingItems(query, source, display, sort)
			results <- result{items: items, err: err}
		}()
	}

	wg.Wait()
	close(results)

	var items []briefing.Item
	var errs []string
	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err.Error())
			continue
		}
		items = append(items, r.items...)
	}
	if len(items) == 0 && len(errs) > 0 {
		return briefing.Output{}, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	if dedupe {
		items = briefing.DedupeItems(items)
	}
	briefing.SortByQueryOrder(items, queries)
	return briefing.NewOutput(queries, source, items), nil
}

func searchBriefingItems(query, source string, display int, sort string) ([]briefing.Item, error) {
	switch source {
	case "naver":
		return searchNaverBriefingItems(query, display, sort)
	case "google":
		return searchGoogleBriefingItems(query, display)
	case "auto":
		items, err := searchNaverBriefingItems(query, display, sort)
		if err == nil {
			return items, nil
		}
		return searchGoogleBriefingItems(query, display)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}

func searchNaverBriefingItems(query string, display int, sort string) ([]briefing.Item, error) {
	items, err := naver.Search(query, display, sort, 1)
	if err != nil {
		return nil, fmt.Errorf("%s naver: %w", query, err)
	}
	out := make([]briefing.Item, 0, len(items))
	for _, item := range items {
		out = append(out, briefing.Item{
			Query:       query,
			Source:      "naver",
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			NaverURL:    item.NaverURL,
			PubDate:     item.PubDate,
		})
	}
	return out, nil
}

func searchGoogleBriefingItems(query string, display int) ([]briefing.Item, error) {
	items, err := google.Search(query, display)
	if err != nil {
		return nil, fmt.Errorf("%s google: %w", query, err)
	}
	out := make([]briefing.Item, 0, len(items))
	for _, item := range items {
		out = append(out, briefing.Item{
			Query:       query,
			Source:      "google",
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			NaverURL:    item.NaverURL,
			PubDate:     item.PubDate,
		})
	}
	return out, nil
}

func writeBriefingOutput(out briefing.Output, format string) {
	if format == "json" {
		if err := briefing.WriteJSON(os.Stdout, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}
	fmt.Print(briefing.Markdown(out))
}

func publishContent(target string, content []byte, webhookURL, parentID, title, pageID string, maxItems int) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Errorf("--target is required")
	}

	switch target {
	case "slack", "discord", "webhook":
		if webhookURL == "" {
			webhookURL = webhookURLFromEnv(target)
		}
		text := strings.TrimSpace(string(content))
		if out, ok := briefing.ParseOutput(content); ok {
			text = briefing.PlainText(out, maxItems)
		}
		if len([]rune(text)) > 3500 {
			runes := []rune(text)
			text = string(runes[:3500]) + "\n...(truncated)"
		}
		return webhook.Send(webhook.Target(target), webhookURL, text)
	case "notion":
		if pageID == "" && parentID == "" {
			return fmt.Errorf("--page-id or --parent-id is required for notion target")
		}
		blocks, err := blocksFromContent(content)
		if err != nil {
			return err
		}
		var pageURL string
		if pageID != "" {
			pageURL, err = notion.AppendBlocks(pageID, blocks)
		} else {
			pageURL, err = notion.CreatePage(parentID, title, blocks)
		}
		if err != nil {
			return err
		}
		fmt.Printf("노션 발행 완료: %s\n", pageURL)
		return nil
	default:
		return fmt.Errorf("--target must be slack, discord, webhook, or notion")
	}
}

func blocksFromContent(content []byte) ([]notion.Block, error) {
	if out, ok := briefing.ParseOutput(content); ok {
		return notion.ParseMarkdownToBlocks(briefing.Markdown(out)), nil
	}
	trimmed := strings.TrimSpace(string(content))
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return notion.ParseSearchJSONToBlocks(content)
	}
	return notion.ParseMarkdownToBlocks(string(content)), nil
}

func webhookURLFromEnv(target string) string {
	switch target {
	case "slack":
		return os.Getenv("SLACK_WEBHOOK_URL")
	case "discord":
		return os.Getenv("DISCORD_WEBHOOK_URL")
	default:
		return os.Getenv("WEBHOOK_URL")
	}
}
