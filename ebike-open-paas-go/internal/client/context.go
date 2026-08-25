package client

import "ebike-open-paas-go/internal/pkg/config"

// CommandContext builds the M2M commandContext for upstream calls.
func CommandContext(tenantID, traceID string) map[string]interface{} {
	pin := config.GlobalConfig().Open.CommandPin
	if pin == "" {
		pin = "open_paas"
	}
	if traceID == "" {
		traceID = NewTraceID()
	}
	return map[string]interface{}{
		"tenantId": tenantID,
		"traceId":  traceID,
		"pin":      pin,
		"platform": "other",
	}
}
