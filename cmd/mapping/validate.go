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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openfga/mapper"
	"github.com/spf13/cobra"
)

var (
	errMappingInvalid        = errors.New("mapping is invalid")
	errModelInconsistent     = errors.New("mapping is inconsistent with authorization model")
	errUnknownValidateFormat = errors.New("unknown format")
)

// collectErrorMessages recursively flattens a joined error into individual message strings.
func collectErrorMessages(err error) []string {
	type multiErr interface{ Unwrap() []error }

	if multi, ok := err.(multiErr); ok {
		msgs := make([]string, 0, len(multi.Unwrap()))

		for _, e := range multi.Unwrap() {
			msgs = append(msgs, collectErrorMessages(e)...)
		}

		return msgs
	}

	return []string{err.Error()}
}

type validateResult struct {
	Valid       bool               `json:"valid"`
	RuleCount   int                `json:"rule_count,omitempty"`
	Rules       []string           `json:"rules,omitempty"`
	Errors      mapper.Diagnostics `json:"errors,omitempty"`
	ModelErrors []string           `json:"model_errors,omitempty"`
}

func ruleNames(rules []mapper.RuleSummary) []string {
	names := make([]string, len(rules))

	for i, rule := range rules {
		names[i] = rule.Name
	}

	return names
}

// writeRuleStatuses writes per-rule ✓/✗ indicators to out.
// Error messages have the leading "rule "name": " prefix stripped to avoid redundancy.
func writeRuleStatuses(out io.Writer, results []ruleCheckResult) {
	for _, result := range results {
		if result.Err == nil {
			fmt.Fprintf(out, "  ✓ %s\n", result.Name)

			continue
		}

		prefix := fmt.Sprintf("rule %q: ", result.Name)
		lines := strings.Split(result.Err.Error(), "\n")
		stripped := make([]string, len(lines))

		for idx, line := range lines {
			stripped[idx] = strings.TrimPrefix(line, prefix)
		}

		fmt.Fprintf(out, "  ✗ %s: %s\n", result.Name, strings.Join(stripped, "; "))
	}
}

func validateMapping( //nolint:cyclop,gocognit
	path, format, modelFile string,
	verbose bool,
	out, errOut io.Writer,
) error {
	if format != "text" && format != "json" {
		return fmt.Errorf("%q: %w (must be text or json)", format, errUnknownValidateFormat)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	compiled, compileErr := mapper.Compile(data)
	if compileErr != nil {
		diags := mapper.DiagnosticsFrom(compileErr)
		if format == "json" {
			result := validateResult{Valid: false, Errors: diags}
			if encErr := json.NewEncoder(out).Encode(result); encErr != nil {
				return fmt.Errorf("encoding output: %w", encErr)
			}
		} else {
			fmt.Fprintln(errOut, diags)
		}

		return errMappingInvalid
	}

	if modelFile != "" { //nolint:nestif
		model, loadErr := loadAuthzModel(modelFile)
		if loadErr != nil {
			return fmt.Errorf("loading model: %w", loadErr)
		}

		ruleResults := validateRulesAgainstModel(compiled.Rules(), model)

		var modelErrs []error

		for _, r := range ruleResults {
			if r.Err != nil {
				modelErrs = append(modelErrs, r.Err)
			}
		}

		if len(modelErrs) > 0 {
			if format == "json" {
				result := validateResult{
					Valid:       false,
					RuleCount:   compiled.RuleCount(),
					ModelErrors: collectErrorMessages(errors.Join(modelErrs...)),
				}
				if encErr := json.NewEncoder(out).Encode(result); encErr != nil {
					return fmt.Errorf("encoding output: %w", encErr)
				}
			} else {
				if verbose {
					writeRuleStatuses(errOut, ruleResults)
				} else {
					fmt.Fprintln(errOut, errors.Join(modelErrs...))
				}
			}

			return errModelInconsistent
		}
	}

	if format == "json" {
		result := validateResult{Valid: true, RuleCount: compiled.RuleCount(), Rules: ruleNames(compiled.Rules())}
		if encErr := json.NewEncoder(out).Encode(result); encErr != nil {
			return fmt.Errorf("encoding output: %w", encErr)
		}

		return nil
	}

	if verbose {
		for _, rule := range compiled.Rules() {
			fmt.Fprintf(out, "  ✓ %s\n", rule.Name)
		}
	}

	fmt.Fprintf(out, "mapping is valid (%d rules)\n", compiled.RuleCount())

	return nil
}

var (
	validateFormat    string
	validateModelFile string
	validateVerbose   bool
)

var validateCmd = &cobra.Command{
	Use:   "validate <mapping-file>",
	Short: "Validate a mapping file",
	Long: `Validates that a mapping file is syntactically correct and all expressions compile.
With --model-file, also checks that every tuple template is consistent with the
authorization model: object types, relations, and user types must exist and be valid.`,
	Example: `  fga mapping validate mapping.yaml
  fga mapping validate --format json mapping.yaml
  fga mapping validate --model-file model.fga mapping.yaml`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := promptMappingFile(args, cmd.ErrOrStderr())

		err := validateMapping(
			path, validateFormat, validateModelFile, validateVerbose,
			cmd.OutOrStdout(), cmd.ErrOrStderr(),
		)
		if errors.Is(err, errUnknownValidateFormat) {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(2)
		}

		if errors.Is(err, errMappingInvalid) || errors.Is(err, errModelInconsistent) {
			os.Exit(2)
		}

		return err
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateFormat, "format", "text", `Output format: "text" or "json"`)
	validateCmd.Flags().StringVar(
		&validateModelFile, "model-file", "",
		"Path to FGA authorization model file (DSL, JSON, or modular)",
	)
	validateCmd.Flags().BoolVarP(
		&validateVerbose, "verbose", "v", false,
		"Show per-rule validation status (text format only)",
	)
}
