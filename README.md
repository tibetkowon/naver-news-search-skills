# naver-news-search-skills

네이버 뉴스 검색 API와 Exa Contents API를 활용해 한국어 뉴스를 검색하고 기사 본문을 가져오는 Go CLI 도구입니다. OpenClaw 에이전트가 `SKILL.md`를 통해 이 도구를 호출하고, 검색 결과를 바탕으로 뉴스를 요약하거나 Notion 페이지로 저장합니다.

## 설치

Go 1.22 이상이 필요합니다.

```bash
git clone https://github.com/kowon/naver-news-search-skills.git
cd naver-news-search-skills
go build -o naver-news .
```

## 환경 변수

`.env` 파일에 작성하면 자동으로 읽습니다 (`.env.example` 참고).

| 변수명 | 설명 | 발급처 |
|--------|------|--------|
| `NAVER_CLIENT_ID` | 네이버 API 클라이언트 ID | [네이버 개발자 센터](https://developers.naver.com) |
| `NAVER_CLIENT_SECRET` | 네이버 API 클라이언트 Secret | [네이버 개발자 센터](https://developers.naver.com) |
| `EXA_API_KEY` | Exa AI API 키 | [Exa AI](https://exa.ai) |
| `NOTION_API_KEY` | Notion Integration 토큰 | [Notion Developers](https://www.notion.so/my-integrations) |

## 커맨드

### `search` — 뉴스 검색

```bash
./naver-news search --query "인공지능" --display 5
./naver-news search --query "AI 반도체" --sort date --display 10
./naver-news search --query "테슬라" --display 3 --fetch       # 전문 포함
./naver-news search --query "인공지능" --display 10 --highlights  # 빠른 브라우징
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim` 정확도순 / `date` 날짜순 |
| `--fetch` | false | 각 기사 전문을 Exa로 함께 가져오기 (요약·저장용) |
| `--highlights` | false | Exa로 핵심 문장 3~5개 추출 (브라우징용) |
| `--highlight-query` | "" | highlights 추출 방향 지정 |
| `--highlight-chars` | 500 | URL당 최대 문자 수 |

> **`--fetch` vs `--highlights`**
> - `--highlights`: 어떤 기사를 읽을지 고를 때. 빠르지만 한국 뉴스 사이트 특성상 노이즈(댓글 수, 날짜 등)가 섞일 수 있음.
> - `--fetch`: 기사 전문을 가져와 에이전트가 직접 요약할 때. 정확도가 높음. 노션 저장 시 권장.

### `fetch` — 특정 URL 본문 가져오기

```bash
./naver-news fetch --url "https://n.news.naver.com/..."
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--url` | (필수) | 기사 URL |

### `notion` — Notion 페이지 생성

stdin으로 Markdown을 받아 지정한 부모 페이지 하위에 새 페이지를 생성합니다.

```bash
echo "$summary_markdown" | ./naver-news notion \
  --parent-id <페이지ID> \
  --title "2026년 3월 2일 뉴스 브리핑"
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--parent-id` | (필수) | 부모 페이지 ID (Notion URL의 마지막 32자리 hex) |
| `--title` | (필수) | 새 페이지 제목 |

성공 시 출력:
```
노션 페이지 생성 완료: https://notion.so/...
```

**`notion` 커맨드가 인식하는 Markdown 규칙:**

| 패턴 | Notion 블록 |
|------|------------|
| `# 카테고리` | `heading_1` |
| `## [제목](url)` | `heading_2` + 제목에 URL 하이퍼링크 |
| `## 제목` | `heading_2` |
| `N. 텍스트` | `numbered_list_item` |
| `---` | `divider` |
| 그 외 | `paragraph` (`**bold**` 인라인 지원) |

`search` 출력을 그대로 파이프하면 카테고리 이모지 자동 매핑, 기사 URL 하이퍼링크 적용 등 추가 처리가 됩니다.

## 에이전트 워크플로우

### 빠른 브라우징

기사 목록을 훑어보고 읽을 기사를 고를 때:

```bash
./naver-news search --query "인공지능" --display 10 --highlights
./naver-news search --query "경제 주식" --display 10 --highlights
```

### 뉴스 요약 후 Notion 저장 (권장)

에이전트가 기사 전문을 읽고 직접 요약한 뒤, 요약만 Notion에 올립니다.

```
1. search --fetch  →  에이전트가 전문 읽고 요약 Markdown 작성  →  notion
```

에이전트가 작성하는 요약 Markdown 형식:

```markdown
# 🤖 인공지능

## [삼성전자, AI 반도체 출시](https://example.com/article/1)

삼성전자가 차세대 AI 반도체를 출시했다.

1. 전작 대비 성능 40% 향상
2. 하반기 글로벌 출시 예정

---

# 💰 경제 주식

## 코스피 2600선 회복

오늘 코스피 지수가 2600선을 회복했다.

1. 외국인 순매수 지속
2. 반도체 섹터 강세

---
```

## 출력 형식 (`search`)

```markdown
# 네이버 뉴스 검색 결과: "인공지능"

총 3개 기사

## 1. 삼성전자, AI 반도체 신제품 출시

- **날짜**: Mon, 02 Mar 2026 09:00:00 +0900
- **원문 링크**: https://www.example.com/article/123
- **네이버 링크**: https://n.news.naver.com/article/123

삼성전자가 새로운 AI 반도체 제품을 출시했다...

---
```

## 프로젝트 구조

```
naver-news-search-skills/
├── CLAUDE.md               # 프로젝트 개요 및 빌드 가이드
├── SKILL.md                # OpenClaw 에이전트용 스킬 매니페스트
├── README.md
├── main.go                 # CLI 진입점
├── go.mod
├── .env.example            # 환경 변수 예시
├── internal/
│   ├── naver/
│   │   └── client.go       # 네이버 뉴스 API 클라이언트
│   ├── exa/
│   │   └── client.go       # Exa Contents API 클라이언트
│   └── notion/
│       └── client.go       # Notion API 클라이언트 + Markdown 파서
├── .claude/
│   └── skills/             # Claude 워크플로우 스킬
└── docs/
    └── apis/               # API 명세
```

## 라이선스

MIT
