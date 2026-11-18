package execsafe

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAllowedBinariesValidation(t *testing.T) {
	ctx := context.Background()

	// Prohibited binary should be rejected immediately before execution
	_, err := RunDirect(ctx, 2*time.Second, "/bin/sh", "-c", "echo test")
	if err == nil {
		t.Fatal("expected error for /bin/sh execution, got nil")
	}
	if !strings.Contains(err.Error(), "tidak terdaftar pada allowlist") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// Another non-whitelisted binary
	_, err = RunDirect(ctx, 2*time.Second, "/usr/bin/curl", "http://evil.com")
	if err == nil {
		t.Fatal("expected error for unwhitelisted binary, got nil")
	}
}

func TestCleanEnvironment(t *testing.T) {
	// Verify clean environment constants
	foundPath := false
	for _, env := range cleanEnv {
		if strings.HasPrefix(env, "PATH=") {
			foundPath = true
			if env != "PATH=/usr/sbin:/usr/bin:/sbin:/bin" {
				t.Fatalf("unexpected PATH in cleanEnv: %s", env)
			}
		}
	}
	if !foundPath {
		t.Fatal("PATH not found in cleanEnv")
	}
}
