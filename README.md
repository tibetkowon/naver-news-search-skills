# naver-news-search-skills

네이버 뉴스 검색 API와 Exa Contents API를 활용해 한국어 뉴스를 검색하고 기사 본문을 가져오는 Go CLI 도구입니다. OpenClaw 에이전트가 `SKILL.md`를 통해 이 도구를 호출하고, 검색 결과를 바탕으로 뉴스를 요약합니다.

## 설치

Go 1.22 이상이 필요합니다.

```bash
git clone https://github.com/kowon/naver-news-search-skills.git
cd naver-news-search-skills
go build -o naver-news .
```

## 환경 변수

| 변수명 | 설명 | 발급처 |
|--------|------|--------|
| `NAVER_CLIENT_ID` | 네이버 API 클라이언트 ID | [네이버 개발자 센터](https://developers.naver.com) |
| `NAVER_CLIENT_SECRET` | 네이버 API 클라이언트 Secret | [네이버 개발자 센터](https://developers.naver.com) |
| `EXA_API_KEY` | Exa AI API 키 | [Exa AI](https://exa.ai) |

## 사용법

### 뉴스 검색

```bash
NAVER_CLIENT_ID=<id> NAVER_CLIENT_SECRET=<secret> \
  ./naver-news search --query "인공지능" --display 5
```

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--query` | (필수) | 검색어 |
| `--display` | 10 | 결과 개수 (1-100) |
| `--sort` | sim | `sim` 정확도순 / `date` 날짜순 |
| `--fetch` | false | 각 기사 본문을 Exa로 함께 가져오기 |

### 기사 본문 가져오기

```bash
EXA_API_KEY=<key> ./naver-news fetch --url "https://n.news.naver.com/..."
```

### 검색 + 본문 통합

```bash
NAVER_CLIENT_ID=<id> NAVER_CLIENT_SECRET=<secret> EXA_API_KEY=<key> \
  ./naver-news search --query "AI 반도체" --display 3 --fetch
```

## 출력 형식

모든 출력은 Markdown 형식이며, 에이전트가 직접 읽고 처리할 수 있습니다.

```markdown
# 네이버 뉴스 검색 결과: "인공지능"

총 5개 기사

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
├── internal/
│   ├── naver/
│   │   └── client.go       # 네이버 뉴스 API 클라이언트
│   └── exa/
│       └── client.go       # Exa Contents API 클라이언트
├── .claude/
│   └── skills/             # Claude 워크플로우 스킬
└── docs/
    ├── apis/               # API 명세
    ├── plans/              # 기능 계획 문서
    └── reviews/            # 코드 리뷰 및 해설
```

## 라이선스

MIT
