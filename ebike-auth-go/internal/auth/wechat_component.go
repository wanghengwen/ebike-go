package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/utils"
)

type componentAccessTokenResp struct {
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int    `json:"expires_in"`
	Errcode              int    `json:"errcode"`
	Errmsg               string `json:"errmsg"`
}

func getWechatComponentAccessToken(ctx context.Context) (string, error) {
	cached, _ := redis.Get(ctx, redis.GetComponentAccessTokenKey())
	if cached != "" {
		return cached, nil
	}

	componentAppId := config.AppConfig.WechatComponentAppId
	componentAppSecret := config.AppConfig.WechatComponentAppSecret
	if componentAppId == "" || componentAppSecret == "" {
		return "", fmt.Errorf("wechat component config not found")
	}

	ticket, _ := redis.Get(ctx, redis.GetComponentVerifyTicketKey())
	if ticket == "" {
		return "", fmt.Errorf("component verify ticket not exist")
	}

	token, err := utils.FetchWechatComponentAccessToken(componentAppId, componentAppSecret, ticket)
	if err != nil {
		return "", err
	}
	_ = redis.Set(ctx, redis.GetComponentAccessTokenKey(), token, 110*time.Minute)
	return token, nil
}

func wechatThirdMiniAppSession(ctx context.Context, appId, jsCode string) (*utils.WechatSession, error) {
	componentToken, err := getWechatComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return utils.WechatComponentJscode2Session(
		appId,
		jsCode,
		config.AppConfig.WechatComponentAppId,
		componentToken,
	)
}

func decryptWechatPhone(sessionKey, encryptedData, iv string) (string, error) {
	decrypted, err := utils.DecryptWechatData(sessionKey, encryptedData, iv)
	if err != nil {
		return "", err
	}
	var wechatPhone struct {
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	}
	if err := json.Unmarshal(decrypted, &wechatPhone); err != nil {
		return "", err
	}
	return fmt.Sprintf("+%s-%s", wechatPhone.CountryCode, wechatPhone.PurePhoneNumber), nil
}
