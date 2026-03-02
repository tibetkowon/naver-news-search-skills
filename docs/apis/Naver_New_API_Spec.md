# 네이버 뉴스 검색 API 명세서

네이버 뉴스 검색 API는 네이버 뉴스 검색 결과를 XML 또는 JSON 형식으로 반환하는 RESTful API입니다.

## 1. 기본 정보
* **프로토콜**: HTTPS
* **HTTP 메서드**: GET
* **인증 방식**: 비로그인 방식 (HTTP 헤더에 클라이언트 아이디와 시크릿 전송)
* **호출 한도**: 일일 25,000회

## 2. 요청 (Request)

### 요청 URL
| 반환 형식 | URL |
| :--- | :--- |
| **JSON** | `https://openapi.naver.com/v1/search/news.json` |
| **XML** | `https://openapi.naver.com/v1/search/news.xml` |

### 요청 헤더
| 헤더 이름 | 값 | 설명 |
| :--- | :--- | :--- |
| `X-Naver-Client-Id` | {발급받은 클라이언트 아이디} | 애플리케이션 등록 시 발급받은 ID |
| `X-Naver-Client-Secret` | {발급받은 클라이언트 시크릿} | 애플리케이션 등록 시 발급받은 Secret |

### 요청 파라미터
| 파라미터 | 타입 | 필수 여부 | 설명 |
| :--- | :--- | :--- | :--- |
| `query` | String | **Y** | 검색어 (UTF-8 인코딩 필수) |
| `display` | Integer | N | 한 번에 표시할 검색 결과 개수 (기본 10, 최대 100) |
| `start` | Integer | N | 검색 시작 위치 (기본 1, 최대 1000) |
| `sort` | String | N | 정렬 방법: `sim` (정확도순, 기본값), `date` (날짜순) |

---

## 3. 응답 (Response)

### 응답 필드 (JSON/XML 공통)
| 필드명 | 타입 | 설명 |
| :--- | :--- | :--- |
| `lastBuildDate` | dateTime | 검색 결과가 생성된 시간 |
| `total` | Integer | 총 검색 결과 개수 |
| `start` | Integer | 검색 시작 위치 |
| `display` | Integer | 한 번에 표시된 검색 결과 개수 |
| `items` | Array/Item | 개별 검색 결과를 포함하는 컨테이너 |
| `items/title` | String | 뉴스 기사의 제목 (<b> 태그 포함 가능) |
| `items/originallink` | String | 뉴스 기사 원문의 URL |
| `items/link` | String | 뉴스 기사의 네이버 뉴스 URL |
| `items/description` | String | 뉴스 기사의 내용을 요약한 패시지 정보 |
| `items/pubDate` | dateTime | 뉴스 기사가 네이버에 제공된 시간 |

---

## 4. 오류 코드
| 오류 코드 | 상태 코드 | 설명 |
| :--- | :--- | :--- |
| `SE01` | 400 | 잘못된 쿼리 요청 (파라미터 등 확인) |
| `SE02` | 400 | 부적절한 display 값 (1~100 범위 밖) |
| `SE03` | 400 | 부적절한 start 값 (1~100 범위 밖) |
| `SE06` | 400 | 잘못된 인코딩 (UTF-8 미사용) |
| `SE05` | 404 | 존재하지 않는 API 주소 |
| `SE99` | 500 | 시스템 에러 |
