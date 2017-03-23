package logger

import "context"

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

// NewContext creates a new context object containing a logger value.
func NewContext(ctx context.Context, logger KayveeLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger value contained in a context.
// For convenience, if the context does not contain a logger, a new logger is
// created and returned. This allows users of this method to use the logger
// immediately, e.g.
//   logger.FromContext(ctx).Info("...")
func FromContext(ctx context.Context) KayveeLogger {
	logger := ctx.Value(loggerKey)
	if lggr, ok := logger.(KayveeLogger); ok {
		return lggr
	}
	return New("")
}
