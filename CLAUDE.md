# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go backend application template standardizing folder structure for systems with REST/SSE + TCP dual transport, in-memory caching, and multi-database support. The canonical implementation is in `go/gin/`. The `go/gorilla/` directory is a deprecated predecessor — refer only to `go/gin/`.

## Commands

All commands run from `go/gin/`.

**Run the server** (must run from `cmd/server/` so `env.ini` and `log/` resolve correctly):
```bash
cd cmd/server && go run main.go
```

**Build binary:**
```bash
cd cmd/server && go build -o ./main main.go
```

**Build Docker image** (from `go/gin/`):
```bash
./docker_build.sh
```

**Run all tests:**
```bash
go test ./...
```

**Run a single test package** (tests also resolve `env.ini` relative to their working directory — each test package has an `envfile_here.txt` placeholder showing where `env.ini` belongs):
```bash
go test ./internal/transport/http-rest/test/
go test ./internal/db/db-handler/test/
go test ./internal/service/test/
go test ./internal/logger/test/
```

## Configuration

Config is loaded from `env.ini` in the working directory at startup (parsed by `go-ini`). Copy `cmd/server/envfile_here.txt` → `cmd/server/env.ini` and fill in values. The `[DB] TYPE` key selects the driver: `oracle`, `maria`, `mysql`, `postgres`, `tibero`.

## Architecture

### Startup sequence

`main.go` → `container.NewContainer()` wires every dependency → `app.NewApplication()` → `app.Start()` launches goroutines → main loop blocks on context cancellation.

`container.go` is the single-file DI root. All components are constructed there in dependency order; no global singletons elsewhere (except the logger and Redis client, which are set once via `SetLogger`/`InitRedis`).

### Goroutine model

Shutdown order is hard-coded in `app.Shutdown()`: SystemMonitoring → REST → DB log manager → Workers → TCP → services → EventBus → Redis → logger. Reversing this order risks panics.

| Goroutine | Purpose |
|-----------|---------|
| REST server | Gin HTTP/SSE |
| TCP server | Binary protocol over raw TCP |
| CacheWorker | Polls cache every 1s, syncs to DB |
| TCPWorker | Subscribes to `tcpclient.send` EventBus topic, drives outbound TCP |
| CronWorker | Scheduled tasks |
| Log cleanup | Rotates daily log files, deletes files older than 3 days |
| DB log manager | Buffers and flushes event logs to DB asynchronously |

### EventBus

EventBus (`github.com/Jaeun-Choi98/modules/eventbus`) is the internal message broker. The deprecated local copy at `internal/eventbus/eventbus.go` is commented-out dead code — do not use it. Custom event types and topics are defined in `internal/eventbus/event-define/custom_event.go`.

Key topics:
- `TCPNoReplyType` / `TCPWithReplyType` — REST handler publishes, TCPWorker subscribes
- `cache.data` / `system.state` / `system.event` — CacheWorker/Healthcheck publishes, REST/SSE handler subscribes and streams to browsers

### Transport layer

**REST (`internal/transport/http-rest/`)**: Gin router initialized in `controller.RoutePath()`. Middleware stack: logging → error recovery → CORS → (per-group) JWT or SSE headers. JWT is delivered via HTTP-only cookie. SPA static files are served from `cmd/server/build/` (root) and `cmd/server/spa-test/` (`/spatest` prefix).

**TCP (`internal/transport/tcp/`)**: Custom binary protocol. `server/parser/` handles frame parsing (RTMS and text formats); `server/serializer/` writes binary frames. `server/client/client_manager.go` tracks connected clients by `uint32` ID. A standalone mock client lives at `server/client-mock/` (its own `go.mod`).

### Database layer

`DBHandlerInterface` in `internal/db/db-handler/db_handler.go` abstracts all DB access. `NewDBHandler` selects the concrete driver based on `config.DBType`. Implementations are in `maria/` and `oracle/` subdirectories. Add new query methods to the interface and implement them in each driver file.

### In-memory cache (`internal/infra/cache/`)

`Cache` wraps an in-memory store with operations split into separate files (`get.go`, `set.go`, `delete.go`, `load.go`, `ect.go`). It is loaded from DB on startup and kept in sync by CacheWorker.

### Logger

Global logger is file-based (not structured). Writes to `log/YYYY/MM/DD/info` and `log/YYYY/MM/DD/debug`. The `log/` directory is created automatically relative to the working directory. Each test package needs its own `log/` path — the `test/log/logfile_here.txt` placeholders mark where these will appear.
