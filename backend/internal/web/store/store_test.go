package store_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
)

func TestStoreAndAuditLogChain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aether-store-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	migrationPath := filepath.Join("..", "..", "..", "migrations", "001_init.sql")

	st, err := store.NewStore(dbPath, migrationPath)
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()

	// Append 3 audit logs and verify hash chaining
	if err := st.AppendAuditLog(ctx, "usr-1", nil, "domain.create", "domain:example.com", "127.0.0.1"); err != nil {
		t.Fatalf("Failed to append audit log 1: %v", err)
	}
	if err := st.AppendAuditLog(ctx, "usr-1", nil, "domain.ssl", "domain:example.com", "127.0.0.1"); err != nil {
		t.Fatalf("Failed to append audit log 2: %v", err)
	}
	if err := st.AppendAuditLog(ctx, "admin", nil, "user.suspend", "user:usr-1", "127.0.0.1"); err != nil {
		t.Fatalf("Failed to append audit log 3: %v", err)
	}

	logs, err := st.ListAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list audit logs: %v", err)
	}

	if len(logs) != 3 {
		t.Fatalf("Expected 3 audit logs, got %d", len(logs))
	}

	// Logs are ordered DESC: log[0] is #3, log[1] is #2, log[2] is #1
	// log[0].PrevHash must equal log[1].EntryHash
	if logs[0].PrevHash != logs[1].EntryHash {
		t.Fatalf("Audit log hash chain broken between log 3 and log 2")
	}
	// log[1].PrevHash must equal log[2].EntryHash
	if logs[1].PrevHash != logs[2].EntryHash {
		t.Fatalf("Audit log hash chain broken between log 2 and log 1")
	}
	// log[2] (genesis) must have 64 zeros as prev_hash
	if logs[2].PrevHash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("Genesis audit log prev_hash must be 64 zeros, got %s", logs[2].PrevHash)
	}
}
