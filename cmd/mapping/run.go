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

	"github.com/mattn/go-isatty"
	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
	"github.com/spf13/cobra"
)

var (
	errUnknownRunFormat = errors.New("unknown format")
	errRecordsFailed    = errors.New("one or more input records failed")
	errNonObjectRecord  = errors.New("record is not a JSON object")
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

// filterOperationOutput is the JSON shape for a rule with tuple_filters: the
// filter conditions needed to expand the rule against a store, plus the desired
// state tuples to write once the filter is resolved.
type filterOperationOutput struct {
	Filters []language.TupleFilter `json:"filters"`
	Tuples  []writeTupleOutput     `json:"tuples"`
}

// batchOutput is the --format json shape without --writes-only: writes and
// deletes split by action, plus filter operations that cannot be expanded offline.
type batchOutput struct {
	Writes                []writeTupleOutput      `json:"writes"`
	Deletes               []writeTupleOutput      `json:"deletes"`
	TupleFilterOperations []filterOperationOutput `json:"tuple_filter_operations"`
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

// toFilterOpOutputs converts mapper filter operations to their JSON output shape,
// preserving each operation's desired-state tuples alongside its filter conditions.
func toFilterOpOutputs(ops []mapper.TupleFilterOperation) []filterOperationOutput {
	result := make([]filterOperationOutput, 0, len(ops))
	for _, filterOp := range ops {
		tuples := make([]writeTupleOutput, 0, len(filterOp.Tuples))
		for _, t := range filterOp.Tuples {
			tuples = append(tuples, toWriteTuple(t))
		}

		result = append(result, filterOperationOutput{Filters: filterOp.Filters, Tuples: tuples})
	}

	return result
}

// dedupFilterOps removes duplicate filter operations by full identity, preserving
// order. Used in the --aggregate path so an operation produced identically across
// records collapses to one, mirroring mapper.Compact's tuple dedup.
func dedupFilterOps(ops []mapper.TupleFilterOperation) []mapper.TupleFilterOperation {
	if len(ops) <= 1 {
		return ops
	}

	seen := make(map[string]struct{}, len(ops))
	deduped := make([]mapper.TupleFilterOperation, 0, len(ops))

	for _, filterOp := range ops {
		key, err := json.Marshal(filterOp)
		if err != nil {
			// TupleFilterOperation only contains JSON-safe types via JSON unmarshal;
			// treat any unexpected error as a unique entry to avoid silent data loss.
			deduped = append(deduped, filterOp)

			continue
		}

		k := string(key)
		if _, exists := seen[k]; exists {
			continue
		}

		seen[k] = struct{}{}

		deduped = append(deduped, filterOp)
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
// object splitting writes/deletes and listing unresolved filter operations.
func emitJSON(tuples []language.Tuple, ops []mapper.TupleFilterOperation, writesOnly bool, out io.Writer) error {
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
		Writes:                make([]writeTupleOutput, 0, len(tuples)),
		Deletes:               make([]writeTupleOutput, 0),
		TupleFilterOperations: toFilterOpOutputs(ops),
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

// warnUnresolvedFilters reports, one per line on stderr, each filter within each
// filter operation that cannot be expanded without a store.
func warnUnresolvedFilters(ops []mapper.TupleFilterOperation, errOut io.Writer) {
	for _, filterOp := range ops {
		for _, f := range filterOp.Filters {
			fmt.Fprintf(errOut,
				"WARN unresolved tuple filter (needs a store to expand): action=%s user=%s relation=%s object=%s\n",
				f.Action, f.User, f.Relation, f.Object)
		}
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

	buffered, ops, hadFailure, err := evalRecords(ctx, compiled, inputReader, opts, out, errOut)
	if err != nil {
		return err
	}

	// Streaming mode emitted per-record; ops is empty and warnUnresolvedFilters is a no-op.
	if isStreaming(opts) {
		warnUnresolvedFilters(ops, errOut)
	} else if err := emitBuffered(buffered, ops, opts, out, errOut); err != nil {
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
// evaluated; otherwise the tuples are buffered and returned. Unresolved filter
// operations are accumulated across all records (streaming mode emits warnings
// immediately and accumulates nothing). With --continue-on-error a malformed or
// evaluation-failing record is warned to errOut and skipped, and the boolean
// return reports whether any record was skipped; without it the first failure
// is returned as a fatal error.
func evalRecords(
	ctx context.Context,
	compiled *mapper.Mapping,
	inputReader io.Reader,
	opts runMappingOptions,
	out, errOut io.Writer,
) ([]language.Tuple, []mapper.TupleFilterOperation, bool, error) {
	buffered := make([]language.Tuple, 0)
	ops := make([]mapper.TupleFilterOperation, 0)
	reader := bufio.NewReader(inputReader)
	lineNum := 0
	hadFailure := false

	for {
		line, readErr := reader.ReadString('\n')

		if line != "" {
			lineNum++

			tuples, recOps, skipped, fatal := processLine(
				ctx, compiled, strings.TrimSpace(line), lineNum, opts, out, errOut)
			if fatal != nil {
				return nil, nil, false, fatal
			}

			hadFailure = hadFailure || skipped

			ops = append(ops, recOps...)
			buffered = append(buffered, tuples...)
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return nil, nil, false, fmt.Errorf("reading input: %w", readErr)
		}
	}

	return buffered, ops, hadFailure, nil
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
) ([]language.Tuple, []mapper.TupleFilterOperation, bool, error) {
	if trimmed == "" {
		return nil, nil, false, nil
	}

	recTuples, recOps, recErr := evalRecord(ctx, compiled, trimmed)
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

		// Emit filter warnings immediately so they are not held in memory until
		// EOF and are still surfaced if a later record causes a fatal error.
		warnUnresolvedFilters(recOps, errOut)

		return nil, nil, false, nil
	}

	return recTuples, recOps, false, nil
}

// evalRecord parses one JSONL line into an event and evaluates it against the
// mapping, returning the produced tuples and unresolved filter operations. It
// performs no output; the caller decides how to emit. A malformed line or an
// evaluation error is returned as err.
func evalRecord(
	ctx context.Context,
	compiled *mapper.Mapping,
	line string,
) ([]language.Tuple, []mapper.TupleFilterOperation, error) {
	var event map[string]any

	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, nil, fmt.Errorf("reading input JSON: %w", err)
	}

	// A bare `null` unmarshals into a nil map without error; reject it here so a
	// non-object record fails at the input boundary like any other malformed line.
	if event == nil {
		return nil, nil, fmt.Errorf("reading input JSON: %w", errNonObjectRecord)
	}

	result, err := compiled.Evaluate(ctx, event)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating mapping: %w", err)
	}

	return result.Tuples, result.TupleFilterOperations, nil
}

// emitBuffered emits a whole-run result as a single document. With --aggregate
// the buffered tuples are collapsed (cross-record dedup and write/delete
// conflict detection) before emitting; a conflict is a runtime error.
func emitBuffered(
	tuples []language.Tuple,
	ops []mapper.TupleFilterOperation,
	opts runMappingOptions,
	out, errOut io.Writer,
) error {
	if opts.aggregate {
		compacted, err := mapper.Compact(tuples)
		if err != nil {
			return fmt.Errorf("aggregating tuples: %w", err)
		}

		tuples = compacted
		ops = dedupFilterOps(ops)
	}

	if opts.format == "json" {
		if err := emitJSON(tuples, ops, opts.writesOnly, out); err != nil {
			return err
		}

		// The batch object carries tuple_filter_operations; the writes-only flat array does not.
		if opts.writesOnly {
			warnUnresolvedFilters(ops, errOut)
		}

		return nil
	}

	// jsonl with --aggregate: collapsed tuples emitted as lines after buffering.
	if err := emitJSONL(tuples, opts.writesOnly, out); err != nil {
		return err
	}

	warnUnresolvedFilters(ops, errOut)

	return nil
}

var (
	runFormat          string
	runWritesOnly      bool
	runInputFile       string
	runAggregate       bool
	runContinueOnError bool
	runInteractive     bool
)

var runCmd = &cobra.Command{
	Use:   "run <mapping-file>",
	Short: "Evaluate a mapping against JSON input and emit tuple operations",
	Long: `Reads JSONL from stdin (or --input) and evaluates it against the mapping file.
Input is JSON Lines: one JSON object per line. Outputs tuple operations as JSONL (default)
or a JSON batch (--format json).
Runs entirely offline — no store reads, no credentials, no network.
Rules using tuple_filters cannot be expanded offline and are reported as warnings on stderr
(or under tuple_filter_operations in the --format json batch).

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
  fga mapping run --writes-only mapping.yaml > out.jsonl && fga tuple write --store-id $STORE_ID --file out.jsonl
  fga mapping run mapping.yaml -i`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		errStream := cmd.ErrOrStderr()

		// Reject incompatible interactive flags before any prompt, so a bad flag
		// combination never blocks on asking for a mapping path first.
		if runInteractive {
			if err := checkInteractiveFlags(runWritesOnly, runInputFile); err != nil {
				fmt.Fprintln(errStream, "Error: "+err.Error())
				os.Exit(2)
			}
		}

		path := promptMappingFile(args, errStream)

		if runInteractive {
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				fmt.Fprintln(errStream, "Error: --interactive requires an interactive terminal")
				os.Exit(2)
			}

			err := runMappingInteractive(cmd.Context(), path, cmd.InOrStdin(), cmd.OutOrStdout(), errStream)
			if errors.Is(err, errMappingInvalid) {
				os.Exit(2)
			}

			return err
		}

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
			cmd.Context(), path,
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
	runCmd.Flags().StringVar(
		&runInputFile, "input", "", "Path to JSONL input file, one JSON object per line (default: stdin)")
	runCmd.Flags().BoolVar(
		&runAggregate, "aggregate", false,
		"Buffer all records and collapse them (dedup tuples and filters, detect write/delete conflicts) before emitting",
	)
	runCmd.Flags().BoolVar(
		&runContinueOnError, "continue-on-error", false,
		"Skip input records that fail to parse or evaluate (warn to stderr) and exit non-zero if any were skipped",
	)
	runCmd.Flags().BoolVarP(
		&runInteractive, "interactive", "i", false,
		"Explore the mapping in a terminal loop: paste JSON documents and see the tuples they produce (requires a TTY)",
	)
}
