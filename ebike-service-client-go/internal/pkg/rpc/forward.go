package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"ebike-service-client-go/internal/api/dto"
)

// Downstream service names, matching the Java @FeignClient(name = ...) values.
const (
	ServiceUser       = "ebike-user"
	ServiceFence      = "ebike-fence"
	ServiceManagement = "ebike-management"
	ServiceOrder      = "ebike-order"
	ServicePay        = "ebike-pay"
	ServiceAccount    = "ebike-account"
	ServiceMarketing  = "ebike-marketing"
	ServiceOperation  = "ebike-operation"
	ServiceDevicePaas = "ebike-device-paas"
	ServiceMapService = "map-service"
)

// ForwardCommand mirrors the Java gateway pattern:
//
//	xxxApi.method(CommandContextHolder.genParam(XxxCmd.class, dto))
//
// i.e. BeanUtils.copyProperties(dto, cmd) + cmd.setCommandContext(ctx), then a
// Feign POST. Behavioral equivalence note: BeanUtils copies same-named properties
// only, and the downstream Jackson ignores unknown properties (Spring Boot
// default), so serializing ALL bound DTO fields plus commandContext is
// behaviorally identical from the downstream's point of view. Fields the client
// sent but the DTO does not declare are dropped by the typed Go struct, exactly
// like in Java.
//
// The downstream Result is returned for verbatim passthrough.
func ForwardCommand(ctx context.Context, serviceName, path string, dtoObj interface{}, cmdCtx *dto.CommandContext) (*dto.Result, error) {
	body, err := mergeCommandContext(dtoObj, cmdCtx)
	if err != nil {
		return nil, err
	}
	var result dto.Result
	if err := PostToService(ctx, serviceName, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ForwardCommandRaw is like ForwardCommand but returns the raw downstream
// response body. Used for endpoints whose Java controller returns the
// downstream object directly instead of a Result wrapper (e.g. pay callbacks).
func ForwardCommandRaw(ctx context.Context, serviceName, path string, dtoObj interface{}, cmdCtx *dto.CommandContext) (json.RawMessage, error) {
	body, err := mergeCommandContext(dtoObj, cmdCtx)
	if err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if err := PostToService(ctx, serviceName, path, body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// mergeCommandContext serializes the DTO and injects the commandContext field,
// mirroring Java's CommandContextHolder.genParam.
func mergeCommandContext(dtoObj interface{}, cmdCtx *dto.CommandContext) (map[string]json.RawMessage, error) {
	b, err := json.Marshal(dtoObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request dto: %v", err)
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("failed to convert request dto: %v", err)
	}
	ctxBytes, err := json.Marshal(cmdCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal commandContext: %v", err)
	}
	m["commandContext"] = ctxBytes
	return m, nil
}
