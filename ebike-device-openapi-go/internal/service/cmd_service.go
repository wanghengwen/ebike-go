package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ecuHTTPClient     *http.Client
	ecuHTTPClientOnce sync.Once
)

// InitECUHTTPClient initializes the shared HTTP client for ECU gateway calls.
// Timeouts align with Java RestTemplateConfig (connect 2s, read 15s).
func InitECUHTTPClient() {
	ecuHTTPClientOnce.Do(func() {
		refreshECUHTTPClient()
	})
}

func refreshECUHTTPClient() {
	connectTimeout := config.GetConnectTimeout()
	readTimeout := config.GetReadTimeout()
	ecuHTTPClient = &http.Client{
		Timeout: readTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: connectTimeout,
			}).DialContext,
		},
	}
}

func getECUHTTPClient() *http.Client {
	if ecuHTTPClient == nil {
		InitECUHTTPClient()
	}
	return ecuHTTPClient
}

// DoAction mimics EbikeCmdServiceImpl.doAction.
// ecuLogin selects transport: xiaoan → gateway /ecu/wild; luoping → EMQX publish.
func DoAction(req interface{}, ebikeReq *dto.EBikeRequest, cmdID int16, ecuLogin *dto.EcuLogin, dryRun bool, shadowMode bool) (interface{}, error) {
	// Generate TraceID if empty
	ctx := ebikeReq.CommandContext
	if ctx == nil {
		ctx = &dto.CommandContext{}
	}
	if ctx.TraceId == "" {
		if ebikeReq.TraceId != "" {
			ctx.TraceId = ebikeReq.TraceId
		} else {
			ctx.TraceId = strings.ReplaceAll(uuid.New().String(), "-", "")
		}
	}

	jobId := strings.ReplaceAll(uuid.New().String(), "-", "")

	// Build WildCmd (also used as the canonical cmd/params carrier for luoping)
	wildCmd := &dto.WildCmd{
		Imei:           ebikeReq.Imei,
		Cmd:            cmdID,
		Async:          ebikeReq.Async,
		JobId:          jobId,
		CommandContext: ctx,
		Ip:             config.GetLocalIP(), // use actual pod IP so gateway can callback
		Port:           config.GetHTTPServerPort(),
	}

	// Java EbikeCmdServiceImpl: dt/tm are set only when BOTH are non-null.
	if ebikeReq.Dt != nil && ebikeReq.Tm != nil {
		wildCmd.Dt = ebikeReq.Dt
		wildCmd.Tm = ebikeReq.Tm
	}

	if ebikeReq.Payload != nil {
		switch v := ebikeReq.Payload.(type) {
		case string:
			wildCmd.Payload = v
		default:
			b, err := json.Marshal(v)
			if err == nil {
				wildCmd.Payload = string(b)
			} else {
				wildCmd.Payload = fmt.Sprintf("%v", v)
			}
		}
	}

	// Extract subclasses specific fields
	params := dto.ExtractParams(req)
	if len(params) > 0 {
		wildCmd.Params = params
	}

	// === 新增：如果是 DryRun 模式，直接返回生成的 WildCmd，不进行实际下发 ===
	if dryRun {
		return wildCmd, nil
	}

	if shadowMode {
		logger.Log.Info("shadow mode intercept command", zap.Any("wildCmd", wildCmd))
		// Mock success response
		transmissionBody := &dto.TransmissionBody{
			Async:   wildCmd.Async,
			JobId:   jobId,
			Result:  "",
			Payload: "",
		}
		return transmissionBody, nil
	}

	deviceType := dto.DeviceTypeXiaoan
	if ecuLogin != nil && ecuLogin.Type != "" {
		deviceType = ecuLogin.Type
	}

	if strings.EqualFold(deviceType, dto.DeviceTypeLuoping) {
		return sendLuopingMsg(wildCmd, ecuLogin)
	}

	if ecuLogin == nil {
		return nil, fmt.Errorf("设备离线或操作失败")
	}

	// Send to ECU Gateway (xiaoan)
	wildResult, err := SendMsg(wildCmd, ecuLogin.Host, ecuLogin.Port)
	if err != nil {
		return nil, fmt.Errorf("设备离线或操作失败: %w", err)
	}
	if wildResult == nil {
		return nil, fmt.Errorf("设备离线或操作失败")
	}

	transmissionBody := &dto.TransmissionBody{
		Async:  wildCmd.Async,
		JobId:  jobId,
		Result: wildResult.Json,
	}
	if wildResult.Payload != "" {
		transmissionBody.Payload = wildResult.Payload
	}

	return transmissionBody, nil
}

func sendLuopingMsg(wildCmd *dto.WildCmd, ecuLogin *dto.EcuLogin) (interface{}, error) {
	if ecuLogin == nil {
		return nil, fmt.Errorf("luoping ecuLogin missing")
	}

	// V8: same 小安 wild frame as gateway /ecu/wild (AA55|cmd=0|seq|JSON),
	// encrypted and published to ecu/bcmd/ebike/{deviceId}. Sync wait uses seq→jobId.
	jobId, err := PublishLuopingWildCmd(context.Background(), ecuLogin, wildCmd)
	if err != nil {
		logger.Log.Error("luoping publish failed",
			zap.String("imei", wildCmd.Imei),
			zap.String("deviceId", ecuLogin.DeviceId),
			zap.Int16("cmd", wildCmd.Cmd),
			zap.String("jobId", wildCmd.JobId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("luoping publish failed: %w", err)
	}

	body := &dto.TransmissionBody{
		Async:  wildCmd.Async,
		JobId:  jobId,
		Result: "",
	}
	if wildCmd.Async {
		return body, nil
	}

	timeout := config.GetMqttSyncWait()
	if wildCmd.Timeout != nil && *wildCmd.Timeout > 0 {
		timeout = time.Duration(*wildCmd.Timeout) * time.Millisecond
	}
	reply, waitErr := waitLuopingReply(jobId, timeout)
	if waitErr != nil {
		logger.Log.Error("luoping sync wait failed",
			zap.String("imei", wildCmd.Imei),
			zap.String("deviceId", ecuLogin.DeviceId),
			zap.Int16("cmd", wildCmd.Cmd),
			zap.String("jobId", jobId),
			zap.Duration("timeout", timeout),
			zap.Error(waitErr),
		)
		return nil, waitErr
	}
	body.Result = reply
	body.Payload = reply
	return body, nil
}

func SendMsg(wildCmd *dto.WildCmd, host string, port int) (*dto.WildResult, error) {
	url := fmt.Sprintf("http://%s:%d/ecu/wild", host, port)

	bodyBytes, err := json.Marshal(wildCmd)
	if err != nil {
		return nil, err
	}

	if wildCmd.Cmd == 33 {
		logger.Log.Info("lock_command_ecu_request",
			zap.String("url", url),
			zap.String("imei", wildCmd.Imei),
			zap.String("jobId", wildCmd.JobId),
			zap.Bool("async", wildCmd.Async),
			zap.Any("params", wildCmd.Params),
			zap.String("traceId", traceIDFromWildCmd(wildCmd)),
			zap.ByteString("body", bodyBytes),
		)
	} else {
		logger.Log.Info("sendMsg request",
			zap.String("url", url),
			zap.String("imei", wildCmd.Imei),
			zap.Int16("cmd", wildCmd.Cmd),
			zap.ByteString("body", bodyBytes),
		)
	}

	resp, err := getECUHTTPClient().Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if wildCmd.Cmd == 33 {
		logger.Log.Info("lock_command_ecu_response",
			zap.Int("status", resp.StatusCode),
			zap.String("imei", wildCmd.Imei),
			zap.String("jobId", wildCmd.JobId),
			zap.ByteString("body", respBody),
		)
	} else {
		logger.Log.Info("sendMsg response",
			zap.Int("status", resp.StatusCode),
			zap.String("imei", wildCmd.Imei),
			zap.Int16("cmd", wildCmd.Cmd),
			zap.ByteString("body", respBody),
		)
	}

	var dsResult dto.DownstreamResult
	if err := json.Unmarshal(respBody, &dsResult); err != nil {
		return nil, err
	}

	if dsResult.Code != "200" {
		return nil, fmt.Errorf("gateway returned error: code=%s message=%s", dsResult.Code, dsResult.Message)
	}

	if dsResult.Data != nil && dsResult.Data.Json != "" {
		dsResult.Data.Json = utils.AddWGS84(dsResult.Data.Json)
	}

	return dsResult.Data, nil
}

func traceIDFromWildCmd(wildCmd *dto.WildCmd) string {
	if wildCmd == nil || wildCmd.CommandContext == nil {
		return ""
	}
	return wildCmd.CommandContext.TraceId
}
