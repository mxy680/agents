package gcpconsole

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newOAuthListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OAuth clients for a project",
		RunE:  makeRunOAuthList(factory),
	}
	cmd.Flags().String("project-number", "", "GCP project number (required)")
	_ = cmd.MarkFlagRequired("project-number")
	return cmd
}

func makeRunOAuthList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		projectNumber, _ := cmd.Flags().GetString("project-number")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf(
			"/clients?projectNumber=%s&readMask=client_id,redirect_uris,type,creation_time,display_name,client_secrets&returnDisabledClients=true",
			projectNumber,
		)

		var resp ListClientsResponse
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing OAuth clients for project %q: %w", projectNumber, err)
		}

		return printOAuthClients(cmd, resp.Clients)
	}
}

func newOAuthGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of an OAuth client",
		RunE:  makeRunOAuthGet(factory),
	}
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		clientID, _ := cmd.Flags().GetString("client-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf(
			"/clients/%s?readMask=client_id,redirect_uris,type,creation_time,display_name,client_secrets",
			clientID,
		)

		var oauthClient OAuthClient
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &oauthClient); err != nil {
			return fmt.Errorf("getting OAuth client %q: %w", clientID, err)
		}

		return printOAuthClient(cmd, &oauthClient)
	}
}

func newOAuthCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new OAuth client",
		RunE:  makeRunOAuthCreate(factory),
	}
	cmd.Flags().String("project-number", "", "GCP project number (required)")
	cmd.Flags().String("name", "", "Display name for the OAuth client (required)")
	cmd.Flags().String("redirect-uris", "", "Comma-separated list of redirect URIs (required)")
	_ = cmd.MarkFlagRequired("project-number")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("redirect-uris")
	return cmd
}

func makeRunOAuthCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		projectNumber, _ := cmd.Flags().GetString("project-number")
		name, _ := cmd.Flags().GetString("name")
		redirectURIsFlag, _ := cmd.Flags().GetString("redirect-uris")

		redirectURIs := splitAndTrim(redirectURIsFlag)
		if len(redirectURIs) == 0 {
			return fmt.Errorf("at least one redirect URI is required")
		}

		if cli.IsDryRun(cmd) {
			return cli.PrintJSON(map[string]any{
				"action":        "create",
				"projectNumber": projectNumber,
				"displayName":   name,
				"redirectUris":  redirectURIs,
				"type":          "WEB",
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"type":          "WEB",
			"displayName":   name,
			"redirectUris":  redirectURIs,
			"authType":      "SHARED_SECRET",
			"brandId":       projectNumber,
			"projectNumber": projectNumber,
		}

		var oauthClient OAuthClient
		if err := client.doJSON(ctx, http.MethodPost, "/clients", body, &oauthClient); err != nil {
			return fmt.Errorf("creating OAuth client %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(oauthClient)
		}

		fmt.Printf("Created OAuth client: %s\n", oauthClient.ClientID)
		if len(oauthClient.ClientSecrets) > 0 {
			fmt.Printf("Client Secret: %s\n", oauthClient.ClientSecrets[0].ClientSecret)
		}
		return nil
	}
}

func newOAuthUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an OAuth client's redirect URIs",
		RunE:  makeRunOAuthUpdate(factory),
	}
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	cmd.Flags().String("add-redirect", "", "Redirect URI to add")
	cmd.Flags().String("remove-redirect", "", "Redirect URI to remove")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		clientID, _ := cmd.Flags().GetString("client-id")
		addRedirect, _ := cmd.Flags().GetString("add-redirect")
		removeRedirect, _ := cmd.Flags().GetString("remove-redirect")

		if addRedirect == "" && removeRedirect == "" {
			return fmt.Errorf("specify --add-redirect or --remove-redirect")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// Fetch current client state first
		getPath := fmt.Sprintf(
			"/clients/%s?readMask=client_id,redirect_uris,type,creation_time,display_name,client_secrets",
			clientID,
		)
		var existing OAuthClient
		if err := client.doJSON(ctx, http.MethodGet, getPath, nil, &existing); err != nil {
			return fmt.Errorf("fetching OAuth client %q: %w", clientID, err)
		}

		// Build updated redirect URI list
		updatedURIs := applyRedirectChanges(existing.RedirectURIs, addRedirect, removeRedirect)

		if cli.IsDryRun(cmd) {
			return cli.PrintJSON(map[string]any{
				"action":       "update",
				"clientId":     clientID,
				"redirectUris": updatedURIs,
			})
		}

		// Send full client object with updated redirect URIs
		existing.RedirectURIs = updatedURIs
		putPath := fmt.Sprintf("/clients/%s", clientID)
		var updated OAuthClient
		if err := client.doJSON(ctx, http.MethodPut, putPath, existing, &updated); err != nil {
			return fmt.Errorf("updating OAuth client %q: %w", clientID, err)
		}

		return printOAuthClient(cmd, &updated)
	}
}

func newOAuthDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an OAuth client (irreversible)",
		RunE:  makeRunOAuthDelete(factory),
	}
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		clientID, _ := cmd.Flags().GetString("client-id")

		if cli.IsDryRun(cmd) {
			return cli.PrintJSON(map[string]any{
				"action":   "delete",
				"clientId": clientID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/clients/%s", clientID), nil); err != nil {
			return fmt.Errorf("deleting OAuth client %q: %w", clientID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "clientId": clientID})
		}
		fmt.Printf("Deleted OAuth client: %s\n", clientID)
		return nil
	}
}

// splitAndTrim splits a comma-separated string and trims whitespace from each part.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// applyRedirectChanges adds/removes a redirect URI from the existing list.
func applyRedirectChanges(existing []string, add, remove string) []string {
	uriSet := make(map[string]struct{}, len(existing))
	for _, u := range existing {
		uriSet[u] = struct{}{}
	}

	if remove != "" {
		delete(uriSet, remove)
	}
	if add != "" {
		uriSet[add] = struct{}{}
	}

	result := make([]string, 0, len(uriSet))
	// Preserve original order, then append new
	for _, u := range existing {
		if u == remove {
			continue
		}
		result = append(result, u)
	}
	if add != "" {
		if _, exists := uriSet[add]; exists {
			alreadyIn := false
			for _, u := range existing {
				if u == add {
					alreadyIn = true
					break
				}
			}
			if !alreadyIn {
				result = append(result, add)
			}
		}
	}
	return result
}
