package fly

import (
	"fmt"

	"github.com/spf13/cobra"
)

const regionsQuery = `query {
  platform {
    regions {
      name
      code
    }
  }
}`

func newRegionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available Fly.io regions",
		RunE:  makeRunRegionsList(factory),
	}
	return cmd
}

func makeRunRegionsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var result struct {
			Platform struct {
				Regions []Region `json:"regions"`
			} `json:"platform"`
		}
		if err := client.graphQL(ctx, regionsQuery, nil, &result); err != nil {
			return fmt.Errorf("listing regions: %w", err)
		}

		return printRegions(cmd, result.Platform.Regions)
	}
}
