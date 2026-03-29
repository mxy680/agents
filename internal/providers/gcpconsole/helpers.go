package gcpconsole

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// OAuthClient represents a GCP OAuth 2.0 client from the Console API.
type OAuthClient struct {
	ClientID      string         `json:"clientId"`
	ProjectNumber string         `json:"projectNumber"`
	BrandID       string         `json:"brandId"`
	DisplayName   string         `json:"displayName"`
	Type          string         `json:"type"`
	AuthType      string         `json:"authType"`
	RedirectURIs  []string       `json:"redirectUris"`
	ClientSecrets []ClientSecret `json:"clientSecrets,omitempty"`
	CreationTime  string         `json:"creationTime"`
	UpdateTime    string         `json:"updateTime"`
}

// ClientSecret represents a secret associated with an OAuth client.
type ClientSecret struct {
	ClientSecret string `json:"clientSecret"`
	CreateTime   string `json:"createTime"`
	State        string `json:"state"`
	ID           string `json:"id"`
}

// ListClientsResponse is the response from the list clients endpoint.
type ListClientsResponse struct {
	Clients []OAuthClient `json:"clients"`
}

// --- Print helpers ---

func printOAuthClient(cmd *cobra.Command, client *OAuthClient) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(client)
	}

	lines := []string{
		fmt.Sprintf("Client ID:      %s", client.ClientID),
		fmt.Sprintf("Display Name:   %s", client.DisplayName),
		fmt.Sprintf("Type:           %s", client.Type),
		fmt.Sprintf("Auth Type:      %s", client.AuthType),
		fmt.Sprintf("Project Number: %s", client.ProjectNumber),
		fmt.Sprintf("Brand ID:       %s", client.BrandID),
		fmt.Sprintf("Created:        %s", client.CreationTime),
		fmt.Sprintf("Updated:        %s", client.UpdateTime),
	}

	if len(client.RedirectURIs) > 0 {
		lines = append(lines, "Redirect URIs:")
		for _, uri := range client.RedirectURIs {
			lines = append(lines, "  - "+uri)
		}
	}

	if len(client.ClientSecrets) > 0 {
		lines = append(lines, "Client Secrets:")
		for _, s := range client.ClientSecrets {
			lines = append(lines, fmt.Sprintf("  - ID: %s  State: %s  Created: %s", s.ID, s.State, s.CreateTime))
		}
	}

	cli.PrintText(lines)
	return nil
}

func printOAuthClients(cmd *cobra.Command, clients []OAuthClient) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(clients)
	}

	if len(clients) == 0 {
		cli.PrintText([]string{"No OAuth clients found."})
		return nil
	}

	lines := make([]string, 0, len(clients)*3)
	for i, c := range clients {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines,
			fmt.Sprintf("Client ID:    %s", c.ClientID),
			fmt.Sprintf("Display Name: %s", c.DisplayName),
			fmt.Sprintf("Type:         %s", c.Type),
			fmt.Sprintf("Created:      %s", c.CreationTime),
		)
	}
	cli.PrintText(lines)
	return nil
}

// confirmDestructive returns an error if --confirm is not set.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this operation is irreversible; pass --confirm to proceed")
	}
	return nil
}
