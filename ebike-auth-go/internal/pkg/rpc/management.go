package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ebike-auth-go/internal/pkg/logger"

	"go.uber.org/zap"
)

type CommandContext struct {
	TraceId       string `json:"traceId"`
	TenantId      string `json:"tenantId,omitempty"`
	Pin           string `json:"pin,omitempty"`
	Source        string `json:"source,omitempty"`
	StressTesting bool   `json:"stressTesting,omitempty"`
	Ip            string `json:"ip,omitempty"`
	Platform      string `json:"platform,omitempty"`
	DeviceId      string `json:"deviceId,omitempty"`
}

type QueryAllTenantAuthQry struct {
	CommandContext CommandContext `json:"commandContext"`
}

type TenantAuthCo struct {
	TenantId             string `json:"tenantId"`
	TenantSecret         string `json:"tenantSecret"`
	AccessTokenValidity  int    `json:"accessTokenValidity"`
	RefreshTokenValidity int    `json:"refreshTokenValidity"`
	AuthorizedGrantTypes string `json:"authorizedGrantTypes"`
	Scope                string `json:"scope"`
}

type ManagementResult[T any] struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type QueryAllTenantThirdAuthQry struct {
	CommandContext CommandContext `json:"commandContext"`
}

type TenantThirdAuthCo struct {
	TenantId     string `json:"tenantId"`
	ThirdType    int    `json:"thirdType"`
	AppId        string `json:"appId"`
	AppSecret    string `json:"appSecret"`
	PayAppId     string `json:"payAppId"`
	PayAppSecret string `json:"payAppSecret"`
}

func getManagementBaseURL() string {
	if url := os.Getenv("EBIKE_MANAGEMENT_URL"); url != "" {
		return url
	}
	url, err := ResolveServiceAddress("ebike-management")
	if err != nil {
		logger.Log.Debug("Nacos resolve ebike-management failed, using fallback URL", zap.Error(err))
		return "http://ebike-management:8080"
	}
	return url
}

func QueryAllTenantAuth() ([]TenantAuthCo, error) {
	url := fmt.Sprintf("%s/tenant/tenantAuth/queryAll", getManagementBaseURL())

	req := QueryAllTenantAuthQry{
		CommandContext: CommandContext{
			TraceId:  fmt.Sprintf("trace-%d", time.Now().UnixNano()),
			TenantId: "0",
		},
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, err
	}

	var result ManagementResult[[]TenantAuthCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("parse result error: %w, raw response: %s", err, string(resp.Body()))
	}

	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("management RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return result.Data, nil
}

func QueryAllTenantThirdAuth() ([]TenantThirdAuthCo, error) {
	url := fmt.Sprintf("%s/tenant/tenantThirdAuth/queryAll", getManagementBaseURL())

	req := QueryAllTenantThirdAuthQry{
		CommandContext: CommandContext{
			TraceId:  fmt.Sprintf("trace-%d", time.Now().UnixNano()),
			TenantId: "0",
		},
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, err
	}

	var result ManagementResult[[]TenantThirdAuthCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("parse result error: %w, raw response: %s", err, string(resp.Body()))
	}

	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("management RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return result.Data, nil
}
