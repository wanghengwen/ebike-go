package sender

import (
	"encoding/json"
	"strings"

	"push-notification-go/internal/model"
	"push-notification-go/internal/pkg/logger"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunSMSHandler sends SMS via Alibaba Cloud Dysmsapi.
type AliyunSMSHandler struct{}

func (h *AliyunSMSHandler) Send(data *SendData) *model.SendResult {
	ctx := data.Ctx
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(data.ConfigJson), &configMap); err != nil {
		logger.Printf(ctx, "[AliyunSMS] failed to parse configJson: %v", err)
		return model.FailResult("解析配置异常")
	}

	accessKeyId := configMap["accessKeyId"]
	accessKeySecret := configMap["accessKeySecret"]

	client, err := createSMSClient(accessKeyId, accessKeySecret)
	if err != nil {
		logger.Printf(ctx, "[AliyunSMS] failed to create client: %v", err)
		return model.FailResult("创建阿里云SMS客户端异常")
	}

	paramsJSON, _ := json.Marshal(data.Params)

	sendSmsRequest := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(data.Phone),
		SignName:      tea.String(data.SignName),
		TemplateCode:  tea.String(data.SupplierTemplateCode),
		TemplateParam: tea.String(string(paramsJSON)),
	}

	runtime := &util.RuntimeOptions{}
	resp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		logger.Printf(ctx, "[AliyunSMS] call failed phone=%s err=%v", data.Phone, err)
		return model.FailResult("调用阿里云发送短信异常")
	}

	if resp == nil || resp.Body == nil {
		logger.Printf(ctx, "[AliyunSMS] response is nil phone=%s", data.Phone)
		return model.FailResult("阿里云返回响应为null")
	}

	logger.Printf(ctx, "[AliyunSMS] response phone=%s code=%s requestId=%s",
		data.Phone, tea.StringValue(resp.Body.Code), tea.StringValue(resp.Body.RequestId))

	if !strings.EqualFold(tea.StringValue(resp.Body.Code), "OK") {
		return model.FailResultWithID(tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId))
	}

	return model.SuccessResult(tea.StringValue(resp.Body.RequestId))
}

func createSMSClient(accessKeyId, accessKeySecret string) (*dysmsapi.Client, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}
	return dysmsapi.NewClient(cfg)
}
