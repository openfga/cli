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
)

var versionStr = fmt.Sprintf("v`%s` (commit: `%s`, date: `%s`)", build.Version, build.Commit, build.Date)

// versionCmd is the entrypoint for the `fga version“ command.
var versionCmd *cobra.Command = &cobra.Command{
	Use:   "version",
	Short: "Reports the FGA CLI version",
	Long:  "Reports the FGA CLI version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("fga version %s\n", versionStr)

		return nil
	},
	Args: cobra.NoArgs,
}
