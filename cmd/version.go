package cmd

import (
	"fmt"

	"github.com/openfga/cli/internal/build"
	"github.com/spf13/cobra"
)

var versionStr = fmt.Sprintf("version `%s` build from `%s` on `%s` ", build.Version, build.Commit, build.Date)

// versionCmd is the entrypoint for the `fga version“ command.
var versionCmd *cobra.Command = &cobra.Command{
	Use:   "version",
	Short: "Reports the FGA CLI version",
	Long:  "Reports the FGA CLI version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(versionStr)

		return nil
	},
	Args: cobra.NoArgs,
}
