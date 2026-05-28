# naver-news — 에이전트 가이드

한국어 뉴스를 검색하고 기사 본문을 가져오는 CLI 도구입니다. 에이전트가 직접 `naver-news` 바이너리를 호출하고, 수집한 내용을 LLM 능력으로 요약합니다.

## 빌드

```bash
go build -o naver-news .
```

Go 1.22 이상 필요. 외부 패키지 없음.

## 환경 변수

| 변수명 | 필수 조건 | 발급처 |
|--------|-----------|--------|
| `NAVER_CLIENT_ID` | `--source naver` (기본) | 네이버 개발자 센터 |
| `NAVER_CLIENT_SECRET` | `--source naver` (기본) | 네이버 개발자 센터 |
| `EXA_API_KEY` | `fetch`, `search --fetch` | exa.ai |
| `NOTION_API_KEY` | `notion` 커맨드 | Notion Integrations |

`.env` 파일에 작성하면 자동으로 읽습니다.

---

## 커맨드 레퍼런스

### `search` — 뉴스 검색

```bash
./naver-news search --query <검색어> [옵션...]
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1–100) |
| `--sort` | sim | `sim` 정확도순 / `date` 날짜순 — naver 전용 |
| `--start` | 1 | 시작 위치 (1-based, 페이지네이션) — naver 전용 |
| `--source` | naver | `naver` \| `google` |
| `--fetch` | false | 각 기사 전문을 Exa로 가져오기 |
| `--format` | json | `json` \| `markdown` |

**JSON 출력 구조:**
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
      "pub_date": "Mon, 02 Jun 2025 09:00:00 +0900",
      "content": "기사 전문... (--fetch 시에만 포함)"
    }
  ]
}
```

### `fetch` — 기사 본문 가져오기

```bash
./naver-news fetch --url <URL> [--url <URL> ...] [--format json|markdown]
```

단일 URL:
```json
{"url": "https://...", "content": "기사 전문..."}
```

복수 URL:
```json
{"results": [{"url": "https://...", "content": "..."}, ...]}
```

### `notion` — Notion 페이지 저장

stdin(JSON 또는 Markdown)을 Notion 페이지로 저장합니다. JSON/Markdown 자동 감지.

```bash
# 새 페이지 생성
./naver-news search --query "AI" | ./naver-news notion --parent-id <ID> --title "뉴스 브리핑"

# 기존 페이지에 추가
./naver-news search --query "경제" | ./naver-news notion --page-id <ID>
```

---

## 워크플로우 패턴

### 패턴 A — 빠른 목록 확인

API 키가 없거나 가볍게 훑어볼 때.

```bash
# Google RSS (API 키 불필요)
./naver-news search --query "인공지능" --display 10 --source google

# Naver (NAVER_CLIENT_ID, NAVER_CLIENT_SECRET 필요)
./naver-news search --query "인공지능" --display 10
```

에이전트는 `items` 배열의 `title`과 `description`을 보고 읽을 기사를 선택합니다.

### 패턴 B — 기사 전문 요약

```bash
# 검색과 동시에 전문 수집 (EXA_API_KEY 필요)
./naver-news search --query "인공지능" --display 3 --fetch

# 또는 선택한 URL만 개별 수집
./naver-news fetch --url "https://..."
```

에이전트는 `content` 필드를 읽고 직접 요약합니다.

### 패턴 C — 뉴스 브리핑 → Notion 저장

```bash
# 여러 주제 검색 후 하나의 Notion 페이지로 정리
{
  ./naver-news search --query "인공지능" --display 3 --fetch
  ./naver-news search --query "경제 주식" --display 3 --fetch
} | ./naver-news notion --parent-id <ID> --title "2026년 6월 1일 브리핑"
```

에이전트가 직접 요약 Markdown을 작성해 notion에 넘기는 것도 가능:

```bash
echo "# 🤖 인공지능\n\n## [기사 제목](https://url)\n\n요약 내용\n\n---" \
  | ./naver-news notion --parent-id <ID> --title "브리핑"
```

### 패턴 D — 페이지네이션

```bash
./naver-news search --query "AI" --display 10 --start 1   # 1~10번째
./naver-news search --query "AI" --display 10 --start 11  # 11~20번째
```

---

## notion 커맨드가 인식하는 Markdown 규칙

에이전트가 직접 Markdown을 작성해 notion에 넘길 때 참고합니다.

| 패턴 | Notion 블록 |
|------|------------|
| `# 제목` | `heading_1` |
| `## [제목](url)` | `heading_2` + URL 하이퍼링크 |
| `## 제목` | `heading_2` |
| `N. 텍스트` | `numbered_list_item` |
| `---` | `divider` |
| 그 외 | `paragraph` (`**bold**` 인라인 지원) |
