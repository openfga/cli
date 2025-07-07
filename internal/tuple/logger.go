package tuple

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openfga/go-sdk/client"
	"gopkg.in/yaml.v3"
)

// TupleLogger writes tuple logs in various formats.
type TupleLogger struct {
	file          *os.File
	writer        *bufio.Writer
	format        string
	headerWritten bool
}

// NewTupleLogger creates a logger for the given file path.
func NewTupleLogger(path string) (*TupleLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to stat log file %s: %w", path, err)
	}

	return &TupleLogger{
		file:          f,
		writer:        bufio.NewWriter(f),
		format:        strings.ToLower(filepath.Ext(path)),
		headerWritten: info.Size() > 0,
	}, nil
}

func (l *TupleLogger) Close() error {
	if l == nil {
		return nil
	}
	_ = l.writer.Flush()
	_ = l.file.Sync()
	return l.file.Close()
}

func (l *TupleLogger) flush() {
	_ = l.writer.Flush()
	_ = l.file.Sync()
}

// LogSuccess writes a successful tuple key.
func (l *TupleLogger) LogSuccess(key client.ClientTupleKey) {
	if l == nil {
		return
	}
	switch l.format {
	case ".csv":
		l.writeCSV([]string{key.User, key.Relation, key.Object})
	case ".yaml", ".yml":
		record := []client.ClientTupleKey{key}
		b, _ := yaml.Marshal(record)
		l.writer.Write(b)
	default: // json and jsonl
		b, _ := json.Marshal(key)
		l.writer.Write(append(b, '\n'))
	}
	l.flush()
}

// LogFailure writes a failed tuple key.
func (l *TupleLogger) LogFailure(key client.ClientTupleKey) {
	if l == nil {
		return
	}
	switch l.format {
	case ".csv":
		l.writeCSV([]string{key.User, key.Relation, key.Object})
	case ".yaml", ".yml":
		record := []client.ClientTupleKey{key}
		b, _ := yaml.Marshal(record)
		l.writer.Write(b)
	default:
		b, _ := json.Marshal(key)
		l.writer.Write(append(b, '\n'))
	}
	l.flush()
}

func (l *TupleLogger) writeCSV(record []string) {
	w := csv.NewWriter(l.writer)
	if !l.headerWritten {
		header := []string{"user", "relation", "object"}
		if len(record) == 4 {
			header = append(header, "reason")
		}
		w.Write(header)
		l.headerWritten = true
	}
	_ = w.Write(record)
	w.Flush()
}

// Context utilities

type logKey struct{ name string }

var successLogKey = &logKey{"successLog"}
var failureLogKey = &logKey{"failureLog"}

func WithSuccessLogger(ctx context.Context, l *TupleLogger) context.Context {
	return context.WithValue(ctx, successLogKey, l)
}

func WithFailureLogger(ctx context.Context, l *TupleLogger) context.Context {
	return context.WithValue(ctx, failureLogKey, l)
}

func getSuccessLogger(ctx context.Context) *TupleLogger {
	logger, _ := ctx.Value(successLogKey).(*TupleLogger)
	return logger
}

func getFailureLogger(ctx context.Context) *TupleLogger {
	logger, _ := ctx.Value(failureLogKey).(*TupleLogger)
	return logger
}
