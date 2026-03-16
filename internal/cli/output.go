package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintText writes lines to stdout, one per line.
func PrintText(lines []string) {
	fmt.Println(strings.Join(lines, "\n"))
}

// PrintResult checks the --json flag and outputs accordingly.
func PrintResult(cmd *cobra.Command, v any, textLines []string) error {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		return PrintJSON(v)
	}
	PrintText(textLines)
	return nil
}

// IsJSONOutput returns whether the --json flag is set.
func IsJSONOutput(cmd *cobra.Command) bool {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	return jsonFlag
}

// IsDryRun returns whether the --dry-run flag is set.
func IsDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}
