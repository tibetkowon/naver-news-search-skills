# Plan: Naver News Search Skills 구현

**작성일:** 2026-03-02
**상태:** 완료 (Completed)

---

## Goal

OpenClaw 에이전트가 한국어 뉴스를 검색하고 요약할 수 있도록, 네이버 뉴스 검색 API와 Exa Content API를 활용하는 Go CLI 도구를 만들고, 이를 위한 SKILL.md 매니페스트를 작성한다.

## Requirements

- Go 표준 라이브러리만 사용 (외부 의존성 없음)
- 에이전트가 Markdown 출력을 직접 파싱할 수 있어야 함
- 네이버 API: GET `/v1/search/news.json` (검색어, 개수, 정렬 지원)
- Exa API: POST `/contents` (URL로 본문 텍스트 추출)
- `<b>` 태그 등 HTML 마크업 제거 후 출력
- 환경 변수로 인증 정보 관리 (하드코딩 금지)

## Affected Files

| 파일 | 작업 |
|------|------|
| `CLAUDE.md` | 신규 작성 |
| `SKILL.md` | 신규 작성 |
| `README.md` | 업데이트 |
| `main.go` | 신규 작성 |
| `go.mod` | 신규 작성 |
| `internal/naver/client.go` | 신규 작성 |
| `internal/exa/client.go` | 신규 작성 |
| `.gitignore` | 신규 작성 |

## Architecture

```
OpenClaw 에이전트
   └── SKILL.md 읽기
       └── naver-news CLI 호출
           ├── search --query "..." → 네이버 뉴스 API → 뉴스 목록(Markdown)
           └── fetch --url "..."   → Exa Contents API → 기사 본문(Markdown)
```

**역할 분리:**
- **Go CLI**: 데이터 수집 (API 호출, 포맷 변환)
- **OpenClaw LLM**: 최종 요약 및 분석

## Implementation Phases

### Phase 1: 프로젝트 설정
- [x] CLAUDE.md 작성 (프로젝트 개요, 환경 변수, 빌드 방법)
- [x] go.mod 초기화 (`github.com/kowon/naver-news-search-skills`)

### Phase 2: API 클라이언트
- [x] `internal/naver/client.go` — 네이버 뉴스 검색 API 클라이언트
  - `NewsItem` 구조체
  - `Search(query, display, sort)` 함수
  - `<b>` 태그 제거 (정규식)
- [x] `internal/exa/client.go` — Exa Contents API 클라이언트
  - `FetchContent(url)` 함수
  - POST `/contents` with `{"ids": [url], "text": true}`

### Phase 3: CLI 진입점
- [x] `main.go` — 서브커맨드 구조
  - `search` 서브커맨드 (`--query`, `--display`, `--sort`, `--fetch`)
  - `fetch` 서브커맨드 (`--url`)
  - Markdown 형식 출력

### Phase 4: 문서화
- [x] SKILL.md 작성 (YAML 프론트매터 + 사용법 + 한국어 예시)
- [x] README.md 업데이트
- [x] docs/plans/ 계획 문서 (이 파일)
- [x] docs/reviews/ 코드 해설 문서

## Verification

```bash
# 빌드 확인
go build -o naver-news .

# 사용법 출력
./naver-news

# 필수 플래그 검증
./naver-news search          # Error: --query is required
./naver-news fetch           # Error: --url is required

# 실제 API 테스트 (환경 변수 필요)
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy \
  ./naver-news search --query "인공지능" --display 5

EXA_API_KEY=zzz \
  ./naver-news fetch --url "https://n.news.naver.com/..."
```
