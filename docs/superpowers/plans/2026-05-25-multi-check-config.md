# Multi-CheckConfig Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow each site to have multiple CheckConfig entries, each assigned to a specific client, replacing the current single-check-type-per-site design.

**Architecture:** CheckConfig is a first-class entity with its own CRUD API. The scheduler iterates check configs instead of sites, dispatching tasks to the assigned client. Check and Report models gain a `CheckConfigID` field for per-config granularity.

**Tech Stack:** Go 1.26.3, Echo v5, MongoDB (mongo-driver v2), gRPC + protobuf, buf v2

---

### Task 1: Proto — Add check_config_id to CheckRequest and CheckResponse

**Files:**
- Modify: `pkg/proto/net/bingyan/hust_uptime/v1/check.proto`
- Regenerate: `pkg/proto/*.go` (via `buf generate`)

- [ ] **Step 1: Add check_config_id field to check.proto**

In `pkg/proto/net/bingyan/hust_uptime/v1/check.proto`, add field 7 to `CheckRequest` and field 8 to `CheckResponse`:

```protobuf
message CheckRequest {
  string id = 1;
  string url = 2;
  string method = 3;
  CheckType check_type = 4;
  optional int32 timeout_seconds = 5;
  optional bytes extra = 6;
  string check_config_id = 7;
}

message CheckResponse {
  string id = 1;
  bool success = 2;
  ErrorType error_type = 3;
  google.protobuf.Timestamp timestamp = 4;
  optional int32 response_time_ms = 5;
  optional bytes extra = 6;
  CheckType check_type = 7;
  string check_config_id = 8;
}
```

- [ ] **Step 2: Regenerate protobuf code**

Run: `buf generate`

Expected: regenerated files in `pkg/proto/` with new `CheckConfigId` fields on `CheckRequest` and `CheckResponse`.

- [ ] **Step 3: Verify generated code compiles**

Run: `go build ./pkg/proto/...`
Expected: success (or only errors from other packages, not proto itself).

- [ ] **Step 4: Commit**

```bash
git add pkg/proto/net/bingyan/hust_uptime/v1/check.proto pkg/proto/
git commit -m "proto: add check_config_id to CheckRequest and CheckResponse"
```

---

### Task 2: Model — Add CheckConfigID to Check and Report

**Files:**
- Modify: `internal/model/check.go`
- Modify: `internal/model/report.go`

- [ ] **Step 1: Add CheckConfigID to Check model**

In `internal/model/check.go`, add the field after `ClientID`:

```go
type Check struct {
	ID            bson.ObjectID     `json:"id" bson:"_id,omitempty"`
	SiteID        bson.ObjectID     `json:"site_id" bson:"site_id"`
	ClientID      bson.ObjectID     `json:"client_id" bson:"client_id"`
	CheckConfigID bson.ObjectID     `json:"check_config_id" bson:"check_config_id"`
	Timestamp     int64             `json:"timestamp" bson:"timestamp"`
	Type          myproto.CheckType `json:"type" bson:"type"`
	Status        myproto.ErrorType `json:"status" bson:"status"`
	Result        string            `json:"result" bson:"result"`
	Delay         int64             `json:"delay" bson:"delay"`
}
```

- [ ] **Step 2: Add CheckConfigID to Report model**

In `internal/model/report.go`, add the field after `SiteID`:

```go
type Report struct {
	SiteID        bson.ObjectID `json:"site_id" bson:"site_id"`
	CheckConfigID bson.ObjectID `json:"check_config_id" bson:"check_config_id"`
	Timeframe     string        `json:"timeframe" bson:"timeframe"`
	Type          int           `json:"type" bson:"type"`
	Checks        int64         `json:"checks" bson:"checks"`
	Successes     int64         `json:"successes" bson:"successes"`
	Uptime        float64       `json:"uptime" bson:"uptime"`
	AvgDelay      float64       `json:"avg_delay" bson:"avg_delay"`
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/model/check.go internal/model/report.go
git commit -m "model: add CheckConfigID to Check and Report"
```

---

### Task 3: DAO — Create CheckConfig DAO

**Files:**
- Create: `internal/dao/check_config.go`

- [ ] **Step 1: Create check_config.go DAO**

Create `internal/dao/check_config.go`:

```go
package dao

import (
	"context"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const checkConfigCollection = "check_configs"

func InsertCheckConfig(ctx context.Context, cc *model.CheckConfig) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).InsertOne(ctx, cc)
	return err
}

func FindCheckConfigByID(ctx context.Context, id bson.ObjectID) (*model.CheckConfig, error) {
	var cc model.CheckConfig
	err := db.MongoDB.Collection(checkConfigCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&cc)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func FindCheckConfigs(ctx context.Context, filter bson.M, page, pageSize int64) ([]model.CheckConfig, int64, error) {
	col := db.MongoDB.Collection(checkConfigCollection)

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"_id": -1})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, 0, err
	}
	return ccs, total, nil
}

func UpdateCheckConfig(ctx context.Context, id bson.ObjectID, update bson.M) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func DeleteCheckConfig(ctx context.Context, id bson.ObjectID) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func FindCheckConfigsBySiteID(ctx context.Context, siteID bson.ObjectID) ([]model.CheckConfig, error) {
	cursor, err := db.MongoDB.Collection(checkConfigCollection).Find(ctx, bson.M{"site_id": siteID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, err
	}
	return ccs, nil
}

func FindCheckConfigsByClientID(ctx context.Context, clientID bson.ObjectID) ([]model.CheckConfig, error) {
	cursor, err := db.MongoDB.Collection(checkConfigCollection).Find(ctx, bson.M{"client_id": clientID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, err
	}
	return ccs, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/dao/...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/dao/check_config.go
git commit -m "dao: add CheckConfig CRUD operations"
```

---

### Task 4: Shared — Create checktype bitmask package

**Files:**
- Create: `internal/checktype/checktype.go`

The `Bit` function is needed by both the scheduler (service package) and the check-config controller. To avoid circular imports, put it in a shared package.

- [ ] **Step 1: Create internal/checktype/checktype.go**

Create directory `internal/checktype/` and file `internal/checktype/checktype.go`:

```go
package checktype

import (
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
)

// Bit returns the bitmask value for a given CheckType.
func Bit(ct myproto.CheckType) int32 {
	switch ct {
	case myproto.CheckType_CHECK_TYPE_HTTP:
		return 1
	case myproto.CheckType_CHECK_TYPE_PING:
		return 2
	case myproto.CheckType_CHECK_TYPE_TCP:
		return 4
	case myproto.CheckType_CHECK_TYPE_OTHER:
		return 8
	default:
		return 0
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/checktype/
git commit -m "shared: add checktype bitmask utility package"
```

---

### Task 5: Controller — Create CheckConfig API

**Files:**
- Create: `internal/controller/check_configs.go`

- [ ] **Step 1: Create check_configs.go controller**

Create `internal/controller/check_configs.go`:

```go
package controller

import (
	"net/http"

	"github.com/BingyanStudio/is-hust-online/internal/checktype"
	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateCheckConfigRequest struct {
	SiteID        string `json:"site_id" validate:"required"`
	ClientID      string `json:"client_id" validate:"required"`
	CheckType     int32  `json:"check_type" validate:"required"`
	CheckInterval string `json:"check_interval" validate:"required"`
	CheckExtra    string `json:"check_extra"`
}

type UpdateCheckConfigRequest struct {
	ClientID      *string `json:"client_id"`
	CheckType     *int32  `json:"check_type"`
	CheckInterval *string `json:"check_interval"`
	CheckExtra    *string `json:"check_extra"`
	Status        *int32  `json:"status"`
}

func ListCheckConfigs(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

	filter := bson.M{}
	if siteID := c.QueryParam("site_id"); siteID != "" {
		filter["site_id"] = siteID
	}
	if clientID := c.QueryParam("client_id"); clientID != "" {
		filter["client_id"] = clientID
	}

	ccs, total, err := dao.FindCheckConfigs(c.Request().Context(), filter, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list check configs", err)
	}

	return param.SuccessWithPaging(c, ccs, &pageParam, total)
}

func GetCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	cc, err := dao.FindCheckConfigByID(c.Request().Context(), id)
	if err != nil {
		return param.Error(c, http.StatusNotFound, "check config not found", err)
	}

	return param.Success(c, cc)
}

func CreateCheckConfig(c *echo.Context) error {
	var req CreateCheckConfigRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}
	if err := c.Validate(req); err != nil {
		return param.Error(c, http.StatusBadRequest, "validation failed", err)
	}

	siteID, err := bson.ObjectIDFromHex(req.SiteID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid site_id", err)
	}
	clientID, err := bson.ObjectIDFromHex(req.ClientID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid client_id", err)
	}

	// Validate site exists
	if _, err := dao.FindSiteByID(c.Request().Context(), siteID); err != nil {
		return param.Error(c, http.StatusBadRequest, "site not found", err)
	}

	// Validate client exists and has matching capability
	client, err := dao.FindClientByID(c.Request().Context(), clientID)
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "client not found", err)
	}

	requiredBit := checktype.Bit(myproto.CheckType(req.CheckType))
	if client.Capabilities&requiredBit == 0 {
		return param.Error(c, http.StatusBadRequest, "client does not support this check type", nil)
	}

	cc := &model.CheckConfig{
		ID:            bson.NewObjectID(),
		SiteID:        siteID,
		ClientID:      clientID,
		Status:        model.CHECK_ENABLED,
		CheckType:     req.CheckType,
		CheckInterval: req.CheckInterval,
		CheckExtra:    req.CheckExtra,
	}

	if err := dao.InsertCheckConfig(c.Request().Context(), cc); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to create check config", err)
	}

	return param.Success(c, cc)
}

func UpdateCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	var req UpdateCheckConfigRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	update := bson.M{}
	if req.ClientID != nil {
		update["client_id"] = *req.ClientID
	}
	if req.CheckType != nil {
		update["check_type"] = *req.CheckType
	}
	if req.CheckInterval != nil {
		update["check_interval"] = *req.CheckInterval
	}
	if req.CheckExtra != nil {
		update["check_extra"] = *req.CheckExtra
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}

	if len(update) == 0 {
		return param.Error(c, http.StatusBadRequest, "no fields to update", nil)
	}

	if err := dao.UpdateCheckConfig(c.Request().Context(), id, update); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to update check config", err)
	}

	return param.Success(c, nil)
}

func DeleteCheckConfig(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	if err := dao.DeleteCheckConfig(c.Request().Context(), id); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to delete check config", err)
	}

	return param.Success(c, nil)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/controller/check_configs.go
git commit -m "controller: add CheckConfig CRUD API handlers"
```

---

### Task 6: Controller — Clean up Site controller

**Files:**
- Modify: `internal/controller/sites.go`

- [ ] **Step 1: Remove check fields from CreateSiteRequest and UpdateSiteRequest**

Replace the entire content of `internal/controller/sites.go`:

```go
package controller

import (
	"net/http"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateSiteRequest struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type UpdateSiteRequest struct {
	Name        *string `json:"name"`
	URL         *string `json:"url"`
	Type        *string `json:"type"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}

func ListSites(c *echo.Context) error {
	var pageParam param.PageParam
	if err := c.Bind(&pageParam); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid params", err)
	}

	sites, total, err := dao.FindSites(c.Request().Context(), bson.M{}, pageParam.Page, pageParam.PageSize)
	if err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to list sites", err)
	}

	return param.SuccessWithPaging(c, sites, &pageParam, total)
}

func GetSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	site, err := dao.FindSiteByID(c.Request().Context(), id)
	if err != nil {
		return param.Error(c, http.StatusNotFound, "site not found", err)
	}

	return param.Success(c, site)
}

func CreateSite(c *echo.Context) error {
	var req CreateSiteRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}
	if err := c.Validate(req); err != nil {
		return param.Error(c, http.StatusBadRequest, "validation failed", err)
	}

	site := &model.Site{
		Name:        req.Name,
		URL:         req.URL,
		Type:        req.Type,
		Description: req.Description,
		Status:      model.SITE_STATUS_ENABLED,
		CreatedAt:   time.Now().Unix(),
	}

	if err := dao.InsertSite(c.Request().Context(), site); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to create site", err)
	}

	return param.Success(c, site)
}

func UpdateSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	var req UpdateSiteRequest
	if err := c.Bind(&req); err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid request body", err)
	}

	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.URL != nil {
		update["url"] = *req.URL
	}
	if req.Type != nil {
		update["type"] = *req.Type
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}

	if len(update) == 0 {
		return param.Error(c, http.StatusBadRequest, "no fields to update", nil)
	}

	if err := dao.UpdateSite(c.Request().Context(), id, update); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to update site", err)
	}

	return param.Success(c, nil)
}

func DeleteSite(c *echo.Context) error {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return param.Error(c, http.StatusBadRequest, "invalid id", err)
	}

	if err := dao.DeleteSite(c.Request().Context(), id); err != nil {
		return param.Error(c, http.StatusInternalServerError, "failed to delete site", err)
	}

	return param.Success(c, nil)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/controller/sites.go
git commit -m "controller: remove check fields from Site API"
```

---

### Task 7: Views — Register CheckConfig routes

**Files:**
- Modify: `internal/views/index.go`

- [ ] **Step 1: Add check-config routes**

Add the following block after the `// Clients` section in `internal/views/index.go`:

```go
	// Check Configs
	api.GET("/check-configs", controller.ListCheckConfigs)
	api.GET("/check-configs/:id", controller.GetCheckConfig)
	api.POST("/check-configs", controller.CreateCheckConfig)
	api.PUT("/check-configs/:id", controller.UpdateCheckConfig)
	api.DELETE("/check-configs/:id", controller.DeleteCheckConfig)
```

- [ ] **Step 2: Commit**

```bash
git add internal/views/index.go
git commit -m "views: register CheckConfig API routes"
```

---

### Task 8: Service — Rewrite Scheduler for CheckConfig

**Files:**
- Modify: `internal/service/scheduler.go`

- [ ] **Step 1: Rewrite scheduler.go**

Replace the entire content of `internal/service/scheduler.go`:

```go
package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/checktype"
	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Scheduler struct {
	ctx        context.Context
	dispatcher *TaskDispatcher
	stopCh     chan struct{}
	mu         sync.Mutex
	lastRun    map[string]time.Time
}

func NewScheduler(ctx context.Context, dispatcher *TaskDispatcher) *Scheduler {
	return &Scheduler{
		ctx:        ctx,
		dispatcher: dispatcher,
		stopCh:     make(chan struct{}),
		lastRun:    make(map[string]time.Time),
	}
}

func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.tick()
			case <-s.stopCh:
				return
			case <-s.ctx.Done():
				return
			}
		}
	}()
	slog.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	slog.Info("scheduler stopped")
}

// parseSchedule parses a check interval as either a Go duration ("5m", "30s")
// or a cron expression ("*/5 * * * *").
func parseSchedule(expr string) (cron.Schedule, error) {
	if d, err := time.ParseDuration(expr); err == nil {
		return cron.Every(d), nil
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return parser.Parse(expr)
}

func (s *Scheduler) tick() {
	sites, err := dao.FindAllEnabledSites(s.ctx)
	if err != nil {
		slog.Error("scheduler: failed to find sites", "error", err)
		return
	}

	clientIDs := s.dispatcher.GetOnlineClientIDsWithCapabilities()
	if len(clientIDs) == 0 {
		return
	}

	// Build a map of online client IDs for quick lookup
	onlineClients := make(map[string]int32, len(clientIDs))
	for _, c := range clientIDs {
		onlineClients[c.ID] = c.Capabilities
	}

	now := time.Now()

	for _, site := range sites {
		configs, err := dao.FindCheckConfigsBySiteID(s.ctx, site.ID)
		if err != nil {
			slog.Error("scheduler: failed to find check configs", "site", site.Name, "error", err)
			continue
		}

		for _, config := range configs {
			if config.Status != model.CHECK_ENABLED {
				continue
			}

			// Check if assigned client is online
			clientIDHex := config.ClientID.Hex()
			caps, clientOnline := onlineClients[clientIDHex]
			if !clientOnline {
				slog.Debug("scheduler: client offline, skipping check config",
					"site", site.Name, "client", clientIDHex)
				continue
			}

			// Verify client has the required capability
			requiredBit := checktype.Bit(myproto.CheckType(config.CheckType))
			if caps&requiredBit == 0 {
				slog.Warn("scheduler: client lacks required capability, skipping",
					"site", site.Name, "client", clientIDHex, "check_type", config.CheckType)
				continue
			}

			// Parse schedule and check if it's time to run
			schedule, err := parseSchedule(config.CheckInterval)
			if err != nil {
				slog.Warn("scheduler: invalid check_interval, skipping config",
					"config", config.ID.Hex(), "interval", config.CheckInterval, "error", err)
				continue
			}

			configIDHex := config.ID.Hex()
			s.mu.Lock()
			last, exists := s.lastRun[configIDHex]
			nextRun := schedule.Next(last)
			if exists && now.Before(nextRun) {
				s.mu.Unlock()
				continue
			}
			s.lastRun[configIDHex] = now
			s.mu.Unlock()

			task := &myproto.CheckTask{
				TaskId: bson.NewObjectID().Hex(),
				Check: &myproto.CheckRequest{
					Id:            site.ID.Hex(),
					Url:           site.URL,
					CheckType:     myproto.CheckType(config.CheckType),
					Method:        "GET",
					CheckConfigId: config.ID.Hex(),
				},
				AssignedAt: now.Unix(),
			}

			if s.dispatcher.Dispatch(clientIDHex, task) {
				slog.Debug("task dispatched", "site", site.Name, "client", clientIDHex, "config", configIDHex)
			} else {
				slog.Warn("task dispatch failed (channel full)", "site", site.Name, "client", clientIDHex)
			}
		}
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/service/...`
Expected: success (other packages may still have errors).

- [ ] **Step 3: Commit**

```bash
git add internal/service/scheduler.go
git commit -m "service: rewrite scheduler to use CheckConfig"
```

---

### Task 9: Service — Update CheckService for CheckConfigID

**Files:**
- Modify: `internal/service/check_service.go`
- Modify: `internal/dao/report.go`

- [ ] **Step 1: Update check_service.go**

Replace the entire content of `internal/service/check_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CheckServiceService struct {
	myproto.UnimplementedCheckServiceServer
	dispatcher *TaskDispatcher
}

func NewCheckServiceService(dispatcher *TaskDispatcher) *CheckServiceService {
	return &CheckServiceService{dispatcher: dispatcher}
}

func (s *CheckServiceService) WatchTasks(req *myproto.WatchTasksRequest, stream myproto.CheckService_WatchTasksServer) error {
	clientID := req.ClientId
	taskCh, exists := s.dispatcher.GetChannel(clientID)
	if !exists {
		return status.Error(codes.NotFound, "client not registered")
	}

	slog.Info("client watching tasks", "client_id", clientID)

	defer func() {
		s.dispatcher.UnregisterClient(clientID)
		slog.Info("client stopped watching tasks", "client_id", clientID)
	}()

	for {
		select {
		case task, ok := <-taskCh:
			if !ok {
				return nil
			}
			if err := stream.Send(task); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *CheckServiceService) ReportResult(ctx context.Context, req *myproto.CheckResultRequest) (*myproto.CheckResultResponse, error) {
	result := req.Result
	if result == nil {
		return nil, status.Error(codes.InvalidArgument, "missing result")
	}

	var responseTimeMs int32
	if result.ResponseTimeMs != nil {
		responseTimeMs = result.GetResponseTimeMs()
	}

	siteID, err := bson.ObjectIDFromHex(result.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid site id")
	}
	clientID, err := bson.ObjectIDFromHex(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client id")
	}

	checkConfigID, _ := bson.ObjectIDFromHex(result.CheckConfigId)

	check := &model.Check{
		SiteID:        siteID,
		ClientID:      clientID,
		CheckConfigID: checkConfigID,
		Type:          result.CheckType,
		Status:        result.ErrorType,
		Result:        buildResultString(result),
		Delay:         int64(responseTimeMs),
	}

	if err := dao.InsertCheck(ctx, check); err != nil {
		slog.Error("failed to insert check", "error", err)
		return nil, status.Error(codes.Internal, "failed to save check result")
	}

	// Update report aggregation
	if err := updateReports(ctx, siteID, checkConfigID, result.Success, float64(responseTimeMs)); err != nil {
		slog.Error("failed to update report", "error", err)
	}

	return &myproto.CheckResultResponse{Success: true}, nil
}

func buildResultString(resp *myproto.CheckResponse) string {
	if resp.Success {
		return "success"
	}
	return fmt.Sprintf("error: %s", resp.ErrorType.String())
}

func updateReports(ctx context.Context, siteID, checkConfigID bson.ObjectID, success bool, delay float64) error {
	now := time.Now()
	successCount := int64(0)
	if success {
		successCount = 1
	}

	reportTypes := []struct {
		reportType int
		timeframe  string
	}{
		{model.REPORT_TYPE_HOURLY, now.Format("2006-01-02 15:00:00")},
		{model.REPORT_TYPE_DAILY, now.Format("2006-01-02")},
		{model.REPORT_TYPE_MONTHLY, now.Format("2006-01")},
	}

	for _, rt := range reportTypes {
		report := &model.Report{
			SiteID:        siteID,
			CheckConfigID: checkConfigID,
			Timeframe:     rt.timeframe,
			Type:          rt.reportType,
			Successes:     successCount,
			Uptime:        0,
			AvgDelay:      delay,
		}

		if err := dao.UpsertReport(ctx, report); err != nil {
			return err
		}

		RecalculateReportUptime(ctx, siteID, checkConfigID, rt.timeframe, rt.reportType)
	}
	return nil
}

func RecalculateReportUptime(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int) {
	rpt, err := dao.FindReport(ctx, siteID, checkConfigID, timeframe, reportType)
	if err != nil || rpt == nil || rpt.Checks == 0 {
		return
	}
	uptime := (float64(rpt.Successes) / float64(rpt.Checks)) * 100
	uptime = math.Round(uptime*100) / 100
	dao.SetReportUptime(ctx, siteID, checkConfigID, timeframe, reportType, uptime)
}
```

- [ ] **Step 2: Update dao/report.go to include CheckConfigID in filter**

Replace the entire content of `internal/dao/report.go`:

```go
package dao

import (
	"context"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const reportCollection = "reports"

func UpsertReport(ctx context.Context, report *model.Report) error {
	filter := bson.M{
		"site_id":         report.SiteID,
		"check_config_id": report.CheckConfigID,
		"timeframe":       report.Timeframe,
		"type":            report.Type,
	}

	update := bson.M{
		"$inc": bson.M{
			"checks":    1,
			"successes": report.Successes,
		},
		"$set": bson.M{
			"uptime":   report.Uptime,
			"avg_delay": report.AvgDelay,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := db.MongoDB.Collection(reportCollection).UpdateOne(ctx, filter, update, opts)
	return err
}

func FindReport(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int) (*model.Report, error) {
	filter := bson.M{
		"site_id":         siteID,
		"check_config_id": checkConfigID,
		"timeframe":       timeframe,
		"type":            reportType,
	}
	var report model.Report
	err := db.MongoDB.Collection(reportCollection).FindOne(ctx, filter).Decode(&report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func SetReportUptime(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int, uptime float64) error {
	filter := bson.M{
		"site_id":         siteID,
		"check_config_id": checkConfigID,
		"timeframe":       timeframe,
		"type":            reportType,
	}
	_, err := db.MongoDB.Collection(reportCollection).UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{"uptime": uptime},
	})
	return err
}

func FindReportsBySiteID(ctx context.Context, siteID string, reportType *int, page, pageSize int64) ([]model.Report, error) {
	filter := bson.M{"site_id": siteID}
	if reportType != nil {
		filter["type"] = *reportType
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"timeframe": -1})

	cursor, err := db.MongoDB.Collection(reportCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []model.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/service/... ./internal/dao/...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add internal/service/check_service.go internal/dao/report.go
git commit -m "service: handle CheckConfigID in ReportResult and report aggregation"
```

---

### Task 10: Client — Pass check_config_id through check flow

**Files:**
- Modify: `cmd/client/main.go`

- [ ] **Step 1: Update performCheck to pass check_config_id**

In `cmd/client/main.go`, update `performCheck` to set `CheckConfigId` on the response. Change the response construction after the switch block:

Find this code:

```go
	resp.CheckType = check.CheckType
	elapsed := int32(time.Since(start).Milliseconds())
```

Replace with:

```go
	resp.CheckType = check.CheckType
	resp.CheckConfigId = check.CheckConfigId
	elapsed := int32(time.Since(start).Milliseconds())
```

Also update the default case in the switch to include CheckConfigId:

Find:

```go
	default:
		resp = &myproto.CheckResponse{
			Id:        check.Id,
			Success:   false,
			ErrorType: myproto.ErrorType_ERROR_TYPE_NO_ERROR,
			CheckType: check.CheckType,
		}
```

Replace with:

```go
	default:
		resp = &myproto.CheckResponse{
			Id:            check.Id,
			Success:       false,
			ErrorType:     myproto.ErrorType_ERROR_TYPE_NO_ERROR,
			CheckType:     check.CheckType,
			CheckConfigId: check.CheckConfigId,
		}
```

- [ ] **Step 2: Commit**

```bash
git add cmd/client/main.go
git commit -m "client: pass check_config_id through check flow"
```

---

### Task 11: Build — Verify full project compiles

**Files:** none (verification only)

- [ ] **Step 1: Run go build**

Run: `go build ./...`
Expected: success with no errors.

If there are compilation errors, fix them before proceeding.

- [ ] **Step 2: Commit any fixes**

```bash
git add -A
git commit -m "fix: resolve remaining compilation issues"
```
