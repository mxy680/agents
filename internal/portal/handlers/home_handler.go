package handlers

import (
	"html/template"
	"net/http"

	mw "github.com/emdash-projects/agents/internal/portal/middleware"
)

// HomeHandler renders the home page.
type HomeHandler struct {
	store     *mw.SessionStore
	templates *template.Template
}

// NewHomeHandler creates a new HomeHandler.
func NewHomeHandler(store *mw.SessionStore, templates *template.Template) *HomeHandler {
	return &HomeHandler{store: store, templates: templates}
}

// HandleHome renders the home page, passing login state to the template.
func (h *HomeHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"LoggedIn": mw.GetUserID(h.store, r) != "",
	}
	h.templates.ExecuteTemplate(w, "layout", data)
}
