package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

// ctxKey is a private type for context keys to avoid collisions across packages.
// Go best practice: never use built-in types (string, int) as context keys.
type ctxKey string

const (
	// ctxKeyAcceptLang is the typed context key for Accept-Language header propagation
	ctxKeyAcceptLang ctxKey = "Accept-Language"
)

// ContextKeys for Gin Context
const (
	KeyCommandContext = "commandContext"
	KeyUserPin        = "userPin"
	KeyTenantId       = "tenantId"
	KeyAcceptLanguage = "acceptLanguage"
)

// AuthoritiesPayload mirrors the JWT payload structure set by ebike-gateway-client's SecurityGatewayFilter.
// The gateway decodes the JWT, then Base64-encodes the payload JSON and puts it in the "authorities" header.
// Java field names use underscore (user_name, client_id) because they come from OAuth2 JWT standard claims.
type AuthoritiesPayload struct {
	UserName  string   `json:"user_name"` // This is the user's "pin" (phone number / user ID)
	ClientId  string   `json:"client_id"` // This is the "tenantId"
	Nickname  string   `json:"nickname"`  // This is the "name"
	Platform  string   `json:"platform"`
	DeviceId  string   `json:"deviceId"`
	Exp       int      `json:"exp"`
	GrantType string   `json:"grantType"`
	Scope     []string `json:"scope"`
	Jti       string   `json:"jti"`
}

// AuthContextMiddleware mirrors Java's FilterAutoAssemble (xyy-starter-gatewayfilter).
// It reads the "authorities" header (Base64-encoded JWT payload), parses it,
// and constructs a CommandContext that is stored in the Gin context.
//
// Java flow:
//  1. ebike-gateway-client's SecurityGatewayFilter verifies JWT token
//  2. It Base64-encodes the JWT payload and puts it in "authorities" header
//     using Java's Base64.getUrlEncoder() (URL-safe)
//  3. ebike-service-client's FilterAutoAssemble (from xyy-starter-gatewayfilter) reads this header
//  4. It decodes and extracts: client_id -> tenantId, user_name -> pin, nickname -> name,
//     and uses request.getLocalAddr() (the SERVER-side local IP) as ip
//  5. It validates tenantId is not empty; any failure short-circuits the request with
//     {success:false, code:"00001", msg:"网关解析authorities异常: ..."}
//  6. It sets CommandContextHolder (ThreadLocal) so downstream services can use it
func AuthContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAuthExcluded(c.Request.URL.Path) {
			c.Next()
			return
		}

		authHeader := c.GetHeader("authorities")

		abortWithAuthError := func(reason string) {
			// Java: returnJson(response, String.format("网关解析authorities异常: authorities=%s msg=%s", ...))
			// wrapped by ResultHelper.exception(MsgCodeEnum.EXCEPTION, ...) -> code "00001", HTTP 200
			msg := fmt.Sprintf("网关解析authorities异常: authorities=%s msg=%s", authHeader, reason)
			c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
			c.Abort()
		}

		if authHeader == "" {
			// Java's FilterAutoAssemble unconditionally parses the header for all
			// non-excluded paths; a missing header fails parsing and is rejected.
			abortWithAuthError("authorities header is missing")
			return
		}

		// Decode Base64 payload.
		// Java uses Base64.getUrlDecoder() which is URL-safe and accepts input
		// with or without padding. Try both Go URL-safe variants accordingly.
		var payloadBytes []byte
		var err error
		for _, enc := range [](*base64.Encoding){
			base64.RawURLEncoding,
			base64.URLEncoding,
		} {
			payloadBytes, err = enc.DecodeString(authHeader)
			if err == nil {
				break
			}
		}
		if err != nil {
			abortWithAuthError(err.Error())
			return
		}

		var payload AuthoritiesPayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			abortWithAuthError(err.Error())
			return
		}

		// Java: Assert.notEmpty(commandContext.getTenantId(), "authorities.tenantId")
		if payload.ClientId == "" {
			abortWithAuthError("authorities.tenantId 不能为空")
			return
		}

		// Build CommandContext exactly as Java's FilterAutoAssemble does:
		//   tenantId = json.getString("client_id")
		//   pin = json.getString("user_name")
		//   name = json.getString("nickname")
		//   ip = request.getLocalAddr()  // server-side local address, NOT the client IP
		cmdCtx := &dto.CommandContext{
			TenantId: payload.ClientId,
			Pin:      payload.UserName,
			Name:     payload.Nickname,
			Ip:       localAddr(c),
		}

		// Store in Gin context for downstream handlers
		c.Set(KeyCommandContext, cmdCtx)
		c.Set(KeyUserPin, payload.UserName)
		c.Set(KeyTenantId, payload.ClientId)

		// Also preserve Accept-Language for downstream header passthrough
		acceptLang := c.GetHeader("Accept-Language")
		c.Set(KeyAcceptLanguage, acceptLang)

		// Inject into standard Go context using typed key so it propagates to RPC calls
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, ctxKeyAcceptLang, acceptLang)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

// localAddr mirrors Java's request.getLocalAddr(): the IP of the local interface
// on which the request was received.
func localAddr(c *gin.Context) string {
	if addr, ok := c.Request.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		if host, _, err := net.SplitHostPort(addr.String()); err == nil {
			return host
		}
		return addr.String()
	}
	return ""
}

// CompleteCommandContext mirrors Java's CommandContextAspect (xyy-starter-gatewayfilter):
// after the request body is bound, the aspect completes the CommandContext with
// fields taken from the ClientDTO:
//
//	context.setSource(appName);
//	context.setTraceId(dto.getTraceId());
//	context.setStressTesting(dto.isStressTesting());
//	context.setPlatform(dto.getPlatform());
//	context.setDeviceId(dto.getDeviceId());
//
// If no context exists yet (Java: excluded no-login paths), a new one is created
// with tenantId taken from the request body.
func CompleteCommandContext(c *gin.Context, client *dto.ClientDTO) *dto.CommandContext {
	cmdCtx := GetCommandContext(c)
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{TenantId: client.TenantId}
	}
	cmdCtx.Source = config.GlobalConfig.Server.Name
	cmdCtx.TraceId = client.TraceId
	cmdCtx.StressTesting = client.StressTesting
	cmdCtx.Platform = client.Platform
	cmdCtx.DeviceId = client.DeviceId
	return cmdCtx
}

// GetAcceptLanguageFromCtx extracts Accept-Language from a standard context.Context.
// This is used by the RPC client layer to forward headers to downstream services,
// mirroring Java's FeignHeaderInterceptor.
func GetAcceptLanguageFromCtx(ctx context.Context) string {
	if val, ok := ctx.Value(ctxKeyAcceptLang).(string); ok {
		return val
	}
	return ""
}

// GetCommandContext retrieves the CommandContext from the Gin context
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
