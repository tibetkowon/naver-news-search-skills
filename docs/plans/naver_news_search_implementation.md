# naver-news 구현 계획

## 목표

한국어 뉴스를 검색하고 에이전트가 요약 후보를 고를 수 있도록 제목, 설명, 날짜, 원문 링크, 네이버 뉴스 링크를 제공하는 CLI 도구를 만든다.

기사 본문 추출은 봇 차단과 언론사별 HTML 구조 차이 때문에 CLI 범위 밖으로 둔다.

## 구성

| 파일 | 역할 |
|------|------|
| `main.go` | CLI 서브커맨드, 출력 포맷 처리 |
| `internal/naver/client.go` | 네이버 뉴스 검색 API 클라이언트 |
| `internal/google/client.go` | Google News RSS 클라이언트 |
| `internal/notion/client.go` | 검색 결과/Markdown을 Notion 블록으로 변환 및 저장 |
| `internal/dotenv/dotenv.go` | `.env` 로딩 |
| `README.md`, `AGENT.md`, `SKILL.md` | 사용자 및 에이전트용 문서 |

## 커맨드

```bash
./naver-news search --query "인공지능" --display 5
./naver-news search --query "AI" --sort date --display 10
./naver-news search --query "인공지능" --source google
./naver-news search --query "AI" --format markdown
./naver-news search --query "AI" | ./naver-news notion --parent-id <ID> --title "뉴스 브리핑"
```

## 출력

기본 출력은 JSON이다.

```json
{
  "query": "인공지능",
  "source": "naver",
  "items": [
    {
      "title": "기사 제목",
      "description": "검색 결과 설명",
      "url": "https://원문URL",
      "naver_url": "https://n.news.naver.com/...",
      "pub_date": "Mon, 02 Jun 2025 09:00:00 +0900"
    }
  ]
}
```

## 완료 항목

- [x] Naver 검색
- [x] Google News RSS 검색
- [x] JSON/Markdown 출력
- [x] Notion 새 페이지 생성
- [x] 기존 Notion 페이지 append
- [x] 본문 추출 기능을 범위 밖으로 제거
