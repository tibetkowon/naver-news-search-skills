---
name: naver-news-search
description: 네이버 뉴스 검색 API와 Exa로 뉴스를 검색하고 기사 본문을 가져오는 도구
version: 1.0.0
binary: ./naver-news
env:
  - NAVER_CLIENT_ID
  - NAVER_CLIENT_SECRET
  - EXA_API_KEY
capabilities:
  - news_search
  - article_fetch
---

# naver-news-search 스킬

한국어 뉴스를 검색하고 기사 본문을 가져오는 CLI 도구입니다. 네이버 뉴스 검색 API로 뉴스 목록을 검색하고, Exa Contents API로 기사 원문을 가져옵니다.

## 필수 환경 변수

| 변수명 | 설명 | 필요한 커맨드 |
|--------|------|---------------|
| `NAVER_CLIENT_ID` | 네이버 개발자 센터 클라이언트 ID | `search` |
| `NAVER_CLIENT_SECRET` | 네이버 개발자 센터 클라이언트 Secret | `search` |
| `EXA_API_KEY` | Exa AI API 키 | `fetch`, `search --fetch` |

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

> `--fetch`와 `--highlights`는 동시에 사용하지 않습니다. `--highlights`는 목록 브라우징용, `--fetch`는 전체 본문이 필요할 때 사용합니다.

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

에이전트는 다음 순서로 이 도구를 사용합니다:

1. **검색**: `search` 커맨드로 관련 뉴스 목록을 가져옵니다.
2. **선택**: 관련성 높은 기사의 URL을 선택합니다.
3. **본문 가져오기**: `fetch` 커맨드 또는 `search --fetch`로 본문을 가져옵니다.
4. **요약**: 가져온 내용을 바탕으로 에이전트 자신이 요약을 생성합니다.

## 카테고리별 뉴스 브라우징

네이버 API는 카테고리 파라미터를 제공하지 않습니다. 카테고리별 정리가 필요하면 주제별로 `search`를 여러 번 호출하세요. `--highlights`를 사용하면 각 기사의 핵심 문장만 간결하게 확인할 수 있습니다:

```bash
./naver-news search --query "인공지능" --display 5 --highlights
./naver-news search --query "경제 주식" --display 5 --highlights
./naver-news search --query "정치" --display 5 --highlights
./naver-news search --query "스포츠" --display 5 --highlights
```

에이전트는 각 결과를 카테고리 제목과 함께 묶어 사용자에게 정리된 형태로 제공합니다.

## 한국어 예시 프롬프트

| 사용자 요청 | 권장 커맨드 |
|------------|-------------|
| "오늘 IT 뉴스 알려줘" | `search --query "IT" --sort date --display 5 --highlights` |
| "인공지능 관련 최신 뉴스" | `search --query "인공지능" --sort date --display 10 --highlights` |
| "테슬라 뉴스 자세히 알고 싶어" | `search --query "테슬라" --display 3 --fetch` |
| "이 기사 내용 요약해줘: [URL]" | `fetch --url "[URL]"` |
| "반도체 산업 동향 분석해줘" | `search --query "반도체" --sort date --display 10 --fetch` |
| "카테고리별 오늘 뉴스 정리해줘" | 주제별로 `search --highlights` 여러 번 호출 |
