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
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
	"github.com/openfga/cli/internal/output"
	"github.com/openfga/cli/internal/storetest"
)

var (
	errNoTestFilesSpecified = errors.New("no test files specified")
	errNoTestFilesFound     = errors.New("no test files found")
	errTestFileDoesNotExist = errors.New("test file does not exist")
)

// modelTestCmd represents the test command.
var modelTestCmd = &cobra.Command{
	Use:     "test",
	Short:   "Test an Authorization Model",
	Long:    "Run a set of tests against a particular Authorization Model.",
	Example: `fga model test --tests model.fga.yaml`,
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read and validate all flags
		testsFilePatterns, err := cmd.Flags().GetStringArray("tests")
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

		// Expand test file patterns (handles both glob patterns and shell-expanded files)
		fileNames, err := expandTestFilePatterns(testsFilePatterns, args)
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

// expandTestFilePatterns takes test file patterns (which can be literal file paths or glob patterns)
// and positional arguments (from shell expansion), and returns a list of resolved file paths.
// It handles both quoted glob patterns (where the CLI does the expansion) and shell-expanded
// arguments (where the shell expands the glob before passing to the CLI).
func expandTestFilePatterns(patterns []string, posArgs []string) ([]string, error) {
	// Combine flag values and positional args
	// This handles shell expansion: when the shell expands ./example/*.fga.yaml to
	// ./example/file1.yaml ./example/file2.yaml, the first file goes to the --tests flag
	// and the rest end up as positional arguments
	allPatterns := make([]string, 0, len(patterns)+len(posArgs))
	allPatterns = append(allPatterns, patterns...)
	allPatterns = append(allPatterns, posArgs...)

	if len(allPatterns) == 0 {
		return nil, errNoTestFilesSpecified
	}

	fileNames := []string{}

	for _, pattern := range allPatterns {
		// First, check if it's a literal file that exists
		if _, err := os.Stat(pattern); err == nil {
			fileNames = append(fileNames, pattern)

			continue
		}

		// Otherwise, try to expand it as a glob pattern
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid tests pattern %s due to %w", pattern, err)
		}

		if len(matches) > 0 {
			fileNames = append(fileNames, matches...)
		} else {
			// If glob didn't match and file doesn't exist, report error
			return nil, fmt.Errorf("%w: %s", errTestFileDoesNotExist, pattern)
		}
	}

	if len(fileNames) == 0 {
		return nil, errNoTestFilesFound
	}

	return fileNames, nil
}

func init() {
	modelTestCmd.Flags().String("store-id", "", "Store ID")
	modelTestCmd.Flags().String("model-id", "", "Model ID")
	modelTestCmd.Flags().StringArray("tests", []string{},
		"Path or glob of YAML test files. Can be specified multiple times or use glob patterns")
	modelTestCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	modelTestCmd.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")

	if err := modelTestCmd.MarkFlagRequired("tests"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/test", err)
		os.Exit(1)
	}
}
