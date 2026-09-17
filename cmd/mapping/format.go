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
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
)

// testRunResult holds the outcome of a full test run for formatting.
type testRunResult struct {
	results     []mapper.TestResult
	filtered    int
	skipped     int
	stopped     bool
	filterUsed  string
	mappingFile string
}

// colorSet holds ANSI color functions for terminal output.
type colorSet struct {
	pass  func(string) string
	fail  func(string) string
	err   func(string) string
	plus  func(string) string
	minus func(string) string
}

func newColorSet(enabled bool) colorSet {
	id := func(s string) string { return s }
	if !enabled {
		return colorSet{pass: id, fail: id, err: id, plus: id, minus: id}
	}

	wrap := func(code string) func(string) string {
		return func(s string) string { return code + s + "\033[0m" }
	}

	return colorSet{
		pass:  wrap("\033[32m"),
		fail:  wrap("\033[31m"),
		err:   wrap("\033[33m"),
		plus:  wrap("\033[32m"),
		minus: wrap("\033[31m"),
	}
}

// sameFile reports whether src and dst refer to the same file.
// Returns false if either path cannot be stat'd.
func sameFile(src, dst string) bool {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}

	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false
	}

	return os.SameFile(srcInfo, dstInfo)
}

// isTTY reports whether the given writer is a terminal.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice != 0
}

// colorEnabled determines whether color output should be used.
func colorEnabled(writer io.Writer, noColorFlag bool, toFile bool) bool {
	if noColorFlag || toFile {
		return false
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return isTTY(writer)
}

// missingTuples returns tuples present in expected but not in actual.
func missingTuples(expected, actual []language.Tuple) []language.Tuple {
	actualCount := make(map[string]int, len(actual))

	for _, tup := range actual {
		actualCount[tup.Key()]++
	}

	var missing []language.Tuple

	for _, tup := range expected {
		key := tup.Key()
		if actualCount[key] > 0 {
			actualCount[key]--
		} else {
			missing = append(missing, tup)
		}
	}

	return missing
}

// extraTuples returns tuples present in actual but not in expected.
func extraTuples(expected, actual []language.Tuple) []language.Tuple {
	expectedCount := make(map[string]int, len(expected))

	for _, tup := range expected {
		expectedCount[tup.Key()]++
	}

	var extra []language.Tuple

	for _, tup := range actual {
		key := tup.Key()
		if expectedCount[key] > 0 {
			expectedCount[key]--
		} else {
			extra = append(extra, tup)
		}
	}

	return extra
}

func tupleFilterKey(f language.TupleFilter) string {
	return f.User + "\x00" + f.Relation + "\x00" + f.Object + "\x00" + string(f.Action)
}

// missingTupleFilters returns filters present in expected but not in actual.
func missingTupleFilters(expected, actual []language.TupleFilter) []language.TupleFilter {
	actualCount := make(map[string]int, len(actual))

	for _, f := range actual {
		actualCount[tupleFilterKey(f)]++
	}

	var missing []language.TupleFilter

	for _, f := range expected {
		key := tupleFilterKey(f)
		if actualCount[key] > 0 {
			actualCount[key]--
		} else {
			missing = append(missing, f)
		}
	}

	return missing
}

// extraTupleFilters returns filters present in actual but not in expected.
func extraTupleFilters(expected, actual []language.TupleFilter) []language.TupleFilter {
	expectedCount := make(map[string]int, len(expected))

	for _, f := range expected {
		expectedCount[tupleFilterKey(f)]++
	}

	var extra []language.TupleFilter

	for _, f := range actual {
		key := tupleFilterKey(f)
		if expectedCount[key] > 0 {
			expectedCount[key]--
		} else {
			extra = append(extra, f)
		}
	}

	return extra
}

// sortTupleFilters returns a sorted copy of the given filters.
func sortTupleFilters(filters []language.TupleFilter) []language.TupleFilter {
	sorted := make([]language.TupleFilter, len(filters))
	copy(sorted, filters)
	sort.Slice(sorted, func(i, j int) bool {
		return tupleFilterKey(sorted[i]) < tupleFilterKey(sorted[j])
	})

	return sorted
}

// sortTuples returns a copy of tuples sorted by (user, relation, object, action, condition).
func sortTuples(tuples []language.Tuple) []language.Tuple {
	sorted := make([]language.Tuple, len(tuples))
	copy(sorted, tuples)
	sort.Slice(sorted, func(left, right int) bool {
		if sorted[left].User != sorted[right].User {
			return sorted[left].User < sorted[right].User
		}

		if sorted[left].Relation != sorted[right].Relation {
			return sorted[left].Relation < sorted[right].Relation
		}

		if sorted[left].Object != sorted[right].Object {
			return sorted[left].Object < sorted[right].Object
		}

		if sorted[left].Action != sorted[right].Action {
			return sorted[left].Action < sorted[right].Action
		}

		return sorted[left].Condition < sorted[right].Condition
	})

	return sorted
}

// errWriter wraps an io.Writer and captures the first write error.
// Once an error occurs, subsequent writes are no-ops.
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}

	n, err := ew.w.Write(p)
	ew.err = err

	return n, err //nolint:wrapcheck
}

func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintf(ew, format, args...)
}

func (ew *errWriter) println(s string) {
	if ew.err != nil {
		return
	}

	_, ew.err = fmt.Fprintln(ew, s)
}

// writeTupleBlock writes a set of tuples using tabwriter for column alignment.
func writeTupleBlock(dest *errWriter, tuples []language.Tuple, indent string) {
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

			fmt.Fprintf(tabw, "%s%s\t%s\t%s\t[%s]\t(%s%s)\n",
				indent, tup.User, tup.Relation, tup.Object, tup.Action, tup.Condition, ctx)
		} else {
			fmt.Fprintf(tabw, "%s%s\t%s\t%s\t[%s]\n",
				indent, tup.User, tup.Relation, tup.Object, tup.Action)
		}
	}

	if err := tabw.Flush(); err != nil && dest.err == nil {
		dest.err = err
	}
}

// countResults returns passed, failed, and error counts from results.
func countResults(results []mapper.TestResult) (int, int, int) {
	var passed, failed, errCount int

	for _, result := range results {
		switch {
		case result.Error != nil:
			errCount++
		case result.Passed:
			passed++
		default:
			failed++
		}
	}

	return passed, failed, errCount
}

// totalDuration sums the duration of all test results.
func totalDuration(results []mapper.TestResult) time.Duration {
	var total time.Duration

	for _, result := range results {
		total += result.Duration
	}

	return total
}

// formatDuration formats a duration as milliseconds, e.g. "8ms".
func formatDuration(dur time.Duration) string {
	return fmt.Sprintf("%dms", dur.Milliseconds())
}

// truncateInput marshals an event to compact JSON and truncates at 500 runes.
func truncateInput(input map[string]any) string {
	b, err := json.Marshal(input)
	if err != nil {
		return "(unrepresentable)"
	}

	runes := []rune(string(b))
	if len(runes) > 500 {
		return string(runes[:500]) + "..."
	}

	return string(runes)
}
