package fileops

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectCpanelArchive(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "cpmove-charlie.tar.gz")

	// Create test cpmove archive
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	files := []struct {
		name string
		body string
	}{
		{"cpmove-charlie/cp/charlie", "charlie"},
		{"cpmove-charlie/userdata/charlie.com", "documentroot: /home/charlie/public_html"},
		{"cpmove-charlie/mysql/charlie_app.sql", "CREATE TABLE test (id INT);"},
		{"cpmove-charlie/homedir/public_html/index.php", "<?php echo 'hello'; ?>"},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: 0644,
			Size: int64(len(file.body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		if _, err := tw.Write([]byte(file.body)); err != nil {
			t.Fatalf("failed to write body: %v", err)
		}
	}

	tw.Close()
	gw.Close()
	f.Close()

	// Inspect archive
	meta, err := InspectCpanelArchive(archivePath)
	if err != nil {
		t.Fatalf("InspectCpanelArchive failed: %v", err)
	}

	if meta.Username != "charlie" {
		t.Fatalf("expected username charlie, got %s", meta.Username)
	}

	if len(meta.Domains) != 1 || meta.Domains[0] != "charlie.com" {
		t.Fatalf("expected domain charlie.com, got %v", meta.Domains)
	}

	if len(meta.Databases) != 1 || meta.Databases[0] != "charlie_app" {
		t.Fatalf("expected database charlie_app, got %v", meta.Databases)
	}

	if !meta.HasFiles {
		t.Fatal("expected has_files to be true")
	}
}
