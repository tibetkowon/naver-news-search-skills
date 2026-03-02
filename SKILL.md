---
name: naver-news-search
description: 네이버 뉴스 검색 API와 Exa로 뉴스를 검색하고 기사 본문을 가져오는 도구
version: 1.2.0
binary: ./naver-news
build: go build -o naver-news .
env:
  - NAVER_CLIENT_ID
  - NAVER_CLIENT_SECRET
  - EXA_API_KEY
  - NOTION_API_KEY
capabilities:
  - news_search
  - article_fetch
  - notion_publish
---

# naver-news-search 스킬

한국어 뉴스를 검색하고 기사 본문을 가져오는 CLI 도구입니다. 네이버 뉴스 검색 API로 뉴스 목록을 검색하고, Exa Contents API로 기사 원문을 가져옵니다.

> **사용 전 빌드 필요**: 처음 사용하거나 소스가 업데이트된 경우 반드시 빌드합니다.
> ```bash
> go build -o naver-news .
> ```

## 필수 환경 변수

| 변수명 | 설명 | 필요한 커맨드 |
|--------|------|---------------|
| `NAVER_CLIENT_ID` | 네이버 개발자 센터 클라이언트 ID | `search` |
| `NAVER_CLIENT_SECRET` | 네이버 개발자 센터 클라이언트 Secret | `search` |
| `EXA_API_KEY` | Exa AI API 키 | `fetch`, `search --fetch`, `search --highlights` |
| `NOTION_API_KEY` | Notion Integration 토큰 | `notion` |

## 커맨드

### `search` — 뉴스 검색

네이버 뉴스 API로 뉴스를 검색하여 Markdown 형식으로 출력합니다.

```bash
./naver-news search --query <검색어> [--display <수>] [--sort <sim|date>] [--fetch] [--highlights] [--highlight-query <검색어>] [--highlight-chars <N>]
```

**플래그:**

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim`: 정확도순, `date`: 날짜순 |
| `--fetch` | false | 각 기사의 전체 본문을 Exa로 가져오기 |
| `--highlights` | false | 각 기사 핵심 문장 3~5개를 Exa highlights로 가져오기 |
| `--highlight-query` | "" | highlights 추출 방향 지정 (빈 문자열이면 자동) |
| `--highlight-chars` | 500 | URL당 최대 문자 수 (한국어 약 3~5문장) |

> **`--highlights`와 `--fetch`는 목적이 다릅니다.**
> - `--highlights`: Exa가 자동 추출한 3~5문장. 빠르지만 한국 뉴스 사이트의 댓글·날짜 등 노이즈가 섞일 수 있음. **기사 목록을 훑어보고 읽을 기사를 고를 때** 사용합니다.
> - `--fetch`: 기사 전문을 가져와 에이전트 LLM이 직접 정리. 정확도가 높음. **실제로 내용을 요약하거나 노션에 저장할 때** 사용합니다.
> - 두 플래그를 동시에 사용하지 않습니다.

**출력 예시 (기본):**

```markdown
# 네이버 뉴스 검색 결과: "인공지능"

총 5개 기사

## 1. 삼성전자, AI 반도체 신제품 출시

- **날짜**: Mon, 03 Mar 2026 09:00:00 +0900
- **원문 링크**: https://www.example.com/article/123
- **네이버 링크**: https://n.news.naver.com/article/123

삼성전자가 새로운 AI 반도체 제품을 출시했다...

---
```

**출력 예시 (`--highlights` 사용 시):**

```markdown
## 1. 삼성전자, AI 반도체 신제품 출시

- **날짜**: Mon, 02 Mar 2026 09:00:00 +0900
- **원문 링크**: https://www.example.com/article/123
- **네이버 링크**: https://n.news.naver.com/article/123

삼성전자가 새로운 AI 반도체 제품을 출시했다...

**핵심 내용:**
1. 삼성전자가 새로운 AI 반도체 제품을 출시했다.
2. 이 제품은 전작 대비 성능이 40% 향상됐다.
3. 하반기 글로벌 출시를 목표로 하고 있다.

---
```

### `notion` — 노션 페이지 생성

`search` 커맨드의 Markdown 출력을 stdin으로 받아 Notion 페이지를 생성합니다.

```bash
{ ./naver-news search ...; ./naver-news search ...; } \
  | NOTION_API_KEY=<token> ./naver-news notion \
      --parent-id <페이지ID> \
      --title "2026년 3월 2일 뉴스 브리핑"
```

**플래그:**

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--parent-id` | (필수) | 부모 페이지 ID (URL의 마지막 32자리 hex) |
| `--title` | (필수) | 새 페이지 제목 |

**성공 시 출력:**

```
노션 페이지 생성 완료: https://notion.so/...
```

**동작:**
- stdin의 Markdown을 파싱해 Notion 블록으로 변환합니다.
- `# 네이버 뉴스 검색 결과: "쿼리"` → `🤖 인공지능` 형태의 `heading_1` (카테고리 이모지 자동 매핑)
- `## N. 제목` → 원문 URL 하이퍼링크가 적용된 `heading_2`
- 날짜·링크 메타 줄은 건너뜁니다.
- 페이지 상단에 생성 일시 callout을 자동 삽입합니다.
- 100블록 초과 시 배치로 분할 추가합니다.

---

### `fetch` — 기사 본문 가져오기

Exa Contents API로 특정 URL의 기사 본문을 Markdown 형식으로 가져옵니다.

```bash
./naver-news fetch --url <URL>
```

**플래그:**

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--url` | (필수) | 기사 URL |

**출력 예시:**

```markdown
# 기사 본문

**URL**: https://www.example.com/article/123

---

삼성전자가 새로운 AI 반도체 제품을 출시했다. 이번 제품은...
```

## 사용 예시

### 기본 뉴스 검색

```bash
./naver-news search --query "인공지능" --display 5
```

### 날짜순 정렬 검색

```bash
./naver-news search --query "AI 반도체" --sort date --display 10
```

### 검색 + 본문 함께 가져오기

```bash
./naver-news search --query "테슬라" --display 3 --fetch
```

### 검색 + 핵심 문장 가져오기 (highlights)

```bash
./naver-news search --query "AI 반도체" --display 5 --highlights
```

### highlights + 추출 방향 지정

```bash
./naver-news search --query "테슬라" --display 3 --highlights --highlight-query "주가 투자 영향"
```

### 특정 기사 본문 가져오기

```bash
./naver-news fetch --url "https://n.news.naver.com/article/001/0123456789"
```

## 에이전트 활용 패턴

### 패턴 A — 빠른 브라우징 후 선택 요약

기사가 많을 때 먼저 훑어보고, 관련성 높은 기사만 상세히 읽는 패턴입니다.

```
1단계 (브라우징): search --highlights  → 여러 기사 제목·설명 빠르게 확인
2단계 (선택):     에이전트가 관련 기사 URL 선택
3단계 (본문):     fetch --url <URL>     → 선택한 기사 전문 가져오기
4단계 (정리):     에이전트 LLM이 전문을 읽고 직접 요약
```

> `--highlights`는 **어떤 기사를 읽을지 고를 때**만 씁니다. Exa 자동 추출 특성상 한국 뉴스 사이트에서 노이즈(댓글 수, 날짜 등)가 섞일 수 있어, 최종 요약에는 사용하지 않습니다.

### 패턴 B — 검색 + 즉시 요약 (정확도 우선)

소수의 기사를 처음부터 정확하게 가져올 때 사용합니다.

```
1단계: search --fetch --display 3  → 검색 + 전문 동시에 가져오기
2단계: 에이전트 LLM이 전문을 읽고 직접 요약
```

### 패턴 C — 에이전트가 요약 후 노션 저장 (권장)

노션에 저장할 때는 **에이전트가 기사 전문을 읽고 직접 요약한 뒤**, 그 요약만 노션에 올립니다. 네이버 API 설명 스니펫이나 Exa 전문이 그대로 들어가지 않아 훨씬 깔끔합니다.

```
1단계: search --fetch           → 기사 전문 가져오기
2단계: 에이전트 LLM이 전문 읽고 요약 Markdown 작성
3단계: 에이전트 요약 → notion  → 노션 페이지 저장
```

에이전트가 작성하는 요약 Markdown 형식:

```markdown
# 🤖 인공지능

## [삼성전자, AI 반도체 출시](https://example.com/article/1)

삼성전자가 차세대 AI 반도체를 출시했다.

1. 전작 대비 성능 40% 향상
2. 하반기 글로벌 출시 예정

---

## [OpenAI, GPT-5 공개](https://example.com/article/2)

OpenAI가 GPT-5를 공개했다.

1. 멀티모달 성능 대폭 개선
2. API 가격 인하 예정

---

# 💰 경제 주식

## 코스피 2600선 회복

오늘 코스피 지수가 2600선을 회복했다.

1. 외국인 순매수 지속
2. 반도체 섹터 강세

---
```

**`notion` 커맨드가 인식하는 Markdown 규칙:**

| 패턴 | Notion 블록 |
|------|------------|
| `# 카테고리` | `heading_1` (이모지 포함 시 그대로, 없으면 📰 추가 안 함) |
| `## [제목](url)` | `heading_2` + 제목에 URL 하이퍼링크 |
| `## 제목` | `heading_2` (링크 없음) |
| `N. 텍스트` | `numbered_list_item` |
| `---` | `divider` |
| 그 외 | `paragraph` (`**bold**` 인라인 지원) |

## 카테고리별 뉴스 브라우징

네이버 API는 카테고리 파라미터를 제공하지 않습니다. 카테고리별 정리가 필요하면 주제별로 `search`를 여러 번 호출합니다.

**어떤 기사가 있는지 빠르게 확인할 때** (`--highlights`):

```bash
./naver-news search --query "인공지능" --display 5 --highlights
./naver-news search --query "경제 주식" --display 5 --highlights
```

에이전트는 이 결과를 보고 읽을 기사를 선택한 뒤, `fetch`로 전문을 가져와 직접 정리합니다.

## 카테고리별 뉴스 → 노션 자동화 패턴

에이전트가 기사 전문을 읽고 직접 요약한 뒤, 그 요약만 노션에 올립니다.

**에이전트가 수행할 단계:**

1. `search --fetch`로 각 카테고리의 기사 전문을 가져옵니다.
2. 가져온 기사 전문을 읽고, 위 "패턴 C" 형식에 맞춰 요약 Markdown을 직접 작성합니다.
3. 작성한 요약 Markdown을 `notion` 커맨드에 stdin으로 전달합니다.

```bash
# 1단계: 기사 전문 가져오기 (카테고리별로 실행)
./naver-news search --query "인공지능" --display 3 --fetch
./naver-news search --query "경제 주식" --display 3 --fetch

# 3단계: 에이전트가 작성한 요약 Markdown을 notion으로 전달
./naver-news notion --parent-id <page_id> --title "2026년 3월 2일 뉴스 브리핑"
```

> 2단계(요약 작성)는 에이전트 LLM이 직접 수행합니다. `notion` 커맨드에 전달되는 내용은 에이전트가 작성한 요약이어야 하며, `search` 출력을 그대로 파이프하지 않습니다.

> 기사 수가 많으면 토큰 사용량이 늘어납니다. 카테고리당 `--display 3~5`가 적당합니다.

## 한국어 예시 프롬프트

| 사용자 요청 | 방법 |
|------------|------|
| "오늘 IT 뉴스 어떤 거 있어?" | `search --highlights` → 에이전트가 목록 훑어 정리 |
| "인공지능 관련 최신 뉴스 요약해줘" | `search --fetch` → 에이전트가 전문 읽고 요약 |
| "테슬라 뉴스 자세히 알고 싶어" | `search --fetch` 또는 `fetch --url` |
| "이 기사 내용 요약해줘: [URL]" | `fetch --url "[URL]"` → 에이전트가 요약 |
| "반도체 산업 동향 분석해줘" | `search --fetch --sort date` → 에이전트가 분석 |
| "카테고리별 오늘 뉴스 노션에 정리해줘" | `search --fetch` × 카테고리 수 → 에이전트 요약 Markdown → `notion` |
