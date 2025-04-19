/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/build"
	"github.com/openfga/cli/internal/cmdutils"
)

var versionStr = fmt.Sprintf("v`%s` (commit: `%s`, date: `%s`)", build.Version, build.Commit, build.Date)

var forceServerVersion bool

// versionCmd is the entrypoint for the `fga version` command.
var versionCmd *cobra.Command = &cobra.Command{
	Use:   "version",
	Short: "Reports the FGA CLI version",
	Long:  "Reports the FGA CLI version and OpenFGA server version if configured.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		fmt.Printf("fga CLI version %s\n", versionStr)

		// Try to get server version if configured or forced
		clientConfig := cmdutils.GetClientConfig(cmd)
		if clientConfig.ApiUrl != "" || forceServerVersion {
			if clientConfig.ApiUrl == "" {
				fmt.Println("Warning: No API URL configured. Use --force to check server version anyway.")

				return nil
			}

			fgaClient, err := clientConfig.GetFgaClient()
			if err != nil {
				fmt.Printf("Warning: Could not connect to OpenFGA server: %v\n", err)

				return nil
			}

			serverVersion, err := clientConfig.GetServerVersion(fgaClient)
			if err != nil {
				fmt.Printf("Warning: Could not get OpenFGA server version: %v\n", err)
				fmt.Printf("openfga version v`unknown`\n")

				return nil
			}

			fmt.Printf("openfga version v`%s`\n", serverVersion)
		}

		return nil
	},
	Args: cobra.NoArgs,
}

func init() {
	versionCmd.Flags().BoolVar(
		&forceServerVersion,
		"force",
		false,
		"Force checking server version even if API URL is not configured",
	)
	rootCmd.AddCommand(versionCmd)
}
