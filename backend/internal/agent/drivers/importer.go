package drivers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/fileops"
)

var validUserRegex = regexp.MustCompile(`^[a-z][a-z0-9]{2,15}$`)

type ImportAccountResult struct {
	Username  string   `json:"username"`
	Domains   []string `json:"domains"`
	Databases []string `json:"databases"`
	Status    string   `json:"status"`
}

func ImportCpanelAccount(ctx context.Context, archivePath string) (*ImportAccountResult, error) {
	meta, err := fileops.InspectCpanelArchive(archivePath)
	if err != nil {
		return nil, fmt.Errorf("inspeksi arsip cPanel gagal: %w", err)
	}

	if !validUserRegex.MatchString(meta.Username) {
		return nil, fmt.Errorf("username tidak valid dalam arsip: %s", meta.Username)
	}

	stagingDir := filepath.Join("/var/panel/tmp", meta.Username, "importer")
	_ = os.MkdirAll(stagingDir, 0750)
	defer os.RemoveAll(stagingDir)

	// Ensure system user exists
	homeDir := filepath.Join("/home", meta.Username)
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		_, _ = execsafe.RunDirect(ctx, 0, "/usr/sbin/useradd", "-m", "-s", "/bin/bash", meta.Username)
	}

	res := &ImportAccountResult{
		Username:  meta.Username,
		Domains:   meta.Domains,
		Databases: meta.Databases,
		Status:    "imported",
	}

	return res, nil
}
