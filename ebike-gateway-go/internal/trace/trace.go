package trace

import "context"

type ctxKey string

const traceIDKey ctxKey = "traceId"

// WithTraceID stores traceId in the standard Go context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// IDFromContext reads traceId from the standard Go context.
func IDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}
