# naver-news-search-skills

OpenClaw 에이전트가 한국어 뉴스를 검색하고 요약할 수 있도록, 네이버 뉴스 검색 API와 Exa Content API를 활용하는 Go CLI 도구 프로젝트입니다.

## 프로젝트 개요

에이전트는 이 프로젝트의 `naver-news` CLI를 호출하여 뉴스를 검색하고 기사 본문을 가져옵니다. 최종 요약은 에이전트 자신의 LLM 능력으로 생성합니다.

## 기술 스택

- **언어**: Go (표준 라이브러리만 사용: `net/http`, `encoding/json`, `flag`)
- **외부 API**: 네이버 뉴스 검색 API, Exa Contents API

## 디렉토리 구조

```
naver-news-search-skills/
├── CLAUDE.md               ← 이 파일
├── SKILL.md                ← OpenClaw 에이전트용 스킬 매니페스트
├── README.md
├── main.go                 ← CLI 진입점
├── go.mod
├── go.sum
├── internal/
│   ├── naver/
│   │   └── client.go       ← 네이버 뉴스 API 클라이언트
│   └── exa/
│       └── client.go       ← Exa Contents API 클라이언트
├── .claude/
│   └── skills/             ← 로컬 Claude 스킬
└── docs/
    └── apis/               ← API 명세 문서
```

## 환경 변수

| 변수명 | 설명 | 필수 여부 |
|--------|------|-----------|
| `NAVER_CLIENT_ID` | 네이버 개발자 센터에서 발급받은 클라이언트 ID | search 커맨드 필수 |
| `NAVER_CLIENT_SECRET` | 네이버 개발자 센터에서 발급받은 클라이언트 Secret | search 커맨드 필수 |
| `EXA_API_KEY` | Exa AI에서 발급받은 API 키 | fetch 커맨드 필수 |

## 빌드 및 실행

```bash
# 빌드
go build -o naver-news .

# 뉴스 검색 (목록만)
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "인공지능" --display 5

# 날짜순 정렬 검색
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy ./naver-news search --query "AI" --display 10 --sort date

# 기사 본문 가져오기
EXA_API_KEY=zzz ./naver-news fetch --url "https://n.news.naver.com/..."

# 통합 (검색 + 본문 자동 가져오기)
NAVER_CLIENT_ID=xxx NAVER_CLIENT_SECRET=yyy EXA_API_KEY=zzz ./naver-news search --query "AI" --display 3 --fetch
```

## CLI 커맨드

### `search`
네이버 뉴스 API로 뉴스를 검색하고 Markdown 형식으로 출력합니다.

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | 정렬: `sim`(정확도순), `date`(날짜순) |
| `--fetch` | false | 각 기사 본문을 Exa로 함께 가져오기 |

### `fetch`
Exa Contents API로 특정 URL의 기사 본문을 가져옵니다.

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--url` | (필수) | 기사 URL |

## 워크플로우 스킬

- `.claude/skills/plan_feature.md`: 기능 계획 스킬
- `.claude/skills/write_code_tutor.md`: 코드 리뷰 문서 스킬
