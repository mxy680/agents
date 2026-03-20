package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

// Server is the HTTP API server for the orchestrator.
type Server struct {
	cfg    Config
	store  *Store
	k8s    *K8sClient
	creds  *CredentialResolver
	router chi.Router
}

// NewServer creates and configures a new Server.
func NewServer(cfg Config, store *Store, k8s *K8sClient, creds *CredentialResolver) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
		k8s:   k8s,
		creds: creds,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Get("/templates", s.handleListTemplates)
		r.Get("/templates/{id}", s.handleGetTemplate)
		r.Post("/agents/deploy", s.handleDeploy)
		r.Get("/agents", s.handleListInstances)
		r.Get("/agents/{id}", s.handleGetInstance)
		r.Get("/agents/{id}/logs", s.handleGetLogs)
		r.Post("/agents/{id}/stop", s.handleStopAgent)
		r.Delete("/agents/{id}", s.handleDeleteInstance)
	})

	// Health check (no auth)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.router = r
	return s
}

// Router returns the HTTP handler for the server.
func (s *Server) Router() http.Handler {
	return s.router
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Printf("orchestrator listening on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// contextKey is a private type for context keys.
type contextKey string

const userIDKey contextKey = "user_id"

// authMiddleware validates Supabase JWT and extracts user ID.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			writeError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.cfg.SupabaseJWTSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid claims")
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			writeError(w, http.StatusUnauthorized, "missing subject")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserID(r *http.Request) string {
	return r.Context().Value(userIDKey).(string)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
