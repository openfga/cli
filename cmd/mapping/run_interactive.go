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
	"text/tabwriter"

	"github.com/openfga/mapper"
	"github.com/openfga/mapper/language"
	"golang.org/x/term"
)

var (
	errInteractiveWithWritesOnly      = errors.New("--interactive cannot be combined with --writes-only")
	errInteractiveWithInput           = errors.New("--interactive cannot be combined with --input")
	errInteractiveWithFormat          = errors.New("--interactive cannot be combined with --format")
	errInteractiveWithAggregate       = errors.New("--interactive cannot be combined with --aggregate")
	errInteractiveWithContinueOnError = errors.New("--interactive cannot be combined with --continue-on-error")
)

const (
	promptPrimary      = "> "
	promptContinuation = "... "
)

// checkInteractiveFlags reports the first flag that is incompatible with the
// interactive explorer. Interactive output is a human-readable table evaluated
// one pasted document at a time, so the batch flags have no meaning in the loop:
// --writes-only and --format select machine output, --input reads from a file
// instead of the prompt, and --aggregate and --continue-on-error act on a whole
// input stream. Rejecting them is clearer than silently ignoring them.
//
// formatChanged reports whether --format was set on the command line: it has a
// non-empty default ("jsonl"), so its value alone cannot distinguish an unset
// flag from one the user explicitly passed. Any explicit --format is rejected.
func checkInteractiveFlags(opts runMappingOptions, inputFile string, formatChanged bool) error {
	switch {
	case opts.writesOnly:
		return errInteractiveWithWritesOnly
	case inputFile != "":
		return errInteractiveWithInput
	case formatChanged:
		return errInteractiveWithFormat
	case opts.aggregate:
		return errInteractiveWithAggregate
	case opts.continueOnError:
		return errInteractiveWithContinueOnError
	}

	return nil
}

// interactiveSession holds the state of a single `run --interactive` loop: the
// currently loaded mapping, whether per-evaluation tracing is displayed, and the
// output streams. compiled is replaced in place by :reload.
type interactiveSession struct {
	ctx         context.Context //nolint:containedctx
	path        string
	compiled    *mapper.Mapping
	traceOn     bool
	interactive bool
	out         io.Writer
	errOut      io.Writer
}

// lineReader abstracts reading one line of input so the raw-terminal explorer
// (arrow-key editing, history) and the plain reader used by tests and pipes can
// share the same evaluation loop. SetPrompt switches between the primary and
// continuation prompts as a multi-line document is assembled.
type lineReader interface {
	ReadLine() (string, error)
	SetPrompt(prompt string)
}

// termLineReader drives a golang.org/x/term terminal: full line editing, cursor
// movement, and history on a real TTY.
type termLineReader struct{ terminal *term.Terminal }

func (r *termLineReader) ReadLine() (string, error) {
	line, err := r.terminal.ReadLine()
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	return line, nil
}

func (r *termLineReader) SetPrompt(prompt string) { r.terminal.SetPrompt(prompt) }

// scannerLineReader reads whole lines from a plain io.Reader (tests, pipes). It
// has no line editing; it prints the current prompt and returns the next line,
// reporting io.EOF once the input is exhausted.
type scannerLineReader struct {
	scanner *bufio.Scanner
	out     io.Writer
	prompt  string
}

func (r *scannerLineReader) ReadLine() (string, error) {
	fmt.Fprint(r.out, r.prompt)

	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}

	if err := r.scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	return "", io.EOF
}

func (r *scannerLineReader) SetPrompt(prompt string) { r.prompt = prompt }

// readWriter pairs an input reader with an output writer so a term.Terminal can
// read keystrokes from one stream and render to another.
type readWriter struct {
	io.Reader
	io.Writer
}

// runMappingInteractive compiles the mapping (with tracing enabled so :trace can
// toggle display without recompiling) and runs the explore loop, reading pasted
// JSON documents from input and rendering the resulting tuple operations. A real
// terminal gets full line editing via raw mode; anything else (tests, pipes)
// uses a plain line scanner. A compile failure is reported as errMappingInvalid
// before the loop starts.
func runMappingInteractive(ctx context.Context, path string, input io.Reader, out, errOut io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	compiled, compileErr := mapper.Compile(data, mapper.WithTrace(true))
	if compileErr != nil {
		fmt.Fprintln(errOut, mapper.DiagnosticsFrom(compileErr))

		return errMappingInvalid
	}

	session := &interactiveSession{ctx: ctx, path: path, compiled: compiled, out: out, errOut: errOut}

	if file, ok := input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		return session.runRaw(file)
	}

	return session.runPlain(input)
}

// runRaw puts the terminal into raw mode and drives the explorer with a
// term.Terminal, giving cursor movement, history, and line editing. Session
// output is routed through the terminal so it interleaves correctly with the
// editing line and gets CRLF translation. The terminal state is always restored
// on exit.
func (s *interactiveSession) runRaw(file *os.File) error {
	fileDescriptor := int(file.Fd())

	oldState, err := term.MakeRaw(fileDescriptor)
	if err != nil {
		return fmt.Errorf("entering raw mode: %w", err)
	}

	defer func() { _ = term.Restore(fileDescriptor, oldState) }()

	terminal := term.NewTerminal(readWriter{Reader: file, Writer: s.out}, promptPrimary)

	// term.NewTerminal assumes 80x24; seed the real dimensions so cursor and
	// repaint maths are correct on wider terminals when editing wrapped lines.
	if width, height, sizeErr := term.GetSize(fileDescriptor); sizeErr == nil {
		_ = terminal.SetSize(width, height)
	}

	stopResize := watchResize(fileDescriptor, terminal)
	defer stopResize()

	s.out = terminal
	s.errOut = terminal
	s.interactive = true

	s.banner()

	return s.loop(&termLineReader{terminal: terminal})
}

// runPlain drives the explorer over a plain reader without raw mode, used by
// tests and non-terminal input.
func (s *interactiveSession) runPlain(input io.Reader) error {
	s.banner()

	scanner := bufio.NewScanner(input)
	// Allow pasted documents well beyond the 64KB default line cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	return s.loop(&scannerLineReader{scanner: scanner, out: s.out, prompt: promptPrimary})
}

func (s *interactiveSession) banner() {
	fmt.Fprintf(s.out,
		"mapping loaded: %d rules. Type or paste a JSON document; it is evaluated once complete.  "+
			"commands: :reload  :trace on|off  :quit\n\n",
		s.compiled.RuleCount())
}

// loop reads input line by line. A line starting with ":" while no document is
// buffered is a command. Otherwise lines accumulate into a document that is
// evaluated as soon as it forms a complete JSON value; an incomplete document
// switches to the continuation prompt and keeps reading. A blank line forces
// evaluation of whatever is buffered.
func (s *interactiveSession) loop(reader lineReader) error {
	var doc []string

	nudged := false

	submit := func() {
		s.evaluate(strings.Join(doc, "\n"))
		doc = doc[:0]
		nudged = false

		reader.SetPrompt(promptPrimary)
	}

	for {
		line, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err //nolint:wrapcheck // already wrapped by the lineReader implementation
		}

		trimmed := strings.TrimSpace(line)

		if len(doc) == 0 && strings.HasPrefix(trimmed, ":") {
			if s.handleCommand(trimmed) {
				return nil
			}

			continue
		}

		if trimmed == "" {
			if len(doc) > 0 {
				submit()
			}

			continue
		}

		doc = append(doc, line)

		if awaitingMoreInput(strings.Join(doc, "\n")) {
			nudged = s.nudgeIncomplete(nudged)

			reader.SetPrompt(promptContinuation)

			continue
		}

		submit()
	}
}

// nudgeIncomplete prints a one-time hint, on an interactive terminal only, that
// the buffered document is not yet valid JSON and how to force evaluation. It
// returns the updated nudged flag so the hint is shown at most once per document.
func (s *interactiveSession) nudgeIncomplete(nudged bool) bool {
	if nudged || !s.interactive {
		return nudged
	}

	fmt.Fprintln(s.errOut,
		"  … incomplete JSON — keep typing, or press Enter on a blank line to evaluate as-is")

	return true
}

type jsonCompleteness int

const (
	jsonComplete jsonCompleteness = iota
	jsonIncomplete
	jsonInvalid
)

// classifyJSON reports whether s is a complete JSON value, an incomplete one
// (more input needed), or invalid. A truncated document decodes with an
// unexpected-EOF error, which is the signal to keep reading rather than reject;
// any other decode error means the document will never parse, so it is submitted
// immediately and the evaluation step reports the error.
func classifyJSON(s string) jsonCompleteness {
	var value any

	err := json.NewDecoder(strings.NewReader(s)).Decode(&value)

	switch {
	case err == nil:
		return jsonComplete
	case errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF):
		return jsonIncomplete
	default:
		return jsonInvalid
	}
}

// awaitingMoreInput reports whether the buffered document is incomplete in a way
// that more lines can legitimately complete. JSON is whitespace-insensitive
// between tokens, so a newline is a valid separator only where whitespace is
// legal. A document is therefore continuable only if it is incomplete and
// appending a newline keeps it incomplete rather than making it invalid: an
// unterminated string or a half-typed literal/number becomes invalid with a
// newline appended (a raw newline cannot appear mid-token), so it is evaluated
// and its error reported instead of silently waiting for input that can never
// make it valid.
func awaitingMoreInput(s string) bool {
	if classifyJSON(s) != jsonIncomplete {
		return false
	}

	return classifyJSON(s+"\n") != jsonInvalid
}

// handleCommand runs a `:` command and reports whether the loop should quit.
func (s *interactiveSession) handleCommand(cmd string) bool {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return false
	}

	switch fields[0] {
	case ":quit", ":q", ":exit":
		return true
	case ":reload", ":r":
		s.reload()
	case ":trace":
		s.setTrace(fields)
	default:
		fmt.Fprintf(s.errOut, "unknown command %q; use :reload, :trace on|off, :quit\n", fields[0])
	}

	return false
}

func (s *interactiveSession) setTrace(fields []string) {
	if len(fields) != 2 || (fields[1] != "on" && fields[1] != "off") {
		fmt.Fprintln(s.errOut, "usage: :trace on|off")

		return
	}

	s.traceOn = fields[1] == "on"

	fmt.Fprintf(s.out, "trace %s\n", fields[1])
}

// reload re-reads and recompiles the mapping from disk. On failure the loaded
// mapping is kept so the session stays usable.
func (s *interactiveSession) reload() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		fmt.Fprintf(s.errOut, "reload failed: %v\n", err)

		return
	}

	compiled, compileErr := mapper.Compile(data, mapper.WithTrace(true))
	if compileErr != nil {
		fmt.Fprintln(s.errOut, mapper.DiagnosticsFrom(compileErr))
		fmt.Fprintln(s.errOut, "reload failed; keeping the previously loaded mapping")

		return
	}

	s.compiled = compiled

	fmt.Fprintf(s.out, "mapping reloaded: %d rules\n", compiled.RuleCount())
}

// evaluate parses and evaluates one pasted document, rendering its tuples (and,
// when tracing is on, the per-rule summary). Parse and evaluation errors are
// reported inline and do not end the session.
func (s *interactiveSession) evaluate(doc string) {
	var event map[string]any

	if err := json.Unmarshal([]byte(doc), &event); err != nil {
		s.reportJSONError(doc, err)

		return
	}

	if event == nil {
		fmt.Fprintf(s.errOut, "Error: %v\n", errNonObjectRecord)

		return
	}

	result, err := s.compiled.Evaluate(s.ctx, event)
	if err != nil {
		fmt.Fprintf(s.errOut, "Error: evaluating mapping: %v\n", err)

		// Evaluate returns a populated trace alongside the error, so under
		// :trace on the failing rule is still surfaced instead of no trace.
		if s.traceOn && result != nil {
			s.renderTrace(result.Trace)
		}

		return
	}

	s.renderTuples(result.Tuples, result.TupleFilterOperations)

	if s.traceOn {
		s.renderTrace(result.Trace)
	}
}

// reportJSONError prints a parse failure. For a syntax error it locates the
// offending byte and points a caret at it (compiler style); other errors fall
// back to a plain message.
func (s *interactiveSession) reportJSONError(doc string, err error) {
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		fmt.Fprintf(s.errOut, "Error: reading input JSON: %v\n", err)

		return
	}

	line, column, text := locateOffset(doc, int(syntaxErr.Offset))

	fmt.Fprintf(s.errOut, "Error: invalid JSON at line %d, column %d: %v\n", line, column, err)
	fmt.Fprintf(s.errOut, "  %s\n  %s^\n", text, strings.Repeat(" ", column-1))
}

// locateOffset maps a byte offset within doc to a 1-based line and column and
// returns the text of that line, so an error can be shown with a caret under the
// offending character. offset is a json.SyntaxError.Offset, which points just
// past the byte that triggered the error, so the caret targets offset-1.
func locateOffset(doc string, offset int) (int, int, string) {
	if offset > len(doc) {
		offset = len(doc)
	}

	caret := max(offset-1, 0)

	prefix := doc[:caret]
	line := strings.Count(prefix, "\n") + 1
	lineStart := strings.LastIndex(prefix, "\n") + 1
	column := caret - lineStart + 1

	text := doc[lineStart:]
	if end := strings.IndexByte(text, '\n'); end >= 0 {
		text = text[:end]
	}

	return line, column, text
}

// renderTuples prints one aligned row per tuple as `op user relation object`.
// Tuple-filter operations render their filter conditions as `filter` rows,
// followed by the desired-state tuples that operation reconciles toward as
// indented `desired` rows; these cannot be applied offline (they drive a
// read-diff-write against a store) so they are shown for inspection, not as
// guaranteed writes.
func (s *interactiveSession) renderTuples(tuples []language.Tuple, ops []mapper.TupleFilterOperation) {
	if len(tuples) == 0 && len(ops) == 0 {
		fmt.Fprintln(s.out, "  (no tuples)")

		return
	}

	writer := tabwriter.NewWriter(s.out, 0, 0, 3, ' ', 0)

	for _, tuple := range tuples {
		operation := string(tuple.Action)
		if operation == "" {
			operation = string(language.ActionWrite)
		}

		fmt.Fprintf(writer, "  %s\t%s\t%s\t%s\n", operation, tuple.User, tuple.Relation, tupleObject(tuple))
	}

	for _, filterOp := range ops {
		for _, filter := range filterOp.Filters {
			action := string(filter.Action)
			if action == "" {
				// Match the compiler default so an unset action reads as patch.
				action = string(language.FilterActionPatch)
			}

			fmt.Fprintf(writer, "  filter:%s\t%s\t%s\t%s\n",
				action, orWildcard(filter.User), orWildcard(filter.Relation), orWildcard(filter.Object))
		}

		for _, tuple := range filterOp.Tuples {
			fmt.Fprintf(writer, "    desired\t%s\t%s\t%s\n", tuple.User, tuple.Relation, tupleObject(tuple))
		}
	}

	_ = writer.Flush()
}

// orWildcard renders an empty tuple-filter field as "*", the wildcard it stands
// for: an unset user, relation, or object matches any value, so a blank cell
// would misleadingly read as a literal empty string.
func orWildcard(field string) string {
	if field == "" {
		return "*"
	}

	return field
}

// tupleObject formats a tuple's object, appending its condition name in brackets
// when the tuple is conditioned, together with the rendered context so two
// tuples that differ only by context are distinguishable.
func tupleObject(tuple language.Tuple) string {
	if tuple.Condition == "" {
		return tuple.Object
	}

	if len(tuple.Context) > 0 {
		if ctx, err := json.Marshal(tuple.Context); err == nil {
			return fmt.Sprintf("%s  [%s %s]", tuple.Object, tuple.Condition, ctx)
		}
	}

	return fmt.Sprintf("%s  [%s]", tuple.Object, tuple.Condition)
}

// renderTrace prints the same rule summary the test command emits: matched rules
// with their tuple counts, skipped rules, and any rule errors.
func (s *interactiveSession) renderTrace(trace *mapper.Trace) {
	if trace == nil {
		return
	}

	var matched, skipped, errored []string

	for _, ruleTrace := range trace.Rules {
		switch ruleTrace.Status {
		case mapper.RuleMatched:
			matched = append(matched, fmt.Sprintf("%s -> %d tuples", ruleTrace.Name, ruleTrace.EmittedN))
		case mapper.RuleSkipped:
			skipped = append(skipped, ruleTrace.Name)
		case mapper.RuleErrored:
			errored = append(errored, fmt.Sprintf("%s: %v", ruleTrace.Name, ruleTrace.Error))
		}
	}

	if len(matched) > 0 {
		fmt.Fprintf(s.out, "  rules matched: %s\n", strings.Join(matched, ", "))
	}

	if len(skipped) > 0 {
		fmt.Fprintf(s.out, "  rules skipped: %s (when guard false)\n", strings.Join(skipped, ", "))
	}

	if len(errored) > 0 {
		fmt.Fprintf(s.out, "  rules errored: %s\n", strings.Join(errored, ", "))
	}

	fmt.Fprintf(s.out, "  (evaluated in %s)\n", trace.Duration)
}
