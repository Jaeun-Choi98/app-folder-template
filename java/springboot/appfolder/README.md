# appfolder — Spring Boot Application Template

Spring Boot 3.5 / Java 17 template mirroring the `go/gin/` folder structure.
Supports REST + SSE + TCP dual transport, in-memory cache, Redis, and multi-database (MariaDB / Oracle).

## Commands

```bash
# Run dev server
./gradlew bootRun

# Build JAR
./gradlew build

# Run JAR
java -jar build/libs/appfolder-0.0.1-SNAPSHOT.jar

# Run all tests
./gradlew test

# Run a single test class
./gradlew test --tests "juchoi.template.appfolder.transport.http.ApiControllerTest"
```

## Configuration

Fill in `src/main/resources/application.properties` before running.
`db.type` selects the JDBC driver: `maria` or `oracle`.

Key properties:

| Property              | Purpose                               |
| --------------------- | ------------------------------------- |
| `db.type`             | Driver selection: `maria` / `oracle`  |
| `spring.datasource.*` | JDBC connection (URL, user, password) |
| `spring.data.redis.*` | Redis host, port, password            |
| `tcp.port`            | Inbound TCP server port               |
| `jwt.secret`          | HS256 signing key (min 32 chars)      |

### Startup sequence

`AppfolderApplication.main()` → Spring IoC builds all beans → `AppLifecycle.start()` (SmartLifecycle) → TCP server starts accepting → `@Scheduled` workers activate.

### Shutdown sequence

`SmartLifecycle.stop()` (highest phase = stops first): TcpServer → CacheWorker → TcpWorker → Spring closes remaining beans.

## Package overview

```
config/           @Configuration beans — DI root (mirrors container.go)
app/              SmartLifecycle — startup/shutdown order (mirrors app.go)
transport/
  http/
    controller/   @RestController, SseController
    filter/       LogFilter, JwtFilter (OncePerRequestFilter)
    request/      Request DTOs
    response/     BaseResponse<T>, response DTOs
    util/         JwtUtil
  tcp/
    server/       TcpServer, TcpClient, ClientManager, RtmsParser, BinaryWriter
    client/       TcpOutboundClient
service/          ApiService, TcpService interfaces + impl/
worker/           CacheWorker, TcpWorker, CronWorker (@Scheduled)
eventbus/         EventBus, event/TcpSendEvent, event/CacheDataEvent
db/
  handler/        DbHandler interface, maria/, oracle/
  entity/         SampleEntity (record)
infra/cache/      Cache (ConcurrentHashMap + snapshot)
redis/            RedisModel
healthcheck/      SystemMonitoring
model/            Domain models
```

## TCP Package — Java vs Go

This is the area where Java differs most from Go.

### The core difference

|                     | Go                       | Java (this template)            |
| ------------------- | ------------------------ | ------------------------------- |
| Concurrency unit    | goroutine (~2 KB stack)  | thread (~1 MB stack)            |
| Per-connection cost | near-zero                | significant                     |
| Approach            | goroutine per connection | `ExecutorService` (thread pool) |
| Upgrade path        | already optimal          | Java 21+: virtual threads       |

Go spawns a goroutine per connection without hesitation because they are cheap.
Java threads are 500× heavier, so we use `Executors.newCachedThreadPool()` to reuse threads across connections. The code structure is identical — one read loop per connection — but the underlying OS resource usage is very different.

**On Java 21+**, replace the executor in `TcpServer` with:

```java
Executors.newVirtualThreadPerTaskExecutor()
```

This restores goroutine-level weight per connection with no other code changes.

### How a frame travels through the TCP layer

```
1. Client sends bytes over TCP
        ↓
2. TcpServer.acceptLoop() — ServerSocket.accept() returns a Socket
        ↓
3. TcpClient is created; executor.submit(client::readLoop) spawns a thread
        ↓
4. readLoop() calls RtmsParser.parse(InputStream)
   Frame format: [1B opcode][4B big-endian length][NB body]
        ↓
5. TcpClient.dispatch(frame) — switch on opcode → TcpService.handle0x01() etc.
        ↓
6. TcpService updates Cache or DB
```

### How REST triggers an outbound TCP send

```
1. GET /send-work-reply
        ↓
2. ApiController creates TcpSendEvent(clientId, opcode, data)
   publishes to EventBus topic "tcpclient.send"
        ↓
3. TcpWorker handler (running on EventBus executor thread) receives event
        ↓
4. ClientManager.get(clientId) → TcpClient
   BinaryWriter.build(opcode, data) → byte[]
   TcpClient.send(frame) → synchronized write to OutputStream
        ↓
5. event.complete(reply) → CompletableFuture resolved
        ↓
6. ApiController's future.get(10s) unblocks → returns reply to browser
```

Go uses a channel for step 5-6 (`Response chan TCPResponse`).
Java uses `CompletableFuture<byte[]>` — same blocking/unblocking semantics.

### Key classes

| Class               | Responsibility                                                  |
| ------------------- | --------------------------------------------------------------- |
| `TcpServer`         | `ServerSocket` accept loop, spawns one thread per connection    |
| `TcpClient`         | Wraps a `Socket`; blocking `readLoop()` + synchronized `send()` |
| `ClientManager`     | `ConcurrentHashMap<Integer, TcpClient>` — thread-safe registry  |
| `RtmsParser`        | Reads bytes from `InputStream`, returns typed `Frame` records   |
| `BinaryWriter`      | Builds outbound byte arrays from opcode + body                  |
| `TcpOutboundClient` | Connects to external TCP endpoints (server-as-client role)      |
| `TcpWorker`         | EventBus subscriber; bridges REST → TCP send path               |
| `TcpSendEvent`      | Payload struct; holds `CompletableFuture` for reply-type sends  |
