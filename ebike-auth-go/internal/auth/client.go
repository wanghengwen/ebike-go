package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/rpc"
	"ebike-auth-go/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type ClientAuthCmd struct {
	GrantType     string `json:"grant_type"`
	Phone         string `json:"phone,omitempty"`
	MessageCode   string `json:"messageCode,omitempty"`
	Email         string `json:"email,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Secret        string `json:"secret,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	IdentityToken string `json:"identityToken,omitempty"` // Apple
	FullName      string `json:"fullName,omitempty"`      // Apple
	Code          string `json:"code,omitempty"`          // Wechat/Google

	// ClientDTO fields
	TraceId  string `json:"traceId,omitempty"`
	TenantId string `json:"tenantId,omitempty"`
	Platform string `json:"platform,omitempty"`
	DeviceId string `json:"deviceId,omitempty"`

	IdToken       string `json:"idToken,omitempty"`       // Google
	AccessToken   string `json:"accessToken,omitempty"`   // Facebook / WeChat
	AppId         string `json:"appId,omitempty"`         // WeChat / Alipay
	JsCode        string `json:"js_code,omitempty"`       // WeChat
	EncryptedData string `json:"encryptedData,omitempty"` // WeChat / Alipay
	Iv            string `json:"iv,omitempty"`            // WeChat
	AuthCode      string `json:"authCode,omitempty"`      // Alipay
	YudaoxingCode string `json:"yudaoxingCode,omitempty"` // Yudaoxing
	ThirdLoginId  int64  `json:"thirdLoginId,omitempty"`  // Bound phone
}

type ClientTokenCo struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	TokenType    string   `json:"tokenType"`
	ExpiresIn    int      `json:"expiresIn"`
	Scope        []string `json:"scope"`
	Pin          string   `json:"pin"`
	Nickname     string   `json:"nickname"`
	Avatar       string   `json:"avatar"`
	Openid       string   `json:"openid"`
}

type LogoutCmd struct {
	TenantId string `json:"tenantId"`
	Phone    string `json:"phone"`
}

type AppleClaims struct {
	Iss           string       `json:"iss"`
	Aud           string       `json:"aud"`
	Sub           string       `json:"sub"`
	Email         string       `json:"email"`
	EmailVerified FlexibleBool `json:"email_verified"`
	jwt.RegisteredClaims
}

type GoogleClaims struct {
	Iss           string       `json:"iss"`
	Aud           string       `json:"aud"`
	Sub           string       `json:"sub"`
	Email         string       `json:"email"`
	EmailVerified FlexibleBool `json:"email_verified"`
	Name          string       `json:"name"`
	Picture       string       `json:"picture"`
	jwt.RegisteredClaims
}

func ClientTokenHandler(c *gin.Context) {
	tenantId, err := verifyClient(c)
	if err != nil {
		logger.Log.Warn("Client verification failed", zap.Error(err))
		sendError(c, CodeAccessUnauthorized, err.Error())
		return
	}

	var cmd ClientAuthCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		sendError(c, CodeIllegalArgument, "Invalid request body format")
		return
	}

	tenantAuth, _ := GetTenantAuth(tenantId)
	var user *rpc.UserDO
	var scopes []string
	if tenantAuth.Scope != "" {
		scopes = []string{tenantAuth.Scope}
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

	// Set tracing properties on context
	c.Set("tenantId", tenantId)
	c.Set("traceId", cmd.TraceId)

	// Getting access token for: 1
	logger.WithContext(c).Info(fmt.Sprintf("Getting access token for: %s", tenantId))

	cmdCtx := rpc.NewCommandContext(tenantId, cmd.TraceId, platform, deviceId, "")

	var openid string

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
		// In Java, if secret == "xyy@2022", skip SMS verification
		if cmd.Secret != "" {
			if cmd.Secret != "xyy@2022" {
				logger.WithContext(c).Warn("Secret not match", zap.String("phone", cmd.Phone))
				sendError(c, CodeException, "手机秘钥不匹配")
				return
			}
			logger.WithContext(c).Info("Getting access token via phone_code with secret bypass", zap.String("phone", cmd.Phone), zap.String("tenantId", tenantId))
		} else {
			// 88888888 is a debug code that skips verification
			if cmd.MessageCode != "88888888" {
				logger.WithContext(c).Info("Getting access token via phone_code", zap.String("phone", cmd.Phone), zap.String("tenantId", tenantId))
				ok, err := rpc.VerifyPhoneMessageCode(c, cmdCtx, "1", cmd.Phone, cmd.MessageCode)
				if err != nil || !ok {
					logger.WithContext(c).Warn("Message code verify failed", zap.String("phone", cmd.Phone), zap.Error(err))
					sendError(c, CodeMessageCodeVerifyFailed, "验证码校验失败")
					return
				}
			} else {
				logger.WithContext(c).Info("Getting access token via phone_code with debug code bypass", zap.String("phone", cmd.Phone), zap.String("tenantId", tenantId))
			}
		}

		user, err = rpc.ClientGetUserByPhone(c, cmdCtx, cmd.Phone)
		if err != nil || user == nil {
			// Auto register user if doesn't exist
			user, err = rpc.ClientUserPhoneRegister(c, cmdCtx, cmd.Phone)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "手机号自动注册失败")
				return
			}
		}
		c.Set("pin", user.Pin)
		logger.WithContext(c).Info("User fetched successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("phone", user.Phone))

	case "email_code":
		if cmd.Email == "" || cmd.MessageCode == "" {
			sendError(c, CodeIllegalArgument, "email and messageCode are required")
			return
		}
		ok, err := rpc.VerifyEmailMessageCode(c, cmdCtx, "1", cmd.Email, cmd.MessageCode)
		if err != nil || !ok {
			sendError(c, CodeMessageCodeVerifyFailed, "验证码校验失败")
			return
		}
		user, err = rpc.ClientGetUserByEmail(c, cmdCtx, cmd.Email)
		if err != nil || user == nil {
			user, err = rpc.ClientUserEmailRegister(c, cmdCtx, cmd.Email)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "邮箱自动注册失败")
				return
			}
		}
		c.Set("pin", user.Pin)

	case "password":
		if cmd.Username == "" || cmd.Password == "" {
			sendError(c, CodeIllegalArgument, "username and password are required")
			return
		}
		logger.WithContext(c).Info("Getting access token via password", zap.String("username", cmd.Username), zap.String("tenantId", tenantId))
		user, err = rpc.ClientGetUserByPin(c, cmdCtx, cmd.Username)
		if err != nil || user == nil {
			logger.WithContext(c).Warn("User not found by phone", zap.String("username", cmd.Username), zap.Error(err))
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		logger.WithContext(c).Info("User fetched successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("phone", user.Phone))
		if !utils.MatchesSHA256(cmd.Password, user.Password) {
			sendError(c, CodeAuthenticationFailed, "用户名和密码不匹配")
			return
		}

	case "phone_secret":
		if cmd.Phone == "" || cmd.Secret == "" {
			sendError(c, CodeIllegalArgument, "phone and secret are required")
			return
		}
		expectedSecret := utils.Sha256(cmd.Phone + "_xyy@2022@" + tenantId)
		if cmd.Secret != expectedSecret {
			sendError(c, CodeException, "手机秘钥不匹配")
			return
		}
		user, err = rpc.ClientGetUserByPhone(c, cmdCtx, cmd.Phone)
		if err != nil || user == nil {
			user, err = rpc.ClientUserPhoneRegister(c, cmdCtx, cmd.Phone)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "手机号自动注册失败")
				return
			}
		}
		c.Set("pin", user.Pin)

	case "facebook":
		if cmd.AccessToken == "" {
			sendError(c, CodeIllegalArgument, "accessToken is required")
			return
		}
		fbInfo, err := utils.VerifyFacebookAccessToken(cmd.AccessToken)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, err.Error())
			return
		}
		openid = fbInfo.UserID
		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, fbInfo.AppId, openid, 3)
		if err != nil || user == nil {
			thirdUserInfo, err := rpc.ClientAddSocialUser(c, cmdCtx, fbInfo.AppId, openid, fbInfo.Name, fbInfo.Avatar, fbInfo.Email, "", "", 3)
			if err != nil {
				sendError(c, CodeSocialLoginFailed, "Register social account failed")
				return
			}
			sendMustBindPhone(c, thirdUserInfo.Id)
			return
		}
		c.Set("pin", user.Pin)

	case "apple":
		if cmd.IdentityToken == "" {
			sendError(c, CodeIllegalArgument, "identityToken is required")
			return
		}
		claims, err := verifyAppleIdentityToken(cmd.IdentityToken, GetApplePublicKey)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "Apple identityToken validation failed: "+err.Error())
			return
		}
		openid = claims.Sub

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, claims.Aud, openid, 2) // APPLE ThirdType is 2
		if err != nil || user == nil {
			// Auto register or ask to bind phone
			if claims.Email != "" {
				user, err = rpc.ClientSocialUserRegister(c, cmdCtx, claims.Aud, openid, cmd.FullName, "", claims.Email, "", "", 2)
			} else {
				// Must bind phone, register temp third-party account first
				thirdUserInfo, err := rpc.ClientAddSocialUser(c, cmdCtx, claims.Aud, openid, cmd.FullName, "", "", "", "", 2)
				if err != nil {
					sendError(c, CodeSocialLoginFailed, "Register social account failed")
					return
				}
				sendMustBindPhone(c, thirdUserInfo.Id)
				return
			}
		}
		c.Set("pin", user.Pin)

	case "google":
		if cmd.IdToken == "" {
			sendError(c, CodeIllegalArgument, "idToken is required")
			return
		}
		claims, err := verifyGoogleIdToken(cmd.IdToken, GetGooglePublicKey)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "Google idToken validation failed: "+err.Error())
			return
		}
		openid = claims.Sub

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, claims.Aud, openid, 1) // GOOGLE ThirdType is 1
		if err != nil || user == nil {
			if claims.Email != "" {
				user, err = rpc.ClientSocialUserRegister(c, cmdCtx, claims.Aud, openid, claims.Name, claims.Picture, claims.Email, "", "", 1)
			} else {
				thirdUserInfo, err := rpc.ClientAddSocialUser(c, cmdCtx, claims.Aud, openid, claims.Name, claims.Picture, "", "", "", 1)
				if err != nil {
					sendError(c, CodeSocialLoginFailed, "Register social account failed")
					return
				}
				sendMustBindPhone(c, thirdUserInfo.Id)
				return
			}
		}
		c.Set("pin", user.Pin)

	case "wechat_miniapp", "wechat_third_miniapp":
		if cmd.AppId == "" || cmd.JsCode == "" || cmd.EncryptedData == "" || cmd.Iv == "" {
			sendError(c, CodeIllegalArgument, "appId, js_code, encryptedData, and iv are required")
			return
		}

		thirdType := 4 // WECHAT_MINI_APP is 4
		if cmd.GrantType == "wechat_third_miniapp" {
			thirdType = 5 // WECHAT_THIRD_MINI_APP is 5
		}

		thirdAuth, err := GetTenantThirdAuth(tenantId, thirdType)
		if err != nil || thirdAuth == nil {
			sendError(c, CodeSocialLoginFailed, "WeChat config not found for tenant: "+tenantId)
			return
		}

		if cmd.AppId != thirdAuth.AppId {
			sendError(c, CodeSocialLoginFailed, "AppId mismatch")
			return
		}

		// Execute jscode2Session
		var session *utils.WechatSession
		if cmd.GrantType == "wechat_third_miniapp" {
			session, err = wechatThirdMiniAppSession(c, cmd.AppId, cmd.JsCode)
			if err != nil {
				if strings.Contains(err.Error(), "component verify ticket") {
					sendError(c, CodeComponentVerifyTicketMissing, "")
					return
				}
				sendError(c, CodeSocialLoginFailed, "WeChat component jscode2session failed: "+err.Error())
				return
			}
		} else {
			session, err = utils.WechatJscode2Session(cmd.AppId, thirdAuth.AppSecret, cmd.JsCode)
			if err != nil {
				sendError(c, CodeSocialLoginFailed, "WeChat jscode2session failed: "+err.Error())
				return
			}
		}
		openid = session.OpenId

		fullPhone, err := decryptWechatPhone(session.SessionKey, cmd.EncryptedData, cmd.Iv)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "WeChat payload decryption failed")
			return
		}

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, cmd.AppId, openid, thirdType)
		if err != nil || user == nil {
			user, err = rpc.ClientSocialUserRegister(c, cmdCtx, cmd.AppId, openid, "WeChat User", "", "", fullPhone, "", thirdType)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "WeChat auto registration failed")
				return
			}
		}
		c.Set("pin", user.Pin)

	case "alipay_miniapp":
		if cmd.AppId == "" || cmd.AuthCode == "" || cmd.EncryptedData == "" {
			sendError(c, CodeIllegalArgument, "appId, authCode, and encryptedData are required")
			return
		}

		thirdAuth, err := GetTenantThirdAuth(tenantId, 6) // ALIPAY_MINI_APP is 6
		if err != nil || thirdAuth == nil {
			sendError(c, CodeSocialLoginFailed, "Alipay config not found for tenant: "+tenantId)
			return
		}

		if cmd.AppId != thirdAuth.AppId {
			sendError(c, CodeSocialLoginFailed, "AppId mismatch")
			return
		}

		// Split secret: privateKey@alipayPublicKey@decryptKey
		secretParts := strings.Split(thirdAuth.AppSecret, "@")
		if len(secretParts) < 3 {
			sendError(c, CodeSocialLoginFailed, "Invalid Alipay AppSecret config format")
			return
		}
		privateKey := secretParts[0]
		alipayPublicKey := secretParts[1]
		decryptKey := secretParts[2]

		// Call Alipay token API
		aliToken, err := utils.AlipayGetSystemOauthToken(cmd.AppId, privateKey, alipayPublicKey, cmd.AuthCode)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "Alipay auth failed: "+err.Error())
			return
		}
		openid = aliToken.UserId

		// Decrypt mobile
		decryptedMobile, err := utils.Decrypt3DESECBBase64(cmd.EncryptedData, decryptKey)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "Alipay mobile decryption failed")
			return
		}
		fullPhone := "+86-" + decryptedMobile

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, cmd.AppId, openid, 6)
		if err != nil || user == nil {
			user, err = rpc.ClientSocialUserRegister(c, cmdCtx, cmd.AppId, openid, "Alipay User", "", "", fullPhone, "", 6)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "Alipay auto registration failed")
				return
			}
		}
		c.Set("pin", user.Pin)

	case "social_bound_phone":
		if cmd.ThirdLoginId == 0 || cmd.Phone == "" || cmd.MessageCode == "" {
			sendError(c, CodeIllegalArgument, "thirdLoginId, phone, and messageCode are required")
			return
		}

		ok, err := rpc.VerifyPhoneMessageCode(c, cmdCtx, "1", cmd.Phone, cmd.MessageCode)
		if err != nil || !ok {
			sendError(c, CodeMessageCodeVerifyFailed, "验证码校验失败")
			return
		}

		// Retrieve third-party info
		thirdInfo, err := rpc.ClientSelectThirdUserInfoById(c, cmdCtx, cmd.ThirdLoginId)
		if err != nil || thirdInfo == nil {
			sendError(c, CodeSocialLoginFailed, "Third party auth record not found")
			return
		}
		openid = thirdInfo.OpenId

		// Register and bind
		user, err = rpc.ClientSocialUserRegister(c, cmdCtx, thirdInfo.AppId, thirdInfo.OpenId, "Social User", "", "", cmd.Phone, "", thirdInfo.ThirdType)
		if err != nil || user == nil {
			sendError(c, CodeSocialLoginFailed, "Binding and register failed")
			return
		}
		c.Set("pin", user.Pin)

	case "union_pay_miniapp":
		if cmd.JsCode == "" { // Code maps to js_code or authCode in unionpay
			sendError(c, CodeIllegalArgument, "code/js_code is required")
			return
		}
		thirdAuth, err := GetTenantThirdAuth(tenantId, 7) // UNION_PAY_MINI_APP is 7
		if err != nil || thirdAuth == nil {
			sendError(c, CodeSocialLoginFailed, "UnionPay config not found")
			return
		}
		secretParts := strings.Split(thirdAuth.AppSecret, "@")
		if len(secretParts) < 2 {
			sendError(c, CodeSocialLoginFailed, "Invalid UnionPay AppSecret config format")
			return
		}
		secret := secretParts[0]
		dcSecret := secretParts[1]

		backendToken, err := utils.GetUnionPayBackendToken(tenantId, thirdAuth.AppId, secret)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "UnionPay get backend token failed: "+err.Error())
			return
		}

		accessToken, openIdVal, err := utils.GetUnionPayAccessTokenAndOpenId(thirdAuth.AppId, backendToken, cmd.JsCode)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "UnionPay get access token failed: "+err.Error())
			return
		}
		openid = openIdVal

		phone, err := utils.GetUnionPayMobile(thirdAuth.AppId, backendToken, openid, accessToken, dcSecret)
		if err != nil {
			sendError(c, CodeSocialLoginFailed, "UnionPay get phone failed: "+err.Error())
			return
		}

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, thirdAuth.AppId, openid, 7)
		if err != nil || user == nil {
			user, err = rpc.ClientSocialUserRegister(c, cmdCtx, thirdAuth.AppId, openid, "UnionPay User", "", "", phone, "", 7)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "UnionPay register user failed")
				return
			}
		}
		c.Set("pin", user.Pin)

	case "yudaoxing_app":
		if cmd.YudaoxingCode == "" {
			sendError(c, CodeIllegalArgument, "yudaoxingCode is required")
			return
		}
		// Load Yudaoxing properties from configuration map
		// We fallback to standard config fields if not configured in Nacos
		// Note: properties (baseUrl, account, privateKey, publicVerifyKey)
		yudaoProps := YudaoxingProperties{
			BaseUrl:            config.AppConfig.Yudaoxing.BaseUrl,
			Account:            config.AppConfig.Yudaoxing.Account,
			MerchantPrivateKey: config.AppConfig.Yudaoxing.MerchantPrivateKey,
			ServerPublicKey:    config.AppConfig.Yudaoxing.ServerPublicKey,
		}
		// Fallback check
		if yudaoProps.BaseUrl == "" {
			sendError(c, CodeSocialLoginFailed, "Yudaoxing login config not found")
			return
		}

		userInfo, err := VerifyYudaoxingCode(cmd.YudaoxingCode, yudaoProps)
		if err != nil || userInfo == nil {
			sendError(c, CodeSocialLoginFailed, "Yudaoxing authentication failed: "+err.Error())
			return
		}

		// 玉环公共电单车客户定制（可选，见 auth/yuhuan_yudaoxing.go）
		yuhuanCustom := config.AppConfig.YuhuanYudaoxingCustom
		if blockMsg := userInfo.YuhuanYudaoxingLoginBlockMsg(yuhuanCustom); blockMsg != "" {
			sendError(c, CodeSocialLoginFailed, blockMsg)
			return
		}

		openid = userInfo.OpenId

		user, err = rpc.ClientGetUserBySocial(c, cmdCtx, yudaoProps.Account, openid, 8) // YUDAOXING_APP is 8
		if err != nil || user == nil {
			user, err = rpc.ClientSocialUserRegister(c, cmdCtx, yudaoProps.Account, openid, userInfo.Nickname, userInfo.Avatar, "", userInfo.PhoneForSocialRegister(yuhuanCustom), "", 8)
			if err != nil || user == nil {
				sendError(c, CodeSocialLoginFailed, "Yudaoxing register user failed")
				return
			}
			// Add Real-name info matching Java Provider
			_ = rpc.ClientAddAuthInfo(c, cmdCtx, userInfo.UserName, userInfo.IdCard, user.Pin)
		}
		c.Set("pin", user.Pin)

	case "refresh_token":
		if cmd.RefreshToken == "" {
			sendError(c, CodeIllegalArgument, "refresh_token is required")
			return
		}

		claims, err := ParseToken(cmd.RefreshToken)
		if err != nil {
			sendError(c, CodeTokenInvalidOrExpired, "Token invalid or expired")
			return
		}

		redisKey := redis.GetLoginDeviceKey(tenantId, claims.UserName)
		oldKey := redis.GetLoginDeviceOldKey(claims.UserName)

		cachedDevice, _ := redis.Get(c, redisKey)
		cachedDeviceOld, _ := redis.Get(c, oldKey)

		if cachedDevice == "" && cachedDeviceOld == "" {
			sendError(c, CodeTokenInvalidOrExpired, "Token invalid or expired")
			return
		}

		deviceToCheck := cachedDevice
		if deviceToCheck == "" {
			deviceToCheck = cachedDeviceOld
		}

		if deviceToCheck != claims.DeviceId {
			sendError(c, CodeOtherDeviceLogin, "账号在别处登录")
			return
		}

		user, err = rpc.ClientGetUserByPin(c, cmdCtx, claims.UserName)
		if err != nil || user == nil {
			sendError(c, CodeUserNotFound, "用户不存在")
			return
		}
		c.Set("pin", user.Pin)
		// Restore params from claims
		deviceId = claims.DeviceId

	default:
		sendError(c, CodeInvalidGrant, "Unsupported grant type: "+cmd.GrantType)
		return
	}

	if user.Status != 1 {
		sendError(c, CodeUserDisabled, "该账号已被禁用")
		return
	}

	// Generate Access and Refresh Token
	accessTokenValidity := tenantAuth.AccessTokenValidity
	if accessTokenValidity <= 0 {
		accessTokenValidity = 7200
	}
	refreshTokenValidity := tenantAuth.RefreshTokenValidity
	if refreshTokenValidity <= 0 {
		refreshTokenValidity = 2592000
	}

	accessToken, refreshToken, err := GenerateTokenPair(
		user.Pin,
		tenantId,
		platform,
		deviceId,
		user.Nickname,
		user.Avatar,
		openid,
		scopes,
		user.Authorities,
		cmd.GrantType,
		accessTokenValidity,
		refreshTokenValidity,
	)

	if err != nil {
		sendError(c, CodeException, "Failed to generate tokens")
		return
	}
	logger.WithContext(c).Info("Token generated successfully", zap.String("pin", user.Pin), zap.String("nickname", user.Nickname), zap.String("tenantId", tenantId), zap.String("platform", platform), zap.String("deviceId", deviceId))

	// Update Login State in Redis
	// In client mode, do not update device login state if phone_code has secret (flush stats script)
	if cmd.GrantType == "phone_code" && cmd.Secret != "" {
		// Skip session registration to avoid device kick prompt for sync operations
	} else {
		redisKey := redis.GetLoginDeviceKey(tenantId, user.Pin)
		_ = redis.Set(c, redisKey, deviceId, time.Duration(refreshTokenValidity)*time.Second)
	}

	// Gray Sync logic matching Java AuthService
	if cmd.GrantType != "refresh_token" && cmd.MessageCode != "88888888" {
		// check if tenant is gray
		// Normally read from applicationProperties.getGrayTenantIds()
		// We mock/check if env matches or if tenantId is set to gray
		isGray := isGrayTenant(tenantId)
		if isGray {
			logger.WithContext(c).Info("Syncing gray tenant user data to SaaS", zap.String("tenantId", tenantId), zap.String("pin", user.Pin))
			_ = rpc.ClientSyncUserData(c, cmdCtx, user.Pin)
		}
	}

	// Update last login asynchronously
	logger.WithContext(c).Info(fmt.Sprintf("start update user last login pin=%s", user.Pin))
	bgCtx := logger.DetachContext(c)
	bgCmdCtx := cmdCtx
	bgCmdCtx.Pin = user.Pin
	go rpc.ClientUpdateLastLogin(bgCtx, bgCmdCtx, user.Pin)

	tokenResponse := ClientTokenCo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    accessTokenValidity,
		Scope:        scopes,
		Pin:          user.Pin,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Openid:       openid,
	}

	sendSuccess(c, tokenResponse)
}

func ClientLogoutHandler(c *gin.Context) {
	authHeader := c.GetHeader("authorities")
	if authHeader == "" {
		sendError(c, CodeAccessUnauthorized, "Missing authorities header")
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		decodedBytes, err = base64.URLEncoding.DecodeString(authHeader)
	}
	if err != nil {
		sendError(c, CodeAccessUnauthorized, "Invalid authorities encoding")
		return
	}

	var claims struct {
		UserName string `json:"user_name"`
		ClientId string `json:"client_id"` // client_id contains tenantId
	}
	if err := json.Unmarshal(decodedBytes, &claims); err != nil {
		sendError(c, CodeAccessUnauthorized, "Invalid authorities format")
		return
	}

	if claims.UserName != "" && claims.ClientId != "" {
		redisKey := redis.GetLoginDeviceKey(claims.ClientId, claims.UserName)
		_ = redis.Delete(c, redisKey)
	}

	sendSuccess(c, "成功")
}

func ClientPhoneLogoutHandler(c *gin.Context) {
	var cmd LogoutCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		sendError(c, CodeIllegalArgument, "Invalid request body")
		return
	}

	if cmd.TenantId == "" || cmd.Phone == "" {
		sendError(c, CodeIllegalArgument, "tenantId and phone are required")
		return
	}

	fullPhone := "+86-" + cmd.Phone
	cmdCtx := rpc.NewCommandContext(cmd.TenantId, "", "", "", "")
	user, err := rpc.ClientGetUserByPhone(c, cmdCtx, fullPhone)
	if err != nil || user == nil {
		logger.Log.Warn("Phone logout user not found", zap.String("phone", fullPhone))
		sendSuccess(c, "成功")
		return
	}

	redisKey := redis.GetLoginDeviceKey(cmd.TenantId, user.Pin)
	_ = redis.Delete(c, redisKey)

	// Set user login tag (node tag = 2)
	tagKey := redis.GetUserLoginTagKey(cmd.TenantId, user.Pin)
	_ = redis.Set(c, tagKey, "2", 30*24*time.Hour) // long expiration

	sendSuccess(c, "成功")
}
