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
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
)

// formatText writes results in human-readable text format.
func formatText(out io.Writer, run testRunResult, verbose bool, color bool) error { //nolint:cyclop
	dest := &errWriter{w: out}
	colors := newColorSet(color)

	passed, failed, errCount := countResults(run.results)
	dur := totalDuration(run.results)

	if run.filterUsed != "" && len(run.results) == 0 {
		dest.printf("warning: no tests matched filter %q\n\n", run.filterUsed)
	}

	for _, result := range run.results {
		label := testLabel(result, colors)
		dest.printf("%s%s (%s)\n", label, result.Name, formatDuration(result.Duration))

		if verbose {
			writeVerboseTrace(dest, result)
		}
	}

	hasFailures := failed > 0 || errCount > 0

	if hasFailures {
		dest.println("")

		for _, result := range run.results {
			if result.Error != nil {
				dest.printf("--- %s: %s\n", colors.err("ERR"), result.Name)
				dest.printf("    %s\n", result.Error.Error())
				dest.println("")

				continue
			}

			if !result.Passed {
				writeFailureBlock(dest, result, colors, verbose)
			}
		}
	}

	if run.stopped && run.skipped > 0 {
		if !hasFailures {
			dest.println("")
		}

		dest.printf("stopped after first failure (%d tests not run)\n", run.skipped)
	}

	notRun := run.filtered + run.skipped

	dest.println("")
	writeSummaryLine(dest, passed, failed, errCount, notRun, dur)

	return dest.err
}

// testLabel returns the colored, pre-padded label for a test result.
// All labels are padded to 6 visible chars so test names align.
func testLabel(result mapper.TestResult, colors colorSet) string {
	switch {
	case result.Error != nil:
		return colors.err("ERR") + "   " // 3 + 3 = 6
	case result.Passed:
		return colors.pass("PASS") + "  " // 4 + 2 = 6
	default:
		return colors.fail("FAIL") + "  " // 4 + 2 = 6
	}
}

// writeFailureBlock writes the detailed diff for a single failing test.
func writeFailureBlock(dest *errWriter, result mapper.TestResult, colors colorSet, verbose bool) {
	missing := sortTuples(missingTuples(result.Expected, result.Actual))
	extra := sortTuples(extraTuples(result.Expected, result.Actual))

	dest.printf("--- %s: %s\n", colors.fail("FAIL"), result.Name)

	if verbose {
		dest.printf("    input:\n      %s\n\n", truncateInput(result.Input))
	}

	dest.printf("    expected %d tuples, got %d\n\n", len(result.Expected), len(result.Actual))

	dest.println("    missing (expected but not produced):")

	if len(missing) == 0 {
		dest.println("      (none)")
	} else {
		writeDiffTuples(dest, missing, colors.minus("- "), "      ")
	}

	dest.println("")

	dest.println("    extra (produced but not expected):")

	if len(extra) == 0 {
		dest.println("      (none)")
	} else {
		writeDiffTuples(dest, extra, colors.plus("+ "), "      ")
	}

	if len(result.ExpectedTupleFilters) > 0 || len(result.ActualTupleFilters) > 0 {
		missingFilters := sortTupleFilters(missingTupleFilters(result.ExpectedTupleFilters, result.ActualTupleFilters))
		extraFilters := sortTupleFilters(extraTupleFilters(result.ExpectedTupleFilters, result.ActualTupleFilters))

		dest.println("")

		nExpected, nActual := len(result.ExpectedTupleFilters), len(result.ActualTupleFilters)
		dest.printf("    expected %d tuple filters, got %d\n\n", nExpected, nActual)
		dest.println("    missing filters (expected but not produced):")

		if len(missingFilters) == 0 {
			dest.println("      (none)")
		} else {
			writeDiffTupleFilters(dest, missingFilters, colors.minus("- "), "      ")
		}

		dest.println("")
		dest.println("    extra filters (produced but not expected):")

		if len(extraFilters) == 0 {
			dest.println("      (none)")
		} else {
			writeDiffTupleFilters(dest, extraFilters, colors.plus("+ "), "      ")
		}
	}

	if verbose && len(result.Actual) > 0 {
		dest.println("")
		dest.println("    all actual tuples:")
		writeTupleBlock(dest, result.Actual, "      ")
	}

	dest.println("")
}

// writeVerboseTrace writes rule trace information for a single test result.
func writeVerboseTrace(dest *errWriter, result mapper.TestResult) {
	if result.Trace != nil {
		var matched, skipped []string

		for _, rt := range result.Trace.Rules {
			switch rt.Status {
			case mapper.RuleMatched:
				matched = append(matched, fmt.Sprintf("%s -> %d tuples", rt.Name, rt.EmittedN))
			case mapper.RuleSkipped:
				skipped = append(skipped, rt.Name)
			case mapper.RuleErrored:
			}
		}

		if len(matched) > 0 {
			dest.printf("        rules matched: %s\n", strings.Join(matched, ", "))
		}

		if len(skipped) > 0 {
			dest.printf("        rules skipped: %s (when guard false)\n", strings.Join(skipped, ", "))
		}
	}

	if result.Passed && len(result.Actual) > 0 {
		dest.println("        actual tuples:")
		writeTupleBlock(dest, result.Actual, "          ")
	}
}

// writeDiffTupleFilters writes tuple filters with a prefix marker, tab-aligned.
func writeDiffTupleFilters(dest *errWriter, filters []language.TupleFilter, prefix, indent string) {
	if dest.err != nil {
		return
	}

	tabw := tabwriter.NewWriter(dest, 0, 0, 2, ' ', 0)

	for _, filter := range filters {
		user := filter.User
		if user == "" {
			user = "*"
		}

		relation := filter.Relation
		if relation == "" {
			relation = "*"
		}

		object := filter.Object
		if object == "" {
			object = "*"
		}

		fmt.Fprintf(tabw, "%s%s%s\t%s\t%s\t[%s]\n", indent, prefix, user, relation, object, filter.Action)
	}

	if err := tabw.Flush(); err != nil && dest.err == nil {
		dest.err = err
	}
}

// writeDiffTuples writes tuples with a prefix marker, tab-aligned.
// Tuples with a condition are rendered with the condition name and optional context JSON.
func writeDiffTuples(dest *errWriter, tuples []language.Tuple, prefix, indent string) {
	if dest.err != nil {
		return
	}

	tabw := tabwriter.NewWriter(dest, 0, 0, 2, ' ', 0)

	for _, tup := range tuples {
		if tup.Condition != "" {
			ctx := ""

			if len(tup.Context) > 0 {
				if b, jsonErr := json.Marshal(tup.Context); jsonErr == nil {
					ctx = " | " + string(b)
				}
			}

			fmt.Fprintf(tabw, "%s%s%s\t%s\t%s\t[%s]\t(%s%s)\n",
				indent, prefix, tup.User, tup.Relation, tup.Object, tup.Action, tup.Condition, ctx)
		} else {
			fmt.Fprintf(tabw, "%s%s%s\t%s\t%s\t[%s]\n",
				indent, prefix, tup.User, tup.Relation, tup.Object, tup.Action)
		}
	}

	if err := tabw.Flush(); err != nil && dest.err == nil {
		dest.err = err
	}
}

// writeSummaryLine writes the final count summary.
func writeSummaryLine(dest *errWriter, passed, failed, errCount, notRun int, duration time.Duration) {
	parts := []string{
		fmt.Sprintf("%d passed", passed),
		fmt.Sprintf("%d failed", failed),
	}

	if errCount > 0 {
		noun := "error"
		if errCount > 1 {
			noun = "errors"
		}

		parts = append(parts, fmt.Sprintf("%d %s", errCount, noun))
	}

	if notRun > 0 {
		parts = append(parts, fmt.Sprintf("%d not run", notRun))
	}

	dest.printf("%s (%s)\n", strings.Join(parts, ", "), formatDuration(duration))
}
