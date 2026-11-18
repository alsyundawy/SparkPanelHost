package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/agentclient"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/api"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/workers"
)

var (
	Version   = "0.1.0"
	Commit    = "local"
	BuildTime = "unknown"
)

func main() {
	listenAddr := flag.String("listen", ":2087", "HTTP listen address")
	socketPath := flag.String("agent-socket", "/run/panel/agent.sock", "Path to panel-agent Unix Domain Socket")
	dbPath := flag.String("db", "/var/panel/data/sparkpanel.db", "SQLite database file path")
	migrationPath := flag.String("migrations", "migrations/001_init.sql", "SQL migrations file path")
	staticDir := flag.String("static", "dist/web", "Frontend static SPA distribution directory")
	recipeDir := flag.String("recipes", "recipes", "App installer recipes directory")
	flag.Parse()

	log.Printf("[panel-web] Starting SparkPanelHost Web Control Plane %s (Port: %s)", Version, *listenAddr)

	// Fallback database path if /var/panel/data is not writable
	actualDBPath := *dbPath
	if err := os.MkdirAll("/var/panel/data", 0750); err != nil {
		actualDBPath = "./sparkpanel.db"
	}

	// Fallback migration path
	actualMigrationPath := *migrationPath
	if _, err := os.Stat(actualMigrationPath); os.IsNotExist(err) {
		if _, err := os.Stat("../migrations/001_init.sql"); err == nil {
			actualMigrationPath = "../migrations/001_init.sql"
		} else if _, err := os.Stat("backend/migrations/001_init.sql"); err == nil {
			actualMigrationPath = "backend/migrations/001_init.sql"
		}
	}

	st, err := store.NewStore(actualDBPath, actualMigrationPath)
	if err != nil {
		log.Fatalf("[panel-web] Database initialization failed: %v", err)
	}
	defer st.Close()
	log.Printf("[panel-web] Database connected (SQLite WAL) at %s", actualDBPath)

	// Ensure default superadmin exists
	ensureSuperadmin(st)

	// Connect to panel-agent via Unix Domain Socket
	var agent *agentclient.Client
	ag, err := agentclient.NewClient(*socketPath)
	if err != nil {
		log.Printf("[panel-web] Warning: Could not connect to panel-agent (%v). Running in offline/standalone mode.", err)
	} else {
		agent = ag
		defer agent.Close()
		log.Printf("[panel-web] Connected to panel-agent daemon on %s", *socketPath)
	}

	// Recipe directory fallback
	actualRecipeDir := *recipeDir
	if _, err := os.Stat(actualRecipeDir); os.IsNotExist(err) {
		if _, err := os.Stat("../recipes"); err == nil {
			actualRecipeDir = "../recipes"
		}
	}

	// Static directory fallback
	actualStaticDir := *staticDir
	if _, err := os.Stat(actualStaticDir); os.IsNotExist(err) {
		if _, err := os.Stat("../frontend/dist"); err == nil {
			actualStaticDir = "../frontend/dist"
		} else {
			actualStaticDir = ""
		}
	}

	srv := api.NewServer(st, agent, actualRecipeDir)
	router := srv.Routes(actualStaticDir)

	// Background worker manager
	workerMgr := workers.NewWorkerManager(st)
	workerMgr.Start()
	defer workerMgr.Stop()

	httpServer := &http.Server{
		Addr:         *listenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[panel-web] Shutting down gracefully...")
		workerMgr.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
		os.Exit(0)
	}()

	log.Printf("[panel-web] Control plane ready and serving on http://%s", *listenAddr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[panel-web] Server error: %v", err)
	}
}

func ensureSuperadmin(st *store.Store) {
	ctx := context.Background()
	existing, err := st.ListUsers(ctx, nil)
	if err != nil || len(existing) > 0 {
		return
	}

	log.Println("[panel-web] Initializing default superadmin account...")
	defaultPass := "AetherAdmin@2026!"
	hash, err := auth.HashPassword(defaultPass)
	if err != nil {
		log.Fatalf("Failed to hash default admin password: %v", err)
	}

	adminUser := &store.User{
		ID:           "usr-superadmin",
		Username:     "admin",
		Email:        "admin@aetherpanel.local",
		PasswordHash: hash,
		Role:         "superadmin",
		LinuxUID:     0,
		LinuxGID:     0,
		Locale:       "id",
		Theme:        "dark",
		Status:       "active",
	}

	if err := st.CreateUser(ctx, adminUser); err != nil {
		log.Printf("Warning: failed to create initial admin user: %v", err)
		return
	}

	// Create default package
	defaultPkg := &store.Package{
		ID:               "pkg-default",
		Name:             "Default Hosting",
		DiskQuotaMB:      10240,
		InodeQuota:       500000,
		BandwidthQuotaMB: 204800,
		MaxDomains:       10,
		MaxDatabases:     10,
		MaxMailboxes:     20,
		MaxCronJobs:      10,
		MaxFTPAccounts:   5,
		CPUQuotaPercent:  100,
		MemoryLimitMB:    2048,
		IOWeight:         100,
		TasksMax:         256,
		ShellAccess:      true,
		AllowNodeJS:      true,
		AllowPython:      true,
		AllowGitDeploy:   true,
		AllowPostgres:    false,
	}
	_ = st.CreatePackage(ctx, defaultPkg)

	log.Println("================================================================")
	log.Println(" AetherPanel v8.1 Initial Admin Credentials:")
	log.Println(" Username: admin")
	log.Printf(" Password: %s", defaultPass)
	log.Println(" IMPORTANT: Please change this password immediately after login!")
	log.Println("================================================================")

	_ = st.AppendAuditLog(ctx, "system", nil, "init.superadmin", "user:usr-superadmin", "127.0.0.1")
}
