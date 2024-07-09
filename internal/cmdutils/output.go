package cmdutils

import "github.com/spf13/cobra"

func GetOutputConfig(cmd *cobra.Command) (bool, bool) {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	noPretty, _ := cmd.Flags().GetBool("no-pretty")

	return jsonOutput, noPretty
}
