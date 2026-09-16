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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
	"github.com/spf13/cobra"
)

var (
	errUnknownRunFormat = errors.New("unknown format")
	errRecordsFailed    = errors.New("one or more input records failed")
)

type runMappingOptions struct {
	format          string
	writesOnly      bool
	aggregate       bool
	continueOnError bool
}

// opTupleOutput is the default JSONL shape: one operation per line, the tuple's
// action rendered as `op`. This is the full-fidelity shape (both writes and
// deletes) intended for a future action-aware apply.
type opTupleOutput struct {
	Op        string         `json:"op"`
	User      string         `json:"user"`
	Relation  string         `json:"relation"`
	Object    string         `json:"object"`
	Condition string         `json:"condition,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
}

// writeTupleOutput is the ClientTupleKey-compatible shape emitted under
// --writes-only, consumed verbatim by `fga tuple write --file`. The condition is
// an object {name, context} to match the OpenFGA Write API.
type writeTupleOutput struct {
	User      string           `json:"user"`
	Relation  string           `json:"relation"`
	Object    string           `json:"object"`
	Condition *conditionOutput `json:"condition,omitempty"`
}

type conditionOutput struct {
	Name    string         `json:"name"`
	Context map[string]any `json:"context,omitempty"`
}

// batchOutput is the --format json shape without --writes-only: writes and
// deletes split by action, plus filters that cannot be expanded offline.
type batchOutput struct {
	Writes            []writeTupleOutput     `json:"writes"`
	Deletes           []writeTupleOutput     `json:"deletes"`
	UnresolvedFilters []language.TupleFilter `json:"unresolved_filters"`
}

func toOpTuple(tuple language.Tuple) opTupleOutput {
	opValue := string(tuple.Action)
	if opValue == "" {
		opValue = string(language.ActionWrite)
	}

	return opTupleOutput{
		Op:        opValue,
		User:      tuple.User,
		Relation:  tuple.Relation,
		Object:    tuple.Object,
		Condition: tuple.Condition,
		Context:   tuple.Context,
	}
}

func toWriteTuple(tuple language.Tuple) writeTupleOutput {
	w := writeTupleOutput{User: tuple.User, Relation: tuple.Relation, Object: tuple.Object}
	if tuple.Condition != "" {
		w.Condition = &conditionOutput{Name: tuple.Condition, Context: tuple.Context}
	}

	return w
}

func isWrite(tuple language.Tuple) bool {
	return tuple.Action == "" || tuple.Action == language.ActionWrite
}

func flattenFilters(ops []mapper.TupleFilterOperation) []language.TupleFilter {
	filters := make([]language.TupleFilter, 0)
	for _, op := range ops {
		filters = append(filters, op.Filters...)
	}

	return filters
}

// dedupFilters removes duplicate unresolved filters by full identity, preserving
// order. Used in the --aggregate path so a filter produced identically across
// records collapses to one, mirroring mapper.Compact's tuple dedup.
func dedupFilters(filters []language.TupleFilter) []language.TupleFilter {
	if len(filters) <= 1 {
		return filters
	}

	seen := make(map[language.TupleFilter]struct{}, len(filters))
	deduped := make([]language.TupleFilter, 0, len(filters))

	for _, filter := range filters {
		if _, ok := seen[filter]; ok {
			continue
		}

		seen[filter] = struct{}{}

		deduped = append(deduped, filter)
	}

	return deduped
}

// emitJSONL writes one JSON object per line. With writesOnly, only write-action
// tuples are emitted in the bare ClientTupleKey shape; otherwise every tuple is
// emitted with its `op`.
func emitJSONL(tuples []language.Tuple, writesOnly bool, out io.Writer) error {
	enc := json.NewEncoder(out)

	for _, tuple := range tuples {
		var value any

		if writesOnly {
			if !isWrite(tuple) {
				continue
			}

			value = toWriteTuple(tuple)
		} else {
			value = toOpTuple(tuple)
		}

		if err := enc.Encode(value); err != nil {
			return fmt.Errorf("encoding output: %w", err)
		}
	}

	return nil
}

// emitJSON writes a single JSON document. With writesOnly it is a flat array of
// bare write tuples (consumable by `fga tuple write --file`); otherwise a batch
// object splitting writes/deletes and listing unresolved filters.
func emitJSON(tuples []language.Tuple, filters []language.TupleFilter, writesOnly bool, out io.Writer) error {
	enc := json.NewEncoder(out)

	if writesOnly {
		writes := make([]writeTupleOutput, 0, len(tuples))

		for _, tuple := range tuples {
			if isWrite(tuple) {
				writes = append(writes, toWriteTuple(tuple))
			}
		}

		if err := enc.Encode(writes); err != nil {
			return fmt.Errorf("encoding output: %w", err)
		}

		return nil
	}

	batch := batchOutput{
		Writes:            make([]writeTupleOutput, 0, len(tuples)),
		Deletes:           make([]writeTupleOutput, 0),
		UnresolvedFilters: filters,
	}

	for _, tuple := range tuples {
		if isWrite(tuple) {
			batch.Writes = append(batch.Writes, toWriteTuple(tuple))
		} else {
			batch.Deletes = append(batch.Deletes, toWriteTuple(tuple))
		}
	}

	if err := enc.Encode(batch); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}

	return nil
}

// warnUnresolvedFilters reports, one per line on stderr, each delete-by-filter
// operation that cannot be expanded without a store.
func warnUnresolvedFilters(filters []language.TupleFilter, errOut io.Writer) {
	for _, f := range filters {
		fmt.Fprintf(errOut,
			"WARN unresolved tuple filter (needs a store to expand): action=%s user=%s relation=%s object=%s\n",
			f.Action, f.User, f.Relation, f.Object)
	}
}

func runMapping(
	ctx context.Context,
	path string,
	opts runMappingOptions,
	inputReader io.Reader,
	out, errOut io.Writer,
) error {
	if opts.format != "jsonl" && opts.format != "json" {
		return fmt.Errorf("%q: %w (must be jsonl or json)", opts.format, errUnknownRunFormat)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	compiled, compileErr := mapper.Compile(data)
	if compileErr != nil {
		fmt.Fprintln(errOut, mapper.DiagnosticsFrom(compileErr))

		return errMappingInvalid
	}

	buffered, filters, hadFailure, err := evalRecords(ctx, compiled, inputReader, opts, out, errOut)
	if err != nil {
		return err
	}

	// Streaming (default jsonl, no aggregation) has already emitted per record;
	// only the accumulated filter warnings remain.
	if isStreaming(opts) {
		warnUnresolvedFilters(filters, errOut)
	} else if err := emitBuffered(buffered, filters, opts, out, errOut); err != nil {
		return err
	}

	// With --continue-on-error a skipped record is not fatal mid-run, but the
	// command still exits non-zero so callers can detect partial failure.
	if hadFailure {
		return errRecordsFailed
	}

	return nil
}

// isStreaming reports whether output is emitted per record as it is evaluated.
// Only the default JSONL path streams; --format json and --aggregate both need
// the whole run buffered before emitting.
func isStreaming(opts runMappingOptions) bool {
	return opts.format == "jsonl" && !opts.aggregate
}

// evalRecords reads the JSONL input stream (one JSON object per line) and
// evaluates each record. Streaming output is written to out as each record is
// evaluated; otherwise the tuples are buffered and returned. Unresolved filters
// are accumulated across all records. With --continue-on-error a malformed or
// evaluation-failing record is warned to errOut and skipped, and the boolean
// return reports whether any record was skipped; without it the first failure
// is returned as a fatal error.
func evalRecords(
	ctx context.Context,
	compiled *mapper.Mapping,
	inputReader io.Reader,
	opts runMappingOptions,
	out, errOut io.Writer,
) ([]language.Tuple, []language.TupleFilter, bool, error) {
	buffered := make([]language.Tuple, 0)
	filters := make([]language.TupleFilter, 0)
	reader := bufio.NewReader(inputReader)
	lineNum := 0
	hadFailure := false

	for {
		line, readErr := reader.ReadString('\n')

		if line != "" {
			lineNum++

			tuples, recFilters, skipped, fatal := processLine(
				ctx, compiled, strings.TrimSpace(line), lineNum, opts, out, errOut)
			if fatal != nil {
				return nil, nil, false, fatal
			}

			hadFailure = hadFailure || skipped

			filters = append(filters, recFilters...)
			buffered = append(buffered, tuples...)
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return nil, nil, false, fmt.Errorf("reading input: %w", readErr)
		}
	}

	return buffered, filters, hadFailure, nil
}

// processLine evaluates one input line and either streams its output or returns
// its tuples for buffering. A blank line is a no-op. On a record failure it
// warns and reports skipped=true when --continue-on-error is set; otherwise the
// failure is returned as fatal. An output-write failure is always fatal — it is
// not a per-record data problem and would recur on every line.
func processLine(
	ctx context.Context,
	compiled *mapper.Mapping,
	trimmed string,
	lineNum int,
	opts runMappingOptions,
	out, errOut io.Writer,
) ([]language.Tuple, []language.TupleFilter, bool, error) {
	if trimmed == "" {
		return nil, nil, false, nil
	}

	recTuples, recFilters, recErr := evalRecord(ctx, compiled, trimmed)
	if recErr != nil {
		if !opts.continueOnError {
			return nil, nil, false, recErr
		}

		fmt.Fprintf(errOut, "WARN skipping record on line %d: %v\n", lineNum, recErr)

		return nil, nil, true, nil
	}

	if isStreaming(opts) {
		if err := emitJSONL(recTuples, opts.writesOnly, out); err != nil {
			return nil, nil, false, err
		}

		return nil, recFilters, false, nil
	}

	return recTuples, recFilters, false, nil
}

// evalRecord parses one JSONL line into an event and evaluates it against the
// mapping, returning the produced tuples and unresolved filters. It performs no
// output; the caller decides how to emit. A malformed line or an evaluation
// error is returned as err.
func evalRecord(
	ctx context.Context,
	compiled *mapper.Mapping,
	line string,
) ([]language.Tuple, []language.TupleFilter, error) {
	var event map[string]any

	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, nil, fmt.Errorf("reading input JSON: %w", err)
	}

	result, err := compiled.Evaluate(ctx, event)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating mapping: %w", err)
	}

	return result.Tuples, flattenFilters(result.TupleFilterOperations), nil
}

// emitBuffered emits a whole-run result as a single document. With --aggregate
// the buffered tuples are collapsed (cross-record dedup and write/delete
// conflict detection) before emitting; a conflict is a runtime error.
func emitBuffered(
	tuples []language.Tuple,
	filters []language.TupleFilter,
	opts runMappingOptions,
	out, errOut io.Writer,
) error {
	if opts.aggregate {
		compacted, err := mapper.Compact(tuples)
		if err != nil {
			return fmt.Errorf("aggregating tuples: %w", err)
		}

		tuples = compacted
		filters = dedupFilters(filters)
	}

	if opts.format == "json" {
		if err := emitJSON(tuples, filters, opts.writesOnly, out); err != nil {
			return err
		}

		// The batch object carries unresolved_filters; the writes-only flat array does not.
		if opts.writesOnly {
			warnUnresolvedFilters(filters, errOut)
		}

		return nil
	}

	// jsonl with --aggregate: collapsed tuples emitted as lines after buffering.
	if err := emitJSONL(tuples, opts.writesOnly, out); err != nil {
		return err
	}

	warnUnresolvedFilters(filters, errOut)

	return nil
}

var (
	runFormat          string
	runWritesOnly      bool
	runInputFile       string
	runAggregate       bool
	runContinueOnError bool
)

var runCmd = &cobra.Command{
	Use:   "run <mapping-file>",
	Short: "Evaluate a mapping against JSON input and emit tuple operations",
	Long: `Reads JSONL from stdin (or --input) and evaluates it against the mapping file.
Input is JSON Lines: one JSON object per line. Outputs tuple operations as JSONL (default)
or a JSON batch (--format json).
Runs entirely offline — no store reads, no credentials, no network.
Rules using tuple_filters cannot be expanded offline and are reported as warnings on stderr
(or under unresolved_filters in the --format json batch).

Default JSONL emits one operation per line with an "op" field (write or delete).
With --writes-only only write operations are emitted, in the bare ClientTupleKey shape
consumable directly by fga tuple write --file (JSONL lines, or a JSON array with --format json).

Each input record is evaluated and streamed in order. --format json instead collects the whole
run into a single document. --aggregate buffers all records and collapses them (cross-record
dedup of tuples and unresolved filters, plus write/delete conflict detection) before emitting;
a conflict is a runtime error. Both --format json and --aggregate buffer the whole run in memory
before emitting; the default streaming JSONL does not.
--continue-on-error skips any record that fails to parse or evaluate (warning to stderr) and
continues; the command still exits non-zero if any record was skipped.`,
	Example: `  echo '{"id":"anne","org":"acme"}' | fga mapping run mapping.yaml
  fga mapping run mapping.yaml --input event.json --format json
  fga mapping run --writes-only mapping.yaml > out.jsonl && fga tuple write --store-id $STORE_ID --file out.jsonl`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputReader := cmd.InOrStdin()

		if runInputFile != "" {
			file, err := os.Open(runInputFile)
			if err != nil {
				return fmt.Errorf("opening input %s: %w", runInputFile, err)
			}

			defer file.Close()

			inputReader = file
		}

		err := runMapping(
			cmd.Context(), args[0],
			runMappingOptions{
				format:          runFormat,
				writesOnly:      runWritesOnly,
				aggregate:       runAggregate,
				continueOnError: runContinueOnError,
			},
			inputReader, cmd.OutOrStdout(), cmd.ErrOrStderr(),
		)
		if errors.Is(err, errUnknownRunFormat) {
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
	runCmd.Flags().StringVar(&runFormat, "format", "jsonl", `Output format: "jsonl" or "json"`)
	runCmd.Flags().BoolVar(
		&runWritesOnly, "writes-only", false,
		"Emit only write operations in ClientTupleKey format (consumable by fga tuple write)",
	)
	runCmd.Flags().StringVar(&runInputFile, "input", "", "Path to JSON input file (default: stdin)")
	runCmd.Flags().BoolVar(
		&runAggregate, "aggregate", false,
		"Buffer all records and collapse them (dedup tuples and filters, detect write/delete conflicts) before emitting",
	)
	runCmd.Flags().BoolVar(
		&runContinueOnError, "continue-on-error", false,
		"Skip input records that fail to parse or evaluate (warn to stderr) and exit non-zero if any were skipped",
	)
}
