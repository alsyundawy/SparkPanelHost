<!-- markdownlint-disable -->
# 01. Kontrak gRPC Protobuf & Daemon Go Zero-Shell (`panel-agent` v8.1)

Dokumen ini menjabarkan arsitektur keamanan tingkat sistem operasi pada `panel-agent` untuk meniadakan **CWE-78** (*OS Command Injection*), **CWE-426** (*Untrusted Search Path / Env Injection*), **CWE-22 / CWE-59 / CWE-362** (*Path Traversal & Symlink TOCTOU*), **CWE-409** (*Zip Bomb*), dan **CWE-918 / CWE-494** (*SSRF & Unverified Code Download*).

---

## 1. Autentikasi Unix Domain Socket (`SO_PEERCRED`)

`panel-agent` mendengarkan `/run/panel/agent.sock` (`0660`, `root:panel-web`) dan memverifikasi kredensial proses kernel pada setiap koneksi masuk menggunakan `unix.GetsockoptUcred`.

```go
package server

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc/credentials"
)

type PeerAuthInfo struct {
	UID uint32
	PID int32
}

func (PeerAuthInfo) AuthType() string { return "unix-peercred" }

type PeerCreds struct {
	AllowedUID uint32
}

func (p *PeerCreds) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	unixConn, ok := rawConn.(*net.UnixConn)
	if !ok {
		return nil, nil, errors.New("hanya koneksi unix domain socket yang diizinkan")
	}

	sysConn, err := unixConn.SyscallConn()
	if err != nil {
		return nil, nil, err
	}

	var ucred *unix.Ucred
	var sockErr error
	err = sysConn.Control(func(fd uintptr) {
		ucred, sockErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil {
		return nil, nil, err
	}
	if sockErr != nil {
		return nil, nil, sockErr
	}

	if ucred.Uid != p.AllowedUID && ucred.Uid != 0 {
		return nil, nil, fmt.Errorf("security violation: akses socket ditolak untuk UID %d (PID %d)", ucred.Uid, ucred.Pid)
	}

	return rawConn, PeerAuthInfo{UID: ucred.Uid, PID: ucred.Pid}, nil
}
```

---

## 2. Zero-Shell Execution Runner dengan Sanitized Environment (`execsafe/runner.go`)

Tidak ada satu pun pemanggilan `sh -c` atau `bash -c`. Selain daftar biner eksplisit (*allowlist*), `cmd.Env` selalu ditimpa dengan environment minimal yang deterministik untuk mencegah injeksi `LD_PRELOAD` maupun manipulasi `PATH`.

```go
package execsafe

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

var allowedBinaries = map[string]bool{
	"/usr/sbin/nginx":          true,
	"/usr/bin/systemctl":       true,
	"/usr/bin/systemd-run":     true,
	"/usr/sbin/useradd":        true,
	"/usr/sbin/usermod":        true,
	"/usr/sbin/userdel":        true,
	"/usr/sbin/setquota":       true,
	"/usr/bin/rspamadm":        true,
	"/usr/bin/doveadm":         true,
	"/usr/sbin/postmap":        true,
	"/usr/bin/mariadb-dump":    true,
	"/usr/bin/mariadb-backup":  true,
	"/usr/bin/pg_dump":         true,
	"/usr/bin/certbot":         true,
	"/usr/bin/restic":          true,
	"/usr/bin/rclone":          true,
	"/usr/sbin/nft":            true,
	"/usr/bin/cscli":           true,
	"/usr/bin/pdnsutil":        true,
	"/usr/bin/crontab":         true,
	"/usr/bin/git":             true,
}

var cleanEnv = []string{
	"PATH=/usr/sbin:/usr/bin:/sbin:/bin",
	"LC_ALL=C.UTF-8",
	"LANG=C.UTF-8",
}

func RunDirect(ctx context.Context, timeout time.Duration, binary string, args ...string) (string, error) {
	if !allowedBinaries[binary] {
		return "", fmt.Errorf("security violation: biner %q tidak terdaftar pada allowlist", binary)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, binary, args...)
	cmd.Env = cleanEnv

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("eksekusi %s gagal (%v): %s", binary, err, stderr.String())
	}
	return stdout.String(), nil
}
```

---

## 3. Kernel-Enforced Safe File & Archive Extraction (`openat2` + Anti-Zip-Bomb)

Pada v6.0, penggunaan `unix.O_NOFOLLOW` hanya memeriksa komponen nama file terakhir, sehingga masih rentan terhadap *intermediate directory symlink* (misalnya `public_html/subdir -> /etc`). Pada **v8.1**, setiap pembukaan direktori dan file di dalam direktori tenant menggunakan syscall Linux **`unix.Openat2`** dengan **`unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS`**, didukung untuk arsip **`.tar.gz` maupun `.zip`**, serta dibatasi ukuran total ekstraksinya untuk mencegah **Zip Bomb (CWE-409)**.

```go
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
	maxSingleFileBytes  int64 = 512 * 1024 * 1024  // 512 MB per file
	maxTotalExtractSize int64 = 4096 * 1024 * 1024 // 4 GB total arsip (anti Zip-Bomb CWE-409)
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

// openFileBeneath menggunakan Linux openat2(RESOLVE_BENEATH | RESOLVE_NO_SYMLINKS)
// sehingga kernel VFS menolak traversal maupun symlink pada komponen parent mana pun.
func openFileBeneath(baseDirFD int, relPath string, mode uint32) (*os.File, error) {
	cleanRel := path.Clean("/" + relPath)
	cleanRel = strings.TrimPrefix(cleanRel, "/")
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

	// Buat dan buka setiap subdirektori parent secara aman
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
			// Tolak symlink di dalam arsip ZIP
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
				rc.Close()
				return err
			}
			n, copyErr := io.Copy(outFile, io.LimitReader(rc, maxSingleFileBytes))
			closeErr := outFile.Close()
			rc.Close()
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
```

---

## 4. SSRF-Safe Upstream Downloader (`netfetch/safehttp.go`)

Mencegah **CWE-918 (SSRF)** saat `panel-agent` mengunduh arsip aplikasi dari URL manifes dengan memvalidasi host HTTPS dan memblokir seluruh alamat IP non-publik langsung pada lapisan socket dialer (*anti-DNS-rebinding*).

```go
package netfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"
)

var allowedUpstreamHosts = map[string]bool{
	"wordpress.org":                  true,
	"downloads.wordpress.org":        true,
	"github.com":                     true,
	"api.github.com":                 true,
	"codeload.github.com":            true,
	"objects.githubusercontent.com":  true,
	"downloads.joomla.org":           true,
	"ftp.drupal.org":                 true,
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	return true
}

func DownloadVerifiedUpstream(ctx context.Context, rawURL, destFile string, maxBytes int64) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || !allowedUpstreamHosts[parsed.Hostname()] {
		return fmt.Errorf("security violation: URL upstream tidak diizinkan %q", rawURL)
	}

	dialer := &net.Dialer{
		Timeout: 15 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if !isPublicIP(ip) {
				return fmt.Errorf("SSRF blocked: koneksi ke IP non-publik %s ditolak", host)
			}
			return nil
		},
	}

	client := &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			DialContext:       dialer.DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("terlalu banyak redirect HTTP")
			}
			if req.URL.Scheme != "https" || !allowedUpstreamHosts[req.URL.Hostname()] {
				return fmt.Errorf("redirect ke host tidak diizinkan: %s", req.URL.Hostname())
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream HTTP status %d", resp.StatusCode)
	}

	out, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(resp.Body, maxBytes))
	return err
}
```
