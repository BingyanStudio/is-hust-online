# gRPC Auth Proto Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update the proto files to implement the dual-service architecture (ClientManager + CheckService) with client authentication, metadata, heartbeat, and task assignment.

**Architecture:** Split proto into domain-specific files: `check.proto` (existing check types), `client.proto` (client messages/enums), `task.proto` (task/result messages), `client_service.proto` (ClientManager RPC), `check_service.proto` (CheckService RPC). Use `buf` for proto linting and Go code generation.

**Tech Stack:** protobuf, gRPC, buf, Go

---

## File Structure

| File | Responsibility |
|------|---------------|
| `pkg/proto/check.proto` | CheckType, ErrorType, CheckRequest, CheckResponse, ServiceStatus (existing) |
| `pkg/proto/client.proto` | ClientInfo, ClientStatus, Register/Heartbeat/Deregister messages |
| `pkg/proto/task.proto` | WatchTasksRequest, CheckTask, CheckResultRequest, CheckResultResponse |
| `pkg/proto/client_service.proto` | ClientManager service definition |
| `pkg/proto/check_service.proto` | CheckService service definition |
| `buf.yaml` | buf configuration for linting and generation |
| `buf.gen.yaml` | buf code generation config |

---

### Task 1: Create client.proto

**Files:**
- Create: `pkg/proto/client.proto`

- [ ] **Step 1: Write client.proto**

```protobuf
syntax = "proto3";
package net.bingyan.hust_uptime.v1;

option go_package = "github.com/BingyanStudio/is-hust-online/pkg/proto";

import "check.proto";

enum ClientStatus {
    CLIENT_STATUS_UNKNOWN = 0;
    CLIENT_STATUS_ONLINE = 1;
    CLIENT_STATUS_BUSY = 2;
    CLIENT_STATUS_DEGRADED = 3;
    CLIENT_STATUS_OFFLINE = 4;
}

message ClientInfo {
    string name = 1;
    string location = 2;
    repeated CheckType capabilities = 3;
    map<string, string> labels = 4;
}

message RegisterRequest {
    ClientInfo client_info = 1;
}

message RegisterResponse {
    string client_id = 1;
    bool success = 2;
    string message = 3;
}

message HeartbeatRequest {
    string client_id = 1;
    ClientStatus status = 2;
}

message HeartbeatResponse {
    bool success = 1;
    int64 server_time = 2;
}

message DeregisterRequest {
    string client_id = 1;
}

message DeregisterResponse {
    bool success = 1;
}
```

- [ ] **Step 2: Verify proto compiles**

Run: `protoc --proto_path=pkg/proto --go_out=. --go-grpc_out=. pkg/proto/client.proto`
Expected: No errors (may fail on import if check.proto not in include path — that's OK, we'll fix in Task 4 with buf)

- [ ] **Step 3: Commit**

```bash
git add pkg/proto/client.proto
git commit -m "feat(proto): add client messages and enums"
```

---

### Task 2: Create task.proto

**Files:**
- Create: `pkg/proto/task.proto`

- [ ] **Step 1: Write task.proto**

```protobuf
syntax = "proto3";
package net.bingyan.hust_uptime.v1;

option go_package = "github.com/BingyanStudio/is-hust-online/pkg/proto";

import "check.proto";

message WatchTasksRequest {
    string client_id = 1;
}

message CheckTask {
    string task_id = 1;
    CheckRequest check = 2;
    int64 assigned_at = 3;
    optional int64 deadline = 4;
}

message CheckResultRequest {
    string client_id = 1;
    string task_id = 2;
    CheckResponse result = 3;
}

message CheckResultResponse {
    bool success = 1;
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/proto/task.proto
git commit -m "feat(proto): add task and result messages"
```

---

### Task 3: Create service proto files

**Files:**
- Create: `pkg/proto/client_service.proto`
- Create: `pkg/proto/check_service.proto`

- [ ] **Step 1: Write client_service.proto**

```protobuf
syntax = "proto3";
package net.bingyan.hust_uptime.v1;

option go_package = "github.com/BingyanStudio/is-hust-online/pkg/proto";

import "client.proto";

service ClientManager {
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
    rpc Deregister(DeregisterRequest) returns (DeregisterResponse);
}
```

- [ ] **Step 2: Write check_service.proto**

```protobuf
syntax = "proto3";
package net.bingyan.hust_uptime.v1;

option go_package = "github.com/BingyanStudio/is-hust-online/pkg/proto";

import "task.proto";

service CheckService {
    rpc WatchTasks(WatchTasksRequest) returns (stream CheckTask);
    rpc ReportResult(CheckResultRequest) returns (CheckResultResponse);
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/proto/client_service.proto pkg/proto/check_service.proto
git commit -m "feat(proto): add ClientManager and CheckService definitions"
```

---

### Task 4: Set up buf for proto management

**Files:**
- Create: `buf.yaml`
- Create: `buf.gen.yaml`

- [ ] **Step 1: Install buf**

Run: `go install github.com/bufbuild/buf/cmd/buf@latest`
Expected: buf binary installed

- [ ] **Step 2: Write buf.yaml**

```yaml
version: v2
modules:
  - path: pkg/proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

- [ ] **Step 3: Write buf.gen.yaml**

```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: .
    opt: paths=source_relative
  - remote: buf.build/grpc/go
    out: .
    opt: paths=source_relative
```

- [ ] **Step 4: Run buf lint**

Run: `buf lint`
Expected: Pass or show warnings about enum prefix (we used CLIENT_STATUS_ prefix so should be clean)

- [ ] **Step 5: Commit**

```bash
git add buf.yaml buf.gen.yaml
git commit -m "chore: add buf configuration for proto management"
```

---

### Task 5: Generate Go code and verify

**Files:**
- Generate: `pkg/proto/*.pb.go`, `pkg/proto/*_grpc.pb.go`

- [ ] **Step 1: Generate Go stubs**

Run: `buf generate`
Expected: Generates `.pb.go` and `_grpc.pb.go` files in `pkg/proto/`

- [ ] **Step 2: Verify generated code compiles**

Run: `go build ./pkg/proto/...`
Expected: No errors

- [ ] **Step 3: Commit generated code**

```bash
git add pkg/proto/
git commit -m "feat(proto): generate Go stubs for gRPC services"
```

---

### Task 6: Update go.mod with dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Tidy dependencies**

Run: `go mod tidy`
Expected: Adds google.golang.org/grpc and google.golang.org/protobuf dependencies

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add gRPC and protobuf dependencies"
```
