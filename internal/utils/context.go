package utils

import (
	"context"
	"os"
)

type contextKey struct {
	name string
}

var keyDebug = contextKey{"debug"}

func WithDebugContext(ctx context.Context, debug bool) context.Context {
	return context.WithValue(ctx, keyDebug, debug)
}

func GetDebugContextValue(ctx context.Context) bool {
	debug, _ := ctx.Value(keyDebug).(bool)

	return debug
}

var (
	keySuccessLog = contextKey{"successLog"}
	keyFailureLog = contextKey{"failureLog"}
)

// WithSuccessLog attaches a success log file to the context.
func WithSuccessLog(ctx context.Context, file *os.File) context.Context {
	return context.WithValue(ctx, keySuccessLog, file)
}

// WithFailureLog attaches a failure log file to the context.
func WithFailureLog(ctx context.Context, file *os.File) context.Context {
	return context.WithValue(ctx, keyFailureLog, file)
}

// GetSuccessLog retrieves the success log file from the context.
func GetSuccessLog(ctx context.Context) *os.File {
	file, _ := ctx.Value(keySuccessLog).(*os.File)
	return file
}

// GetFailureLog retrieves the failure log file from the context.
func GetFailureLog(ctx context.Context) *os.File {
	file, _ := ctx.Value(keyFailureLog).(*os.File)
	return file
}
