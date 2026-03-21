package x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
)

const (
	xBaseURL      = "https://x.com"
	uploadBaseURL = "https://upload.x.com"
	capsBaseURL   = "https://caps.x.com"

	// xBearerToken is the static public bearer token used by X's web client.
	xBearerToken = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
)

// DefaultFeatures contains the common GraphQL feature flags expected by X's API.
var DefaultFeatures = map[string]bool{
	"rweb_tipjar_consumption_enabled":                                         true,
	"responsive_web_graphql_exclude_directive_enabled":                        true,
	"verified_phone_label_enabled":                                            false,
	"creator_subscriptions_tweet_preview_api_enabled":                         true,
	"responsive_web_graphql_timeline_navigation_enabled":                      true,
	"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
	"communities_web_enable_tweet_community_results_fetch":                    true,
	"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
	"articles_preview_enabled":                                                true,
	"responsive_web_edit_tweet_api_enabled":                                   true,
	"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
	"view_counts_everywhere_api_enabled":                                      true,
	"longform_notetweets_consumption_enabled":                                 true,
	"responsive_web_twitter_article_tweet_consumption_enabled":                true,
	"tweet_awards_web_tipping_enabled":                                        false,
	"creator_subscriptions_quote_tweet_preview_enabled":                       false,
	"freedom_of_speech_not_reach_fetch_enabled":                               true,
	"standardized_nudges_misinfo":                                             true,
	"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
	"rweb_video_timestamps_enabled":                                           true,
	"longform_notetweets_rich_text_read_enabled":                              true,
	"longform_notetweets_inline_media_enabled":                                true,
	"responsive_web_enhance_cards_enabled":                                    false,
}

// RateLimitError is returned when X responds with HTTP 429.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("x rate limit exceeded; retry after %s", e.RetryAfter)
	}
	return "x rate limit exceeded"
}

// AccountLockedError is returned when X locks the account (e.g. suspicious activity).
type AccountLockedError struct {
	Message string
}

func (e *AccountLockedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("x account locked: %s", e.Message)
	}
	return "x account locked"
}

// Client is an HTTP client wrapper for X's internal GraphQL and v1.1 APIs.
type Client struct {
	http      *http.Client
	session   *auth.XSession
	baseURL   string
	uploadURL string // base URL for upload.x.com (defaults to uploadBaseURL; overridable in tests)
	capsURL   string // base URL for caps.x.com (defaults to capsBaseURL; overridable in tests)

	mu   sync.Mutex
	csrf string // ct0 cookie value, also sent as X-CSRF-Token header
}

// ClientFactory is the function signature for creating a Client.
type ClientFactory func(ctx context.Context) (*Client, error)

// DefaultClientFactory returns a ClientFactory that reads real credentials
// from environment variables.
func DefaultClientFactory() ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session, err := auth.NewXSession()
		if err != nil {
			return nil, fmt.Errorf("x auth: %w", err)
		}
		return &Client{
			http:      &http.Client{Timeout: 30 * time.Second},
			session:   session,
			baseURL:   xBaseURL,
			uploadURL: uploadBaseURL,
			capsURL:   capsBaseURL,
			csrf:      session.CSRFToken,
		}, nil
	}
}

// newClientWithBase creates a Client targeting a custom base URL (used in tests).
// uploadURL and capsURL also point to the same base URL so tests can mock all endpoints.
func newClientWithBase(session *auth.XSession, httpClient *http.Client, base string) *Client {
	return &Client{
		http:      httpClient,
		session:   session,
		baseURL:   base,
		uploadURL: base,
		capsURL:   base,
		csrf:      session.CSRFToken,
	}
}

// applyHeaders sets all required X API request headers.
func (c *Client) applyHeaders(req *http.Request) {
	c.mu.Lock()
	csrf := c.csrf
	c.mu.Unlock()

	req.Header.Set("Authorization", "Bearer "+xBearerToken)
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("Cookie", c.session.CookieString())
	req.Header.Set("User-Agent", c.session.UserAgent)
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Client-Language", "en")
	req.Header.Set("Content-Type", "application/json")
}

// captureResponseHeaders rotates the ct0 CSRF token from Set-Cookie headers.
func (c *Client) captureResponseHeaders(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sc := range resp.Header["Set-Cookie"] {
		for _, part := range strings.Split(sc, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "ct0=") {
				newCSRF := strings.TrimPrefix(part, "ct0=")
				if newCSRF != "" {
					c.csrf = newCSRF
				}
			}
		}
	}
}

// GraphQL performs a GET request to an X GraphQL endpoint.
// variables and features are JSON-encoded and passed as URL query params.
// Returns the "data" field from the response JSON.
func (c *Client) GraphQL(ctx context.Context, queryHash, operationName string, variables map[string]any, features map[string]bool) (json.RawMessage, error) {
	varsJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql variables: %w", err)
	}
	featJSON, err := json.Marshal(features)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql features: %w", err)
	}

	params := url.Values{}
	params.Set("variables", string(varsJSON))
	params.Set("features", string(featJSON))

	path := fmt.Sprintf("/i/api/graphql/%s/%s?%s", queryHash, operationName, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build GraphQL request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL %s: %w", operationName, err)
	}
	c.captureResponseHeaders(resp)

	return c.extractData(resp, operationName)
}

// GraphQLPost performs a POST request to an X GraphQL endpoint.
// Used for mutations (create, delete, etc.).
// Returns the "data" field from the response JSON.
func (c *Client) GraphQLPost(ctx context.Context, queryHash, operationName string, variables map[string]any, features map[string]bool) (json.RawMessage, error) {
	body := map[string]any{
		"variables": variables,
		"features":  features,
		"queryId":   queryHash,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal GraphQL POST body: %w", err)
	}

	path := fmt.Sprintf("/i/api/graphql/%s/%s", queryHash, operationName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build GraphQL POST request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL POST %s: %w", operationName, err)
	}
	c.captureResponseHeaders(resp)

	return c.extractData(resp, operationName)
}

// Get performs a GET request to a v1.1 API path.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// Post performs a POST request with a form-encoded body.
func (c *Client) Post(ctx context.Context, path string, body url.Values) (*http.Response, error) {
	encoded := ""
	if body != nil {
		encoded = body.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("build POST request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// PostJSON performs a POST request with a JSON-encoded body.
func (c *Client) PostJSON(ctx context.Context, path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal POST body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build POST JSON request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST JSON %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build DELETE request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DELETE %s: %w", path, err)
	}
	c.captureResponseHeaders(resp)
	return resp, nil
}

// DecodeJSON reads resp.Body into target, first checking for error status codes.
func (c *Client) DecodeJSON(resp *http.Response, target any) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return parseRateLimitError(resp)
	}
	if resp.StatusCode >= 400 {
		return handleXError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// extractData reads the response body and returns the "data" field.
func (c *Client) extractData(resp *http.Response, operationName string) (json.RawMessage, error) {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, parseRateLimitError(resp)
	}
	if resp.StatusCode >= 400 {
		return nil, handleXError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read GraphQL response for %s: %w", operationName, err)
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode GraphQL envelope for %s: %w", operationName, err)
	}

	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("x GraphQL error in %s: %s", operationName, envelope.Errors[0].Message)
	}

	return envelope.Data, nil
}

// handleXError reads an error response body and returns a typed error.
func handleXError(resp *http.Response) error {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("x API error (http=%d): could not read response body: %w", resp.StatusCode, readErr)
	}

	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"errors"`
	}
	if jsonErr := json.Unmarshal(body, &envelope); jsonErr == nil && len(envelope.Errors) > 0 {
		msg := envelope.Errors[0].Message
		code := envelope.Errors[0].Code
		// Code 326 = account locked
		if code == 326 {
			return &AccountLockedError{Message: msg}
		}
		return fmt.Errorf("x API error (code=%d, http=%d): %s", code, resp.StatusCode, msg)
	}

	return fmt.Errorf("x API error (http=%d): no details available", resp.StatusCode)
}

// parseRateLimitError builds a RateLimitError from a 429 response.
func parseRateLimitError(resp *http.Response) *RateLimitError {
	io.Copy(io.Discard, resp.Body) //nolint:errcheck
	resp.Body.Close()

	e := &RateLimitError{}
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil {
			e.RetryAfter = time.Duration(secs) * time.Second
		}
	}
	return e
}
