package portal

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emdash-projects/agents/internal/portal/database"
	"github.com/emdash-projects/agents/internal/portal/handlers"
	mw "github.com/emdash-projects/agents/internal/portal/middleware"
	portaloauth "github.com/emdash-projects/agents/internal/portal/oauth"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	Router  *chi.Mux
	DB      *pgxpool.Pool
	Queries *database.Queries
	Config  *Config
}

// parseTemplate parses the layout plus one page template from the embedded FS.
func parseTemplate(names ...string) *template.Template {
	return template.Must(template.ParseFS(templateFS, names...))
}

// NewServer creates a new Server with configured routes.
func NewServer(cfg *Config, pool *pgxpool.Pool) *Server {
	s := &Server{
		Router:  chi.NewRouter(),
		DB:      pool,
		Queries: database.New(pool),
		Config:  cfg,
	}

	s.Router.Use(chimiddleware.Logger)
	s.Router.Use(chimiddleware.Recoverer)
	s.Router.Use(chimiddleware.RealIP)
	s.Router.Use(chimiddleware.Timeout(30 * time.Second))

	// Session store and OAuth config
	store := mw.NewSessionStore(cfg.SessionSecret)
	googleLoginCfg := portaloauth.NewGoogleLoginConfig(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.BaseURL)

	// Parse templates (layout + page-specific content)
	homeTpl := parseTemplate("templates/layout.html", "templates/home.html")
	loginTpl := parseTemplate("templates/layout.html", "templates/login.html")

	// Handlers
	homeH := handlers.NewHomeHandler(store, homeTpl)
	authH := handlers.NewAuthHandler(store, s.Queries, googleLoginCfg, loginTpl)

	// Static files from embedded FS
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("portal: failed to create static sub-fs: %v", err)
	}
	s.Router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Public routes
	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	s.Router.Get("/", homeH.HandleHome)
	s.Router.Get("/login", authH.HandleLogin)
	s.Router.Get("/auth/google", authH.HandleGoogleLogin)
	s.Router.Get("/auth/google/callback", authH.HandleGoogleCallback)
	s.Router.Post("/logout", authH.HandleLogout)

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
