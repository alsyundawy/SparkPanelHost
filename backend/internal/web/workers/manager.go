package workers

import (
	"context"
	"sync"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

type WorkerManager struct {
	store  *store.Store
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewWorkerManager(s *store.Store) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerManager{
		store:  s,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start launches all background workers in separate goroutines.
func (m *WorkerManager) Start() {
	// 1. DNS Cluster sync worker (runs every 60s)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		RunDNSClusterWorker(m.ctx, m.store, 60*time.Second)
	}()

	// 2. RBL Blacklist monitoring worker (runs every 6 hours)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		RunRBLWorker(m.ctx, m.store, "127.0.0.1", 6*time.Hour)
	}()

	// 3. Backup automated test-restore worker (runs daily)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		RunBackupVerifyWorker(m.ctx, m.store, 24*time.Hour)
	}()
}

// Stop gracefully shuts down all running workers.
func (m *WorkerManager) Stop() {
	m.cancel()
	m.wg.Wait()
}
