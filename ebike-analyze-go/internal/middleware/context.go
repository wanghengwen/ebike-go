package middleware

import (
	"context"
	"fmt"
	"net/http"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const (
	ctxKeyAcceptLang ctxKey = "Accept-Language"
	ctxKeyTraceId    ctxKey = "X-Trace-Id"

	KeyCommandContext = "commandContext"
)

// CommandContextMiddleware propagates HTTP headers into request context for outbound RPC.
// Tenant/pin are taken from the request body's commandContext after handler binding
// (mirrors Java xyy-starter-moduleaop CommandContextAspect on controller entry).
func CommandContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLang := c.GetHeader("Accept-Language")
		traceId := c.GetHeader("X-Trace-Id")

		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, ctxKeyAcceptLang, acceptLang)
		reqCtx = context.WithValue(reqCtx, ctxKeyTraceId, traceId)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

// ApplyCommand validates and stores commandContext from a Command DTO.
// Mirrors Java CommandContextAspect.doAround: Assert.notNull(commandContext),
// Assert.notNull(traceId). Note tenantId is intentionally NOT validated here,
// matching Java (the aspect reads tenantId/pin but never asserts on them).
func ApplyCommand(c *gin.Context, cmd *dto.Command) (*dto.CommandContext, bool) {
	if cmd == nil || cmd.CommandContext == nil {
		writeCommandContextError(c, "commandContext 不能为null")
		return nil, false
	}
	ctx := cmd.CommandContext
	if ctx.TraceId == "" {
		writeCommandContextError(c, "traceId 不能为null")
		return nil, false
	}
	ctx.Source = config.GlobalConfig.Server.Name
	c.Set(KeyCommandContext, ctx)

	reqCtx := c.Request.Context()
	reqCtx = context.WithValue(reqCtx, ctxKeyTraceId, ctx.TraceId)
	c.Request = c.Request.WithContext(reqCtx)

	return ctx, true
}

func writeCommandContextError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeNotNull, msg)) // MsgCodeEnum.NOT_NULL = "00018"
	c.Abort()
}

// GetAcceptLanguageFromCtx extracts Accept-Language for outbound Feign-style RPC.
func GetAcceptLanguageFromCtx(ctx context.Context) string {
	if val, ok := ctx.Value(ctxKeyAcceptLang).(string); ok {
		return val
	}
	return ""
}

// GetTraceIdFromCtx extracts trace id from context (header or commandContext).
func GetTraceIdFromCtx(ctx context.Context) string {
	if val, ok := ctx.Value(ctxKeyTraceId).(string); ok {
		return val
	}
	return ""
}

// GetCommandContext retrieves CommandContext stored by ApplyCommand.
func GetCommandContext(c *gin.Context) *dto.CommandContext {
	val, exists := c.Get(KeyCommandContext)
	if !exists {
		return nil
	}
	ctx, ok := val.(*dto.CommandContext)
	if !ok {
		return nil
	}
	return ctx
}

// TenantID returns tenantId from stored CommandContext or an explicit fallback.
func TenantID(c *gin.Context, fallback string) (string, error) {
	if ctx := GetCommandContext(c); ctx != nil && ctx.TenantId != "" {
		return ctx.TenantId, nil
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("tenantId must not be empty")
}
