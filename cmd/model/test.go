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

	"github.com/openfga/cli/internal/authorizationmodel"
	"github.com/openfga/cli/internal/cmdutils"
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

		var format authorizationmodel.ModelFormat
		var storeData *storetest.StoreData
		if testsFileName != "" {
			f, sd, err := storetest.ReadFromFile(testsFileName, path.Dir(testsFileName))
			if err != nil {
				return err //nolint:wrapcheck
			}
			format = f
			storeData = sd
		}

		features, err := cmd.Flags().GetString("features")
		if err != nil {
			return err //nolint:wrapcheck
		}

		reporter, err := cmd.Flags().GetString("reporter")
		if err != nil {
			return err //nolint:wrapcheck
		}

		status, err := storetest.RunCucumberTests(
			features,
			fgaClient,
			storeData,
			format,
			reporter,
		)
		if err != nil {
			return err //nolint:wrapcheck
		}

		os.Exit(status)

		return nil
	},
}

func init() {
	testCmd.Flags().String("store-id", "", "Store ID")
	testCmd.Flags().String("model-id", "", "Model ID")
	testCmd.Flags().String("tests", "", "Tests file Name. The file should have the OpenFGA tests in a valid YAML or JSON format") //nolint:lll
	testCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	testCmd.Flags().String("features", "", "Features directory.")
	testCmd.Flags().String("reporter", "simple", "Reporter to use for the tests")
}
