package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/rpc"
	"ebike-auth-go/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthCmd struct {
	GrantType    string `json:"grant_type"`
	Phone        string `json:"phone,omitempty"`
	MessageCode  string `json:"messageCode,omitempty"`
	Email        string `json:"email,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Key          string `json:"key,omitempty"`
	Secret       string `json:"secret,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`

	// ClientDTO fields
	TraceId  string `json:"traceId,omitempty"`
	TenantId string `json:"tenantId,omitempty"`
	Platform string `json:"platform,omitempty"`
	DeviceId string `json:"deviceId,omitempty"`
}

type TokenCo struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	TokenType    string   `json:"tokenType"`
	ExpiresIn    int      `json:"expiresIn"`
	Scope        []string `json:"scope"`
	Pin          string   `json:"pin"`
	Nickname     string   `json:"nickname"`
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
}

func getPlatformName(platform string) string {
	if platform == "android" || platform == "ios" || platform == "Android" || platform == "iOS" || platform == "IOS" {
		return "app"
	}
	if platform == "" {
		return "other"
	}
	return platform
}

func verifyClient(c *gin.Context) (string, error) {
	var clientId, clientSecret string

	// 1. Basic Auth Header
	username, password, ok := c.Request.BasicAuth()
	if ok {
		clientId = username
		clientSecret = password
	} else {
		// 2. Query / Form Params
		clientId = c.PostForm("client_id")
		if clientId == "" {
			clientId = c.Query("client_id")
		}
		clientSecret = c.PostForm("client_secret")
		if clientSecret == "" {
			clientSecret = c.Query("client_secret")
		}
	}

	if clientId == "" {
		return "", fmt.Errorf("missing client_id")
	}

	tenantAuth, err := GetTenantAuth(clientId)
	if err != nil || tenantAuth == nil {
		return "", fmt.Errorf("invalid client_id: %s", clientId)
	}

	// Password encoder matching "xyy"
	if !utils.MatchesSHA256(clientSecret, utils.Sha256("xyy")) {
		return "", fmt.Errorf("invalid client_secret")
	}

	return clientId, nil
}

func BusinessTokenHandler(c *gin.Context) {
	clientId, err := verifyClient(c)
	if err != nil {
		logger.Log.Warn("Client verification failed", zap.Error(err))
		sendError(c, CodeAccessUnauthorized, err.Error())
		return
	}

	var cmd AuthCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		sendError(c, CodeIllegalArgument, "Invalid request body format")
		return
	}

	// Get tenant configuration
	tenantAuth, _ := GetTenantAuth(clientId)

	var user *rpc.UserDO
	var scopes []string
	if tenantAuth.Scope != "" {
		scopes = []string{tenantAuth.Scope} // Simplify scopes mapping
	}

	platform := cmd.Platform
	if platform == "" {
		platform = c.GetHeader("platform")
	}

	deviceId := cmd.DeviceId
	if deviceId == "" {
		deviceId = c.GetHeader("deviceId")
	}
	if deviceId == "" {
		deviceId = "normal"
	}

	tenantId := cmd.TenantId
	if tenantId == "" {
		tenantId = c.GetHeader("tenantId")
	}
	if tenantId == "" {
		tenantId = clientId
	}

	// Set tracing properties on context
	c.Set("tenantId", tenantId)
	c.Set("traceId", cmd.TraceId)

	// Getting access token for: 1
	logger.WithContext(c).Info(fmt.Sprintf("Getting access token for: %s", tenantId))

	cmdCtx := rpc.NewCommandContext(tenantId, cmd.TraceId, platform, deviceId, "")

	if err := validateAuthorizedGrantType(tenantAuth, cmd.GrantType); err != nil {
		sendError(c, CodeInvalidGrant, err.Error())
		return
	}

	switch cmd.GrantType {
	case "phone_code":
		if cmd.Phone == "" || cmd.MessageCode == "" {
			sendError(c, CodeIllegalArgument, "phone and messageCode are required")
			return
		}
		logger.WithContext(c).Info("Getting access token via phone_code", zap.String("phone", cmd.Phone), zap.String("tenantId", tenantId))
		ok, err := rpc.VerifyPhoneMessageCode(c, cmdCtx, "1", cmd.Phone, cmd.MessageCode)
		if err != nil || !ok {
			logger.WithContext(c).Warn("Message code verify failed", zap.String("phone", cmd.Phone), zap.Error(err))
			sendError(c, CodeMessageCodeVerifyFailed, "验证码校验失败")
			return
		}
		user, err = rpc.BusinessGetUserByPhone(c, cmdCtx, cmd.Phone)
		if err != nil || user == nil {
			logger.WithContext(c).Warn("User not found by phone", zap.String("phone", cmd.Phone), zap.Error(err))
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		logger.WithContext(c).Info("User fetched successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("phone", user.Phone))
	case "email_code":
		if cmd.Email == "" || cmd.MessageCode == "" {
			sendError(c, CodeIllegalArgument, "email and messageCode are required")
			return
		}
		logger.WithContext(c).Info("Getting access token via email_code", zap.String("email", cmd.Email), zap.String("tenantId", tenantId))
		ok, err := rpc.VerifyEmailMessageCode(c, cmdCtx, "1", cmd.Email, cmd.MessageCode)
		if err != nil || !ok {
			logger.WithContext(c).Warn("Email code verify failed", zap.String("email", cmd.Email), zap.Error(err))
			sendError(c, CodeMessageCodeVerifyFailed, "验证码校验失败")
			return
		}
		user, err = rpc.BusinessGetUserByEmail(c, cmdCtx, cmd.Email)
		if err != nil || user == nil {
			logger.WithContext(c).Warn("User not found by email", zap.String("email", cmd.Email), zap.Error(err))
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		logger.WithContext(c).Info("User fetched successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("email", user.Email))
	case "password":
		if cmd.Username == "" || cmd.Password == "" {
			sendError(c, CodeIllegalArgument, "username and password are required")
			return
		}
		logger.WithContext(c).Info("Getting access token via password", zap.String("username", cmd.Username), zap.String("tenantId", tenantId))
		user, err = rpc.BusinessGetUserByPin(c, cmdCtx, cmd.Username)
		if err != nil || user == nil {
			logger.WithContext(c).Warn("User not found by pin", zap.String("username", cmd.Username), zap.Error(err))
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		logger.WithContext(c).Info("User fetched successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("phone", user.Phone))
		if !utils.MatchesSHA256(cmd.Password, user.Password) {
			sendError(c, CodeInvalidGrant, "用户名和密码不匹配")
			return
		}

	case "ak":
		if cmd.Key == "" || cmd.Secret == "" {
			sendError(c, CodeIllegalArgument, "key and secret are required")
			return
		}
		// In business mode, load aks map from configurations.
		if config.AppConfig.Aks == nil || config.AppConfig.Aks[cmd.Key] != cmd.Secret {
			sendError(c, CodeAuthenticationFailed, "AK/SK验证失败")
			return
		}
		user = &rpc.UserDO{
			Pin:         cmd.Key,
			Nickname:    cmd.Key,
			Authorities: []string{"NORMAL"},
			Status:      1,
		}
		c.Set("pin", user.Pin)

	case "phone_secret":
		if cmd.Phone == "" || cmd.Secret == "" {
			sendError(c, CodeIllegalArgument, "phone and secret are required")
			return
		}
		// String secret2 = Hashing.sha256().newHasher().putString(phone+"_xyy@2022", Charsets.UTF_8).hash().toString();
		expectedSecret := utils.Sha256(cmd.Phone + "_xyy@2022")
		if cmd.Secret != expectedSecret {
			sendError(c, CodeException, "手机秘钥不匹配")
			return
		}
		user, err = rpc.BusinessGetUserByPhone(c, cmdCtx, cmd.Phone)
		if err != nil || user == nil {
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)

	case "phone_password":
		if cmd.Phone == "" || cmd.Password == "" {
			sendError(c, CodeIllegalArgument, "phone and password are required")
			return
		}
		user, err = rpc.BusinessGetUserByPhone(c, cmdCtx, cmd.Phone)
		if err != nil || user == nil {
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		if !utils.MatchesSHA256(cmd.Password, user.Password) {
			sendError(c, CodeMessageCodeVerifyFailed, "用户名和密码不匹配")
			return
		}

	case "refresh_token":
		if cmd.RefreshToken == "" {
			sendError(c, CodeIllegalArgument, "refresh_token is required")
			return
		}

		claims, err := ParseToken(cmd.RefreshToken)
		if err != nil {
			sendError(c, CodeInvalidToken, "登录已过期，请重新登录(03)")
			return
		}

		// Redis verification matching login device session
		redisKey := redis.GetLoginDeviceKey(getPlatformName(claims.Platform), claims.UserName)
		cachedDevice, err := redis.Get(c, redisKey)
		if err != nil || cachedDevice == "" {
			sendError(c, CodeInvalidToken, "登录已过期，请重新登录(04)")
			return
		}
		if cachedDevice == "role-change" {
			sendError(c, CodeUserRoleChanged, "账号权限已变更，请重新登录(02)")
			return
		}
		if cachedDevice != claims.DeviceId {
			sendError(c, CodeOtherDeviceLogin, "账号已在别处登录(02)")
			return
		}

		user, err = rpc.BusinessGetUserByPin(c, cmdCtx, claims.UserName)
		if err != nil || user == nil {
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		// Set device and platform from claims for token generation
		deviceId = claims.DeviceId
		platform = claims.Platform

	default:
		sendError(c, CodeInvalidGrant, "Unsupported grant type: "+cmd.GrantType)
		return
	}

	if !checkBusinessPCPermission(platform, user) {
		sendError(c, CodeUserNoPermission, "")
		return
	}

	if user.Status != 1 {
		sendError(c, CodeUserDisabled, "该账号已被禁用")
		return
	}

	// Generate Access and Refresh Token
	// Expire values derived from Tenant settings
	accessTokenValidity := tenantAuth.AccessTokenValidity
	if accessTokenValidity <= 0 {
		accessTokenValidity = 7200 // 2 hours
	}
	refreshTokenValidity := tenantAuth.RefreshTokenValidity
	if refreshTokenValidity <= 0 {
		refreshTokenValidity = 2592000 // 30 days
	}

	accessToken, refreshToken, err := GenerateTokenPair(
		user.Pin,
		clientId,
		platform,
		deviceId,
		user.Nickname,
		user.Avatar,
		"",
		scopes,
		user.Authorities,
		cmd.GrantType,
		accessTokenValidity,
		refreshTokenValidity,
	)

	if err != nil {
		logger.WithContext(c).Error("Failed to generate tokens", zap.Error(err), zap.String("pin", user.Pin))
		sendError(c, CodeException, "Failed to generate tokens")
		return
	}

	logger.WithContext(c).Info("Token generated successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("tenantId", tenantId), zap.String("platform", platform), zap.String("deviceId", deviceId))

	// Update Login State in Redis
	redisKey := redis.GetLoginDeviceKey(getPlatformName(platform), user.Pin)
	_ = redis.Set(c, redisKey, deviceId, time.Duration(refreshTokenValidity)*time.Second)

	// Update Last Login Time asynchronously
	if cmd.GrantType != "refresh_token" {
		logger.WithContext(c).Info(fmt.Sprintf("start update user last login pin=%s", user.Pin))
		bgCtx := logger.DetachContext(c)
		bgCmdCtx := cmdCtx
		bgCmdCtx.Pin = user.Pin
		go rpc.BusinessUpdateLastLogin(bgCtx, bgCmdCtx, user.Pin)
	}

	tokenResponse := TokenCo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    accessTokenValidity,
		Scope:        scopes,
		Pin:          user.Pin,
		Nickname:     user.Nickname,
		Name:         user.Nickname,
		Avatar:       user.Avatar,
	}

	sendSuccess(c, tokenResponse)
}

func BusinessLogoutHandler(c *gin.Context) {
	authHeader := c.GetHeader("authorities")
	if authHeader == "" {
		sendError(c, CodeAccessUnauthorized, "Missing authorities header")
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		// Try base64 URL encoding if standard fails
		decodedBytes, err = base64.URLEncoding.DecodeString(authHeader)
	}

	if err != nil {
		sendError(c, CodeAccessUnauthorized, "Invalid authorities encoding")
		return
	}

	var claims struct {
		UserName string `json:"user_name"`
		Platform string `json:"platform"`
	}

	if err := json.Unmarshal(decodedBytes, &claims); err != nil {
		sendError(c, CodeAccessUnauthorized, "Invalid authorities format")
		return
	}

	if claims.UserName != "" {
		redisKey := redis.GetLoginDeviceKey(getPlatformName(claims.Platform), claims.UserName)
		_ = redis.Delete(c, redisKey)
	}

	sendSuccess(c, "成功")
}
