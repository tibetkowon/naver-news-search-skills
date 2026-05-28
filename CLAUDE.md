# naver-news-search-skills

한국어 뉴스를 검색하고 요약할 수 있도록, 네이버 뉴스 검색 API와 Exa Content API를 활용하는 Go CLI 도구 프로젝트입니다.

## 프로젝트 개요

에이전트는 이 프로젝트의 `naver-news` CLI를 호출하여 뉴스를 검색하고 기사 본문을 가져옵니다. 최종 요약은 에이전트 자신의 LLM 능력으로 생성합니다.

## 기술 스택

- **언어**: Go (표준 라이브러리만 사용: `net/http`, `encoding/json`, `encoding/xml`, `flag`, `regexp`)
- **외부 API**: 네이버 뉴스 검색 API, Google News RSS, Exa Contents API, Notion API

## 디렉토리 구조

```
naver-news-search-skills/
├── CLAUDE.md               ← 이 파일
├── AGENT.md                ← 범용 에이전트 가이드
├── SKILL.md                ← 에이전트용 스킬 매니페스트
├── README.md
├── main.go                 ← CLI 진입점
├── go.mod
├── internal/
│   ├── naver/
│   │   └── client.go       ← 네이버 뉴스 API 클라이언트
│   ├── google/
│   │   └── client.go       ← Google News RSS 클라이언트
│   ├── exa/
│   │   └── client.go       ← Exa Contents API 클라이언트
│   └── notion/
│       └── client.go       ← Notion API 클라이언트 + 파서
├── .claude/
│   └── skills/             ← 로컬 Claude 스킬
└── docs/
    └── apis/               ← API 명세 문서
```

## 환경 변수

| 변수명 | 설명 | 필수 여부 |
|--------|------|-----------|
| `NAVER_CLIENT_ID` | 네이버 개발자 센터 클라이언트 ID | `--source naver` 필수 |
| `NAVER_CLIENT_SECRET` | 네이버 개발자 센터 클라이언트 Secret | `--source naver` 필수 |
| `EXA_API_KEY` | Exa AI API 키 | `fetch`, `search --fetch` 필수 |
| `NOTION_API_KEY` | Notion Integration 토큰 | `notion` 커맨드 필수 |

## 빌드 및 실행

```bash
# 빌드
go build -o naver-news .

# 뉴스 검색 (JSON 기본 출력)
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "인공지능" --display 5

# Markdown 출력
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "AI" --display 10 --format markdown

# Google News RSS (API 키 불필요)
./naver-news search --query "인공지능" --display 5 --source google

# 페이지네이션
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "AI" --display 10 --start 11

# 기사 본문 가져오기
EXA_API_KEY=zzz ./naver-news fetch --url "https://n.news.naver.com/..."

# 다중 URL fetch
EXA_API_KEY=zzz ./naver-news fetch --url "https://url1" --url "https://url2"

# 검색 + 전체 본문 가져오기
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy EXA_API_KEY=zzz ./naver-news search --query "AI" --display 3 --fetch

# 검색 결과를 Notion 페이지로 저장
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "인공지능" --display 5 \
  | NOTION_API_KEY=nnn ./naver-news notion --parent-id <page_id> --title "2026년 3월 2일 뉴스 브리핑"

# 기존 Notion 페이지에 추가
./naver-news search --query "경제" --display 5 \
  | NOTION_API_KEY=nnn ./naver-news notion --page-id <page_id>
```

## CLI 커맨드

### `search`
뉴스를 검색하고 JSON(기본) 또는 Markdown 형식으로 출력합니다.

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim`(정확도순), `date`(날짜순) — naver 전용 |
| `--start` | 1 | 시작 위치 (1-based) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
| `--fetch` | false | 각 기사 전문을 Exa로 함께 가져오기 |
| `--format` | json | `json` \| `markdown` |

### `fetch`
Exa Contents API로 기사 본문을 가져옵니다. `--url`을 여러 번 사용해 다중 URL 처리 가능.

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--url` | (필수, 반복 가능) | 기사 URL |
| `--format` | json | `json` \| `markdown` |

### `notion`
stdin(JSON 또는 Markdown)을 Notion 페이지로 저장합니다.

| 플래그 | 설명 |
|--------|------|
| `--parent-id` + `--title` | 새 페이지 생성 |
| `--page-id` | 기존 페이지에 블록 append |

## 워크플로우 스킬

- `.claude/skills/plan_feature.md`: 기능 계획 스킬
- `.claude/skills/write_code_tutor.md`: 코드 리뷰 문서 스킬
