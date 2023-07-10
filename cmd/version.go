package cmd

import (
	"log"

	"github.com/openfga/cli/internal/build"
	"github.com/spf13/cobra"
)

// versionCmd is the entrypoint for the `fga version“ command.
var versionCmd *cobra.Command = &cobra.Command{
	Use:   "version",
	Short: "Reports the FGA CLI version",
	Long:  "Reports the FGA CLI version.",
	RunE:  version,
	Args:  cobra.NoArgs,
}

func version(_ *cobra.Command, _ []string) error {
	log.Printf("version `%s` build from `%s` on `%s` ", build.Version, build.Commit, build.Date)

	return nil
}
