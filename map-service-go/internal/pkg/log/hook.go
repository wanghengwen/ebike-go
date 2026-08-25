package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type ContextKey string

const (
	TenantIdKey ContextKey = "tenantId"
	TraceIdKey  ContextKey = "traceId"
)

type ContextHook struct{}

func NewContextHook() *ContextHook {
	return &ContextHook{}
}

func (h *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if v := entry.Context.Value(TenantIdKey); v != nil {
			if s, ok := v.(string); ok && s != "" {
				entry.Data["tenantId"] = s
			}
		} else if v := entry.Context.Value("tenantId"); v != nil {
			if s, ok := v.(string); ok && s != "" {
				entry.Data["tenantId"] = s
			}
		}

		if v := entry.Context.Value(TraceIdKey); v != nil {
			if s, ok := v.(string); ok && s != "" {
				entry.Data["traceId"] = s
			}
		} else if v := entry.Context.Value("traceId"); v != nil {
			if s, ok := v.(string); ok && s != "" {
				entry.Data["traceId"] = s
			}
		}
	}
	return nil
}

func WithTenantAndTrace(ctx context.Context, tenantId, traceId string) context.Context {
	ctx = context.WithValue(ctx, TenantIdKey, tenantId)
	return context.WithValue(ctx, TraceIdKey, traceId)
}
