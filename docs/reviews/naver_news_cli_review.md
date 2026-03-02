# 코드 리뷰: naver-news CLI 구현 해설

**작성일:** 2026-03-02
**대상 파일:** `main.go`, `internal/naver/client.go`, `internal/exa/client.go`

---

## 개요

이 문서는 `naver-news` CLI 도구의 Go 코드를 단계별로 해설합니다. 네이버 뉴스 검색 API와 Exa Contents API를 호출하는 HTTP 클라이언트 구현, 서브커맨드 CLI 패턴, Go의 에러 핸들링 관용구를 중심으로 설명합니다.

---

## 상세 해설

### 1. `internal/naver/client.go` — 네이버 뉴스 API 클라이언트

#### 패키지 구조와 구조체 정의

```go
package naver

type NewsItem struct {
    Title        string `json:"title"`
    OriginalLink string `json:"originallink"`
    Link         string `json:"link"`
    Description  string `json:"description"`
    PubDate      string `json:"pubDate"`
}
```

- **`package naver`**: Go에서 패키지는 디렉토리 단위로 구성됩니다. `internal/` 하위에 두면 이 모듈 밖에서는 임포트할 수 없어 캡슐화가 강화됩니다.
- **구조체 태그 `json:"..."`**: `encoding/json` 패키지가 JSON 키 이름을 매핑할 때 사용합니다. 네이버 API는 `originallink`(소문자)를 반환하므로, Go의 관례인 `OriginalLink`와 태그로 연결합니다.

#### `<b>` 태그 제거 — 정규식 활용

```go
var boldTagRe = regexp.MustCompile(`</?b>`)

func stripBoldTags(s string) string {
    return boldTagRe.ReplaceAllString(s, "")
}
```

- **`regexp.MustCompile`**: 패키지 수준 변수로 정규식을 미리 컴파일합니다. 함수 내부에서 매번 컴파일하면 비용이 반복되므로, 이렇게 전역 변수로 한 번만 컴파일하는 것이 Go 관용구입니다.
- **`MustCompile` vs `Compile`**: `MustCompile`은 잘못된 정규식이면 패닉을 발생시킵니다. 정규식이 하드코딩된 상수라면 `MustCompile`이 적합합니다 — 런타임이 아닌 코드 작성 시점에 오류를 잡기 때문입니다.
- **`</?b>`**: `?`는 앞의 `/`가 0개 또는 1개임을 뜻합니다. `<b>`와 `</b>` 모두 매칭합니다.

#### HTTP 요청 구성

```go
params := url.Values{}
params.Set("query", query)
params.Set("display", strconv.Itoa(display))

reqURL := "https://openapi.naver.com/v1/search/news.json?" + params.Encode()

req, err := http.NewRequest(http.MethodGet, reqURL, nil)
req.Header.Set("X-Naver-Client-Id", clientID)
req.Header.Set("X-Naver-Client-Secret", clientSecret)
```

- **`url.Values`**: 쿼리 파라미터를 안전하게 URL 인코딩합니다. 한글 검색어처럼 특수문자가 포함된 경우 `%EC%9D%B8%EA%B3%B5%EC%A7%80%EB%8A%A5` 형태로 자동 인코딩됩니다. 직접 문자열을 이어붙이면 버그가 생길 수 있습니다.
- **`http.NewRequest`**: 헤더나 바디 등 세부 설정이 필요할 때 사용합니다. 단순 GET이라도 `http.Get`은 헤더를 추가할 수 없으므로 `NewRequest`를 씁니다.

#### 에러 핸들링 패턴

```go
resp, err := http.DefaultClient.Do(req)
if err != nil {
    return nil, fmt.Errorf("calling Naver API: %w", err)
}
defer resp.Body.Close()
```

- **`fmt.Errorf("context: %w", err)`**: Go 1.13부터 도입된 에러 래핑 패턴입니다. `%w`로 감싼 에러는 `errors.Is()`/`errors.As()`로 원본 에러를 추출할 수 있습니다. 에러 메시지에 문맥을 추가해 디버깅을 쉽게 합니다.
- **`defer resp.Body.Close()`**: HTTP 응답 바디는 반드시 닫아야 합니다. `defer`로 함수 종료 시 자동으로 닫히게 하는 것이 Go 관용구입니다. `err != nil` 체크 뒤에 바로 defer를 선언해야 nil 포인터 패닉을 피할 수 있습니다.

#### JSON 디코딩

```go
var result searchResponse
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, fmt.Errorf("decoding response: %w", err)
}
```

- **`json.NewDecoder` vs `json.Unmarshal`**: `NewDecoder`는 `io.Reader`를 스트리밍 방식으로 읽습니다. HTTP 응답 바디처럼 이미 `io.Reader`로 제공되는 경우에 적합합니다. `json.Unmarshal`은 `[]byte`를 메모리에 전부 읽은 후 파싱하므로 불필요한 메모리 사용이 발생합니다.

---

### 2. `internal/exa/client.go` — Exa Contents API 클라이언트

#### POST 요청과 JSON 바디

```go
type contentsRequest struct {
    IDs  []string `json:"ids"`
    Text bool     `json:"text"`
}

body := contentsRequest{IDs: []string{pageURL}, Text: true}
bodyBytes, err := json.Marshal(body)

req, err := http.NewRequest(http.MethodPost, "https://api.exa.ai/contents", bytes.NewReader(bodyBytes))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("x-api-key", apiKey)
```

- **`bytes.NewReader`**: `[]byte`를 `io.Reader`로 변환합니다. `http.NewRequest`의 세 번째 인자는 `io.Reader` 타입이므로 필요한 변환입니다.
- **`json:"ids"`**: Exa API가 요구하는 JSON 키는 `ids`(복수형 배열)입니다. URL 하나를 배열로 감싸서 전달합니다.
- **`Content-Type: application/json`**: POST 요청에서 바디가 JSON임을 서버에 알려야 합니다. 누락하면 서버가 바디를 올바르게 파싱하지 못합니다.

---

### 3. `main.go` — CLI 서브커맨드 패턴

#### 서브커맨드 라우팅

```go
switch os.Args[1] {
case "search":
    runSearch(os.Args[2:])
case "fetch":
    runFetch(os.Args[2:])
}
```

- **`os.Args`**: 프로그램 실행 시 전달된 인자 배열입니다. `os.Args[0]`은 프로그램 이름, `os.Args[1]`부터 실제 인자입니다.
- **서브커맨드 패턴**: `git commit`, `docker run`처럼 서브커맨드마다 다른 플래그 세트를 갖는 CLI의 일반적인 구조입니다. 각 서브커맨드 함수에 `os.Args[2:]`를 전달해 각자 `flag.FlagSet`으로 파싱합니다.

#### `flag.FlagSet` — 서브커맨드별 독립 플래그

```go
fs := flag.NewFlagSet("search", flag.ExitOnError)
query := fs.String("query", "", "검색어 (필수)")
display := fs.Int("display", 10, "결과 개수 (1-100, 기본 10)")
fs.Parse(args)
```

- **`flag.FlagSet` vs `flag.CommandLine`**: 전역 `flag` 패키지 함수(`flag.String`, `flag.Parse`)는 단일 플래그 세트만 지원합니다. 서브커맨드처럼 여러 플래그 세트가 필요하면 `flag.NewFlagSet`으로 독립적인 세트를 만들어야 합니다.
- **`flag.ExitOnError`**: 알 수 없는 플래그나 잘못된 값이 입력되면 자동으로 `os.Exit(2)`를 호출합니다.
- **포인터 반환**: `fs.String`, `fs.Int` 등은 값이 아닌 포인터를 반환합니다. `fs.Parse` 호출 후 포인터가 가리키는 값이 채워지므로, 파싱 전에 역참조(`*query`)하면 기본값만 얻습니다.

---

## 핵심 요약

| 개념 | 핵심 포인트 |
|------|------------|
| **패키지 캡슐화** | `internal/` 하위 패키지는 모듈 외부에서 임포트 불가 |
| **정규식 컴파일** | `var re = regexp.MustCompile(...)` 로 패키지 수준에서 한 번만 컴파일 |
| **에러 래핑** | `fmt.Errorf("context: %w", err)` 로 문맥 추가, 원본 에러 보존 |
| **`defer` 리소스 정리** | `defer resp.Body.Close()` — err 체크 직후, 함수 종료 시 자동 실행 |
| **스트리밍 JSON** | `json.NewDecoder(resp.Body).Decode(&v)` — HTTP 바디는 스트리밍 디코딩 |
| **서브커맨드 CLI** | `flag.NewFlagSet`으로 서브커맨드별 독립 플래그 세트 구성 |
| **URL 인코딩** | `url.Values` + `params.Encode()` — 한글 등 특수문자 자동 처리 |
