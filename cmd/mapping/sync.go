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
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/openfga/mapper"
	"github.com/openfga/mapper/apply"
	"github.com/spf13/cobra"

	"github.com/openfga/cli/internal/cmdutils"
)

var errSyncRecordsFailed = errors.New("one or more records failed")

// errStyles holds lipgloss styles bound to a specific output writer.
type errStyles struct {
	time     lipgloss.Style
	writes   lipgloss.Style
	deletes  lipgloss.Style
	dim      lipgloss.Style
	ok       lipgloss.Style
	errStyle lipgloss.Style
}

func newErrStyles(w io.Writer) errStyles {
	r := lipgloss.NewRenderer(w)
	return errStyles{
		time:     r.NewStyle().Faint(true),
		writes:   r.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		deletes:  r.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		dim:      r.NewStyle().Faint(true),
		ok:       r.NewStyle().Foreground(lipgloss.Color("10")),
		errStyle: r.NewStyle().Foreground(lipgloss.Color("9")),
	}
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type syncOptions struct {
	quiet           bool
	dryRun          bool
	continueOnError bool
}

type syncStats struct {
	records int
	writes  int
	deletes int
	errors  int
	frame   int
}

type statusPrinter struct {
	out        io.Writer
	isTTY      bool
	quiet      bool
	styles     errStyles
	lastRender time.Time
}

func newStatusPrinter(out io.Writer, quiet bool) *statusPrinter {
	isTTY := false
	if f, ok := out.(*os.File); ok {
		isTTY = isatty.IsTerminal(f.Fd())
	}
	return &statusPrinter{out: out, isTTY: isTTY, quiet: quiet, styles: newErrStyles(out)}
}

func (p *statusPrinter) render(stats syncStats, suffix string) {
	if p.quiet {
		return
	}

	ts := p.styles.time.Render(time.Now().Format("15:04:05"))
	spin := p.styles.dim.Render(spinnerFrames[stats.frame%len(spinnerFrames)])
	w := p.styles.writes.Render(fmt.Sprintf("+%d", stats.writes))
	d := p.styles.deletes.Render(fmt.Sprintf("-%d", stats.deletes))
	count := p.styles.dim.Render(fmt.Sprintf("%d records", stats.records))

	line := fmt.Sprintf("%s  %s  %s %s  %s", ts, spin, w, d, count)
	if suffix != "" {
		line += "  " + suffix
	}

	if p.isTTY {
		fmt.Fprintf(p.out, "\r%s\x1b[K", line)
	} else if time.Since(p.lastRender) >= time.Second {
		p.lastRender = time.Now()
		fmt.Fprintln(p.out, line)
	}
}

func (p *statusPrinter) done(stats syncStats) {
	if p.quiet {
		return
	}

	ts := p.styles.time.Render(time.Now().Format("15:04:05"))
	tick := p.styles.ok.Render("✓")
	w := p.styles.writes.Render(fmt.Sprintf("+%d", stats.writes))
	d := p.styles.deletes.Render(fmt.Sprintf("-%d", stats.deletes))
	count := p.styles.dim.Render(fmt.Sprintf("%d records", stats.records))

	line := fmt.Sprintf("%s  %s  %s %s  %s", ts, tick, w, d, count)
	if stats.errors > 0 {
		line += "  " + p.styles.errStyle.Render(fmt.Sprintf("%d errors", stats.errors))
	}

	if p.isTTY {
		fmt.Fprintf(p.out, "\r%s\x1b[K\n", line)
	} else {
		fmt.Fprintln(p.out, line)
	}
}

func (p *statusPrinter) errLine(msg string) {
	ts := p.styles.time.Render(time.Now().Format("15:04:05"))
	x := p.styles.errStyle.Render("✗")
	if p.isTTY {
		fmt.Fprintf(p.out, "\r%s  %s  %s\x1b[K\n", ts, x, msg)
	} else {
		fmt.Fprintf(p.out, "%s  %s  %s\n", ts, x, msg)
	}
}

func syncMapping(
	ctx context.Context,
	path string,
	opts syncOptions,
	base apply.TupleClient,
	inputReader io.Reader,
	out, errOut io.Writer,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	compiled, compileErr := mapper.Compile(data)
	if compileErr != nil {
		fmt.Fprintln(errOut, mapper.DiagnosticsFrom(compileErr))
		return errMappingInvalid
	}

	// For dry-run, wrap the base client so writes are printed instead of applied.
	// Only emit JSONL when stdout is redirected — at a TTY it's unreadable at scale.
	writeTarget := base
	if opts.dryRun {
		dryOut := io.Discard
		if f, ok := out.(*os.File); ok && !isatty.IsTerminal(f.Fd()) {
			dryOut = out
		}
		writeTarget = &printOnlyClient{inner: base, out: dryOut}
	}

	// countingClient is shared across records; reset before each call.
	cc := &countingClient{inner: writeTarget}
	rec := apply.New(cc)

	printer := newStatusPrinter(errOut, opts.quiet)
	stats := syncStats{}

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	type recordResult struct {
		writes  int
		deletes int
		err     error
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// The goroutine lets the main select loop remain responsive during blocking reads.
	// ReadString is not context-cancellable — on early return the goroutine is process-scoped
	// and exits after the next line arrives or the input fd closes.
	results := make(chan recordResult)
	go func() {
		defer close(results)
		reader := bufio.NewReader(inputReader)
		for {
			line, readErr := reader.ReadString('\n')
			if line != "" {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					w, d, recErr := processSyncRecord(runCtx, compiled, rec, cc, trimmed)
					select {
					case results <- recordResult{writes: w, deletes: d, err: recErr}:
					case <-runCtx.Done():
						return
					}
				}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) {
					select {
					case results <- recordResult{err: fmt.Errorf("reading input: %w", readErr)}:
					case <-runCtx.Done():
					}
				}
				return
			}
		}
	}()

	hadFailure := false
	for {
		select {
		case res, ok := <-results:
			if !ok {
				printer.done(stats)
				if hadFailure {
					return errSyncRecordsFailed
				}
				return nil
			}
			if res.err != nil {
				// --continue-on-error skips the whole record on any error (validation,
				// read, write). Each record is independent so this is safe; the ✗ line
				// on stderr makes skipped records visible.
				if !opts.continueOnError {
					cancel()
					printer.errLine(res.err.Error())
					return res.err
				}
				hadFailure = true
				stats.errors++
				stats.records++
				stats.frame++
				printer.errLine(res.err.Error())
			} else {
				stats.records++
				stats.writes += res.writes
				stats.deletes += res.deletes
				stats.frame++
				printer.render(stats, "")
			}

		case <-heartbeat.C:
			stats.frame++
			printer.render(stats, printer.styles.dim.Render("waiting..."))

		case <-ctx.Done():
			cancel()
			printer.errLine(ctx.Err().Error())
			return ctx.Err()
		}
	}
}

func processSyncRecord(
	ctx context.Context,
	compiled *mapper.Mapping,
	rec apply.Executor,
	cc *countingClient,
	line string,
) (writes, deletes int, err error) {
	var event map[string]any
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return 0, 0, fmt.Errorf("reading input JSON: %w", err)
	}
	if event == nil {
		return 0, 0, fmt.Errorf("reading input JSON: %w", errNonObjectRecord)
	}

	result, err := compiled.Evaluate(ctx, event)
	if err != nil {
		return 0, 0, fmt.Errorf("evaluating mapping: %w", err)
	}

	cc.writes = 0
	cc.dels = 0

	if err := rec.Execute(ctx, result); err != nil {
		return 0, 0, err
	}

	return cc.writes, cc.dels, nil
}

var (
	syncInputFile       string
	syncDryRun          bool
	syncContinueOnError bool
	syncQuiet           bool
)

var syncCmd = &cobra.Command{
	Use:   "sync [mapping-file]",
	Short: "Apply a mapping against JSON input, writing results to a store",
	Long: `Reads JSONL from stdin (or --input) and evaluates it against the mapping file,
then applies the resulting tuple writes and deletes to the store.

Rules using tuple_filters read current store state, diff against desired state,
and derive concrete writes and deletes before applying them.

Progress is reported to stderr as a single updating line showing running write
and delete counts. On a non-interactive terminal each update is a new line.
Use --quiet to suppress all progress output (errors are always shown).

--dry-run evaluates and expands filters (reading the store) but prints the
planned operations to stdout instead of writing them.
--continue-on-error skips records that fail to evaluate or apply (shown on
stderr) and exits non-zero if any were skipped.`,
	Example: `  echo '{"id":"anne","org":"acme"}' | fga mapping sync mapping.yaml --store-id $STORE_ID
  fga mapping sync mapping.yaml --input events.jsonl --store-id $STORE_ID --dry-run
  fga mapping sync mapping.yaml --store-id $STORE_ID --continue-on-error < events.jsonl`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := cmdutils.GetClientConfig(cmd)
		errStream := cmd.ErrOrStderr()

		fgaSDKClient, err := clientConfig.GetFgaClient()
		if err != nil {
			return fmt.Errorf("configuring FGA client: %w", err)
		}

		path := promptMappingFile(args, errStream)

		inputReader := cmd.InOrStdin()
		if syncInputFile != "" {
			file, err := os.Open(syncInputFile)
			if err != nil {
				return fmt.Errorf("opening input %s: %w", syncInputFile, err)
			}
			defer file.Close()
			inputReader = file
		}

		opts := syncOptions{
			quiet:           syncQuiet,
			dryRun:          syncDryRun,
			continueOnError: syncContinueOnError,
		}

		tupleClient := &fgaTupleClient{inner: fgaSDKClient}

		err = syncMapping(
			cmd.Context(), path, opts,
			tupleClient, inputReader,
			cmd.OutOrStdout(), errStream,
		)
		if errors.Is(err, errMappingInvalid) {
			os.Exit(2)
		}

		return err
	},
}

func init() {
	syncCmd.Flags().String("store-id", "", "Store ID")
	syncCmd.Flags().String("model-id", "", "Authorization Model ID")
	if err := syncCmd.MarkFlagRequired("store-id"); err != nil {
		panic(err)
	}
	syncCmd.Flags().StringVar(
		&syncInputFile, "input", "", "Path to JSONL input file, one JSON object per line (default: stdin)")
	syncCmd.Flags().BoolVar(
		&syncDryRun, "dry-run", false,
		"Read the store and print planned operations to stdout without writing (suppressed at an interactive terminal)")
	syncCmd.Flags().BoolVar(
		&syncContinueOnError, "continue-on-error", false,
		"Skip records that fail to evaluate or apply (shown on stderr) and exit non-zero if any were skipped")
	syncCmd.Flags().BoolVar(
		&syncQuiet, "quiet", false,
		"Suppress progress output; only errors are written to stderr")
}
