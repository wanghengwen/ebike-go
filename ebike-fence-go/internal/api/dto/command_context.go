package dto

// EnsureCommandContext returns a command context suitable for downstream RPC calls.
// It mirrors Java CommandContextHolder.genParam which copies the inbound request context.
func EnsureCommandContext(cmdCtx *CommandContext, tenantID string) *CommandContext {
	if cmdCtx == nil {
		if tenantID == "" {
			return &CommandContext{}
		}
		return &CommandContext{TenantId: tenantID}
	}
	if tenantID != "" && cmdCtx.TenantId == "" {
		cp := *cmdCtx
		cp.TenantId = tenantID
		return &cp
	}
	return cmdCtx
}
