package drivers

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/execsafe"
)

var dbIdentifierRegex = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

func ValidateDBIdentifier(name string) error {
	if !dbIdentifierRegex.MatchString(name) {
		return fmt.Errorf("invalid database/user identifier %q: must match ^[a-z0-9_]{3,32}$", name)
	}
	return nil
}

// CreateDatabase provisions a new database and user on MariaDB 12.3 or PostgreSQL 17+.
func CreateDatabase(ctx context.Context, username, engine, dbName, dbUser, dbPassword string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if err := ValidateDBIdentifier(dbName); err != nil {
		return err
	}
	if err := ValidateDBIdentifier(dbUser); err != nil {
		return err
	}

	if !strings.HasPrefix(dbName, username+"_") {
		return fmt.Errorf("database name %q must begin with %s_", dbName, username)
	}

	switch engine {
	case "mariadb", "mysql":
		// MariaDB 12.3 LTS provisioning with utf8mb4
		sql := fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; "+
				"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s'; "+
				"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'localhost'; "+
				"FLUSH PRIVILEGES;",
			dbName, dbUser, escapeSQL(dbPassword), dbName, dbUser,
		)
		_, err := execsafe.RunDirect(ctx, 30*time.Second, "/usr/bin/mariadb", "-e", sql)
		return err

	case "postgresql", "postgres":
		// PostgreSQL 17/18 user and database creation
		createRoleSQL := fmt.Sprintf("DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN CREATE ROLE \"%s\" WITH LOGIN PASSWORD '%s'; END IF; END $$;", dbUser, dbUser, escapeSQL(dbPassword))
		_, err := execsafe.RunDirect(ctx, 30*time.Second, "/usr/bin/psql", "-U", "postgres", "-c", createRoleSQL)
		if err != nil {
			return fmt.Errorf("failed to create postgres role: %w", err)
		}

		createDbSQL := fmt.Sprintf("SELECT 'CREATE DATABASE \"%s\" OWNER \"%s\" ENCODING ''UTF8''' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '%s')\\gexec", dbName, dbUser, dbName)
		_, err = execsafe.RunDirect(ctx, 30*time.Second, "/usr/bin/psql", "-U", "postgres", "-c", createDbSQL)
		return err

	default:
		return fmt.Errorf("unsupported database engine: %q", engine)
	}
}

// BackupDatabase dumps a database safely into outputDir and calculates its SHA-256 hash.
func BackupDatabase(ctx context.Context, username, engine, dbName, outputDir string) (string, string, error) {
	if err := ValidateUsername(username); err != nil {
		return "", "", err
	}
	if err := ValidateDBIdentifier(dbName); err != nil {
		return "", "", err
	}

	cleanOut := filepath.Clean(outputDir)
	if !strings.HasPrefix(cleanOut, "/var/panel/tmp/"+username) && !strings.HasPrefix(cleanOut, "/home/"+username) {
		return "", "", fmt.Errorf("security violation: output directory %q must be in /var/panel/tmp/%s or /home/%s", cleanOut, username, username)
	}

	if err := os.MkdirAll(cleanOut, 0700); err != nil {
		return "", "", err
	}

	timestamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("%s-%s.sql.gz", dbName, timestamp)
	archivePath := filepath.Join(cleanOut, archiveName)

	switch engine {
	case "mariadb", "mysql":
		rawSQL, err := execsafe.RunDirect(ctx, 10*time.Minute, "/usr/bin/mariadb-dump",
			"--single-transaction",
			"--quick",
			"--skip-lock-tables",
			dbName,
		)
		if err != nil {
			return "", "", fmt.Errorf("mariadb-dump failed: %w", err)
		}

		if err := writeGzipped(archivePath, []byte(rawSQL)); err != nil {
			return "", "", err
		}

	case "postgresql", "postgres":
		rawDump, err := execsafe.RunDirect(ctx, 10*time.Minute, "/usr/bin/pg_dump",
			"-U", "postgres",
			"-d", dbName,
		)
		if err != nil {
			return "", "", fmt.Errorf("pg_dump failed: %w", err)
		}

		if err := writeGzipped(archivePath, []byte(rawDump)); err != nil {
			return "", "", err
		}

	default:
		return "", "", fmt.Errorf("unsupported database engine: %q", engine)
	}

	// Compute SHA-256
	h := sha256.New()
	f, err := os.Open(archivePath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return "", "", err
	}
	digest := hex.EncodeToString(h.Sum(nil))

	return archivePath, digest, nil
}

func writeGzipped(destPath string, data []byte) error {
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	if _, err := gw.Write(data); err != nil {
		_ = gw.Close()
		return err
	}
	return gw.Close()
}

func escapeSQL(val string) string {
	val = strings.ReplaceAll(val, `\`, `\\`)
	val = strings.ReplaceAll(val, `'`, `\'`)
	val = strings.ReplaceAll(val, `"`, `\"`)
	return val
}
