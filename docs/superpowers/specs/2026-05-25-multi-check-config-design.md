# Multi-CheckConfig Design

## Summary

Allow each site to have multiple `CheckConfig` entries, each assigned to a specific client agent. This replaces the current design where check type/interval is stored directly on the Site model and the scheduler picks a random eligible client.

## Decisions

1. **Manual assignment**: CheckConfig is created via API with an explicit `ClientID`.
2. **Remove Site-level check fields**: `CheckType`, `CheckInterval`, `CheckExtra` are removed from the Site model. All check configuration lives in CheckConfig.
3. **Report granularity**: `CheckConfigID` is added to both `Check` and `Report` models for per-config reporting.
4. **Strict client binding**: Scheduler dispatches to the assigned client only. If that client is offline, the check is skipped.

## Data Model Changes

### Site (`internal/model/site.go`)

Remove `CheckType`, `CheckInterval`, `CheckExtra` (these fields don't exist on the current struct but are referenced in controller/scheduler). Final fields:

- `ID`, `Name`, `Description`, `URL`, `Status`, `Type`, `CreatedAt`

### CheckConfig (`internal/model/check_config.go`)

Use as-is:

```go
type CheckConfig struct {
    ID            bson.ObjectID `bson:"_id"`
    SiteID        bson.ObjectID `bson:"site_id"`
    ClientID      bson.ObjectID `bson:"client_id"`
    Status        int32         `bson:"status"`
    CheckType     int32         `bson:"check_type"`
    CheckInterval string        `bson:"check_interval"`
    CheckExtra    string        `bson:"check_extra"`
}
```

### Check (`internal/model/check.go`)

Add field:

```go
CheckConfigID bson.ObjectID `json:"check_config_id" bson:"check_config_id"`
```

### Report (`internal/model/report.go`)

Add field:

```go
CheckConfigID bson.ObjectID `json:"check_config_id" bson:"check_config_id"`
```

Report aggregation key changes from `(SiteID, Type, Timeframe)` to `(SiteID, CheckConfigID, Type, Timeframe)`.

## DAO Layer

New file: `internal/dao/check_config.go`

| Method | Purpose |
|--------|---------|
| `InsertCheckConfig(ctx, *model.CheckConfig) error` | Create |
| `FindCheckConfigByID(ctx, bson.ObjectID) (*model.CheckConfig, error)` | Read by ID |
| `FindCheckConfigs(ctx, filter bson.M, page, pageSize int64) ([]model.CheckConfig, int64, error)` | Paginated list with filter |
| `UpdateCheckConfig(ctx, bson.ObjectID, bson.M) error` | Partial update |
| `DeleteCheckConfig(ctx, bson.ObjectID) error` | Delete |
| `FindCheckConfigsBySiteID(ctx, bson.ObjectID) ([]model.CheckConfig, error)` | All configs for a site (scheduler) |
| `FindCheckConfigsByClientID(ctx, bson.ObjectID) ([]model.CheckConfig, error)` | All configs for a client |

## API Layer

New file: `internal/controller/check_configs.go`

### Routes

| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| `GET` | `/api/check-configs` | `ListCheckConfigs` | Paginated, filterable by `site_id`, `client_id` |
| `GET` | `/api/check-configs/:id` | `GetCheckConfig` | Single config by ID |
| `POST` | `/api/check-configs` | `CreateCheckConfig` | Create config; validate site/client exist and client has matching capability |
| `PUT` | `/api/check-configs/:id` | `UpdateCheckConfig` | Partial update |
| `DELETE` | `/api/check-configs/:id` | `DeleteCheckConfig` | Delete config |

### CreateCheckConfig validation

- `SiteID` must reference an existing site
- `ClientID` must reference an existing client
- Client's `Capabilities` bitmask must include the requested `CheckType`

### Site controller changes

`CreateSite` and `UpdateSite` no longer accept `CheckType`, `CheckInterval`, `CheckExtra`.

## Scheduler Changes (`internal/service/scheduler.go`)

### Current flow (per tick)

1. Load all enabled sites
2. For each site: parse `site.CheckInterval`, filter eligible clients, pick random, dispatch

### New flow (per tick)

1. Load all enabled sites
2. For each site, load `CheckConfigs` via `dao.FindCheckConfigsBySiteID(site.ID)`
3. For each enabled check config:
   - Parse `config.CheckInterval`
   - Look up `config.ClientID` in the dispatcher's online clients map
   - If client is offline → skip
   - If `now < nextRun` → skip
   - Dispatch `CheckTask` with `config.CheckType` and `config.ID` to the assigned client
4. `lastRun` key changes from `siteID` to `checkConfigID`

### Task construction

```go
task := &myproto.CheckTask{
    TaskId: bson.NewObjectID().Hex(),
    Check: &myproto.CheckRequest{
        Id:             config.SiteID.Hex(),
        CheckConfigId:  config.ID.Hex(),  // new field
        Url:            site.URL,
        CheckType:      myproto.CheckType(config.CheckType),
        Method:         "GET",
    },
    AssignedAt: now.Unix(),
}
```

## Proto Changes (`pkg/proto/net/bingyan/hust_uptime/v1/check.proto`)

Add `check_config_id` to `CheckRequest` and `CheckResponse`:

```protobuf
message CheckRequest {
    // ... existing fields ...
    string check_config_id = 7;
}

message CheckResponse {
    // ... existing fields ...
    string check_config_id = 8;
}
```

## Check Result Flow Changes (`internal/service/check_service.go`)

`ReportResult`:

- Parse `CheckConfigID` from the incoming `CheckResponse`
- Store it on the `Check` document
- Use `(SiteID, CheckConfigID, Type, Timeframe)` as the report aggregation key for upsert

## Files Changed

| File | Change |
|------|--------|
| `internal/model/site.go` | Remove check fields (already missing from struct) |
| `internal/model/check.go` | Add `CheckConfigID` field |
| `internal/model/report.go` | Add `CheckConfigID` field |
| `internal/model/check_config.go` | No change |
| `internal/dao/check_config.go` | **New** — CRUD for CheckConfig collection |
| `internal/controller/check_configs.go` | **New** — HTTP handlers for CheckConfig API |
| `internal/controller/sites.go` | Remove check field handling from Create/Update |
| `internal/views/index.go` | Register check-config routes |
| `internal/service/scheduler.go` | Rewrite tick() to iterate check configs |
| `internal/service/check_service.go` | Handle CheckConfigID in ReportResult |
| `pkg/proto/.../check.proto` | Add check_config_id fields |
| `pkg/proto/client.pb.go` | Regenerate |
| `cmd/client/main.go` | Pass check_config_id through check flow |
