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
	"strings"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// testCmd represents the test command.
var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test an Authorization Model",
	Long:    "Run a set of tests against a particular Authorization Model.",
	Example: `fga model test --tests tests.fga.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		testsFileName, err := cmd.Flags().GetString("tests")
		if err != nil {
			return err //nolint:wrapcheck
		}

		var storeData storetest.StoreData

		testFile, err := os.Open(testsFileName)
		if err != nil {
			return fmt.Errorf("failed to read file %s due to %w", testsFileName, err)
		}
		decoder := yaml.NewDecoder(testFile)
		decoder.KnownFields(true)
		err = decoder.Decode(&storeData)
		if err != nil {
			return fmt.Errorf("failed to unmarshal file %s due to %w", testsFileName, err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err //nolint:wrapcheck
		}

		results, err := storetest.RunTests(
			fgaClient,
			storeData,
			path.Dir(testsFileName),
		)
		if err != nil {
			return fmt.Errorf("error running tests due to %w", err)
		}

		if verbose {
			return output.Display(results) //nolint:wrapcheck
		}

		friendlyResults := []string{}

		for index := 0; index < len(results); index++ {
			friendlyResults = append(friendlyResults, results[index].FriendlyDisplay())
		}

		fmt.Printf("%v", strings.Join(friendlyResults, "\n---\n"))

		return nil
	},
}

func init() {
	testCmd.Flags().String("store-id", "", "Store ID")
	testCmd.Flags().String("model-id", "", "Model ID")
	testCmd.Flags().String("tests", "", "Tests file Name. The file should have the OpenFGA tests in a valid YAML or JSON format") //nolint:lll
	testCmd.Flags().Bool("verbose", false, "Print verbose JSON output")

	if err := testCmd.MarkFlagRequired("tests"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/test", err)
		os.Exit(1)
	}
}
