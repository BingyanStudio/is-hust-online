package service

import (
	"sync"

	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
)

const taskChannelBuffer = 16

type TaskDispatcher struct {
	mu       sync.RWMutex
	channels map[string]chan *myproto.CheckTask
}

func NewTaskDispatcher() *TaskDispatcher {
	return &TaskDispatcher{
		channels: make(map[string]chan *myproto.CheckTask),
	}
}

func (d *TaskDispatcher) RegisterClient(clientID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.channels[clientID]; !exists {
		d.channels[clientID] = make(chan *myproto.CheckTask, taskChannelBuffer)
	}
}

func (d *TaskDispatcher) UnregisterClient(clientID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if ch, exists := d.channels[clientID]; exists {
		close(ch)
		delete(d.channels, clientID)
	}
}

func (d *TaskDispatcher) Dispatch(clientID string, task *myproto.CheckTask) bool {
	d.mu.RLock()
	ch, exists := d.channels[clientID]
	d.mu.RUnlock()
	if !exists {
		return false
	}
	select {
	case ch <- task:
		return true
	default:
		return false
	}
}

func (d *TaskDispatcher) GetChannel(clientID string) (<-chan *myproto.CheckTask, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ch, exists := d.channels[clientID]
	return ch, exists
}

func (d *TaskDispatcher) GetOnlineClientIDs() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ids := make([]string, 0, len(d.channels))
	for id := range d.channels {
		ids = append(ids, id)
	}
	return ids
}
