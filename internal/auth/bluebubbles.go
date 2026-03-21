package auth

import "fmt"

// BlueBubblesEnvConfig holds the environment variable names for BlueBubbles auth.
var BlueBubblesEnvConfig = struct {
	URL      string
	Password string
}{
	URL:      "BLUEBUBBLES_URL",
	Password: "BLUEBUBBLES_PASSWORD",
}

// BlueBubblesCredentials holds the server URL and password required to authenticate
// requests to the BlueBubbles REST API.
type BlueBubblesCredentials struct {
	URL      string
	Password string
}

// NewBlueBubblesCredentials reads BlueBubbles credentials from environment variables.
// Required: BLUEBUBBLES_URL, BLUEBUBBLES_PASSWORD.
func NewBlueBubblesCredentials() (*BlueBubblesCredentials, error) {
	serverURL, err := readEnv(BlueBubblesEnvConfig.URL)
	if err != nil {
		return nil, err
	}
	password, err := readEnv(BlueBubblesEnvConfig.Password)
	if err != nil {
		return nil, err
	}
	return &BlueBubblesCredentials{
		URL:      serverURL,
		Password: password,
	}, nil
}

// redactBlueBubblesCredentials returns a log-safe representation of the credentials.
func redactBlueBubblesCredentials(c *BlueBubblesCredentials) string {
	return fmt.Sprintf(
		"BlueBubblesCredentials{url=%s, password=%s}",
		redact(c.URL),
		redact(c.Password),
	)
}
