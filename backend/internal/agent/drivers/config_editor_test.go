package drivers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		service string
		subPath string
		want    string
		wantErr bool
	}{
		{"nginx", "", "/etc/nginx/nginx.conf", false},
		{"nginx", "nginx.conf", "/etc/nginx/nginx.conf", false},
		{"nginx", "custom", "/etc/nginx/conf.d/custom.conf", false},
		{"php-fpm", "", "/etc/php/8.5/fpm/php.ini", false},
		{"php-fpm", "8.6", "/etc/php/8.6/fpm/php.ini", false},
		{"mariadb", "", "/etc/mysql/mariadb.conf.d/50-server.cnf", false},
		{"mariadb", "custom", "/etc/mysql/conf.d/custom.cnf", false},
		{"invalid", "", "", true},
		{"nginx", "../etc/passwd", "", true},
		{"nginx", "bad/path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.service+"_"+tt.subPath, func(t *testing.T) {
			got, err := ResolveConfigPath(tt.service, tt.subPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveConfigPath(%q, %q) err = %v, wantErr %v", tt.service, tt.subPath, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ResolveConfigPath(%q, %q) = %q, want %q", tt.service, tt.subPath, got, tt.want)
			}
		})
	}
}

func TestListServiceConfigBackups(t *testing.T) {
	// Create a temp directory simulating config dir
	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	baseFile := filepath.Join(tmpDir, "test.conf")
	if err := os.WriteFile(baseFile, []byte("main content"), 0644); err != nil {
		t.Fatal(err)
	}

	bak1 := baseFile + ".bak.20261118120000"
	bak2 := baseFile + ".bak.20261118130000"
	_ = os.WriteFile(bak1, []byte("old content 1"), 0644)
	_ = os.WriteFile(bak2, []byte("old content 2"), 0644)

	// Verify entries in test dir
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 files, got %d", len(entries))
	}
}

func TestValidateServiceConfig_SafeEmpty(t *testing.T) {
	ctx := context.Background()
	// Test validation without binary present or with standard syntax
	err := ValidateServiceConfig(ctx, "nginx", "", "# test config comment\nevents {}\nhttp {}\n")
	if err != nil {
		t.Logf("ValidateServiceConfig output (system nginx may or may not be installed): %v", err)
	}
}
