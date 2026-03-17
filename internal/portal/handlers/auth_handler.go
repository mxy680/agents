package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/emdash-projects/agents/internal/portal/database"
	mw "github.com/emdash-projects/agents/internal/portal/middleware"
)

// AuthHandler handles Google OAuth login, callback, and logout.
type AuthHandler struct {
	store       *mw.SessionStore
	queries     *database.Queries
	oauthConfig *oauth2.Config
	templates   *template.Template
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(store *mw.SessionStore, queries *database.Queries, oauthConfig *oauth2.Config, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		store:       store,
		queries:     queries,
		oauthConfig: oauthConfig,
		templates:   templates,
	}
}

// HandleLogin renders the login page. Redirects to /integrations if already authenticated.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if uid := mw.GetUserID(h.store, r); uid != "" {
		http.Redirect(w, r, "/integrations", http.StatusFound)
		return
	}
	h.templates.ExecuteTemplate(w, "layout", nil)
}

// HandleGoogleLogin initiates the Google OAuth2 flow.
func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	if err := mw.SetOAuthState(h.store, w, r, state); err != nil {
		log.Printf("auth: set oauth state: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleGoogleCallback handles the OAuth2 callback, exchanges the code, upserts the user,
// and sets the session.
func (h *AuthHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	expectedState, err := mw.GetOAuthState(h.store, w, r)
	if err != nil || expectedState == "" {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != expectedState {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "no authorization code", http.StatusBadRequest)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("auth: token exchange: %v", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Fetch user info from Google
	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("auth: fetch userinfo: %v", err)
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("auth: decode userinfo: %v", err)
		http.Error(w, "failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Upsert user in database
	user, err := h.queries.CreateUser(r.Context(), database.CreateUserParams{
		GoogleID:   userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		PictureURL: userInfo.Picture,
	})
	if err != nil {
		log.Printf("auth: upsert user: %v", err)
		http.Error(w, "failed to save user", http.StatusInternalServerError)
		return
	}

	// Format UUID bytes as standard UUID string
	b := user.ID.Bytes
	uid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])

	if err := mw.SetUserID(h.store, w, r, uid); err != nil {
		log.Printf("auth: set session: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/integrations", http.StatusFound)
}

// HandleLogout clears the session and redirects to the home page.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if err := mw.ClearSession(h.store, w, r); err != nil {
		log.Printf("auth: clear session: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// randomState generates a cryptographically random hex string for OAuth state.
func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
