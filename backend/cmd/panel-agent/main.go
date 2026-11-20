package main

import (
	"flag"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	agentv1 "github.com/alsyundawy/SparkPanelHost/backend/gen/agent/v1"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/agent/server"
	"google.golang.org/grpc"
)

var (
	Version   = "0.1.0"
	Commit    = "local"
	BuildTime = "unknown"
)

func main() {
	socketPath := flag.String("socket", "/run/panel/agent.sock", "Path to Unix Domain Socket")
	allowedUser := flag.String("allowed-user", "panel-web", "Username permitted to connect over Unix domain socket")
	configPath := flag.String("config", "", "Path to YAML configuration file (optional)")
	flag.Parse()

	log.Printf("[panel-agent] Starting SparkPanelHost System Agent %s (Commit: %s, Build: %s)", Version, Commit, BuildTime)
	if *configPath != "" {
		log.Printf("[panel-agent] Configuration profile: %s", *configPath)
	}

	// Ensure parent directory exists with restrictive permissions
	socketDir := filepath.Dir(*socketPath)
	if err := os.MkdirAll(socketDir, 0750); err != nil {
		log.Fatalf("Failed to create socket directory %s: %v", socketDir, err)
	}

	// Remove existing socket if present
	_ = os.Remove(*socketPath)

	// Resolve allowed UID
	var allowedUID uint32 = 0
	var allowedUIDInt = 0
	if u, err := user.Lookup(*allowedUser); err == nil {
		if uidInt, parseErr := strconv.Atoi(u.Uid); parseErr == nil && uidInt >= 0 && uidInt <= math.MaxInt32 {
			allowedUID = uint32(uidInt)
			allowedUIDInt = uidInt
			log.Printf("[panel-agent] Enforcing SO_PEERCRED access for UID: %d (%s)", allowedUID, *allowedUser)
		}
	} else {
		log.Printf("[panel-agent] Warning: User %s not found on host, allowing root (UID 0) only", *allowedUser)
	}

	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on unix socket %s: %v", *socketPath, err)
	}
	defer listener.Close()

	// Set socket permissions (0660 root:panel-web)
	if err := os.Chmod(*socketPath, 0660); err != nil {
		log.Printf("Warning: failed to chmod socket: %v", err)
	}
	if allowedUIDInt > 0 {
		_ = os.Chown(*socketPath, 0, allowedUIDInt)
	}

	// Create gRPC server with SO_PEERCRED security handshake
	grpcServer := grpc.NewServer(
		grpc.Creds(&server.PeerCreds{AllowedUID: allowedUID}),
	)

	agentServer := server.NewAgentServer()
	agentv1.RegisterSystemAgentServiceServer(grpcServer, agentServer)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[panel-agent] Shutting down gracefully...")
		grpcServer.GracefulStop()
		_ = os.Remove(*socketPath)
		os.Exit(0)
	}()

	log.Printf("[panel-agent] Daemon listening on %s (Zero-Shell, SO_PEERCRED enforced)", *socketPath)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
