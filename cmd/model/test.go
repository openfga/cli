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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/clierrors"
	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
)

// errAllTestFilesNonRegular is returned when a tests glob pattern matches only non-regular
// files (e.g. FIFOs), so there is nothing safe to read.
var errAllTestFilesNonRegular = errors.New("tests pattern matched only non-regular files (e.g. FIFOs); " +
	"pass an explicit regular file path instead")

// errNoTestFilesMatched is returned when a tests glob pattern matches no files at all.
var errNoTestFilesMatched = errors.New("no test files matched pattern")

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

		allowExternalFiles, err := cmd.Flags().GetBool("allow-external-files")
		if err != nil {
			return fmt.Errorf("failed to get allow-external-files flag: %w", err)
		}

		maxTypes, err := cmd.Flags().GetInt("max-types-per-authorization-model")
		if err != nil {
			return fmt.Errorf("failed to get max-types-per-authorization-model flag: %w", err)
		}

		if maxTypes <= 0 {
			return clierrors.ValidationError("model test",
				"max-types-per-authorization-model must be greater than 0")
		}

		serverConfig := storetest.LocalServerConfig{
			MaxTypesPerAuthorizationModel: maxTypes,
		}

		fileNames, err := resolveTestFiles(testsFileName)
		if err != nil {
			return err
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
			format, storeData, err := storetest.ReadFromFile(file, "", allowExternalFiles)
			if err != nil {
				return fmt.Errorf("failed to read test file %s: %w", file, err)
			}

			test, err := storetest.RunTests(
				cmd.Context(),
				fgaClient,
				storeData,
				format,
				serverConfig,
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

// resolveTestFiles turns testsPattern into the list of test files to read.
//
// A pattern with no glob metacharacters (*, ?, [) is treated as an explicit, literal path
// and honored as-is - including non-regular paths such as process substitution
// (--tests <(...)), which resolves to a FIFO like /dev/fd/11. The caller asked for that
// exact path, so it is not subject to the regular-file filter below.
//
// A pattern with glob metacharacters is expanded via filepath.Glob, and non-regular matches
// (FIFOs, devices, sockets) are dropped: a glob is not an explicit request for any single
// file, and reading a FIFO with no writer via os.ReadFile blocks forever. If the glob matches
// only non-regular files, that is an error rather than something to read.
//
// The metacharacter check (rather than, say, comparing the single glob match against the
// pattern) is deliberate: a FIFO literally named "*.fga.yaml" must still be filtered out, not
// mistaken for a literal path.
func resolveTestFiles(testsPattern string) ([]string, error) {
	if !strings.ContainsAny(testsPattern, `*?[`) {
		if _, statErr := os.Stat(testsPattern); statErr != nil {
			return nil, fmt.Errorf("test file %s does not exist: %w", testsPattern, statErr)
		}

		return []string{testsPattern}, nil
	}

	rawMatches, err := filepath.Glob(testsPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid tests pattern %s due to %w", testsPattern, err)
	}

	if len(rawMatches) == 0 {
		return nil, fmt.Errorf("%w: %s", errNoTestFilesMatched, testsPattern)
	}

	regularFileNames := rawMatches[:0]

	for _, name := range rawMatches {
		info, statErr := os.Stat(name)
		if statErr != nil {
			return nil, fmt.Errorf("failed to stat test file %s: %w", name, statErr)
		}

		if info.Mode().IsRegular() {
			regularFileNames = append(regularFileNames, name)
		}
	}

	if len(regularFileNames) == 0 {
		return nil, fmt.Errorf("%w: %s", errAllTestFilesNonRegular, testsPattern)
	}

	return regularFileNames, nil
}

func init() {
	modelTestCmd.Flags().String("store-id", "", "Store ID")
	modelTestCmd.Flags().String("model-id", "", "Model ID")
	modelTestCmd.Flags().String("tests", "", "Path or glob of YAML test files")
	modelTestCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	modelTestCmd.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")
	modelTestCmd.Flags().Int("max-types-per-authorization-model", 100, //nolint:mnd
		"Max allowed number of type definitions per authorization model")
	modelTestCmd.Flags().Bool("allow-external-files", false, "Allow model_file, tuple_file and tuple_files references in the test file to resolve to paths outside the test file's directory. Only enable this for test files you trust.") //nolint:lll

	if err := modelTestCmd.MarkFlagRequired("tests"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/test", err)
		os.Exit(1)
	}
}
