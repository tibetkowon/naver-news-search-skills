# naver-news

한국어 뉴스를 검색하고 원문 링크를 수집하는 Go CLI 도구입니다. 어떤 에이전트(Hermes, Claude Code, OpenClaw 등)에서도 사용할 수 있습니다.

- **네이버 뉴스 API**: 원문 URL과 네이버 뉴스 URL을 함께 제공하는 기본 검색 소스
- **Google News RSS**: API 키 없이 사용할 수 있는 보조 검색 소스
- **멀티 검색어 브리핑**: 여러 검색어를 병렬 수집하고 URL 기준 중복 제거
- **채널/Notion 발행**: Slack, Discord, Generic Webhook, Notion으로 전송
- **주기 실행**: cron/에이전트 스케줄러에 맞는 1회 실행 또는 `--interval` 반복 실행

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
| `NOTION_API_KEY` | Notion Integration 토큰 | `notion` 커맨드 |
| `SLACK_WEBHOOK_URL` | Slack Incoming Webhook URL | `publish/run --target slack` |
| `DISCORD_WEBHOOK_URL` | Discord Webhook URL | `publish/run --target discord` |
| `WEBHOOK_URL` | 범용 JSON Webhook URL | `publish/run --target webhook` |

## 커맨드

### `search` — 뉴스 검색

```bash
./naver-news search --query "인공지능" --display 5
./naver-news search --query "AI" --sort date --display 10
./naver-news search --query "AI" --display 10 --start 11        # 페이지네이션
./naver-news search --query "인공지능" --source google           # API 키 불필요
./naver-news search --query "AI" --format markdown              # Markdown 출력
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim` 정확도순 / `date` 날짜순 — naver 전용 |
| `--start` | 1 | 시작 위치 (페이지네이션) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
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

Naver는 개별 기사 원문 URL(`url`)과 네이버 뉴스 URL(`naver_url`)을 함께 제공합니다. Google News RSS의 `url`은 Google News 링크일 수 있어 원문 직접 링크로 보장하지 않습니다.

### `brief` — 여러 검색어 브리핑

여러 검색어를 한 번에 수집합니다. `--source auto`는 Naver를 먼저 사용하고, 인증 정보가 없거나 실패하면 Google News RSS로 폴백합니다.

```bash
./naver-news brief --queries "AI,반도체,환율" --display 5
./naver-news brief --queries "AI,반도체" --source google --format markdown
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--queries` | (필수) | 쉼표로 구분한 검색어 목록 |
| `--display` | 5 | 검색어별 결과 개수 |
| `--source` | auto | `auto` \| `naver` \| `google` |
| `--sort` | date | `sim` 정확도순 / `date` 날짜순 — naver 전용 |
| `--concurrency` | 4 | 동시 검색 개수 |
| `--dedupe` | true | URL 기준 중복 제거 |
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

### `publish` — 채널 또는 Notion 발행

`brief` 또는 `search` 결과를 stdin으로 받아 Slack/Discord/Webhook/Notion으로 보냅니다. Slack/Discord 채널은 Webhook URL이 연결된 채널로 결정됩니다.

```bash
./naver-news brief --queries "AI,반도체" \
  | ./naver-news publish --target slack

./naver-news brief --queries "경제,환율" \
  | ./naver-news publish --target discord --webhook-url <DISCORD_WEBHOOK_URL>

./naver-news brief --queries "AI,정책" \
  | ./naver-news publish --target notion --page-id <NOTION_PAGE_ID>
```

### `run` — 수집부터 발행까지 한 번에 실행

에이전트 스케줄러나 cron에서는 `--interval` 없이 1회 실행하는 방식을 권장합니다. 장기 실행 프로세스가 필요한 환경에서는 `--interval`을 지정할 수 있습니다.

```bash
# 한 번 실행하고 Slack으로 발행
./naver-news run --queries "AI,반도체" --target slack

# 1시간마다 Notion 기존 페이지에 추가
./naver-news run --queries "경제,환율" --target notion --page-id <NOTION_PAGE_ID> --interval 1h

# cron 예시: 매일 오전 9시에 실행
0 9 * * * cd /path/to/naver-news-search-skills && ./naver-news run --queries "AI,반도체,환율" --target notion --page-id <NOTION_PAGE_ID>
```

## 에이전트 사용법

에이전트 통합 가이드는 **AGENT.md**를 참고하세요. 이 CLI는 기사 본문 추출을 시도하지 않습니다. 봇 차단과 언론사별 HTML 차이를 피하기 위해 검색 결과의 제목, 설명, 날짜, 링크를 안정적으로 제공하는 데 집중합니다.

## 라이선스

MIT
