package utils //nolint:revive

import "context"

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
