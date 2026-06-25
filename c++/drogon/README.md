# appfolder — C++ 애플리케이션 템플릿

C++17 / Drogon / standalone Asio 기반 템플릿으로, `go/gin/` 폴더 구조를 미러링합니다.
REST + SSE + TCP 이중 전송, 인메모리 캐시, Redis(hiredis), 다중 DB(SOCI — MariaDB / Oracle)를 지원합니다.

## 의존성

| 라이브러리 | 용도 |
|-----------|------|
| [Drogon](https://github.com/drogonframework/drogon) | HTTP REST/SSE 서버 |
| [standalone Asio](https://github.com/chriskohlhoff/asio) | TCP 서버, 비동기 I/O |
| [SOCI](https://github.com/SOCI/soci) | DB 접근 (직접 쿼리) |
| [hiredis](https://github.com/redis/hiredis) | Redis 클라이언트 |

### 의존성 설치

모든 외부 패키지는 `third_party/` 디렉터리에 배치합니다. 프로젝트 루트(`c++/drogon/`)에서 실행:

```bash
git clone --depth 1 --recurse-submodules https://github.com/drogonframework/drogon.git third_party/drogon
git clone --depth 1 https://github.com/chriskohlhoff/asio.git third_party/asio
git clone --depth 1 https://github.com/SOCI/soci.git third_party/soci
git clone --depth 1 https://github.com/redis/hiredis.git third_party/hiredis
```

CMakeLists.txt가 `third_party/` → 시스템 패키지 순서로 자동 탐색합니다.
시스템에 이미 설치된 패키지가 있다면 `third_party/`에 넣지 않아도 됩니다.

## 커맨드

```bash
# 빌드
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)

# 실행 (env.ini가 있는 cmd/server/에서)
cd cmd/server
../../build/appfolder

# Docker 빌드
./docker_build.sh
```

## 설정

`cmd/server/env.ini` 파일을 생성하고 아래 항목을 채웁니다.

```ini
[SERVER]
HOST = 0.0.0.0
PORT = 8080

[TCP]
HOST = 0.0.0.0
PORT = 9090

[DB]
TYPE = maria          # maria | oracle
HOST = localhost
PORT = 3306
USER = root
PASSWORD =
NAME = appfolder

[REDIS]
HOST = localhost
PORT = 6379
PASSWORD =

[JWT]
SECRET = your-secret-key-at-least-32-chars!!
EXPIRE_MIN = 60
```

| 항목 | 용도 |
|------|------|
| `DB.TYPE` | 드라이버 선택: `maria` / `oracle` |
| `TCP.PORT` | TCP 서버 인바운드 포트 |
| `JWT.SECRET` | HS256 서명 키 (최소 32자) |

### 시작 순서

`main()` → `Container::build()` 모든 의존성 조립 → `Application::start()` → Monitoring/Worker/TCP/REST 스레드 실행 → 시그널 대기.

### 종료 순서

SIGINT/SIGTERM 수신 → `Application::shutdown()`: Monitoring → REST → CacheWorker → CronWorker → TCP → EventBus → Redis → DB → Logger.

## 패키지 구조

```
cmd/server/               엔트리포인트 (main.cpp, env.ini)
internal/
  app/                    애플리케이션 생명주기 (start/shutdown)
  config/                 env.ini 파서
  container/              DI 컨테이너 (Go의 container.go에 대응)
  transport/
    http_rest/
      controller/         ApiController, SseController (Drogon HttpController)
      filter/             LogFilter, JwtFilter, CorsFilter (Drogon HttpFilter)
      http_utils/         JwtUtil, HttpError 헬퍼
      request/            요청 DTO
      response/           BaseResponse, 응답 DTO
      rest.h/cpp          Drogon 서버 초기화
    tcp/
      server/
        tcp_server.h/cpp      Asio TCP accept 루프
        tcp_client.h/cpp      소켓 래핑, async read + synchronized send
        client/               ClientManager — 스레드 안전 클라이언트 레지스트리
        controller/           TCP 요청 디스패처
        parser/               RtmsParser — 바이너리 프레임 파싱
        serializer/           BinaryWriter — 송신 프레임 빌드
      client/
        tcp_outbound_client   외부 TCP 엔드포인트 연결 (서버→클라이언트)
  service/                ApiService, TcpService 인터페이스 + 구현
  db/
    db_handler/            DbHandler 인터페이스, maria/, oracle/ (SOCI)
    entity/                SampleEntity, LogEntity
  infra/
    cache/                 인메모리 캐시 (std::unordered_map + mutex)
    define/                상수 정의
    object/                캐시 오브젝트 모델
  event/                   EventBus (topic 기반 pub/sub), 이벤트 타입 정의
  redis/                   hiredis 래퍼
  logger/                  파일 기반 로거 (일별 로테이션)
  healthcheck/             DB·Redis 헬스체크
  worker/                  CacheWorker, CronWorker
  utils/                   유틸리티
```
