package auth

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	weixinAccessTokenCacheTTL = 7000 * time.Second
	weixinJsapiTicketCacheTTL = 7000 * time.Second
	defaultWeixinAPIBaseURL   = "https://api.weixin.qq.com"
)

type JsApiSignatureCmd struct {
	URL string `json:"url"`
}

type JsApiSignatureDto struct {
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	URL       string `json:"url"`
}

type weixinAPIBaseRes struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type weixinAccessTokenRes struct {
	weixinAPIBaseRes
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type weixinJsapiTicketRes struct {
	weixinAPIBaseRes
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
}

func ClientJsApiSignatureHandler(c *gin.Context) {
	var cmd JsApiSignatureCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		sendError(c, CodeIllegalArgument, "Invalid request body format")
		return
	}

	// 与 Java 对齐：controller 不校验 url，且获取失败时 service 返回 null，
	// 最终响应为 success=true、data=null（见 JsApiSignatureService.getSignature）。
	dto, err := getJsApiSignature(c.Request.Context(), cmd.URL)
	if err != nil {
		logger.WithContext(c).Error("get jsapi signature failed", zap.Error(err))
		sendSuccess(c, nil)
		return
	}
	sendSuccess(c, dto)
}

func getJsApiSignature(ctx context.Context, pageURL string) (*JsApiSignatureDto, error) {
	cfg := config.AppConfig.WeixinPublicPlatform
	if cfg.AppId == "" || cfg.AppSecret == "" {
		return nil, fmt.Errorf("weixin public platform appid or appSecret is not configured")
	}

	accessToken, err := getWeixinAccessToken(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if accessToken == "" {
		return nil, fmt.Errorf("weixin access_token is empty")
	}

	ticket, err := getWeixinJsapiTicket(ctx, cfg, accessToken)
	if err != nil {
		return nil, err
	}
	if ticket == "" {
		return nil, fmt.Errorf("weixin jsapi_ticket is empty")
	}

	return buildJsApiSignatureDto(ticket, pageURL), nil
}

func getWeixinAccessToken(ctx context.Context, cfg config.WeixinPublicPlatformConfig) (string, error) {
	cacheKey := fmt.Sprintf("weixin:access_token:%s", cfg.AppId)
	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		return cached, nil
	}

	apiURL := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		weixinAPIBaseURL(cfg), url.QueryEscape(cfg.AppId), url.QueryEscape(cfg.AppSecret))
	resp, err := restyClient.R().SetContext(ctx).Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("request weixin access_token: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("request weixin access_token: http %d", resp.StatusCode())
	}

	var tokenRes weixinAccessTokenRes
	if err := json.Unmarshal(resp.Body(), &tokenRes); err != nil {
		return "", fmt.Errorf("decode weixin access_token response: %w", err)
	}
	if tokenRes.ErrCode != 0 {
		return "", fmt.Errorf("weixin access_token error: errcode=%d errmsg=%s", tokenRes.ErrCode, tokenRes.ErrMsg)
	}
	if tokenRes.AccessToken == "" {
		return "", fmt.Errorf("weixin access_token is empty in response")
	}

	_ = redis.Set(ctx, cacheKey, tokenRes.AccessToken, weixinAccessTokenCacheTTL)
	return tokenRes.AccessToken, nil
}

func getWeixinJsapiTicket(ctx context.Context, cfg config.WeixinPublicPlatformConfig, accessToken string) (string, error) {
	md5Hex := md5HexString(accessToken)
	cacheKey := fmt.Sprintf("weixin:jsapi_ticket:%s", md5Hex)
	if cached, err := redis.Get(ctx, cacheKey); err == nil && cached != "" {
		return cached, nil
	}

	apiURL := fmt.Sprintf("%s/cgi-bin/ticket/getticket?access_token=%s&type=jsapi",
		weixinAPIBaseURL(cfg), url.QueryEscape(accessToken))
	resp, err := restyClient.R().SetContext(ctx).Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("request weixin jsapi_ticket: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("request weixin jsapi_ticket: http %d", resp.StatusCode())
	}

	var ticketRes weixinJsapiTicketRes
	if err := json.Unmarshal(resp.Body(), &ticketRes); err != nil {
		return "", fmt.Errorf("decode weixin jsapi_ticket response: %w", err)
	}
	if ticketRes.ErrCode != 0 {
		return "", fmt.Errorf("weixin jsapi_ticket error: errcode=%d errmsg=%s", ticketRes.ErrCode, ticketRes.ErrMsg)
	}
	if ticketRes.Ticket == "" {
		return "", fmt.Errorf("weixin jsapi_ticket is empty in response")
	}

	_ = redis.Set(ctx, cacheKey, ticketRes.Ticket, weixinJsapiTicketCacheTTL)
	return ticketRes.Ticket, nil
}

func buildJsApiSignatureDto(ticket, pageURL string) *JsApiSignatureDto {
	nonce := generateJsApiNonce()
	timestamp := time.Now().Unix()
	return &JsApiSignatureDto{
		Signature: generateJsApiSignature(ticket, nonce, timestamp, pageURL),
		Timestamp: timestamp,
		Nonce:     nonce,
		URL:       pageURL,
	}
}

func generateJsApiSignature(ticket, nonce string, timestamp int64, pageURL string) string {
	raw := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticket, nonce, timestamp, pageURL)
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateJsApiNonce() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func md5HexString(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func weixinAPIBaseURL(cfg config.WeixinPublicPlatformConfig) string {
	if cfg.APIBaseURL != "" {
		return cfg.APIBaseURL
	}
	return defaultWeixinAPIBaseURL
}
