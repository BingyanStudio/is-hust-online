package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/checktype"
	"github.com/BingyanStudio/is-hust-online/internal/config"
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
	// 主调度器: 每 10 秒执行一次
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

	// 清理协程: 每 5 分钟清理一次过期 check 记录
	go func() {
		cleanupTicker := time.NewTicker(5 * time.Minute)
		defer cleanupTicker.Stop()

		for {
			select {
			case <-cleanupTicker.C:
				s.cleanupOldChecks()
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

func (s *Scheduler) cleanupOldChecks() {
	retentionHours := config.C.CheckRetentionHours
	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour).Unix()

	deleted, err := dao.DeleteChecksBefore(s.ctx, cutoff)
	if err != nil {
		slog.Error("scheduler: failed to cleanup old checks", "error", err)
		return
	}

	if deleted > 0 {
		slog.Info("scheduler: cleaned up old checks", "deleted", deleted, "retention_hours", retentionHours)
	}
}

func parseSchedule(expr string) (cron.Schedule, error) {
	if d, err := time.ParseDuration(expr); err == nil {
		return cron.Every(d), nil
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return parser.Parse(expr)
}

func (s *Scheduler) checkHeartbeats() {
	const heartbeatTimeout = 60 * time.Second
	staleBefore := time.Now().Add(-heartbeatTimeout).Unix()

	staleIDs := s.dispatcher.GetStaleClientIDs(staleBefore)
	if len(staleIDs) == 0 {
		return
	}

	ids := make([]bson.ObjectID, 0, len(staleIDs))
	for _, idHex := range staleIDs {
		s.dispatcher.UnregisterClient(idHex)
		id, _ := bson.ObjectIDFromHex(idHex)
		ids = append(ids, id)
		slog.Warn("client marked offline (heartbeat timeout)", "client_id", idHex)
	}

	if err := dao.BatchUpdateClientStatus(s.ctx, ids, model.CLIENT_STATUS_OFFLINE); err != nil {
		slog.Error("scheduler: failed to batch update client status", "error", err)
	}
}

func (s *Scheduler) tick() {
	s.checkHeartbeats()

	sites, err := dao.FindAllEnabledSites(s.ctx)
	if err != nil {
		slog.Error("scheduler: failed to find sites", "error", err)
		return
	}

	clientIDs := s.dispatcher.GetOnlineClientIDsWithCapabilities()
	if len(clientIDs) == 0 {
		return
	}

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

			clientIDHex := config.ClientID.Hex()
			caps, clientOnline := onlineClients[clientIDHex]
			if !clientOnline {
				slog.Debug("scheduler: client offline, skipping check config",
					"site", site.Name, "client", clientIDHex)
				continue
			}

			requiredBit := checktype.Bit(myproto.CheckType(config.CheckType))
			if caps&requiredBit == 0 {
				slog.Warn("scheduler: client lacks required capability, skipping",
					"site", site.Name, "client", clientIDHex, "check_type", config.CheckType)
				continue
			}

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

			var extra []byte

			extra, err = json.Marshal(config.CheckExtra)
			if err != nil {
				slog.Warn("scheduler: failed to parse check_extra, using empty array",
					"config", config.ID.Hex(), "error", err)
				extra = []byte{}
			}

			task := &myproto.CheckTask{
				TaskId: bson.NewObjectID().Hex(),
				Check: &myproto.CheckRequest{
					Id:            site.ID.Hex(),
					Url:           site.URL,
					CheckType:     myproto.CheckType(config.CheckType),
					Method:        "GET",
					CheckConfigId: config.ID.Hex(),
					Extra:         extra,
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
