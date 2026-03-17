// oauth-setup opens a browser for Google OAuth consent flow and prints
// the resulting access and refresh tokens. Use these to populate
// GMAIL_ACCESS_TOKEN and GMAIL_REFRESH_TOKEN in Doppler.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	clientID := os.Getenv("GOOGLE_DESKTOP_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_DESKTOP_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "GOOGLE_DESKTOP_CLIENT_ID and GOOGLE_DESKTOP_CLIENT_SECRET must be set")
		fmt.Fprintln(os.Stderr, "Run: doppler run -- go run ./cmd/oauth-setup")
		os.Exit(1)
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8089/callback",
		Scopes: []string{
			gmail.MailGoogleComScope,
			gmail.GmailSettingsBasicScope,
			gmail.GmailSettingsSharingScope,
		},
	}

	state := randomState()
	tokenCh := make(chan *oauth2.Token, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			http.Error(w, "no code: "+errMsg, http.StatusBadRequest)
			errCh <- fmt.Errorf("OAuth error: %s", errMsg)
			return
		}

		tok, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			errCh <- err
			return
		}

		fmt.Fprint(w, "<html><body><h1>Done! You can close this tab.</h1></body></html>")
		tokenCh <- tok
	})

	server := &http.Server{Addr: ":8089", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("HTTP server: %w", err)
		}
	}()

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Fprintf(os.Stderr, "\nOpening browser for OAuth consent...\n")
	fmt.Fprintf(os.Stderr, "If it doesn't open, visit:\n%s\n\n", authURL)
	openBrowser(authURL)

	select {
	case tok := <-tokenCh:
		server.Shutdown(context.Background())
		fmt.Fprintf(os.Stderr, "\nTokens received. Add to Doppler:\n\n")
		fmt.Fprintf(os.Stderr, "  doppler secrets set GMAIL_ACCESS_TOKEN='%s' GMAIL_REFRESH_TOKEN='%s'\n\n", tok.AccessToken, tok.RefreshToken)

		// Also output as JSON for scripting
		out := map[string]string{
			"access_token":  tok.AccessToken,
			"refresh_token": tok.RefreshToken,
			"expiry":        tok.Expiry.String(),
		}
		json.NewEncoder(os.Stdout).Encode(out)

	case err := <-errCh:
		server.Shutdown(context.Background())
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	cmd.Start()
}
