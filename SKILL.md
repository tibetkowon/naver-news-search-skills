---
name: naver-news
description: 한국어 뉴스 검색/브리핑 CLI. 네이버 뉴스 API와 Google News RSS로 기사 후보를 수집하고 Slack, Discord, Webhook, Notion에 발행합니다.
version: 3.0.0
binary: ./naver-news
build: go build -o naver-news .
env:
  required: []
  optional:
    - NAVER_CLIENT_ID       # --source naver
    - NAVER_CLIENT_SECRET   # --source naver
    - NOTION_API_KEY        # notion 커맨드
    - SLACK_WEBHOOK_URL     # publish/run --target slack
    - DISCORD_WEBHOOK_URL   # publish/run --target discord
    - WEBHOOK_URL           # publish/run --target webhook
capabilities:
  - news_search
  - news_briefing
  - channel_publish
  - notion_publish
---

# naver-news 스킬

한국어 뉴스를 검색하고 기사 후보의 제목, 설명, 날짜, 링크를 가져온 뒤 채널 또는 Notion에 발행하는 CLI 도구입니다.

> **사용 전 빌드 필요**: 처음 사용하거나 소스가 업데이트된 경우 반드시 빌드합니다.
> ```bash
> go build -o naver-news .
> ```

자세한 사용법은 **AGENT.md**를 참고하세요.

## 커맨드 요약

### `search`

```bash
./naver-news search --query <검색어> [--display N] [--sort sim|date] [--start N] [--source naver|google] [--format json|markdown]
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | 정확도순/날짜순 — naver 전용 |
| `--start` | 1 | 시작 위치 (페이지네이션) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
| `--format` | json | `json` \| `markdown` |

### `brief`

```bash
./naver-news brief --queries <검색어,검색어> [--display N] [--source auto|naver|google] [--format json|markdown]
```

여러 검색어를 병렬 수집하고 URL 기준 중복을 제거합니다. `--source auto`는 Naver 실패 시 Google News RSS로 폴백합니다.

### `publish`

```bash
./naver-news publish --target slack|discord|webhook|notion
```

stdin에서 `search`/`brief` JSON 또는 Markdown을 받아 발행합니다. Slack/Discord 채널은 Webhook URL에 묶인 채널로 결정됩니다.

### `notion`

```bash
# 새 페이지 생성
./naver-news notion --parent-id <ID> --title <제목>

# 기존 페이지에 append
./naver-news notion --page-id <ID>
```

stdin에서 JSON(search 기본 출력) 또는 Markdown을 자동 감지합니다.

### `run`

```bash
./naver-news run --queries <검색어,검색어> --target stdout|slack|discord|webhook|notion [--interval 30m]
```

수집부터 발행까지 한 번에 실행합니다. 스케줄러에서는 `--interval` 없이 1회 실행으로 호출하는 방식을 우선 사용합니다.

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
      "pub_date": "Mon, 02 Jun 2025 09:00:00 +0900"
    }
  ]
}
```

## 워크플로우 예시

```bash
# 뉴스 목록만 (API 키 불필요)
./naver-news search --query "인공지능" --source google

# 검색 결과 → Notion 저장
./naver-news search --query "인공지능" --display 5 \
  | ./naver-news notion --parent-id <ID> --title "뉴스 브리핑"

# 여러 검색어 → Slack 채널 발행
./naver-news brief --queries "AI,반도체,환율" \
  | ./naver-news publish --target slack

# 스케줄러용 1회 실행
./naver-news run --queries "AI,반도체" --target notion --page-id <ID>
```
