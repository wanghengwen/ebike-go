package middleware

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"ebike-open-paas-go/internal/contract"
	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/pkg/logger"
	"ebike-open-paas-go/internal/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ctxAgentKey  = "openAgent"
	ctxTenantKey = "tenantId"
	ctxAgentID   = "agentId"
)

// AgentFrom returns the authenticated agent bound by Auth.
func AgentFrom(c *gin.Context) (config.AgentConfig, bool) {
	v, ok := c.Get(ctxAgentKey)
	if !ok {
		return config.AgentConfig{}, false
	}
	a, ok := v.(config.AgentConfig)
	return a, ok
}

// TenantIDFrom returns the resolved internal tenantId.
func TenantIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxTenantKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// AgentIDFrom returns the Xiaoan agentId string.
func AgentIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxAgentID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Auth validates xc-access-token + agentId, then applies the per-agent rate limit.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("xc-access-token"))
		if token == "" {
			contract.Fail(c, contract.ErrUnauthorized, "xc-access-token missing")
			c.Abort()
			return
		}

		agentID := extractAgentID(c)
		if agentID == "" {
			contract.Fail(c, contract.ErrInputParamsMiss, "agentId missing")
			c.Abort()
			return
		}

		agent, ok := config.Agent(agentID)
		// The quota is charged before the credential is judged, so a caller
		// guessing tokens is metered like any other. Charging only successful
		// requests would leave the guessing loop unlimited, which is the one
		// case where a limit matters most.
		rl := config.AgentRateLimit(agent)
		if allowed, msg := repository.AgentAllowed(agentID, rl); !allowed {
			contract.Fail(c, contract.ErrBizError, msg)
			c.Abort()
			return
		}
		if !ok || !agent.IsEnabled() {
			contract.Fail(c, contract.ErrUnauthorized, "agent not found or disabled")
			c.Abort()
			return
		}
		// Constant-time comparison: == returns as soon as two bytes differ, and
		// the timing difference is measurable over enough requests, which turns
		// guessing a 32-char token from infeasible into a per-character search.
		if subtle.ConstantTimeCompare([]byte(agent.Token), []byte(token)) != 1 {
			contract.Fail(c, contract.ErrUnauthorized, "token unauthorized")
			c.Abort()
			return
		}

		c.Set(ctxAgentKey, agent)
		c.Set(ctxTenantKey, agent.TenantID)
		c.Set(ctxAgentID, agentID)
		logger.Log.Debug("open-paas auth ok",
			zap.String("agentId", agentID),
			zap.String("tenantId", agent.TenantID),
			zap.String("path", c.FullPath()),
		)
		c.Next()
	}
}

// InternalAuth guards /internal/* with the configured shared token.
//
// An unset token denies instead of allowing. These routes expose callback
// subscriptions and consumer internals, and the service is published on a public
// ingress — an unset token used to mean "open to anyone who finds the path",
// which is a config omission away from an outage rather than a deliberate mode.
func InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		want := strings.TrimSpace(config.GlobalConfig().Open.InternalToken)
		if want == "" {
			contract.Fail(c, contract.ErrUnauthorized,
				"internal routes are disabled: set open.internalToken (or INTERNAL_TOKEN) to enable them")
			c.Abort()
			return
		}
		got := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
		if got == "" {
			got = strings.TrimSpace(c.GetHeader("xc-access-token"))
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
			contract.Fail(c, contract.ErrUnauthorized, "internal token unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractAgentID(c *gin.Context) string {
	if q := strings.TrimSpace(c.Query("agentId")); q != "" {
		return q
	}
	if c.Request.Body == nil {
		return ""
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 {
		return ""
	}
	var raw map[string]interface{}
	if json.Unmarshal(body, &raw) != nil {
		return ""
	}
	return stringifyID(raw["agentId"])
}

func stringifyID(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}
