//go:build linux

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
	"path"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
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

func openFileBeneath(baseDirFD int, relPath string, mode uint32) (*os.File, error) {
	cleanRel := strings.TrimPrefix(path.Clean("/"+relPath), "/")
	if cleanRel == "" || cleanRel == "." || strings.HasPrefix(cleanRel, "..") {
		return nil, fmt.Errorf("path relatif tidak valid: %q", relPath)
	}

	parts := strings.Split(cleanRel, "/")
	currFD := baseDirFD
	var openedFDs []int
	defer func() {
		for _, fd := range openedFDs {
			_ = unix.Close(fd)
		}
	}()

	for _, dirPart := range parts[:len(parts)-1] {
		if dirPart == "" || dirPart == "." || dirPart == ".." {
			return nil, fmt.Errorf("komponen direktori ilegal: %q", dirPart)
		}
		_ = unix.Mkdirat(currFD, dirPart, 0750)
		nextFD, err := unix.Openat2(currFD, dirPart, &unix.OpenHow{
			Flags:   unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC,
			Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS,
		})
		if err != nil {
			return nil, fmt.Errorf("gagal membuka direktori aman %q: %w", dirPart, err)
		}
		openedFDs = append(openedFDs, nextFD)
		currFD = nextFD
	}

	fileName := parts[len(parts)-1]
	fileFD, err := unix.Openat2(currFD, fileName, &unix.OpenHow{
		Flags:   unix.O_WRONLY | unix.O_CREAT | unix.O_TRUNC | unix.O_CLOEXEC,
		Mode:    uint64(mode),
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat file aman %q: %w", relPath, err)
	}
	return os.NewFile(uintptr(fileFD), relPath), nil
}

func makeDirBeneath(baseDirFD int, relPath string) error {
	cleanRel := strings.TrimPrefix(path.Clean("/"+relPath), "/")
	if cleanRel == "" || cleanRel == "." {
		return nil
	}
	currFD := baseDirFD
	var openedFDs []int
	defer func() {
		for _, fd := range openedFDs {
			_ = unix.Close(fd)
		}
	}()

	for _, dirPart := range strings.Split(cleanRel, "/") {
		if dirPart == "" || dirPart == "." || dirPart == ".." {
			return fmt.Errorf("komponen direktori ilegal: %q", dirPart)
		}
		_ = unix.Mkdirat(currFD, dirPart, 0750)
		nextFD, err := unix.Openat2(currFD, dirPart, &unix.OpenHow{
			Flags:   unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC,
			Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS,
		})
		if err != nil {
			return err
		}
		openedFDs = append(openedFDs, nextFD)
		currFD = nextFD
	}
	return nil
}

func ExtractArchiveAsUser(archivePath, destDir string, uid, gid int, stripComponents int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origEuid := syscall.Geteuid()
	origEgid := syscall.Getegid()

	if err := syscall.Setegid(gid); err != nil {
		return fmt.Errorf("setegid gagal: %w", err)
	}
	if err := syscall.Seteuid(uid); err != nil {
		_ = syscall.Setegid(origEgid)
		return fmt.Errorf("seteuid gagal: %w", err)
	}
	defer func() {
		_ = syscall.Seteuid(origEuid)
		_ = syscall.Setegid(origEgid)
	}()

	baseDirFD, err := unix.Open(destDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("destDir tidak valid atau berupa symlink: %w", err)
	}
	defer unix.Close(baseDirFD)

	var totalWritten int64
	stripPath := func(rawName string) (string, bool) {
		parts := strings.Split(strings.ReplaceAll(rawName, "\\", "/"), "/")
		if len(parts) <= stripComponents {
			return "", false
		}
		joined := strings.Join(parts[stripComponents:], "/")
		if joined == "" || joined == "." {
			return "", false
		}
		return joined, true
	}

	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		zr, err := zip.OpenReader(archivePath)
		if err != nil {
			return err
		}
		defer zr.Close()

		for _, zf := range zr.File {
			if zf.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink di dalam arsip ZIP ditolak: %s", zf.Name)
			}
			relPath, ok := stripPath(zf.Name)
			if !ok {
				continue
			}
			if zf.FileInfo().IsDir() {
				if err := makeDirBeneath(baseDirFD, relPath); err != nil {
					return err
				}
				continue
			}
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			outFile, err := openFileBeneath(baseDirFD, relPath, 0640)
			if err != nil {
				_ = rc.Close()
				return err
			}
			n, copyErr := io.Copy(outFile, io.LimitReader(rc, maxSingleFileBytes))
			closeErr := outFile.Close()
			_ = rc.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			totalWritten += n
			if totalWritten > maxTotalExtractSize {
				return errors.New("zip-bomb terdeteksi: ukuran ekstraksi melampaui batas maksimum")
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
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		relPath, ok := stripPath(hdr.Name)
		if !ok {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := makeDirBeneath(baseDirFD, relPath); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := openFileBeneath(baseDirFD, relPath, 0640)
			if err != nil {
				return err
			}
			n, copyErr := io.Copy(outFile, io.LimitReader(tr, maxSingleFileBytes))
			closeErr := outFile.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			totalWritten += n
			if totalWritten > maxTotalExtractSize {
				return errors.New("archive-bomb terdeteksi: ukuran ekstraksi melampaui batas maksimum")
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("symlink/hardlink di dalam arsip tar ditolak: %s", hdr.Name)
		}
	}
	return nil
}
