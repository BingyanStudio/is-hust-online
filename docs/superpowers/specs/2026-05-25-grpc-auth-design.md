# gRPC Uptime Checker — 鉴权与客户端标识设计

## 概述

为 is-hust-online uptime checker 系统设计 gRPC proto，实现 server-client 鉴权和 client 标识。

## 架构

采用**双服务**方案：

- **ClientManager** — 管理 client 生命周期（注册、心跳、注销）
- **CheckService** — 业务逻辑（任务推送、结果上报）

## 鉴权机制

- client 在 `Register` RPC 调用时通过 gRPC metadata 携带 `authorization: Bearer <token>`
- server 端 interceptor 统一校验 token
- 校验通过后 server 分配 `client_id`，后续心跳和结果上报使用 `client_id` 标识身份
- token 只在注册时校验一次，减轻 server 认证压力

## 消息定义

### Client 元数据

```protobuf
message ClientInfo {
    string name = 1;                    // 可读名称，如 "hk-node-1"
    string location = 2;                // 网络/地理位置，如 "华科校园网"
    repeated CheckType capabilities = 3; // 支持的检查类型
    map<string, string> labels = 4;     // 扩展标签
}

enum ClientStatus {
    UNKNOWN = 0;
    ONLINE = 1;
    BUSY = 2;
    DEGRADED = 3;
    OFFLINE = 4;
}
```

### 注册

```protobuf
message RegisterRequest {
    ClientInfo client_info = 1;
}

message RegisterResponse {
    string client_id = 1;
    bool success = 2;
    string message = 3;
}
```

### 心跳

```protobuf
message HeartbeatRequest {
    string client_id = 1;
    ClientStatus status = 2;
}

message HeartbeatResponse {
    bool success = 1;
    int64 server_time = 2;
}
```

### 注销

```protobuf
message DeregisterRequest {
    string client_id = 1;
}

message DeregisterResponse {
    bool success = 1;
}
```

### 任务与结果

```protobuf
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

## 服务定义

### ClientManager

```protobuf
service ClientManager {
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
    rpc Deregister(DeregisterRequest) returns (DeregisterResponse);
}
```

### CheckService

```protobuf
service CheckService {
    rpc WatchTasks(WatchTasksRequest) returns (stream CheckTask);
    rpc ReportResult(CheckResultRequest) returns (CheckResultResponse);
}
```

## 通信流程

```
Client                          Server
  |                                |
  |--- Register(token, info) ----->|  首次认证，获取 client_id
  |<-- client_id ------------------|
  |                                |
  |--- WatchTasks(client_id) ----->|  打开 server-stream
  |<-- CheckTask ------------------|  server 持续推送任务
  |<-- CheckTask ------------------|
  |                                |
  |--- ReportResult(task, result)->|  unary 上报结果
  |<-- success --------------------|
  |                                |
  |--- Heartbeat(client_id, st) -->|  定期心跳
  |<-- success, server_time -------|
  |                                |
  |--- Deregister(client_id) ----->|  优雅退出
  |<-- success --------------------|
```

## 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 认证方式 | 静态 Token | 简单直接，适合中小规模 |
| Token 校验 | 仅 Register 时 | 减轻 server 压力，后续用 client_id |
| 任务推送 | Server-stream | 实时性好，server 可控调度 |
| 心跳方式 | Unary RPC | 简单可靠，不依赖流状态 |
| 结果上报 | Unary RPC | 每次检查独立上报，简单直接 |
| 扩展性 | labels map + extra bytes | 预留扩展空间 |
