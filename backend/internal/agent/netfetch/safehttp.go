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
	"wordpress.org":                 true,
	"downloads.wordpress.org":       true,
	"github.com":                    true,
	"api.github.com":                true,
	"codeload.github.com":           true,
	"objects.githubusercontent.com": true,
	"downloads.joomla.org":          true,
	"ftp.drupal.org":                true,
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
			DialContext:         dialer.DialContext,
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
