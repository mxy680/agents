package linkedin

import (
	"github.com/spf13/cobra"
)

// newFeedCmd builds the "feed" subcommand group.
func newFeedCmd(factory ClientFactory) *cobra.Command {
	feedCmd := &cobra.Command{
		Use:   "feed",
		Short: "Browse your LinkedIn feed",
	}
	feedCmd.AddCommand(newFeedListCmd(factory))
	feedCmd.AddCommand(newFeedHashtagCmd(factory))
	return feedCmd
}

// newFeedListCmd builds the "feed list" command.
func newFeedListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your LinkedIn home feed",
		Long:  "Retrieve posts from your LinkedIn homepage feed.",
		RunE:  makeRunFeedList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of feed items to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

// newFeedHashtagCmd builds the "feed hashtag" command.
func newFeedHashtagCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hashtag",
		Short: "Browse a hashtag feed",
		Long:  "Retrieve posts from a LinkedIn hashtag feed.",
		RunE:  makeRunFeedHashtag(factory),
	}
	cmd.Flags().String("tag", "", "Hashtag to browse (without the # prefix)")
	_ = cmd.MarkFlagRequired("tag")
	cmd.Flags().Int("limit", 10, "Maximum number of feed items to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

func makeRunFeedList(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunFeedHashtag(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}
