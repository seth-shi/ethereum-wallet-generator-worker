package models

import (
	"sync"
	"time"
)

type WorkerStatusManager struct {
	keys   []string
	values map[string]*WorkStatusRequest
	locker sync.RWMutex
}

func NewNodeStatusManager(workers []*WorkStatusRequest) *WorkerStatusManager {

	manager := &WorkerStatusManager{
		keys:   make([]string, 0, 1024),
		values: make(map[string]*WorkStatusRequest, 1024),
	}

	for _, w := range workers {
		manager.Add(w)
	}

	return manager
}

func (n *WorkerStatusManager) All() []*WorkStatusRequest {

	n.locker.Lock()
	defer n.locker.Unlock()

	data := make([]*WorkStatusRequest, len(n.keys))
	for i, k := range n.keys {

		node, exists := n.values[k]
		if !exists {
			node = &WorkStatusRequest{}
		}
		data[i] = node
	}

	return data
}

func (n *WorkerStatusManager) Get(key string) (*WorkStatusRequest, bool) {

	n.locker.Lock()
	defer n.locker.Unlock()

	val, exists := n.values[key]

	return val, exists
}

func (n *WorkerStatusManager) Add(newStatus *WorkStatusRequest) {

	n.locker.Lock()
	defer n.locker.Unlock()

	key := newStatus.Name

	if oldStatus, exists := n.values[key]; exists {
		newStatus.Found += oldStatus.Found
		newStatus.Count += oldStatus.Count
	} else {
		n.keys = append(n.keys, key)
	}

	newStatus.LastActiveAt = time.Now()
	n.values[key] = newStatus
}
