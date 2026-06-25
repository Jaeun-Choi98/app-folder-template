# appfolder — Spring Boot 애플리케이션 템플릿

Spring Boot 3.5 / Java 17 기반 템플릿으로, `go/gin/` 폴더 구조를 미러링합니다.
REST + SSE + TCP 이중 전송, 인메모리 캐시, Redis, 다중 DB(MariaDB / Oracle)를 지원합니다.

## 커맨드

```bash
# 개발 서버 실행
./gradlew bootRun

# JAR 빌드
./gradlew build

# JAR 실행
java -jar build/libs/appfolder-0.0.1-SNAPSHOT.jar

# 전체 테스트
./gradlew test

# 단일 테스트 클래스 실행
./gradlew test --tests "juchoi.template.appfolder.transport.http.ApiControllerTest"
```

## 설정

실행 전에 `src/main/resources/application.properties`를 채워야 합니다.
`db.type`으로 JDBC 드라이버를 선택합니다: `maria` 또는 `oracle`.

주요 속성:

| 속성                  | 용도                                  |
| --------------------- | ------------------------------------- |
| `db.type`             | 드라이버 선택: `maria` / `oracle`     |
| `spring.datasource.*` | JDBC 연결 (URL, 사용자, 비밀번호)     |
| `spring.data.redis.*` | Redis 호스트, 포트, 비밀번호          |
| `tcp.port`            | TCP 서버 인바운드 포트                |
| `jwt.secret`          | HS256 서명 키 (최소 32자)             |

### 시작 순서

`AppfolderApplication.main()` → Spring IoC가 모든 빈 생성 → `AppLifecycle.start()` (SmartLifecycle) → TCP 서버가 연결 수락 시작 → `@Scheduled` 워커 활성화.

### 종료 순서

`SmartLifecycle.stop()` (가장 높은 phase = 먼저 종료): TcpServer → CacheWorker → TcpWorker → Spring이 나머지 빈 정리.

## 패키지 구조

```
config/           @Configuration 빈 — DI 루트 (Go의 container.go에 대응)
app/              SmartLifecycle — 시작/종료 순서 관리 (Go의 app.go에 대응)
transport/
  http/
    controller/   @RestController, SseController
    filter/       LogFilter, JwtFilter (OncePerRequestFilter)
    request/      요청 DTO
    response/     BaseResponse<T>, 응답 DTO
    util/         JwtUtil
  tcp/
    server/       TcpServer, TcpClient, ClientManager, RtmsParser, BinaryWriter
    client/       TcpOutboundClient
service/          ApiService, TcpService 인터페이스 + impl/
worker/           CacheWorker, TcpWorker, CronWorker (@Scheduled)
eventbus/         EventBus, event/TcpSendEvent, event/CacheDataEvent
db/
  handler/        DbHandler 인터페이스, maria/, oracle/
  entity/         SampleEntity (record)
infra/cache/      Cache (ConcurrentHashMap + snapshot)
redis/            RedisModel
healthcheck/      SystemMonitoring
model/            도메인 모델
```

## TCP 패키지 — Java vs Go 비교

TCP 계층은 Java와 Go 간 차이가 가장 큰 부분입니다.

### 핵심 차이점

|                     | Go                       | Java (이 템플릿)                |
| ------------------- | ------------------------ | ------------------------------- |
| 동시성 단위         | 고루틴 (~2 KB 스택)      | 스레드 (~1 MB 스택)             |
| 연결당 비용         | 거의 없음                | 상당히 큼                       |
| 방식                | 연결마다 고루틴 생성     | `ExecutorService` (스레드 풀)   |
| 업그레이드 경로     | 이미 최적               | Java 21+: 가상 스레드           |

Go는 고루틴이 가볍기 때문에 연결마다 고루틴을 거리낌 없이 생성합니다.
Java 스레드는 약 500배 무겁기 때문에, `Executors.newCachedThreadPool()`을 사용하여 연결 간 스레드를 재사용합니다. 코드 구조는 동일(연결당 하나의 읽기 루프)하지만, 기저의 OS 자원 사용량은 크게 다릅니다.

**Java 21 이상**에서는 `TcpServer`의 executor를 다음과 같이 교체하면 됩니다:

```java
Executors.newVirtualThreadPerTaskExecutor()
```

이렇게 하면 다른 코드 변경 없이 고루틴 수준의 경량 연결 처리가 가능합니다.

### TCP 프레임 처리 흐름

```
1. 클라이언트가 TCP로 바이트를 전송
        ↓
2. TcpServer.acceptLoop() — ServerSocket.accept()가 Socket 반환
        ↓
3. TcpClient 생성; executor.submit(client::readLoop)으로 스레드 할당
        ↓
4. readLoop()가 RtmsParser.parse(InputStream) 호출
   프레임 형식: [1B opcode][4B big-endian 길이][NB 본문]
        ↓
5. TcpClient.dispatch(frame) — opcode에 따라 분기 → TcpService.handle0x01() 등
        ↓
6. TcpService가 Cache 또는 DB를 업데이트
```

### REST에서 TCP 전송을 트리거하는 흐름

```
1. GET /send-work-reply
        ↓
2. ApiController가 TcpSendEvent(clientId, opcode, data) 생성
   EventBus 토픽 "tcpclient.send"에 발행
        ↓
3. TcpWorker 핸들러 (EventBus executor 스레드에서 실행)가 이벤트 수신
        ↓
4. ClientManager.get(clientId) → TcpClient
   BinaryWriter.build(opcode, data) → byte[]
   TcpClient.send(frame) → synchronized로 OutputStream에 쓰기
        ↓
5. event.complete(reply) → CompletableFuture 완료
        ↓
6. ApiController의 future.get(10s) 블로킹 해제 → 브라우저에 응답 반환
```

Go에서는 5-6단계에 채널(`Response chan TCPResponse`)을 사용합니다.
Java에서는 `CompletableFuture<byte[]>`를 사용하며, 블로킹/언블로킹 의미론은 동일합니다.

### 주요 클래스

| 클래스              | 역할                                                           |
| ------------------- | -------------------------------------------------------------- |
| `TcpServer`         | `ServerSocket` accept 루프, 연결당 스레드 하나 생성            |
| `TcpClient`         | `Socket` 래핑; 블로킹 `readLoop()` + synchronized `send()`    |
| `ClientManager`     | `ConcurrentHashMap<Integer, TcpClient>` — 스레드 안전 레지스트리 |
| `RtmsParser`        | `InputStream`에서 바이트를 읽어 타입이 지정된 `Frame` 레코드 반환 |
| `BinaryWriter`      | opcode + body로 송신용 바이트 배열 생성                        |
| `TcpOutboundClient` | 외부 TCP 엔드포인트에 연결 (서버가 클라이언트 역할)            |
| `TcpWorker`         | EventBus 구독자; REST → TCP 전송 경로를 연결                   |
| `TcpSendEvent`      | 페이로드 구조체; 응답형 전송을 위한 `CompletableFuture` 포함   |
