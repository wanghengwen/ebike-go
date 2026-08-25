package shadow

import "context"

type shadowKeyType struct{}

var shadowKey = shadowKeyType{}

// WithShadowTest marks the context as being in shadow test mode.
func WithShadowTest(ctx context.Context) context.Context {
	return context.WithValue(ctx, shadowKey, true)
}

// IsShadowTest returns true if the context is in shadow test mode.
func IsShadowTest(ctx context.Context) bool {
	val, _ := ctx.Value(shadowKey).(bool)
	return val
}
