package sender

import (
	"encoding/json"
	"strings"

	"push-notification-go/internal/model"
	"push-notification-go/internal/pkg/logger"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dyvmsapi "github.com/alibabacloud-go/dyvmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunVoiceHandler sends voice notifications via Alibaba Cloud Dyvmsapi.
type AliyunVoiceHandler struct{}

func (h *AliyunVoiceHandler) Send(data *SendData) *model.SendResult {
	ctx := data.Ctx
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(data.ConfigJson), &configMap); err != nil {
		logger.Printf(ctx, "[AliyunVoice] failed to parse configJson: %v", err)
		return model.FailResult("解析配置异常")
	}

	accessKeyId := configMap["accessKeyId"]
	accessKeySecret := configMap["accessKeySecret"]

	client, err := createVoiceClient(accessKeyId, accessKeySecret)
	if err != nil {
		logger.Printf(ctx, "[AliyunVoice] failed to create client: %v", err)
		return model.FailResult("创建阿里云Voice客户端异常")
	}

	paramsJSON, _ := json.Marshal(data.Params)

	request := &dyvmsapi.SingleCallByTtsRequest{
		CalledNumber: tea.String(data.Phone),
		TtsCode:      tea.String(data.SupplierTemplateCode),
		TtsParam:     tea.String(string(paramsJSON)),
		Speed:        tea.Int32(-200),
	}

	runtime := &util.RuntimeOptions{}
	resp, err := client.SingleCallByTtsWithOptions(request, runtime)
	if err != nil {
		logger.Printf(ctx, "[AliyunVoice] call failed phone=%s err=%v", data.Phone, err)
		return model.FailResult("调用阿里云发送语音异常")
	}

	if resp == nil || resp.Body == nil {
		logger.Printf(ctx, "[AliyunVoice] response is nil phone=%s", data.Phone)
		return model.FailResult("阿里云返回响应为null")
	}

	logger.Printf(ctx, "[AliyunVoice] response phone=%s code=%s requestId=%s",
		data.Phone, tea.StringValue(resp.Body.Code), tea.StringValue(resp.Body.RequestId))

	if !strings.EqualFold(tea.StringValue(resp.Body.Code), "OK") {
		return model.FailResultWithID(tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId))
	}

	return model.SuccessResult(tea.StringValue(resp.Body.RequestId))
}

func createVoiceClient(accessKeyId, accessKeySecret string) (*dyvmsapi.Client, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("dyvmsapi.aliyuncs.com"),
	}
	return dyvmsapi.NewClient(cfg)
}
