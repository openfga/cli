package mapping

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

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
