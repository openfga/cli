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

package model

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/flags"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
)

// testCmd represents the test command.
var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test an Authorization Model",
	Long:    "Run a set of tests against a particular Authorization Model.",
	Example: `fga model test --tests model.fga.yaml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		testsFileName, err := cmd.Flags().GetString("tests")
		if err != nil {
			return err //nolint:wrapcheck
		}

		format, storeData, err := storetest.ReadFromFile(testsFileName, path.Dir(testsFileName))
		if err != nil {
			return err //nolint:wrapcheck
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err //nolint:wrapcheck
		}

		suppressSummary, err := cmd.Flags().GetBool("suppress-summary")
		if err != nil {
			return err //nolint:wrapcheck
		}

		test, err := storetest.RunTests(
			fgaClient,
			storeData,
			format,
		)
		if err != nil {
			return fmt.Errorf("error running tests due to %w", err)
		}

		passing := test.IsPassing()

		if verbose {
			err = output.Display(test.Results)
			if err != nil {
				return fmt.Errorf("error displaying test results due to %w", err)
			}
		}

		if !suppressSummary {
			fmt.Fprintln(os.Stderr, test.FriendlyDisplay())
		}

		if !passing {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	testCmd.Flags().String("store-id", "", "Store ID")
	testCmd.Flags().String("model-id", "", "Model ID")
	testCmd.Flags().String("tests", "", "Tests file Name. The file should have the OpenFGA tests in a valid YAML or JSON format") //nolint:lll
	testCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	testCmd.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")

	if err := flags.SetFlagRequired(testCmd, "tests", "cmd/models/test", false); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
