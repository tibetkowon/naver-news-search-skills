# naver-news CLI 코드 리뷰 노트

**대상 파일:** `main.go`, `internal/naver/client.go`, `internal/google/client.go`, `internal/notion/client.go`

이 문서는 `naver-news` CLI 도구의 현재 구조를 해설합니다. 네이버 뉴스 검색 API, Google News RSS, Notion API 호출과 JSON/Markdown 출력 흐름을 중심으로 봅니다.

## 1. `main.go` — CLI 진입점

`main.go`는 세 가지 일을 담당합니다.

1. `.env` 파일을 로드합니다.
2. `search`, `notion` 서브커맨드를 분기합니다.
3. 검색 결과를 JSON 또는 Markdown으로 출력하고, stdin 입력을 Notion 블록으로 전달합니다.

`search` 기본 출력은 JSON이며, 에이전트가 구조적으로 읽기 좋도록 `query`, `source`, `items`를 포함합니다.

## 2. `internal/naver/client.go` — Naver 검색

Naver News Search API는 개별 기사에 대해 두 링크를 제공합니다.

- `originallink`: 언론사 원문 URL
- `link`: 네이버 뉴스 URL

코드는 이를 각각 `url`, `naver_url`로 매핑합니다. 제목과 설명에는 `<b>` 태그가 포함될 수 있으므로 제거한 뒤 HTML 엔티티를 unescape합니다.

## 3. `internal/google/client.go` — Google News RSS

Google News RSS는 API 키 없이 사용할 수 있는 보조 검색 소스입니다. 다만 RSS의 `<link>`는 Google News 링크일 수 있으므로, 개별 기사 원문 URL이 필요한 흐름에서는 Naver 검색 결과를 우선 사용합니다.

## 4. `internal/notion/client.go` — Notion 저장

`notion` 커맨드는 stdin으로 받은 JSON 또는 Markdown을 Notion 블록으로 변환합니다.

- JSON 입력: `search` 결과를 heading/paragraph/divider 블록으로 변환
- Markdown 입력: 에이전트가 직접 작성한 브리핑을 블록으로 변환
- 100블록이 넘으면 Notion API 제한에 맞춰 배치로 나누어 append

## 설계 메모

기사 본문 추출은 CLI에서 제거했습니다. 한국 뉴스 사이트는 봇 차단, 동적 렌더링, 광고/동의 레이어, 언론사별 HTML 차이가 커서 범용 CLI가 안정적으로 처리하기 어렵습니다. 이 도구는 검색 결과와 링크를 안정적으로 제공하는 데 집중합니다.
