# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**is-hust-online** (华科大在线吗?) is a distributed uptime monitoring system for HUST websites. It uses a server-client architecture where the server dispatches check tasks to registered client agents via gRPC, and clients perform HTTP/Ping/TCP checks and report results back.

Module: `github.com/BingyanStudio/is-hust-online` (Go 1.26.3)

## Build & Run Commands

```bash
# Build the server
go build ./cmd/server/

# Build the client agent
go build ./cmd/client/

# Regenerate protobuf code (requires buf CLI)
buf generate

# Run the server (needs config/default.yaml - not committed to repo)
go run ./cmd/server/

# Run a client agent
go run ./cmd/client/ --server localhost:9090 --token <token> --capabilities http,ping,tcp
```

No test suite, CI/CD, Makefile, or Dockerfile exists yet.

## Architecture

### Dual-Protocol Design

- **REST API** (Echo v5, port from config): Admin/management endpoints under `/api` for sites, clients, checks, and reports. Protected by HTTP Basic Auth for mutations (POST/PUT/DELETE).
- **gRPC** (port from config.C.GRPCPort): Client agent communication. `ClientManager` service handles registration/heartbeat/deregister. `CheckService` streams tasks to clients and receives results. Token-based auth via unary interceptor.

### Data Flow

1. Scheduler (`service/scheduler.go`) ticks every 10s, finds enabled sites, picks a random online client per site, and pushes tasks to `TaskDispatcher`.
2. `TaskDispatcher` (`service/task_dispatcher.go`) holds per-client buffered channels (size 16) and delivers tasks to the `WatchTasks` server-stream.
3. Client performs the check (HTTP/Ping/TCP) and calls `ReportResult`.
4. `ReportResult` saves a `Check` document to MongoDB and upserts `Report` aggregates (hourly/daily/monthly).

### Package Layout

- `cmd/server/` and `cmd/client/` - Two separate binaries (server and agent)
- `internal/config/` - Viper-based config: reads `config/default.yaml` locally, or fetches from config center in production (when `POD_NAME` env var is set)
- `internal/model/` - MongoDB document structs: Site, Client, Check, Report
- `internal/dao/` - MongoDB CRUD operations per collection (sites, clients, checks, reports)
- `internal/service/` - gRPC service implementations and the scheduler/dispatcher
- `internal/controller/` - HTTP request handlers (Echo v5)
- `internal/views/` - Route registration wiring controllers to Echo routes
- `internal/middleware/` - Basic auth, error handler, request logging (slog, warns >500ms)
- `pkg/proto/` - Protobuf definitions and generated Go code

### Storage

- **MongoDB** (4 collections): `sites`, `clients`, `checks`, `reports`
- **Redis**: Initialized but not actively used in current code

### Protobuf

Proto files live in `pkg/proto/net/bingyan/hust_uptime/v1/`. Generated code is committed alongside them. Configured via `buf.yaml` and `buf.gen.yaml` at repo root (buf v2).

## Configuration

Local development requires `config/default.yaml` (gitignored). Key fields: `mongo` (uri, db), `redis` (addr, password, db), `port`, `grpc_port`, `basic_auth_user`, `basic_auth_password`, `captcha` (geetest), `debug`.
