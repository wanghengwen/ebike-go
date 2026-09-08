package logger

import (
	"context"
	"log"
)

type ctxKey string

const (
	// TraceIdKey is the context key for storing the TraceId.
	TraceIdKey ctxKey = "X-Trace-Id"
)

// WithTraceId injects a traceId into the context.
func WithTraceId(ctx context.Context, traceId string) context.Context {
	if traceId == "" {
		return ctx
	}
	return context.WithValue(ctx, TraceIdKey, traceId)
}

// ContextFrom copies the traceId from ctx into a background context for async work.
func ContextFrom(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	if traceId, ok := ctx.Value(TraceIdKey).(string); ok && traceId != "" {
		return WithTraceId(context.Background(), traceId)
	}
	return context.Background()
}

// Printf logs a message, automatically prepending the traceId if it exists in the context.
func Printf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		if traceId, ok := ctx.Value(TraceIdKey).(string); ok && traceId != "" {
			format = "[" + traceId + "] " + format
		}
	}
	log.Printf(format, v...)
}
