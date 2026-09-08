package controller

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/decode/xiaoan"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/metrics"
	"ebike-device-openapi-go/internal/pkg/utils"
	"ebike-device-openapi-go/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var regService *service.RegisterService

func InitXiaoanApi(rs *service.RegisterService) {
	regService = rs
}

func RegisterXiaoanRoutes(r *gin.Engine) {
	api := r.Group("/device-gateway/xiaoan")
	{
		api.POST("/login", handleLogin)
		api.POST("/logout", handleLogout)
		api.POST("/decode", handleDecode)
		api.POST("/replay", handleReplay)
	}
}

func handleLogin(c *gin.Context) {
	var cmd dto.LoginCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: err.Error()})
		return
	}

	isShadowMode := c.GetBool("isShadowMode")
	ret := regService.Register(&cmd, isShadowMode)
	if !ret {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: "login exception"})
		return
	}

	traceId := ""
	if cmd.CommandContext != nil {
		traceId = cmd.CommandContext.TraceId
	}
	versionStr := ""
	if cmd.Version != nil {
		versionStr = strconv.FormatInt(*cmd.Version, 10)
	}

	login := &dto.Login{
		TraceId:    traceId,
		Imei:       cmd.Imei,
		Host:       cmd.Host,
		Port:       cmd.Port,
		Imsi:       cmd.Imsi,
		Version:    versionStr,
		DeviceType: cmd.DeviceType,
		Timestamp:  cmd.Timestamp,
	}

	c.JSON(http.StatusOK, dto.Result{Success: true, Code: "0", Msg: "成功", Data: login})
}

func handleLogout(c *gin.Context) {
	var cmd dto.LogoutCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: err.Error()})
		return
	}

	isShadowMode := c.GetBool("isShadowMode")
	ret := regService.Unregister(&cmd, isShadowMode)
	if !ret {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: "logout exception"})
		return
	}

	traceId := ""
	if cmd.CommandContext != nil {
		traceId = cmd.CommandContext.TraceId
	}
	logout := &dto.Logout{
		TraceId:   traceId,
		Imei:      cmd.Imei,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	c.JSON(http.StatusOK, dto.Result{Success: true, Code: "0", Msg: "成功", Data: logout})
}

func handleDecode(c *gin.Context) {
	var cmd dto.DecodeCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: err.Error()})
		return
	}

	if cmd.HexBody == "" || cmd.MessageHeader == nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "invalid hex body or header"})
		return
	}

	bytes, err := hex.DecodeString(cmd.HexBody)
	if err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: "hex decode error"})
		return
	}

	buf := xiaoan.NewByteBuf(bytes)
	decodedMsg, err := xiaoan.DecodeHex(cmd.MessageHeader, buf)
	if err != nil {
		logger.Log.Warn("xiaoan decode failed",
			zap.String("imei", cmd.Imei),
			zap.Any("header", cmd.MessageHeader),
			zap.String("hexBody", cmd.HexBody),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10001", Msg: err.Error()})
		return
	}

	importJson, _ := json.Marshal(decodedMsg)

	// Create DeviceReportMessage
	reportMsg := &dto.DeviceReportMessage{
		Imei:          cmd.Imei,
		BussinessType: "ebike",
		Data:          string(importJson),
	}

	// Determine msgType dynamically or use interface
	if typMsg, ok := decodedMsg.(interface{ GetMsgType() string }); ok {
		reportMsg.MsgType = typMsg.GetMsgType()
	} else {
		// Hack to extract msgType since we know it's injected by decoder
		var generic map[string]interface{}
		json.Unmarshal(importJson, &generic)
		if t, ok := generic["msgType"].(string); ok {
			reportMsg.MsgType = t
		}
	}

	isShadowMode := c.GetBool("isShadowMode")
	metrics.IncDecode()
	// Send to Kafka
	regService.Pusher.PushMessage(cmd.Imei, reportMsg, isShadowMode)

	// NOTE(Architecture-Mirror): In Java's DeviceGatewayXiaoanController, `replayHex`
	// is obtained via `replayEncodeMap.get("Replay" + cmd + "Encode")`. Since only
	// `DefaultReplayHexEncode` exists in the Spring context (and its bean name doesn't match),
	// it always returns null, resulting in `NeedReplay: false`.
	// We statically mirror this runtime behavior here to avoid unnecessary reflection and memory allocation.
	replayData := &dto.ReplayData{
		Imei:       cmd.Imei,
		NeedReplay: false,
		DecodeBody: string(importJson),
	}

	if cmd.CommandContext != nil {
		replayData.TraceId = cmd.CommandContext.TraceId
	}

	// bin2 (CMD_PING=2) 时自动更新 host（受配置控制）
	// config.GetConfig() 读取 Nacos 热更新后的最新配置
	if config.GetConfig().Xyy.Bin2UpdateHost && cmd.MessageHeader != nil && cmd.MessageHeader.Cmd == 2 {
		if cmd.CommandContext != nil && cmd.CommandContext.Ip != "" {
			regService.UpdateHost(cmd.Imei, cmd.CommandContext.Ip, isShadowMode)
		}
	}

	c.JSON(http.StatusOK, dto.Result{Success: true, Code: "0", Msg: "成功", Data: replayData})
}

func handleReplay(c *gin.Context) {
	var cmd dto.ReplayCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusOK, dto.Result{Success: false, Code: "10002", Msg: err.Error()})
		return
	}

	logger.Log.Info("device replay received",
		zap.String("imei", cmd.Imei),
		zap.String("jobId", cmd.JobId),
		zap.String("data", cmd.Data),
		zap.String("payload", cmd.Payload),
	)

	// In Java, DeviceReplyMessage extends DeviceReportMessage.
	// messagePusher.pushMessage(imei, deviceReplyMessage) sends the DeviceReplyMessage directly,
	// so Kafka receives a flat JSON: {imei, msgType, bussinessType, data, msgId, identifier, payload}.
	replyMsg := &dto.DeviceReplyMessage{
		Imei:          cmd.Imei,
		MsgType:       "event",
		BussinessType: "ebike",
		Data:          utils.AddWGS84(cmd.Data),
		MsgId:         cmd.JobId,
		Payload:       cmd.Payload,
	}

	isShadowMode := c.GetBool("isShadowMode")
	// Push to Kafka - DeviceReplyMessage is pushed directly (not wrapped in DeviceReportMessage)
	regService.Pusher.PushReplyMessage(cmd.Imei, replyMsg, isShadowMode)

	// Java writes Redis unconditionally: redisMapper.set(key, cmd.getData(), timeout, timeUnit)
	// We must do the same — don't skip on empty Data
	if cmd.JobId != "" {
		if isShadowMode {
			logger.Log.Info("shadow mode intercept redis replay write", zap.String("jobId", cmd.JobId))
		} else {
			replayKey := "ecu_replay_" + cmd.JobId
			regService.Rdb.Set(c.Request.Context(), replayKey, cmd.Data, 2*time.Minute)
			logger.Log.Info("replay redis write", zap.String("key", replayKey), zap.String("value", cmd.Data))
		}
	}

	c.JSON(http.StatusOK, dto.Result{Success: true, Code: "0", Msg: "成功", Data: "done"})
}
