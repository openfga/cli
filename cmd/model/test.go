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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
)

// modelTestCmd represents the test command.
var modelTestCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test an Authorization Model",
	Long:    "Run a set of tests against a particular Authorization Model.",
	Example: `fga model test --tests model.fga.yaml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Read and validate all flags
		testsFileName, err := cmd.Flags().GetString("tests")
		if err != nil {
			return fmt.Errorf("failed to get tests flag: %w", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return fmt.Errorf("failed to get verbose flag: %w", err)
		}

		suppressSummary, err := cmd.Flags().GetBool("suppress-summary")
		if err != nil {
			return fmt.Errorf("failed to get suppress-summary flag: %w", err)
		}

		fileNames, err := filepath.Glob(testsFileName)
		if err != nil {
			return fmt.Errorf("invalid tests pattern %s due to %w", testsFileName, err)
		}
		if len(fileNames) == 0 {
			// Check if the literal path exists
			if _, err := os.Stat(testsFileName); err != nil {
				return fmt.Errorf("test file %s does not exist: %w", testsFileName, err)
			}
			fileNames = []string{testsFileName}
		}
		multipleFiles := len(fileNames) > 1

		clientConfig := cmdutils.GetClientConfig(cmd)

		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client due to %w", err)
		}

		aggregateResults := storetest.TestResults{}
		summaries := []string{}

		for _, file := range fileNames {
			format, storeData, err := storetest.ReadFromFile(file, path.Dir(file))
			if err != nil {
				return fmt.Errorf("failed to read test file %s: %w", file, err)
			}
			test, err := storetest.RunTests(
				cmd.Context(),
				fgaClient,
				storeData,
				format,
			)
			if err != nil {
				return fmt.Errorf("error running tests for %s due to %w", file, err)
			}

			aggregateResults.Results = append(aggregateResults.Results, test.Results...)

			if !suppressSummary && multipleFiles {
				fullDisplay := test.FriendlyDisplay()
				// Extract just the summary part (after "# Test Summary #")
				headerIndex := strings.Index(fullDisplay, "# Test Summary #")
				var summaryText string
				if headerIndex != -1 {
					// Get the summary part and remove the "# Test Summary #" header
					summaryPart := fullDisplay[headerIndex:]
					lines := strings.Split(summaryPart, "\n")
					if len(lines) > 1 {
						summaryText = strings.Join(lines[1:], "\n") // Skip the header line
					}
				} else {
					summaryText = fullDisplay
				}
				summary := fmt.Sprintf("# file: %s\n%s", file, summaryText)
				summaries = append(summaries, summary)
			}
		}

		passing := aggregateResults.IsPassing()

		if !suppressSummary {
			if multipleFiles {
				for _, summary := range summaries {
					fmt.Fprintln(os.Stderr, summary)
				}
			}
			fmt.Fprintln(os.Stderr, aggregateResults.FriendlyDisplay())
		}

		if verbose {
			err = output.Display(aggregateResults.Results)
			if err != nil {
				return fmt.Errorf("error displaying test results due to %w", err)
			}
		}

		if !passing {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	modelTestCmd.Flags().String("store-id", "", "Store ID")
	modelTestCmd.Flags().String("model-id", "", "Model ID")
	modelTestCmd.Flags().String("tests", "", "Path or glob of YAML test files")
	modelTestCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	modelTestCmd.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")

	if err := modelTestCmd.MarkFlagRequired("tests"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/test", err)
		os.Exit(1)
	}
}
