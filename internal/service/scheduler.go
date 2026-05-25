package service

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Scheduler struct {
	dispatcher *TaskDispatcher
	stopCh     chan struct{}
	mu         sync.Mutex
	lastRun    map[string]time.Time
}

func NewScheduler(dispatcher *TaskDispatcher) *Scheduler {
	return &Scheduler{
		dispatcher: dispatcher,
		stopCh:     make(chan struct{}),
		lastRun:    make(map[string]time.Time),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.tick(ctx)
			case <-s.stopCh:
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

func (s *Scheduler) tick(ctx context.Context) {
	sites, err := dao.FindAllEnabledSites(ctx)
	if err != nil {
		slog.Error("scheduler: failed to find sites", "error", err)
		return
	}

	clientIDs := s.dispatcher.GetOnlineClientIDs()
	if len(clientIDs) == 0 {
		return
	}

	now := time.Now()

	for _, site := range sites {
		interval, err := time.ParseDuration(site.CheckInterval)
		if err != nil {
			// Try parsing as minutes (e.g. "5" means 5 minutes)
			continue
		}

		s.mu.Lock()
		last, exists := s.lastRun[site.ID.Hex()]
		if exists && now.Sub(last) < interval {
			s.mu.Unlock()
			continue
		}
		s.lastRun[site.ID.Hex()] = now
		s.mu.Unlock()

		// Pick a random online client
		targetClient := clientIDs[rand.Intn(len(clientIDs))]

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

		if s.dispatcher.Dispatch(targetClient, task) {
			slog.Debug("task dispatched", "site", site.Name, "client", targetClient)
		} else {
			slog.Warn("task dispatch failed (channel full)", "site", site.Name, "client", targetClient)
		}
	}
}
