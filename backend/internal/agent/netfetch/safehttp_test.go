package netfetch

import (
	"context"
	"net"
	"strings"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		ip       string
		isPublic bool
	}{
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false},
		{"::1", false},
		{"fc00::1", false},
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"140.82.112.4", true}, // GitHub IP
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if ip == nil {
			t.Fatalf("failed to parse IP %s", tt.ip)
		}
		if got := isPublicIP(ip); got != tt.isPublic {
			t.Errorf("isPublicIP(%s) = %v, want %v", tt.ip, got, tt.isPublic)
		}
	}
}

func TestDownloadVerifiedUpstreamRejection(t *testing.T) {
	ctx := context.Background()

	// 1. Non-https scheme
	err := DownloadVerifiedUpstream(ctx, "http://downloads.wordpress.org/test.zip", "/tmp/test.zip", 1000)
	if err == nil || !strings.Contains(err.Error(), "URL upstream tidak diizinkan") {
		t.Fatalf("expected error for non-https URL, got: %v", err)
	}

	// 2. Untrusted domain / SSRF candidate
	err = DownloadVerifiedUpstream(ctx, "https://169.254.169.254/latest/meta-data/", "/tmp/test.zip", 1000)
	if err == nil || !strings.Contains(err.Error(), "URL upstream tidak diizinkan") {
		t.Fatalf("expected error for untrusted domain/IP, got: %v", err)
	}

	// 3. Private host not in allowlist
	err = DownloadVerifiedUpstream(ctx, "https://internal.corp/secret.tar.gz", "/tmp/test.zip", 1000)
	if err == nil || !strings.Contains(err.Error(), "URL upstream tidak diizinkan") {
		t.Fatalf("expected error for unwhitelisted host, got: %v", err)
	}
}
