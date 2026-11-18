//go:build !linux

package fileops

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxSingleFileBytes  int64 = 512 * 1024 * 1024
	maxTotalExtractSize int64 = 4096 * 1024 * 1024
)

func VerifySHA256(filePath, expectedHex string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expectedHex) {
		return fmt.Errorf("SHA-256 mismatch: expected %s, got %s", expectedHex, actual)
	}
	return nil
}

// ExtractArchiveAsUser on non-Linux fallback checks clean paths and extracts safely.
func ExtractArchiveAsUser(archivePath, destDir string, uid, gid int, stripComponents int) error {
	cleanDest := filepath.Clean(destDir)
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return err
		}
		defer zr.Close()

		var totalWritten int64
		for _, zf := range zr.File {
			if zf.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink in ZIP rejected: %s", zf.Name)
			}
			parts := strings.Split(filepath.ToSlash(zf.Name), "/")
			if len(parts) <= stripComponents {
				continue
			}
			relPath := strings.Join(parts[stripComponents:], "/")
			if relPath == "" || relPath == "." || strings.HasPrefix(relPath, "..") {
				continue
			}
			targetPath := filepath.Join(cleanDest, relPath)
			if !strings.HasPrefix(targetPath, cleanDest) {
				return fmt.Errorf("path traversal rejected: %s", zf.Name)
			}

			if zf.FileInfo().IsDir() {
				_ = os.MkdirAll(targetPath, 0750)
				continue
			}

			_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
			if err != nil {
				_ = rc.Close()
				return err
			}
			n, err := io.Copy(outFile, io.LimitReader(rc, maxSingleFileBytes))
			_ = outFile.Close()
			_ = rc.Close()
			if err != nil {
				return err
			}
			totalWritten += n
			if totalWritten > maxTotalExtractSize {
				return errors.New("zip-bomb detected")
			}
		}
		return nil
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var totalWritten int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(hdr.Name), "/")
		if len(parts) <= stripComponents {
			continue
		}
		relPath := strings.Join(parts[stripComponents:], "/")
		if relPath == "" || relPath == "." || strings.HasPrefix(relPath, "..") {
			continue
		}
		targetPath := filepath.Join(cleanDest, relPath)
		if !strings.HasPrefix(targetPath, cleanDest) {
			return fmt.Errorf("path traversal rejected: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(targetPath, 0750)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(targetPath), 0750)
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
			if err != nil {
				return err
			}
			n, err := io.Copy(outFile, io.LimitReader(tr, maxSingleFileBytes))
			_ = outFile.Close()
			if err != nil {
				return err
			}
			totalWritten += n
			if totalWritten > maxTotalExtractSize {
				return errors.New("archive-bomb detected")
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("symlink in tar rejected: %s", hdr.Name)
		}
	}
	return nil
}
