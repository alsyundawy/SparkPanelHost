package fileops

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImportedAccountMetadata struct {
	Username  string   `json:"username"`
	Domains   []string `json:"domains"`
	Databases []string `json:"databases"`
	HasFiles  bool     `json:"has_files"`
	HasMail   bool     `json:"has_mail"`
}

// InspectCpanelArchive reads the archive headers to discover account username, domains, and database dumps.
func InspectCpanelArchive(archivePath string) (*ImportedAccountMetadata, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("invalid gzip header: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	meta := &ImportedAccountMetadata{
		Domains:   []string{},
		Databases: []string{},
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		cleanName := filepath.Clean(hdr.Name)
		parts := strings.Split(cleanName, string(filepath.Separator))
		if len(parts) == 0 {
			continue
		}

		// Detect username from root directory e.g., cpmove-alice/ or alice/
		if meta.Username == "" && len(parts) > 1 {
			root := parts[0]
			if strings.HasPrefix(root, "cpmove-") {
				meta.Username = strings.TrimPrefix(root, "cpmove-")
			}
		}

		// Look for userdata domain files
		if strings.Contains(cleanName, "userdata/") && !hdr.FileInfo().IsDir() {
			base := filepath.Base(cleanName)
			if base != "main" && !strings.HasPrefix(base, ".") {
				meta.Domains = append(meta.Domains, base)
			}
		}

		// Look for mysql dumps
		if strings.Contains(cleanName, "mysql/") && strings.HasSuffix(cleanName, ".sql") {
			base := strings.TrimSuffix(filepath.Base(cleanName), ".sql")
			meta.Databases = append(meta.Databases, base)
		}

		if strings.Contains(cleanName, "homedir/public_html") {
			meta.HasFiles = true
		}

		if strings.Contains(cleanName, "email/") || strings.Contains(cleanName, "mail/") {
			meta.HasMail = true
		}
	}

	if meta.Username == "" {
		// Fallback to filename prefix e.g. cpmove-alice.tar.gz
		base := filepath.Base(archivePath)
		if strings.HasPrefix(base, "cpmove-") {
			meta.Username = strings.TrimSuffix(strings.TrimPrefix(base, "cpmove-"), ".tar.gz")
		} else {
			return nil, errors.New("cannot determine username from cPanel archive")
		}
	}

	return meta, nil
}
