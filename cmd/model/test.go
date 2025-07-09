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
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

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
		clientConfig := cmdutils.GetClientConfig(cmd)
		fgaClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("failed to initialize FGA Client: %w", err)
		}

		testsPattern, err := cmd.Flags().GetString("tests")
		if err != nil {
			return fmt.Errorf("unable to read tests flag: %w", err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return fmt.Errorf("unable to read verbose flag: %w", err)
		}
		suppressSummary, err := cmd.Flags().GetBool("suppress-summary")
		if err != nil {
			return fmt.Errorf("unable to read suppress-summary flag: %w", err)
		}

		fileNames, err := filepath.Glob(testsPattern)
		if err != nil {
			return fmt.Errorf("invalid tests pattern %s: %w", testsPattern, err)
		}
		if len(fileNames) == 0 {
			if _, statErr := os.Stat(testsPattern); statErr != nil {
				return fmt.Errorf("no files match %s", testsPattern)
			}
			fileNames = []string{testsPattern}
		}

		logger := cmdutils.GetLogger(cmd)
		ctx := cmd.Context()
		multipleFiles := len(fileNames) > 1
		aggregateResults := storetest.TestResults{}
		summaries := make([]string, len(fileNames))

		var mu sync.Mutex
		g, gctx := errgroup.WithContext(ctx)
		sem := make(chan struct{}, 4)

		for idx, file := range fileNames {
			idx := idx
			file := file
			g.Go(func() error {
				select {
				case sem <- struct{}{}:
				case <-gctx.Done():
					return gctx.Err()
				}
				defer func() { <-sem }()

				format, storeData, err := storetest.ReadFromFile(file, path.Dir(file))
				if err != nil {
					return err
				}

				testRes, err := storetest.RunTests(gctx, fgaClient, storeData, format)
				if err != nil {
					return fmt.Errorf("error running tests for %s: %w", file, err)
				}

				mu.Lock()
				aggregateResults.Results = append(aggregateResults.Results, testRes.Results...)
				if !suppressSummary && multipleFiles {
					summaries[idx] = fmt.Sprintf("# %s\n%s", file, testRes.FriendlyBody())
				}
				mu.Unlock()

				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		if !suppressSummary {
			if multipleFiles {
				for _, s := range summaries {
					if s != "" {
						cmd.PrintErrln(s)
					}
				}
			}
			cmd.PrintErrln(aggregateResults.FriendlyDisplay())
		}

		if verbose {
			if err := output.Display(aggregateResults.Results); err != nil {
				return fmt.Errorf("displaying results: %w", err)
			}
		}

		if !aggregateResults.IsPassing() {
			return fmt.Errorf("tests failed")
		}

		logger.Printf("executed %d test files", len(fileNames))

		return nil
	},
}

func init() {
	modelTestCmd.Flags().String("tests", "", "path or glob of YAML/JSON test files")
	modelTestCmd.Flags().Bool("verbose", false, "Print verbose JSON output")
	modelTestCmd.Flags().Bool("suppress-summary", false, "Suppress the plain text summary output")

	if err := modelTestCmd.MarkFlagRequired("tests"); err != nil {
		fmt.Printf("error setting flag as required - %v: %v\n", "cmd/models/test", err)
		os.Exit(1)
	}
}
