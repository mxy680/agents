// Package middleware provides HTTP middleware and session utilities for the portal.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

const (
	SessionName   = "portal_session"
	SessionUserID = "user_id"
)

// sessionData holds the values stored in a session cookie.
type sessionData struct {
	Values map[string]string
}

// SessionStore is a signed cookie-based session store.
// It uses HMAC-SHA256 to sign session data so cookies cannot be forged.
type SessionStore struct {
	secret []byte
	opts   *SessionOptions
}

// SessionOptions configures cookie attributes.
type SessionOptions struct {
	Path     string
	MaxAge   int
	HttpOnly bool
	SameSite http.SameSite
	Secure   bool
}

// NewSessionStore creates a SessionStore with the given secret.
func NewSessionStore(secret string) *SessionStore {
	return &SessionStore{
		secret: []byte(secret),
		opts: &SessionOptions{
			Path:     "/",
			MaxAge:   86400 * 30,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false,
		},
	}
}

// encode serializes sessionData to a base64+HMAC signed string.
func (s *SessionStore) encode(data *sessionData) (string, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(encoded))
	sig := hex.EncodeToString(mac.Sum(nil))

	return encoded + "|" + sig, nil
}

// decode verifies the HMAC and deserializes sessionData from a cookie value.
func (s *SessionStore) decode(value string) (*sessionData, error) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return nil, http.ErrNoCookie
	}
	encoded, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(encoded))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return nil, http.ErrNoCookie
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var data sessionData
	if err := gob.NewDecoder(bytes.NewReader(raw)).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

// getOrNew retrieves an existing session from the request cookie, or returns a new empty session.
func (s *SessionStore) getOrNew(r *http.Request) *sessionData {
	cookie, err := r.Cookie(SessionName)
	if err != nil {
		return &sessionData{Values: make(map[string]string)}
	}
	data, err := s.decode(cookie.Value)
	if err != nil {
		return &sessionData{Values: make(map[string]string)}
	}
	return data
}

// save writes a sessionData to the response as a signed cookie.
func (s *SessionStore) save(w http.ResponseWriter, data *sessionData, maxAge int) error {
	value, err := s.encode(data)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionName,
		Value:    value,
		Path:     s.opts.Path,
		MaxAge:   maxAge,
		HttpOnly: s.opts.HttpOnly,
		SameSite: s.opts.SameSite,
		Secure:   s.opts.Secure,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
	})
	return nil
}

// GetUserID extracts the user ID string from the session; returns empty string if not set.
func GetUserID(store *SessionStore, r *http.Request) string {
	data := store.getOrNew(r)
	return data.Values[SessionUserID]
}

// SetUserID stores the user ID in the session cookie.
func SetUserID(store *SessionStore, w http.ResponseWriter, r *http.Request, userID string) error {
	data := store.getOrNew(r)
	data.Values[SessionUserID] = userID
	return store.save(w, data, store.opts.MaxAge)
}

// ClearSession removes the session cookie.
func ClearSession(store *SessionStore, w http.ResponseWriter, r *http.Request) error {
	return store.save(w, &sessionData{Values: make(map[string]string)}, -1)
}

// SetOAuthState stores an OAuth state value in the session.
func SetOAuthState(store *SessionStore, w http.ResponseWriter, r *http.Request, state string) error {
	data := store.getOrNew(r)
	data.Values["oauth_state"] = state
	return store.save(w, data, store.opts.MaxAge)
}

// GetOAuthState retrieves the OAuth state value from the session and clears it.
func GetOAuthState(store *SessionStore, w http.ResponseWriter, r *http.Request) (string, error) {
	data := store.getOrNew(r)
	state := data.Values["oauth_state"]
	delete(data.Values, "oauth_state")
	if err := store.save(w, data, store.opts.MaxAge); err != nil {
		return "", err
	}
	return state, nil
}
