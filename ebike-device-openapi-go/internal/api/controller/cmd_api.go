package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/crypto"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/metrics"
	"ebike-device-openapi-go/internal/pkg/validator"
	"ebike-device-openapi-go/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var regServiceForCmd *service.RegisterService

func InitCmdApi(rs *service.RegisterService) {
	regServiceForCmd = rs
}

func RegisterRoutes(r *gin.Engine) {
	// Base commands
	r.POST("/ebike/cmd/lock", handleGenericCommand(33))
	r.POST("/ebike/cmd/defend", handleGenericCommand(4))
	r.POST("/ebike/cmd/helmet_lock", handleGenericCommand(82))
	r.POST("/ebike/cmd/set_inner_param", handleGenericCommand(32))
	r.POST("/ebike/cmd/update_inner_fence", handleGenericCommand(50))
	r.POST("/ebike/cmd/upgrade_device", handleGenericCommand(35))
	r.POST("/ebike/cmd/upgrade_voice", handleGenericCommand(57))
	r.POST("/ebike/cmd/broadcast_voice", handleGenericCommand(14))
	r.POST("/ebike/cmd/query_device_info", handleGenericCommand(34))
	r.POST("/ebike/cmd/query_inner_param", handleGenericCommand(31))
	r.POST("/ebike/cmd/query_blueT_beacon", handleGenericCommand(85))
	r.POST("/ebike/cmd/switch_battery_compartment", handleGenericCommand(40))
	r.POST("/ebike/cmd/set_bluetooth", handleGenericCommand(49))
	r.POST("/ebike/cmd/query_ble_helmet_info", handleGenericCommand(108))
	r.POST("/ebike/cmd/scan_ble_helmet", handleGenericCommand(107))
	r.POST("/ebike/cmd/restart", handleGenericCommand(21))
	r.POST("/ebike/cmd/switch_rear_wheel_lock", handleGenericCommand(28))
	r.POST("/ebike/cmd/set_dashboard", handleGenericCommand(103))
	r.POST("/ebike/cmd/set_park_site", handleGenericCommand(115))
	r.POST("/ebike/cmd/triggerTempState", handleGenericCommand(113))
	r.POST("/ebike/cmd/dynamic_voice", handleGenericCommand(121))
	r.POST("/ebike/cmd/cooperate_location", handleGenericCommand(124))

	// Other paths (not prefix /ebike/cmd)
	r.POST("/force_lock", handleGenericCommand(78))
	r.POST("/set_limit_speed", handleGenericCommand(45))
	r.POST("/lift_limit_speed", handleGenericCommand(93))
	r.POST("/shutdown", handleGenericCommand(25))

	// Special handlers
	r.POST("/ebike/cmd/transmission", handleTransmission)
	r.POST("/transmission2", handleTransmission2)
	r.GET("/ebike/cmd/get_job_result", handleGetJobResult)
}

func handleGenericCommand(cmdID int16) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyMap map[string]interface{}
		if err := c.ShouldBindJSON(&bodyMap); err != nil {
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "参数校验失败: " + err.Error()})
			return
		}

		imei, _ := bodyMap["imei"].(string)
		if cmdID == 33 {
			logLockCommand(c, "request_received", imei, zap.ByteString("rawRequest", mustJSON(bodyMap)))
		}

		if imeiErr := validator.ValidateIMEI(imei); imeiErr != nil {
			if cmdID == 33 {
				logLockCommand(c, "imei_invalid", imei, zap.String("code", imeiErr.Code), zap.String("msg", imeiErr.Msg))
			}
			c.JSON(http.StatusOK, imeiErr)
			return
		}

		// Extract base EBikeRequest fields
		async := requestAsync(bodyMap)
		
		var tm *int
		if val, ok := bodyMap["tm"]; ok && val != nil {
			if fval, ok := val.(float64); ok {
				ival := int(fval)
				tm = &ival
			}
		}
		var dt *int
		if val, ok := bodyMap["dt"]; ok && val != nil {
			if fval, ok := val.(float64); ok {
				ival := int(fval)
				dt = &ival
			}
		}

		var cmdCtx *dto.CommandContext
		if ctxVal, ok := bodyMap["commandContext"]; ok && ctxVal != nil {
			if ctxMap, ok := ctxVal.(map[string]interface{}); ok {
				cmdCtx = &dto.CommandContext{
					TraceId:        getString(ctxMap, "traceId"),
					TenantId:       getString(ctxMap, "tenantId"),
					Pin:            getString(ctxMap, "pin"),
					Ip:             getString(ctxMap, "ip"),
					Platform:       getString(ctxMap, "platform"),
					DeviceId:       getString(ctxMap, "deviceId"),
					Source:         getString(ctxMap, "source"),
					Name:           getString(ctxMap, "name"),
					StressTesting:  getBool(ctxMap, "stressTesting"),
				}
			}
		}

		ebikeReq := &dto.EBikeRequest{
			TraceId:        getString(bodyMap, "traceId"),
			Imei:           imei,
			Async:          async,
			Payload:        bodyMap["payload"],
			Tm:             tm,
			Dt:             dt,
			CommandContext: cmdCtx,
		}

		// Remove common fields, what remains in bodyMap are the specific parameters
		delete(bodyMap, "imei")
		delete(bodyMap, "async")
		delete(bodyMap, "payload")
		delete(bodyMap, "traceId")
		delete(bodyMap, "tm")
		delete(bodyMap, "dt")
		delete(bodyMap, "commandContext")
		carID := getString(bodyMap, "carId")
		delete(bodyMap, "carId")
		delete(bodyMap, "izRiskControl")

		if paramErr := validator.ValidateCmdParams(cmdID, bodyMap); paramErr != nil {
			if cmdID == 33 {
				logLockCommand(c, "param_invalid", imei,
					zap.String("code", paramErr.Code),
					zap.String("msg", paramErr.Msg),
					zap.ByteString("cmdParams", mustJSON(bodyMap)),
				)
			}
			c.JSON(http.StatusOK, paramErr)
			return
		}

		// Tenant verification (skip if cmdID == 124)
		if cmdID != 124 {
			if resErr := verifyTenantId(c, imei, cmdCtx); resErr != nil {
				if cmdID == 33 {
					logLockCommand(c, "tenant_rejected", imei,
						zap.String("code", resErr.Code),
						zap.String("msg", resErr.Msg),
						zap.String("carId", carID),
					)
				}
				c.JSON(http.StatusOK, resErr)
				return
			}
		}

		// Dry run?
		dryRun := c.Query("dry_run") == "true"

		ecuLogin, ecuErr := resolveEcuLogin(c, dryRun)
		if ecuErr != nil {
			if cmdID == 33 {
				logLockCommand(c, "ecu_unavailable", imei,
					zap.String("code", ecuErr.Code),
					zap.String("msg", ecuErr.Msg),
					zap.Bool("dryRun", dryRun),
				)
			}
			c.JSON(http.StatusOK, *ecuErr)
			return
		}

		isShadowMode := c.GetBool("isShadowMode")
		metrics.IncCmd()
		ecuHost, ecuPort := "", 0
		if ecuLogin != nil {
			ecuHost, ecuPort = ecuLogin.Host, ecuLogin.Port
		}
		if cmdID == 33 {
			logLockCommand(c, "dispatch",
				imei,
				zap.Bool("async", async),
				zap.String("carId", carID),
				zap.ByteString("cmdParams", mustJSON(bodyMap)),
				zap.String("ecuHost", ecuHost),
				zap.Int("ecuPort", ecuPort),
				zap.String("ecuType", ecuLoginType(ecuLogin)),
				zap.Bool("dryRun", dryRun),
				zap.Bool("isShadowMode", isShadowMode),
				zap.String("traceId", resolveTraceID(ebikeReq)),
				zap.String("tenantId", tenantIDFromContext(cmdCtx)),
			)
		} else {
			logger.Log.Info("handleGenericCommand",
				zap.Int16("cmdID", cmdID),
				zap.String("imei", imei),
				zap.String("ecuHost", ecuHost),
				zap.Int("ecuPort", ecuPort),
				zap.String("ecuType", ecuLoginType(ecuLogin)),
				zap.Bool("isShadowMode", isShadowMode),
			)
		}
		body, err := service.DoAction(bodyMap, ebikeReq, cmdID, ecuLogin, dryRun, isShadowMode)
		if err != nil {
			if cmdID == 33 {
				logLockCommand(c, "dispatch_failed", imei, zap.Error(err))
			} else {
				logger.Log.Error("handleGenericCommand failed",
					zap.Int16("cmdID", cmdID),
					zap.String("imei", imei),
					zap.Error(err),
				)
			}
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: err.Error()})
			return
		}

		result := dto.Result{
			Success: true,
			Code:    "0",
			Msg:     "成功",
			Data:    body,
		}
		resultJSON, _ := json.Marshal(result)
		if cmdID == 33 {
			logLockCommand(c, "response_ok", imei,
				zap.ByteString("response", resultJSON),
				zap.Any("data", body),
			)
		} else {
			logger.Log.Info("handleGenericCommand response",
				zap.Int16("cmdID", cmdID),
				zap.String("imei", imei),
				zap.ByteString("response", resultJSON),
			)
		}
		c.JSON(http.StatusOK, result)
	}
}

func logLockCommand(c *gin.Context, phase, imei string, fields ...zap.Field) {
	base := []zap.Field{
		zap.String("phase", phase),
		zap.String("path", c.Request.URL.Path),
		zap.String("imei", imei),
		zap.String("clientIP", c.ClientIP()),
	}
	if traceID := c.GetString("traceId"); traceID != "" {
		base = append(base, zap.String("traceId", traceID))
	}
	if appID := c.GetHeader("appId"); appID != "" {
		base = append(base, zap.String("appId", appID))
	}
	logger.Log.Info("lock_command", append(base, fields...)...)
}

func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func resolveTraceID(req *dto.EBikeRequest) string {
	if req == nil {
		return ""
	}
	if req.CommandContext != nil && req.CommandContext.TraceId != "" {
		return req.CommandContext.TraceId
	}
	return req.TraceId
}

func tenantIDFromContext(cmdCtx *dto.CommandContext) string {
	if cmdCtx == nil {
		return ""
	}
	return cmdCtx.TenantId
}

func handleTransmission(c *gin.Context) {
	var bodyMap map[string]interface{}
	if err := c.ShouldBindJSON(&bodyMap); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "参数校验失败: " + err.Error()})
		return
	}
	processTransmission(c, bodyMap)
}

// handleTransmission2 mirrors Java transmission2(@RequestBody String request).
func handleTransmission2(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "参数校验失败: " + err.Error()})
		return
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(raw, &bodyMap); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "参数校验失败: " + err.Error()})
		return
	}
	processTransmission(c, bodyMap)
}

func processTransmission(c *gin.Context, bodyMap map[string]interface{}) {
	imei, _ := bodyMap["imei"].(string)
	if imeiErr := validator.ValidateIMEI(imei); imeiErr != nil {
		c.JSON(http.StatusOK, imeiErr)
		return
	}

	cmdVal, ok := bodyMap["cmd"]
	if !ok || cmdVal == nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "cmd不能为空"})
		return
	}

	cmdMap, ok := cmdVal.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "cmd格式不正确"})
		return
	}

	var cmdID int16
	if cVal, ok := cmdMap["c"]; ok && cVal != nil {
		if fc, ok := cVal.(float64); ok {
			cmdID = int16(fc)
		}
	}
	if cmdID == 0 {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "cmd.c不能为空"})
		return
	}

	if cmdID == 11 {
		reqProto, _ := bodyMap["deviceProtocol"].(string)
		if reqProto == "" {
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10016", Msg: "设备协议[deviceProtocol]不能为空"})
			return
		}
		if regServiceForCmd == nil || regServiceForCmd.Rdb == nil {
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: "Redis service not initialized"})
			return
		}
		key := deviceRedisKey(imei, requestBussinessType(bodyMap))
		deviceProtocol, err := regServiceForCmd.Rdb.HGet(c.Request.Context(), key, "deviceProtocol").Result()
		if err != nil || deviceProtocol == "" {
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10016", Msg: "设备注册未初始化协议[deviceProtocol]"})
			return
		}
		deviceProtocol = strings.ReplaceAll(deviceProtocol, "\"", "")
		if !strings.EqualFold(reqProto, deviceProtocol) {
			c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10016", Msg: "参数设备协议[deviceProtocol]与设备注册的协议不一致"})
			return
		}
	}

	async := requestAsync(bodyMap)
	var tm *int
	if val, ok := bodyMap["tm"]; ok && val != nil {
		if fval, ok := val.(float64); ok {
			ival := int(fval)
			tm = &ival
		}
	}
	var dt *int
	if val, ok := bodyMap["dt"]; ok && val != nil {
		if fval, ok := val.(float64); ok {
			ival := int(fval)
			dt = &ival
		}
	}

	var cmdCtx *dto.CommandContext
	if ctxVal, ok := bodyMap["commandContext"]; ok && ctxVal != nil {
		if ctxMap, ok := ctxVal.(map[string]interface{}); ok {
			cmdCtx = &dto.CommandContext{
				TraceId:       getString(ctxMap, "traceId"),
				TenantId:      getString(ctxMap, "tenantId"),
				Pin:           getString(ctxMap, "pin"),
				Ip:            getString(ctxMap, "ip"),
				Platform:      getString(ctxMap, "platform"),
				DeviceId:      getString(ctxMap, "deviceId"),
				Source:        getString(ctxMap, "source"),
				Name:          getString(ctxMap, "name"),
				StressTesting: getBool(ctxMap, "stressTesting"),
			}
		}
	}

	ebikeReq := &dto.EBikeRequest{
		TraceId:        getString(bodyMap, "traceId"),
		Imei:           imei,
		Async:          async,
		Payload:        bodyMap["payload"],
		Tm:             tm,
		Dt:             dt,
		CommandContext: cmdCtx,
	}

	var params map[string]interface{}
	if pVal, ok := cmdMap["param"]; ok && pVal != nil {
		if pm, ok := pVal.(map[string]interface{}); ok {
			params = pm
		}
	}

	if paramErr := validator.ValidateCmdParams(cmdID, params); paramErr != nil {
		c.JSON(http.StatusOK, paramErr)
		return
	}

	if cmdID != 124 {
		if resErr := verifyTenantId(c, imei, cmdCtx); resErr != nil {
			c.JSON(http.StatusOK, resErr)
			return
		}
	}

	dryRun := c.Query("dry_run") == "true"
	ecuLogin, ecuErr := resolveEcuLogin(c, dryRun)
	if ecuErr != nil {
		c.JSON(http.StatusOK, *ecuErr)
		return
	}

	isShadowMode := c.GetBool("isShadowMode")
	metrics.IncCmd()
	body, err := service.DoAction(params, ebikeReq, cmdID, ecuLogin, dryRun, isShadowMode)
	if err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Result{
		Success: true,
		Code:    "0",
		Msg:     "成功",
		Data:    body,
	})
}

func handleGetJobResult(c *gin.Context) {
	jobId := c.Query("jobId")
	if jobId == "" {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "jobId不能为空"})
		return
	}

	if regServiceForCmd == nil || regServiceForCmd.Rdb == nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: "Redis service not initialized"})
		return
	}

	ctx := c.Request.Context()
	replayKey := "ecu_replay_" + jobId
	data, err := regServiceForCmd.Rdb.Get(ctx, replayKey).Result()
	if err != nil || data == "" {
		data, err = regServiceForCmd.Rdb.Get(ctx, jobId).Result()
		if err == nil && data != "" {
			if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
				data = data[1 : len(data)-1]
			}
			data = strings.ReplaceAll(data, "\\", "")
		}
	}

	if data == "" {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10013", Msg: "设备未返回"})
		return
	}

	c.JSON(http.StatusOK, dto.Result{Success: true, Code: "0", Msg: "成功", Data: data})
}

// resolveEcuLogin mirrors Java ServiceFactory.getEbikeCmdService() offline/type checks.
// Supported types: xiaoan (TCP gateway), luoping (EMQX MQTT).
func resolveEcuLogin(c *gin.Context, dryRun bool) (ecuLogin *dto.EcuLogin, errResult *dto.Result) {
	ecuLoginObj, exists := c.Get("ecuLogin")
	if !exists {
		if dryRun {
			return &dto.EcuLogin{Host: "127.0.0.1", Port: 8080, Type: dto.DeviceTypeXiaoan}, nil
		}
		return nil, &dto.Result{Success: false, Code: "10012", Msg: "设备离线或未找到网关信息"}
	}

	ecuLogin = ecuLoginObj.(*dto.EcuLogin)

	// MQTT(luoping) 仅在 Type 显式为 "luoping" 时启用。这样可避免 Redis 中的
	// 历史缓存（老数据没有 Type 字段，反序列化后为空串）被误判为 MQTT 设备。
	if strings.EqualFold(ecuLogin.Type, dto.DeviceTypeLuoping) {
		if isLuopingSessionStale(ecuLogin) {
			return nil, &dto.Result{Success: false, Code: "10012", Msg: "设备离线或未找到网关信息"}
		}
		return ecuLogin, nil
	}

	// xiaoan 或历史缓存（Type 为空）→ 回退到原有 TCP 网关逻辑，保持向后兼容。
	if ecuLogin.Type == "" || strings.EqualFold(ecuLogin.Type, dto.DeviceTypeXiaoan) {
		return ecuLogin, nil
	}

	// 其余非空且未知的设备类型 → 拒绝。
	return nil, &dto.Result{
		Success: false,
		Code:    "10017",
		Msg:     "设备类型不支持[" + ecuLogin.Type + "]",
	}
}

func isLuopingSessionStale(ecuLogin *dto.EcuLogin) bool {
	if ecuLogin == nil || ecuLogin.LastSeenAt == nil || *ecuLogin.LastSeenAt <= 0 {
		return false // no lastSeen yet — rely on EMQX presence / sweeper
	}
	ttl := config.GetMqttOnlineTTL()
	return time.Since(time.Unix(*ecuLogin.LastSeenAt, 0)) > ttl
}

func ecuLoginType(ecuLogin *dto.EcuLogin) string {
	if ecuLogin == nil {
		return ""
	}
	return ecuLogin.Type
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func verifyTenantId(c *gin.Context, imei string, cmdCtx *dto.CommandContext) *dto.Result {
	if regServiceForCmd == nil || regServiceForCmd.Rdb == nil {
		return &dto.Result{Success: false, Code: "10001", Msg: "Redis service not initialized"}
	}

	// 1. Get tenantId from request or appId header
	var tenantId string
	if cmdCtx != nil && cmdCtx.TenantId != "" {
		tenantId = cmdCtx.TenantId
	} else {
		appId := c.GetHeader("appId")
		if appId != "" {
			tenantId = crypto.DesEncryptDefault(appId)
		}
	}

	// 2. Query device tenant info
	key := "device_ebike_" + imei
	cacheTenantId, err := regServiceForCmd.Rdb.HGet(c.Request.Context(), key, "tenantId").Result()
	if err != nil || cacheTenantId == "" {
		return &dto.Result{
			Success: false,
			Code:    "10006",
			Msg:     "设备[" + imei + "]未绑定租户",
		}
	}

	// Strip quotes if any
	cacheTenantId = strings.ReplaceAll(cacheTenantId, "\"", "")

	if tenantId == cacheTenantId {
		return nil
	}

	// 3. Check parent tenant info
	tenantKey := "tenant:" + tenantId
	isParent, err := regServiceForCmd.Rdb.HGet(c.Request.Context(), tenantKey, "isParent").Result()
	if err == nil && isParent != "" {
		isParent = strings.ReplaceAll(isParent, "\"", "")
	}

	if isParent == "1" {
		return nil
	}

	return &dto.Result{
		Success: false,
		Code:    "10006",
		Msg:     "租户ID不匹配(" + tenantId + "-" + cacheTenantId + ")",
	}
}

// requestAsync resolves async from the request body.
//
// Upstream (paas/rent/fence) declares Boolean async; when unset it stays null and Jackson
// NON_NULL omits the field from JSON — the body usually has no "async" key at all.
// Java EBikeRequest uses `boolean async = true`, so openapi Java deserializes a missing
// key as true. Gateway /ecu/wild only blocks for a device reply when async=false.
//
// Go intentionally treats omitted or null async as false so commands like query_device_info
// wait for the device response. This matches production behavior; it is not a strict mirror
// of Java DTO defaults (Gemini's note that primitive boolean defaults to false was wrong:
// EBikeRequest explicitly initializes async to true).
func requestAsync(body map[string]interface{}) bool {
	if val, ok := body["async"]; ok && val != nil {
		if bval, ok := val.(bool); ok {
			return bval
		}
	}
	return false
}

// requestBussinessType mirrors EBikeRequest.requestBussinessType().
// RequestDeviceType: ebike=0, bike=1. Defaults to ebike when deviceType is absent.
func requestBussinessType(body map[string]interface{}) string {
	if body == nil {
		return "ebike"
	}
	deviceType := parseJSONInt(body["deviceType"])
	if deviceType == 1 {
		return "bike"
	}
	return "ebike"
}

// deviceRedisKey mirrors EbikeCmdService.getRedisKey(BussinessType, imei).
func deviceRedisKey(imei, bussinessType string) string {
	switch strings.ToLower(bussinessType) {
	case "bike":
		return "device_bike_" + imei
	case "battery":
		return "device_battery_" + imei
	case "cabinet":
		return "device_cabinet_" + imei
	default:
		return "device_ebike_" + imei
	}
}

func parseJSONInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	default:
		return 0
	}
}
