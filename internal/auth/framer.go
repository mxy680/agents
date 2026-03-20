package auth

import "fmt"

// FramerEnvConfig holds the environment variable names for Framer API key auth.
var FramerEnvConfig = struct {
	APIKey     string
	ProjectURL string
}{
	APIKey:     "FRAMER_API_KEY",
	ProjectURL: "FRAMER_PROJECT_URL",
}

// FramerCredentials holds the API key and project URL required to authenticate
// requests to the Framer API via the Node.js bridge.
type FramerCredentials struct {
	APIKey     string
	ProjectURL string
}

// NewFramerCredentials reads Framer credentials from environment variables.
// Required: FRAMER_API_KEY, FRAMER_PROJECT_URL.
func NewFramerCredentials() (*FramerCredentials, error) {
	apiKey, err := readEnv(FramerEnvConfig.APIKey)
	if err != nil {
		return nil, err
	}
	projectURL, err := readEnv(FramerEnvConfig.ProjectURL)
	if err != nil {
		return nil, err
	}
	return &FramerCredentials{
		APIKey:     apiKey,
		ProjectURL: projectURL,
	}, nil
}

// redactFramerCredentials returns a log-safe representation of the credentials.
func redactFramerCredentials(c *FramerCredentials) string {
	return fmt.Sprintf(
		"FramerCredentials{api_key=%s, project_url=%s}",
		redact(c.APIKey),
		redact(c.ProjectURL),
	)
}
