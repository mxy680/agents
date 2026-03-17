package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/emdash-projects/agents/internal/portal/crypto"
	"github.com/emdash-projects/agents/internal/portal/database"
	mw "github.com/emdash-projects/agents/internal/portal/middleware"
)

// IntegrationsHandler handles the integrations dashboard and provider connect/disconnect flows.
type IntegrationsHandler struct {
	store        *mw.SessionStore
	queries      *database.Queries
	encryptionKey []byte
	googleConfig  *oauth2.Config
	githubConfig  *oauth2.Config // nil if GitHub is not configured
	templates    *template.Template
}

// NewIntegrationsHandler creates a new IntegrationsHandler.
// githubConfig may be nil if GitHub OAuth credentials are not configured.
func NewIntegrationsHandler(
	store *mw.SessionStore,
	queries *database.Queries,
	encryptionKey []byte,
	googleConfig *oauth2.Config,
	githubConfig *oauth2.Config,
	templates *template.Template,
) *IntegrationsHandler {
	return &IntegrationsHandler{
		store:         store,
		queries:       queries,
		encryptionKey: encryptionKey,
		googleConfig:  googleConfig,
		githubConfig:  githubConfig,
		templates:     templates,
	}
}

// integrationsData is the template data for the integrations page.
type integrationsData struct {
	LoggedIn           bool
	GoogleConnected    bool
	GitHubAvailable    bool
	GitHubConnected    bool
	InstagramConnected bool
	ShowInstagramForm  bool
}

// HandleIntegrations renders the integrations dashboard (GET /integrations).
func (h *IntegrationsHandler) HandleIntegrations(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	integrations, err := h.queries.GetIntegrationsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("integrations: fetch integrations: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := &integrationsData{
		LoggedIn:        true,
		GitHubAvailable: h.githubConfig != nil,
	}
	for _, intg := range integrations {
		switch intg.Provider {
		case "google":
			data.GoogleConnected = true
		case "github":
			data.GitHubConnected = true
		case "instagram":
			data.InstagramConnected = true
		}
	}

	if err := h.templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("integrations: render template: %v", err)
	}
}

// HandleGoogleConnect initiates the Google OAuth2 flow (GET /integrations/google/connect).
func (h *IntegrationsHandler) HandleGoogleConnect(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	if err := mw.SetOAuthState(h.store, w, r, state); err != nil {
		log.Printf("integrations: set oauth state: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	url := h.googleConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleGoogleCallback handles the Google OAuth2 callback (GET /integrations/google/callback).
func (h *IntegrationsHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := h.validateOAuthState(w, r); err != nil {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "no authorization code", http.StatusBadRequest)
		return
	}

	token, err := h.googleConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("integrations: google token exchange: %v", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	if err := h.saveOAuthToken(r, userID, "google", token); err != nil {
		log.Printf("integrations: save google token: %v", err)
		http.Error(w, "failed to save integration", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/integrations", http.StatusFound)
}

// HandleGoogleDisconnect removes the Google integration (POST /integrations/google/disconnect).
func (h *IntegrationsHandler) HandleGoogleDisconnect(w http.ResponseWriter, r *http.Request) {
	h.handleDisconnect(w, r, "google")
}

// HandleGitHubConnect initiates the GitHub OAuth2 flow (GET /integrations/github/connect).
func (h *IntegrationsHandler) HandleGitHubConnect(w http.ResponseWriter, r *http.Request) {
	if h.githubConfig == nil {
		http.Error(w, "GitHub integration is not configured", http.StatusServiceUnavailable)
		return
	}

	state := randomState()
	if err := mw.SetOAuthState(h.store, w, r, state); err != nil {
		log.Printf("integrations: set oauth state: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	url := h.githubConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleGitHubCallback handles the GitHub OAuth2 callback (GET /integrations/github/callback).
func (h *IntegrationsHandler) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.githubConfig == nil {
		http.Error(w, "GitHub integration is not configured", http.StatusServiceUnavailable)
		return
	}

	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := h.validateOAuthState(w, r); err != nil {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "no authorization code", http.StatusBadRequest)
		return
	}

	token, err := h.githubConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("integrations: github token exchange: %v", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	if err := h.saveOAuthToken(r, userID, "github", token); err != nil {
		log.Printf("integrations: save github token: %v", err)
		http.Error(w, "failed to save integration", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/integrations", http.StatusFound)
}

// HandleGitHubDisconnect removes the GitHub integration (POST /integrations/github/disconnect).
func (h *IntegrationsHandler) HandleGitHubDisconnect(w http.ResponseWriter, r *http.Request) {
	h.handleDisconnect(w, r, "github")
}

// HandleInstagramConnect shows the Instagram cookie entry form (GET /integrations/instagram/connect).
func (h *IntegrationsHandler) HandleInstagramConnect(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	integrations, err := h.queries.GetIntegrationsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("integrations: fetch integrations: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := &integrationsData{
		LoggedIn:          true,
		GitHubAvailable:   h.githubConfig != nil,
		ShowInstagramForm: true,
	}
	for _, intg := range integrations {
		switch intg.Provider {
		case "google":
			data.GoogleConnected = true
		case "github":
			data.GitHubConnected = true
		case "instagram":
			data.InstagramConnected = true
		}
	}

	if err := h.templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("integrations: render template: %v", err)
	}
}

// HandleInstagramSave stores encrypted Instagram session cookies (POST /integrations/instagram/save).
func (h *IntegrationsHandler) HandleInstagramSave(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	sessionID := strings.TrimSpace(r.FormValue("session_id"))
	csrfToken := strings.TrimSpace(r.FormValue("csrf_token"))
	dsUserID := strings.TrimSpace(r.FormValue("ds_user_id"))
	mid := strings.TrimSpace(r.FormValue("mid"))
	igDid := strings.TrimSpace(r.FormValue("ig_did"))

	if sessionID == "" || csrfToken == "" || dsUserID == "" {
		http.Error(w, "session_id, csrf_token, and ds_user_id are required", http.StatusBadRequest)
		return
	}

	encryptedSessionID, err := crypto.Encrypt(h.encryptionKey, sessionID)
	if err != nil {
		log.Printf("integrations: encrypt session_id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	encryptedCSRF, err := crypto.Encrypt(h.encryptionKey, csrfToken)
	if err != nil {
		log.Printf("integrations: encrypt csrf_token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	encryptedDSUserID, err := crypto.Encrypt(h.encryptionKey, dsUserID)
	if err != nil {
		log.Printf("integrations: encrypt ds_user_id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	meta := map[string]string{
		"session_id": encryptedSessionID,
		"csrf_token": encryptedCSRF,
		"ds_user_id": encryptedDSUserID,
	}

	if mid != "" {
		encryptedMid, err := crypto.Encrypt(h.encryptionKey, mid)
		if err != nil {
			log.Printf("integrations: encrypt mid: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		meta["mid"] = encryptedMid
	}

	if igDid != "" {
		encryptedIgDid, err := crypto.Encrypt(h.encryptionKey, igDid)
		if err != nil {
			log.Printf("integrations: encrypt ig_did: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		meta["ig_did"] = encryptedIgDid
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		log.Printf("integrations: marshal instagram meta: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = h.queries.UpsertIntegration(r.Context(), database.UpsertIntegrationParams{
		UserID:   userID,
		Provider: "instagram",
		Status:   "active",
		Metadata: metaJSON,
	})
	if err != nil {
		log.Printf("integrations: upsert instagram: %v", err)
		http.Error(w, "failed to save integration", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/integrations", http.StatusFound)
}

// HandleInstagramDisconnect removes the Instagram integration (POST /integrations/instagram/disconnect).
func (h *IntegrationsHandler) HandleInstagramDisconnect(w http.ResponseWriter, r *http.Request) {
	h.handleDisconnect(w, r, "instagram")
}

// handleDisconnect is a shared helper for removing an integration from the database.
func (h *IntegrationsHandler) handleDisconnect(w http.ResponseWriter, r *http.Request, provider string) {
	userID, err := h.getUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := h.queries.DeleteIntegration(r.Context(), userID, provider); err != nil {
		log.Printf("integrations: delete %s integration: %v", provider, err)
		http.Error(w, "failed to disconnect integration", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/integrations", http.StatusFound)
}

// saveOAuthToken encrypts and stores an OAuth token in the database.
func (h *IntegrationsHandler) saveOAuthToken(r *http.Request, userID pgtype.UUID, provider string, token *oauth2.Token) error {
	encryptedAccess, err := crypto.Encrypt(h.encryptionKey, token.AccessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}

	var encryptedRefresh string
	if token.RefreshToken != "" {
		encryptedRefresh, err = crypto.Encrypt(h.encryptionKey, token.RefreshToken)
		if err != nil {
			return fmt.Errorf("encrypt refresh token: %w", err)
		}
	}

	expiry := pgtype.Timestamptz{Time: token.Expiry, Valid: !token.Expiry.IsZero()}

	params := database.UpsertIntegrationParams{
		UserID:      userID,
		Provider:    provider,
		Status:      "active",
		AccessToken: pgtype.Text{String: encryptedAccess, Valid: true},
		TokenExpiry: expiry,
	}
	if encryptedRefresh != "" {
		params.RefreshToken = pgtype.Text{String: encryptedRefresh, Valid: true}
	}

	_, err = h.queries.UpsertIntegration(r.Context(), params)
	return err
}

// getUserID retrieves and parses the authenticated user's UUID from the session.
func (h *IntegrationsHandler) getUserID(r *http.Request) (pgtype.UUID, error) {
	uidStr := mw.GetUserID(h.store, r)
	if uidStr == "" {
		return pgtype.UUID{}, fmt.Errorf("not authenticated")
	}
	return parseUUID(uidStr)
}

// validateOAuthState checks the OAuth state parameter from the callback against the session.
func (h *IntegrationsHandler) validateOAuthState(w http.ResponseWriter, r *http.Request) error {
	expectedState, err := mw.GetOAuthState(h.store, w, r)
	if err != nil || expectedState == "" {
		return fmt.Errorf("missing or invalid oauth state")
	}
	if r.URL.Query().Get("state") != expectedState {
		return fmt.Errorf("state mismatch")
	}
	return nil
}

// parseUUID parses a standard UUID string (with or without dashes) into a pgtype.UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: %s", s)
	}
	var uuid pgtype.UUID
	b, err := hex.DecodeString(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	copy(uuid.Bytes[:], b)
	uuid.Valid = true
	return uuid, nil
}
