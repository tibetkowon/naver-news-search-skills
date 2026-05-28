---
name: naver-news
description: 한국어 뉴스 검색 및 기사 본문 수집 CLI. 네이버 뉴스 API와 Google News RSS로 검색하고, Exa로 전문을 가져와 Notion에 저장합니다.
version: 2.0.0
binary: ./naver-news
build: go build -o naver-news .
env:
  required:
    - NAVER_CLIENT_ID       # --source naver (기본)
    - NAVER_CLIENT_SECRET   # --source naver (기본)
  optional:
    - EXA_API_KEY           # fetch, search --fetch
    - NOTION_API_KEY        # notion 커맨드
capabilities:
  - news_search
  - article_fetch
  - notion_publish
---

# naver-news 스킬

한국어 뉴스를 검색하고 기사 본문을 가져오는 CLI 도구입니다.

> **사용 전 빌드 필요**: 처음 사용하거나 소스가 업데이트된 경우 반드시 빌드합니다.
> ```bash
> go build -o naver-news .
> ```

자세한 사용법은 **AGENT.md**를 참고하세요.

## 커맨드 요약

### `search`

```bash
./naver-news search --query <검색어> [--display N] [--sort sim|date] [--start N] [--source naver|google] [--fetch] [--format json|markdown]
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | 정확도순/날짜순 — naver 전용 |
| `--start` | 1 | 시작 위치 (페이지네이션) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
| `--fetch` | false | 기사 전문을 Exa로 가져오기 (EXA_API_KEY 필요) |
| `--format` | json | `json` \| `markdown` |

### `fetch`

```bash
./naver-news fetch --url <URL> [--url <URL> ...] [--format json|markdown]
```

### `notion`

```bash
# 새 페이지 생성
./naver-news notion --parent-id <ID> --title <제목>

# 기존 페이지에 append
./naver-news notion --page-id <ID>
```

stdin에서 JSON(search 기본 출력) 또는 Markdown을 자동 감지합니다.

## JSON 출력 구조

```json
{
  "query": "인공지능",
  "source": "naver",
  "items": [
    {
      "title": "기사 제목",
      "description": "요약 2~3줄",
      "url": "https://원문URL",
      "naver_url": "https://n.news.naver.com/...",
      "pub_date": "Mon, 02 Jun 2025 09:00:00 +0900",
      "content": "기사 전문 (--fetch 시에만)"
    }
  ]
}
```

## 워크플로우 예시

```bash
# 뉴스 목록만 (API 키 불필요)
./naver-news search --query "인공지능" --source google

# 검색 + 전문 수집 → Notion 저장
./naver-news search --query "인공지능" --display 3 --fetch \
  | ./naver-news notion --parent-id <ID> --title "뉴스 브리핑"

# 다중 주제 → 하나의 페이지
{
  ./naver-news search --query "인공지능" --display 3 --fetch
  ./naver-news search --query "경제" --display 3 --fetch
} | ./naver-news notion --parent-id <ID> --title "2026-06-01 브리핑"
```
