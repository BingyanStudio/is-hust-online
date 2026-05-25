package service

import (
	"context"
	"log/slog"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
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

// checkTypeBit returns the bitmask value for a given CheckType.
func checkTypeBit(ct myproto.CheckType) int32 {
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

	now := time.Now()

	for _, site := range sites {
		schedule, err := parseSchedule(site.CheckInterval)
		if err != nil {
			slog.Warn("scheduler: invalid check_interval, skipping site",
				"site", site.Name, "interval", site.CheckInterval, "error", err)
			continue
		}

		// Validate URL
		u, err := url.Parse(site.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			slog.Warn("scheduler: invalid or unsupported URL, skipping site",
				"site", site.Name, "url", site.URL)
			continue
		}

		s.mu.Lock()
		last, exists := s.lastRun[site.ID.Hex()]
		nextRun := schedule.Next(last)
		if exists && now.Before(nextRun) {
			s.mu.Unlock()
			continue
		}
		s.lastRun[site.ID.Hex()] = now
		s.mu.Unlock()

		// Filter clients that support this site's check type
		requiredBit := checkTypeBit(myproto.CheckType(site.CheckType))
		var eligible []OnlineClient
		for _, c := range clientIDs {
			if c.Capabilities&requiredBit != 0 {
				eligible = append(eligible, c)
			}
		}
		if len(eligible) == 0 {
			slog.Warn("scheduler: no client supports check type, skipping site",
				"site", site.Name, "check_type", site.CheckType)
			continue
		}

		// Pick a random eligible client
		target := eligible[rand.Intn(len(eligible))]

		task := &myproto.CheckTask{
			TaskId: bson.NewObjectID().Hex(),
			Check: &myproto.CheckRequest{
				Id:        site.ID.Hex(),
				Url:       site.URL,
				CheckType: myproto.CheckType(site.CheckType),
				Method:    "GET",
			},
			AssignedAt: now.Unix(),
		}

		if s.dispatcher.Dispatch(target.ID, task) {
			slog.Debug("task dispatched", "site", site.Name, "client", target.ID)
		} else {
			slog.Warn("task dispatch failed (channel full)", "site", site.Name, "client", target.ID)
		}
	}
}
