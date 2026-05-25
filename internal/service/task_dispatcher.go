package service

import (
	"sync"

	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
)

const taskChannelBuffer = 16

type clientEntry struct {
	ch           chan *myproto.CheckTask
	capabilities int32
}

type TaskDispatcher struct {
	mu       sync.RWMutex
	clients  map[string]*clientEntry
}

func NewTaskDispatcher() *TaskDispatcher {
	return &TaskDispatcher{
		clients: make(map[string]*clientEntry),
	}
}

func (d *TaskDispatcher) RegisterClient(clientID string, capabilities int32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.clients[clientID]; !exists {
		d.clients[clientID] = &clientEntry{
			ch:           make(chan *myproto.CheckTask, taskChannelBuffer),
			capabilities: capabilities,
		}
	}
}

func (d *TaskDispatcher) UnregisterClient(clientID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entry, exists := d.clients[clientID]; exists {
		close(entry.ch)
		delete(d.clients, clientID)
	}
}

func (d *TaskDispatcher) Dispatch(clientID string, task *myproto.CheckTask) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	entry, exists := d.clients[clientID]
	if !exists {
		return false
	}
	select {
	case entry.ch <- task:
		return true
	default:
		return false
	}
}

func (d *TaskDispatcher) GetChannel(clientID string) (<-chan *myproto.CheckTask, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	entry, exists := d.clients[clientID]
	if !exists {
		return nil, false
	}
	return entry.ch, true
}

func (d *TaskDispatcher) GetOnlineClientIDs() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ids := make([]string, 0, len(d.clients))
	for id := range d.clients {
		ids = append(ids, id)
	}
	return ids
}

type OnlineClient struct {
	ID           string
	Capabilities int32
}

func (d *TaskDispatcher) GetOnlineClientIDsWithCapabilities() []OnlineClient {
	d.mu.RLock()
	defer d.mu.RUnlock()
	clients := make([]OnlineClient, 0, len(d.clients))
	for id, entry := range d.clients {
		clients = append(clients, OnlineClient{
			ID:           id,
			Capabilities: entry.capabilities,
		})
	}
	return clients
}
