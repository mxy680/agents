package portal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emdash-projects/agents/internal/portal/database"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	Router  *chi.Mux
	DB      *pgxpool.Pool
	Queries *database.Queries
	Config  *Config
}

// NewServer creates a new Server with configured routes.
func NewServer(cfg *Config, pool *pgxpool.Pool) *Server {
	s := &Server{
		Router:  chi.NewRouter(),
		DB:      pool,
		Queries: database.New(pool),
		Config:  cfg,
	}

	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Timeout(30 * time.Second))

	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return s
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.Config.Port)
	log.Printf("portal: listening on %s", addr)
	return http.ListenAndServe(addr, s.Router)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) {
	s.DB.Close()
}
