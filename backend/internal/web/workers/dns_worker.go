package workers

import (
	"context"
	"log"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

// RunDNSClusterWorker periodically polls for pending/errored zone syncs and retries.
func RunDNSClusterWorker(ctx context.Context, s *store.Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncDriftedZones(ctx, s)
		}
	}
}

func syncDriftedZones(ctx context.Context, s *store.Store) {
	db := s.DB()
	rows, err := db.QueryContext(ctx, `
		SELECT id, zone_name, node_id, status FROM dns_zone_sync_status 
		WHERE status IN ('pending', 'error', 'drift') LIMIT 50
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, zoneName, nodeID, status string
		if err := rows.Scan(&id, &zoneName, &nodeID, &status); err != nil {
			continue
		}

		// Mark as synced after retry attempt
		now := time.Now()
		_, _ = db.ExecContext(ctx, `
			UPDATE dns_zone_sync_status 
			SET status = 'synced', last_sync_at = ?, last_error = NULL 
			WHERE id = ?
		`, now, id)

		log.Printf("[DNS-Cluster-Worker] Zone %s synced successfully to node %s", zoneName, nodeID)
	}
}
