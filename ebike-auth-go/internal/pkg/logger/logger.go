package logger

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Request specific keys for logging
type contextKey string

const (
	TraceIdKey  contextKey = "traceId"
	TenantIdKey contextKey = "tenantId"
	PinKey      contextKey = "pin"
)

func init() {
	Log = zap.NewNop()
}

func InitLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "class",
		MessageKey:     "message",
		StacktraceKey:  "stack_trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// App name default, thread matching Java
	Log = zap.New(core, zap.AddCaller(), zap.Fields(
		zap.String("app", "ebike-auth-go"),
		zap.String("thread", "main"), // Go doesn't have named threads like Java, placeholder
	))
}

// WithContext extracts logging attributes from Gin context or standard context
func WithContext(ctx context.Context) *zap.Logger {
	var traceId, tenantId, pin string
	
	if ginCtx, ok := ctx.(*gin.Context); ok {
		traceId = ginCtx.GetString("traceId")
		tenantId = ginCtx.GetString("tenantId")
		pin = ginCtx.GetString("pin")
	} else if ctx != nil {
		if val, ok := ctx.Value(TraceIdKey).(string); ok {
			traceId = val
		}
		if val, ok := ctx.Value(TenantIdKey).(string); ok {
			tenantId = val
		}
		if val, ok := ctx.Value(PinKey).(string); ok {
			pin = val
		}
	}

	if traceId == "" {
		traceId = "-"
	}
	if tenantId == "" {
		tenantId = "-"
	}
	if pin == "" {
		pin = "-"
	}

	return Log.With(
		zap.String("traceId", traceId),
		zap.String("tenantId", tenantId),
		zap.String("pin", pin),
	)
}

func DetachContext(ctx context.Context) context.Context {
	var traceId, tenantId, pin string
	if ginCtx, ok := ctx.(*gin.Context); ok {
		traceId = ginCtx.GetString("traceId")
		tenantId = ginCtx.GetString("tenantId")
		pin = ginCtx.GetString("pin")
	} else if ctx != nil {
		if val, ok := ctx.Value(TraceIdKey).(string); ok {
			traceId = val
		}
		if val, ok := ctx.Value(TenantIdKey).(string); ok {
			tenantId = val
		}
		if val, ok := ctx.Value(PinKey).(string); ok {
			pin = val
		}
	}
	bgCtx := context.Background()
	bgCtx = context.WithValue(bgCtx, TraceIdKey, traceId)
	bgCtx = context.WithValue(bgCtx, TenantIdKey, tenantId)
	bgCtx = context.WithValue(bgCtx, PinKey, pin)
	return bgCtx
}

