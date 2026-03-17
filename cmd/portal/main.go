package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	portal "github.com/emdash-projects/agents/internal/portal"
)

func main() {
	cfg, err := portal.Load()
	if err != nil {
		log.Fatalf("portal: config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("portal: database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("portal: database ping: %v", err)
	}

	srv := portal.NewServer(cfg, pool)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("portal: shutting down...")
		srv.Shutdown(ctx)
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("portal: server: %v", err)
	}
}
