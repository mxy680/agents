package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/emdash-projects/agents/internal/orchestrator"
	_ "github.com/lib/pq"
)

func main() {
	cfg := orchestrator.DefaultConfig()

	// Override from environment
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}

	cfg.DatabaseURL = requireEnv("SUPABASE_DB_URL")
	cfg.EncryptionMasterKey = requireEnv("ENCRYPTION_MASTER_KEY")
	cfg.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	cfg.SupabaseJWTSecret = requireEnv("SUPABASE_JWT_SECRET")

	if v := os.Getenv("KUBE_NAMESPACE"); v != "" {
		cfg.KubeNamespace = v
	}
	if v := os.Getenv("AGENT_BASE_IMAGE"); v != "" {
		cfg.AgentBaseImage = v
	}
	if v := os.Getenv("EXPORT_CREDS_IMAGE"); v != "" {
		cfg.ExportCredsImage = v
	}
	if v := os.Getenv("CLAUDE_SESSION_SECRET"); v != "" {
		cfg.ClaudeSessionSecretRef = v
	}
	cfg.AllowedOrigin = os.Getenv("CORS_ALLOWED_ORIGIN")
	cfg.APIKey = os.Getenv("ORCHESTRATOR_API_KEY")
	cfg.AdminUserID = os.Getenv("ADMIN_USER_ID")

	if v := os.Getenv("RUNTIME"); v != "" {
		cfg.Runtime = v
	}
	cfg.ClaudeOAuthToken = os.Getenv("CLAUDE_CODE_OAUTH_TOKEN")

	// Database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	store := orchestrator.NewStore(db)

	// Container runtime
	var runtime orchestrator.ContainerRuntime
	switch cfg.Runtime {
	case "docker":
		log.Println("using Docker runtime")
		runtime = orchestrator.NewDockerRuntime()
	default:
		log.Println("using Kubernetes runtime")
		k8s, err := orchestrator.NewK8sClient(cfg.KubeNamespace)
		if err != nil {
			log.Fatalf("k8s client: %v", err)
		}
		runtime = orchestrator.NewK8sRuntime(k8s, cfg)
	}

	// Credential resolver
	creds := orchestrator.NewCredentialResolver(store, cfg)

	// Server
	srv := orchestrator.NewServer(cfg, store, runtime, creds)

	// Start reconciler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.StartReconciler(ctx, 10*time.Second)

	// Graceful shutdown — cancel context so reconciler stops, then let
	// defers (db.Close, cancel) run naturally.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		cancel()
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func requireEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("required env var %s not set", name)
	}
	return v
}
