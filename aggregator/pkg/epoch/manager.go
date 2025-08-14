package epoch

import (
	"sync"
	"time"
)

type Info struct {
	Epoch        uint64    `json:"epoch"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Root         string    `json:"root"`
	ReceiptCount int       `json:"receipt_count"`
	Finalized    bool      `json:"finalized"`
}

type Manager struct {
	epochs map[uint64]*Info
	mu     sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		epochs: make(map[uint64]*Info),
	}
}

func (m *Manager) CurrentEpoch() uint64 {
	now := time.Now().UTC()
	return uint64(now.Unix() / 86400)
}

func (m *Manager) GetOrCreateEpoch(epoch uint64) *Info {
	m.mu.RLock()
	info, exists := m.epochs[epoch]
	m.mu.RUnlock()

	if exists {
		return info
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if info, exists := m.epochs[epoch]; exists {
		return info
	}

	startTime := time.Unix(int64(epoch*86400), 0).UTC()
	endTime := startTime.Add(24 * time.Hour)

	info = &Info{
		Epoch:     epoch,
		StartTime: startTime,
		EndTime:   endTime,
		Finalized: false,
	}

	m.epochs[epoch] = info
	return info
}

func (m *Manager) FinalizeEpoch(epoch uint64, root string, receiptCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.epochs[epoch]
	if !exists {
		info = &Info{
			Epoch:     epoch,
			StartTime: time.Unix(int64(epoch*86400), 0).UTC(),
			EndTime:   time.Unix(int64((epoch+1)*86400), 0).UTC(),
		}
		m.epochs[epoch] = info
	}

	info.Root = root
	info.ReceiptCount = receiptCount
	info.Finalized = true

	return nil
}

func (m *Manager) GetEpochInfo(epoch uint64) (*Info, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.epochs[epoch]
	return info, exists
}

func (m *Manager) GetEpochCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.epochs)
}
