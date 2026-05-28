# naver-news

한국어 뉴스를 검색하고 기사 본문을 수집하는 Go CLI 도구입니다. 어떤 에이전트(Hermes, Claude Code, OpenClaw 등)에서도 사용할 수 있습니다.

- **네이버 뉴스 API** — 한국어 뉴스 검색 (일 25,000회 무료)
- **Google News RSS** — API 키 없이 검색 가능한 보조 소스
- **Exa Contents API** — 기사 전문 추출
- **Notion API** — 검색 결과 페이지 저장

## 설치

Go 1.22 이상이 필요합니다.

```bash
git clone https://github.com/kowon/naver-news-search-skills.git
cd naver-news-search-skills
go build -o naver-news .
```

## 환경 변수

`.env` 파일에 작성하면 자동으로 읽습니다.

| 변수명 | 설명 | 필수 조건 |
|--------|------|-----------|
| `NAVER_CLIENT_ID` | 네이버 API 클라이언트 ID | `--source naver` (기본) |
| `NAVER_CLIENT_SECRET` | 네이버 API 클라이언트 Secret | `--source naver` (기본) |
| `EXA_API_KEY` | Exa AI API 키 | `fetch`, `search --fetch` |
| `NOTION_API_KEY` | Notion Integration 토큰 | `notion` 커맨드 |

## 커맨드

### `search` — 뉴스 검색

```bash
./naver-news search --query "인공지능" --display 5
./naver-news search --query "AI" --sort date --display 10
./naver-news search --query "AI" --display 10 --start 11        # 페이지네이션
./naver-news search --query "인공지능" --source google           # API 키 불필요
./naver-news search --query "테슬라" --display 3 --fetch         # 전문 포함
./naver-news search --query "AI" --format markdown              # Markdown 출력
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim` 정확도순 / `date` 날짜순 — naver 전용 |
| `--start` | 1 | 시작 위치 (페이지네이션) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
| `--fetch` | false | 각 기사 전문을 Exa로 가져오기 |
| `--format` | json | `json` \| `markdown` |

**JSON 출력 (기본):**
```json
{
  "query": "인공지능",
  "source": "naver",
  "items": [
    {
      "title": "삼성전자 AI 반도체 출시",
      "description": "삼성전자가...",
      "url": "https://www.example.com/article/1",
      "naver_url": "https://n.news.naver.com/article/001/123",
      "pub_date": "Mon, 02 Jun 2025 09:00:00 +0900"
    }
  ]
}
```

### `fetch` — 기사 본문 가져오기

```bash
./naver-news fetch --url "https://n.news.naver.com/..."
./naver-news fetch --url "https://url1" --url "https://url2"    # 다중 URL
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--url` | (필수, 반복 가능) | 기사 URL |
| `--format` | json | `json` \| `markdown` |

### `notion` — Notion 페이지 저장

stdin(JSON 또는 Markdown)을 Notion 페이지로 저장합니다.

```bash
# 새 페이지 생성
./naver-news search --query "AI" --display 5 \
  | ./naver-news notion --parent-id <ID> --title "뉴스 브리핑"

# 기존 페이지에 추가
./naver-news search --query "경제" --display 5 \
  | ./naver-news notion --page-id <ID>
```

## 에이전트 사용법

에이전트 통합 가이드는 **AGENT.md**를 참고하세요.

```bash
# 기본 워크플로우: 검색 → 전문 수집 → Notion 저장
{
  ./naver-news search --query "인공지능" --display 3 --fetch
  ./naver-news search --query "경제 주식" --display 3 --fetch
} | ./naver-news notion --parent-id <ID> --title "2026-06-01 뉴스 브리핑"
```

## 라이선스

MIT
