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
	if v := os.Getenv("ANTHROPIC_API_KEY_SECRET"); v != "" {
		cfg.AnthropicAPIKeyRef = v
	}

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

	// K8s client
	k8s, err := orchestrator.NewK8sClient(cfg.KubeNamespace)
	if err != nil {
		log.Fatalf("k8s client: %v", err)
	}

	// Credential resolver
	creds := orchestrator.NewCredentialResolver(store, cfg)

	// Server
	srv := orchestrator.NewServer(cfg, store, k8s, creds)

	// Start reconciler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.StartReconciler(ctx, 10*time.Second)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		os.Exit(0)
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
