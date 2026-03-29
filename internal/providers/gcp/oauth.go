package gcp

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
		Short: "List IAM Workforce OAuth clients for a project",
		RunE:  makeRunOAuthList(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	return cmd
}

func makeRunOAuthList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients", iamBaseURL, project)
		var resp struct {
			OAuthClients []map[string]any `json:"oauthClients"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing OAuth clients: %w", err)
		}

		summaries := make([]OAuthClientSummary, 0, len(resp.OAuthClients))
		for _, c := range resp.OAuthClients {
			summaries = append(summaries, toOAuthClientSummary(c))
		}

		return printOAuthClientSummaries(cmd, summaries)
	}
}

func newOAuthCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an IAM Workforce OAuth client",
		RunE:  makeRunOAuthCreate(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("client-id", "", "OAuth client ID (required, e.g. my-app-client)")
	cmd.Flags().String("display-name", "", "Display name for the OAuth client")
	cmd.Flags().StringSlice("redirect-uris", []string{
		"http://localhost:3000/callback",
	}, "Allowed redirect URIs (comma-separated)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		clientID, _ := cmd.Flags().GetString("client-id")
		displayName, _ := cmd.Flags().GetString("display-name")
		redirectURIs, _ := cmd.Flags().GetStringSlice("redirect-uris")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if displayName == "" {
			displayName = clientID
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create OAuth client %q on project %q", clientID, project), map[string]any{
				"action":               "create",
				"project":              project,
				"oauthClientId":        clientID,
				"displayName":          displayName,
				"allowedRedirectUris":  redirectURIs,
			})
		}

		body := map[string]any{
			"displayName":         displayName,
			"allowedRedirectUris": redirectURIs,
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients?oauthClientId=%s",
			iamBaseURL, project, clientID)
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, url, body, &data); err != nil {
			return fmt.Errorf("creating OAuth client %q: %w", clientID, err)
		}

		summary := toOAuthClientSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}
		fmt.Printf("Created OAuth client: %s\n", summary.Name)
		return nil
	}
}

func newOAuthUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update allowed redirect URIs on an OAuth client",
		RunE:  makeRunOAuthUpdate(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	cmd.Flags().StringSlice("redirect-uris", nil, "New allowed redirect URIs (comma-separated, required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("redirect-uris")
	return cmd
}

func makeRunOAuthUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		clientID, _ := cmd.Flags().GetString("client-id")
		redirectURIs, _ := cmd.Flags().GetStringSlice("redirect-uris")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update redirect URIs for OAuth client %q", clientID), map[string]any{
				"action":              "update",
				"project":             project,
				"clientId":            clientID,
				"allowedRedirectUris": redirectURIs,
			})
		}

		body := map[string]any{
			"allowedRedirectUris": redirectURIs,
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients/%s?updateMask=allowedRedirectUris",
			iamBaseURL, project, clientID)
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPatch, url, body, &data); err != nil {
			return fmt.Errorf("updating OAuth client %q: %w", clientID, err)
		}

		summary := toOAuthClientSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}
		fmt.Printf("Updated OAuth client: %s\n  Redirect URIs: %s\n",
			summary.Name, strings.Join(summary.RedirectURIs, ", "))
		return nil
	}
}

func newOAuthDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an IAM Workforce OAuth client (irreversible)",
		RunE:  makeRunOAuthDelete(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		clientID, _ := cmd.Flags().GetString("client-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete OAuth client %q from project %q", clientID, project), map[string]any{
				"action":   "delete",
				"project":  project,
				"clientId": clientID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients/%s",
			iamBaseURL, project, clientID)
		if _, err := client.do(ctx, http.MethodDelete, url, nil); err != nil {
			return fmt.Errorf("deleting OAuth client %q: %w", clientID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "clientId": clientID})
		}
		fmt.Printf("Deleted OAuth client: %s\n", clientID)
		return nil
	}
}

func newOAuthCreateCredentialsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-credentials",
		Short: "Create a client secret for an OAuth client (returns client_id and client_secret)",
		RunE:  makeRunOAuthCreateCredentials(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthCreateCredentials(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		clientID, _ := cmd.Flags().GetString("client-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create credentials for OAuth client %q", clientID), map[string]any{
				"action":   "create-credentials",
				"project":  project,
				"clientId": clientID,
			})
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients/%s/credentials",
			iamBaseURL, project, clientID)
		var data map[string]any
		// POST with empty body creates a new credential.
		if err := client.doJSON(ctx, http.MethodPost, url, map[string]any{}, &data); err != nil {
			return fmt.Errorf("creating credentials for OAuth client %q: %w", clientID, err)
		}

		cred := toOAuthCredential(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(cred)
		}
		lines := []string{
			fmt.Sprintf("Name:          %s", cred.Name),
			fmt.Sprintf("Client ID:     %s", cred.ClientID),
			fmt.Sprintf("Client Secret: %s", cred.ClientSecret),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newOAuthListCredentialsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-credentials",
		Short: "List credentials for an OAuth client",
		RunE:  makeRunOAuthListCredentials(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("client-id", "", "OAuth client ID (required)")
	_ = cmd.MarkFlagRequired("client-id")
	return cmd
}

func makeRunOAuthListCredentials(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		clientID, _ := cmd.Flags().GetString("client-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/locations/global/oauthClients/%s/credentials",
			iamBaseURL, project, clientID)
		var resp struct {
			OAuthClientCredentials []map[string]any `json:"oauthClientCredentials"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing credentials for OAuth client %q: %w", clientID, err)
		}

		creds := make([]OAuthCredential, 0, len(resp.OAuthClientCredentials))
		for _, c := range resp.OAuthClientCredentials {
			creds = append(creds, toOAuthCredential(c))
		}

		return printOAuthCredentials(cmd, creds)
	}
}
