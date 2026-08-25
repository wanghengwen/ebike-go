package sender

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"push-notification-go/internal/model"
	"push-notification-go/internal/pkg/config"
	"push-notification-go/internal/pkg/logger"
	"push-notification-go/internal/pkg/templateutil"
)

// ChuanglanSMSHandler sends SMS via the Chuanglan (253) HTTP API.
type ChuanglanSMSHandler struct{}

// chuanglanRequest is the JSON body sent to the Chuanglan API.
type chuanglanRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Msg      string `json:"msg"`
	Phone    string `json:"phone"`
	Report   string `json:"report"`
}

// chuanglanResponse is the JSON body returned by the Chuanglan API.
type chuanglanResponse struct {
	Time       string `json:"time"`
	MsgId      string `json:"msgId"`
	ErrorMsg   string `json:"errorMsg"`
	FailNum    string `json:"failNum"`
	SuccessNum string `json:"successNum"`
	Code       string `json:"code"`
}

func (h *ChuanglanSMSHandler) Send(data *SendData) *model.SendResult {
	ctx := data.Ctx
	apiUrl := config.GlobalConfig.Xyy.Chuanglan.ApiUrl
	if apiUrl == "" {
		apiUrl = "https://smssh1.253.com/msg/v1/send/json"
	}

	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(data.ConfigJson), &configMap); err != nil {
		logger.Printf(ctx, "[ChuanglanSMS] failed to parse configJson: %v", err)
		return model.FailResult("解析配置异常")
	}

	account := configMap["account"]
	password := configMap["password"]

	// Build message: 【签名】+ rendered template content
	msg := "【" + data.SignName + "】" + templateutil.RenderTemplate(data.TemplateContent, data.Params)

	reqBody := &chuanglanRequest{
		Account:  account,
		Password: password,
		Msg:      msg,
		Phone:    data.Phone,
		Report:   "true",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		logger.Printf(ctx, "[ChuanglanSMS] failed to marshal request: %v", err)
		return model.FailResult("创建请求体异常")
	}

	logger.Printf(ctx, "[ChuanglanSMS] sending phone=%s", data.Phone)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(apiUrl, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		logger.Printf(ctx, "[ChuanglanSMS] HTTP request failed phone=%s err=%v", data.Phone, err)
		return model.FailResult("调用创蓝发送短信异常")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf(ctx, "[ChuanglanSMS] HTTP status not OK phone=%s status=%d", data.Phone, resp.StatusCode)
		return model.FailResult("创蓝返回响应为null")
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf(ctx, "[ChuanglanSMS] failed to read response phone=%s err=%v", data.Phone, err)
		return model.FailResult("读取创蓝响应异常")
	}

	logger.Printf(ctx, "[ChuanglanSMS] response phone=%s body=%s", data.Phone, string(respBytes))

	var clResp chuanglanResponse
	if err := json.Unmarshal(respBytes, &clResp); err != nil {
		logger.Printf(ctx, "[ChuanglanSMS] failed to parse response phone=%s err=%v", data.Phone, err)
		return model.FailResult("解析创蓝响应异常")
	}

	if !strings.EqualFold(clResp.Code, "0") {
		return model.FailResultWithID(clResp.ErrorMsg, clResp.MsgId)
	}

	return model.SuccessResult(clResp.MsgId)
}
