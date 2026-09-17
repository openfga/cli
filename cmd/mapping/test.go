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

package mapping

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/openfga/mapper"
	"github.com/spf13/cobra"
)

var (
	errNoTestsDefined       = errors.New("no tests defined")
	errFilterMatchedNothing = errors.New("matched no tests")
	errTestsFailed          = errors.New("test(s) failed")
	errUnknownTestFormat    = errors.New("unknown format")
	errOutputSameAsInput    = errors.New("--output-file must not point to the mapping input file")
)

type runMappingTestsOptions struct {
	filter     string
	failFast   bool
	format     string
	outputFile string
	noColor    bool
	verbose    bool
}

func runMappingTests( //nolint:cyclop,gocognit
	ctx context.Context,
	path string,
	opts runMappingTestsOptions,
	out, errOut io.Writer,
) error {
	if opts.format == "" {
		opts.format = "text"
	}

	if opts.format != "text" && opts.format != "json" && opts.format != "junit" {
		return fmt.Errorf("%q: %w (must be text, json, or junit)", opts.format, errUnknownTestFormat)
	}

	// Create the output writer before compilation so --output-file is
	// honoured even when the mapping fails to compile.
	toFile := opts.outputFile != ""

	var target io.Writer

	if toFile {
		if sameFile(path, opts.outputFile) {
			return errOutputSameAsInput
		}

		outFile, createErr := os.Create(opts.outputFile)
		if createErr != nil {
			return fmt.Errorf("creating output file: %w", createErr)
		}
		defer outFile.Close()

		target = outFile
	} else {
		target = out
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	compiled, compileErr := mapper.Compile(data, mapper.WithTrace(opts.verbose))
	if compileErr != nil {
		switch opts.format {
		case "json":
			if encErr := formatJSONCompileError(target, compileErr); encErr != nil {
				return fmt.Errorf("encoding compile error: %w", encErr)
			}
		case "junit":
			if encErr := formatJUnitCompileError(target, compileErr); encErr != nil {
				return fmt.Errorf("encoding compile error: %w", encErr)
			}
		default:
			diags := mapper.DiagnosticsFrom(compileErr)
			fmt.Fprintln(errOut, diags)
		}

		return errMappingInvalid
	}

	if compiled.TestCount() == 0 {
		return fmt.Errorf("%w in %s", errNoTestsDefined, path)
	}

	run := compiled.RunTestsFiltered(ctx, opts.filter, opts.failFast)

	useColor := colorEnabled(out, opts.noColor, toFile)

	result := testRunResult{
		results:     run.Results,
		filtered:    run.Filtered,
		skipped:     run.Skipped,
		stopped:     run.Stopped(),
		filterUsed:  opts.filter,
		mappingFile: path,
	}

	if opts.filter != "" && run.Filtered > 0 && len(run.Results) == 0 {
		switch opts.format {
		case "json":
			if encErr := formatJSON(target, result); encErr != nil {
				return fmt.Errorf("writing output: %w", encErr)
			}
		case "junit":
			if encErr := formatJUnit(target, result); encErr != nil {
				return fmt.Errorf("writing output: %w", encErr)
			}
		default:
			if encErr := formatText(target, result, opts.verbose, useColor); encErr != nil {
				return fmt.Errorf("writing output: %w", encErr)
			}
		}

		return fmt.Errorf("filter %q: %w", opts.filter, errFilterMatchedNothing)
	}

	var fmtErr error

	switch opts.format {
	case "json":
		fmtErr = formatJSON(target, result)
	case "junit":
		fmtErr = formatJUnit(target, result)
	default:
		fmtErr = formatText(target, result, opts.verbose, useColor)
	}

	if fmtErr != nil {
		return fmt.Errorf("writing output: %w", fmtErr)
	}

	_, failed, errCount := countResults(result.results)
	if failed > 0 || errCount > 0 {
		return fmt.Errorf("%d %w", failed+errCount, errTestsFailed)
	}

	return nil
}

var (
	testRunFilter  string
	testFailFast   bool
	testFormat     string
	testOutputFile string
	testNoColor    bool
	testVerbose    bool
)

var testCmd = &cobra.Command{
	Use:   "test <mapping-file>",
	Short: "Run the embedded tests in a mapping file",
	Long: `Compiles the mapping and runs its embedded test cases, reporting pass/fail per case.
Exits 1 when any test fails, 2 when the mapping file cannot be compiled.
Use --format to choose between human-readable text (default), JSON, or JUnit XML output.`,
	Example: `  fga mapping test mapping.yaml
  fga mapping test --format json mapping.yaml
  fga mapping test --format junit --output-file results.xml mapping.yaml
  fga mapping test --run anne mapping.yaml
  fga mapping test --fail-fast mapping.yaml`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := promptMappingFile(args, cmd.ErrOrStderr())

		err := runMappingTests(cmd.Context(), path, runMappingTestsOptions{
			filter:     testRunFilter,
			failFast:   testFailFast,
			format:     testFormat,
			outputFile: testOutputFile,
			noColor:    testNoColor,
			verbose:    testVerbose,
		}, cmd.OutOrStdout(), cmd.ErrOrStderr())
		if errors.Is(err, errUnknownTestFormat) {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(2)
		}

		if errors.Is(err, errMappingInvalid) {
			os.Exit(2)
		}

		return err
	},
}

func init() {
	testCmd.Flags().StringVar(
		&testRunFilter, "run", "",
		"Run only tests whose name contains this substring (case-sensitive)",
	)
	testCmd.Flags().BoolVar(&testFailFast, "fail-fast", false, "Stop after the first failing test")
	testCmd.Flags().StringVar(&testFormat, "format", "text", `Output format: "text", "json", or "junit"`)
	testCmd.Flags().StringVarP(&testOutputFile, "output-file", "o", "", "Write output to a file instead of stdout")
	testCmd.Flags().BoolVar(&testNoColor, "no-color", false, "Disable color in text output")
	testCmd.Flags().BoolVarP(
		&testVerbose, "verbose", "v", false,
		"Show rule trace and tuple details for each test (text format only)",
	)
}
