package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"

	"go.uber.org/zap"
)

func logRPC(ctx context.Context, apiName string, req interface{}, respBody []byte) {
	cmdBytes, _ := json.Marshal(req)
	logger.WithContext(ctx).Info(fmt.Sprintf("userApi.%s cmd=%s result=%s", apiName, string(cmdBytes), string(respBody)))
}


type UserDO struct {
	Pin         string   `json:"pin"`
	Nickname    string   `json:"nickname"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Authorities []string `json:"authorities"`
	Status      int      `json:"status"` // 1 = active, 0 = inactive
	Avatar      string   `json:"avatar"`
	Codes       []string `json:"codes"` // saasCode for business auth
}

type UserCO struct {
	Pin         string   `json:"pin"`
	Name        string   `json:"name"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Authorities []string `json:"authorities"`
	Status      *bool    `json:"status"`
	SaasCode    []string `json:"saasCode"`
}

type UserAuthInfoCo struct {
	Pin         string   `json:"pin"`
	Nickname    string   `json:"nickname"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Authorities []string `json:"authorities"`
	Status      int      `json:"status"`
	Avatar      string   `json:"avatar"`
}


type ThirdUserInfoCo struct {
	Id        int64  `json:"id"`
	Pin       string `json:"pin"`
	Phone     string `json:"phone"`
	PurePhone string `json:"purePhone"`
	Email     string `json:"email"`
	OpenId    string `json:"openId"`
	UnionId   string `json:"unionId"`
	AppId     string `json:"appId"`
	Profile   string `json:"profile"`
	ThirdType int    `json:"thirdType"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
}

type UserPhoneCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Phone          string         `json:"phone"`
}

type UserEmailCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Email          string         `json:"email"`
}

type UserCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Pin            string         `json:"pin,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	Email          string         `json:"email,omitempty"`
}

type UserByPinQuery struct {
	CommandContext CommandContext `json:"commandContext"`
	Pin            string         `json:"pin"`
}

type UserByPhoneQuery struct {
	CommandContext CommandContext `json:"commandContext"`
	Phone          string         `json:"phone"`
}

type UserByEmailQuery struct {
	CommandContext CommandContext `json:"commandContext"`
	Email          string         `json:"email"`
}

type UserBySocialQuery struct {
	CommandContext CommandContext `json:"commandContext"`
	AppId          string         `json:"appId"`
	OpenId         string         `json:"openId"`
	ThirdType      int            `json:"thirdType"`
}

type SocialUserInfoCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	AppId          string         `json:"appId"`
	OpenId         string         `json:"openId"`
	ThirdType      int            `json:"thirdType"`
	Nickname       string         `json:"nickname,omitempty"`
	Avatar         string         `json:"avatar,omitempty"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	Profile        string         `json:"profile,omitempty"`
}

type UserByThirdIdCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Id             int64          `json:"id"`
}

type GrayOrderQuery struct {
	CommandContext CommandContext `json:"commandContext"`
	TenantId       string         `json:"tenantId"`
	Pin            string         `json:"pin"`
}

type UserAuthInfoCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	AuthName       string         `json:"authName"`
	AuthNo         string         `json:"authNo"`
	Pin            string         `json:"pin"`
}

type CodeVerifyCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Phone          string         `json:"phone"`
	Code           string         `json:"code"`
	Scene          string         `json:"scene"`
}

type EmailCodeVerifyCmd struct {
	CommandContext CommandContext `json:"commandContext"`
	Email          string         `json:"email"`
	Code           string         `json:"code"`
	Scene          string         `json:"scene"`
}

func newCmdCtx(tenantId, traceId string) CommandContext {
	if traceId == "" {
		traceId = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return CommandContext{
		TenantId: tenantId,
		TraceId:  traceId,
		Pin:      "-1",
	}
}

func NewCommandContext(tenantId, traceId, platform, deviceId, pin string) CommandContext {
	if traceId == "" {
		traceId = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	if pin == "" {
		pin = "-1"
	}
	source := "ebike-auth-" + config.AppConfig.AuthMode
	return CommandContext{
		TenantId:      tenantId,
		TraceId:       traceId,
		Pin:           pin,
		Platform:      platform,
		DeviceId:      deviceId,
		Source:        source,
		StressTesting: false,
	}
}

func getUserBaseURL() string {
	if url := os.Getenv("EBIKE_USER_URL"); url != "" {
		return url
	}
	url, err := ResolveServiceAddress("ebike-user")
	if err != nil {
		logger.Log.Debug("Nacos resolve ebike-user failed, using fallback URL", zap.Error(err))
		return "http://ebike-user:8080"
	}
	return url
}

// ---------------- BUSINESS AUTH RPCs (using ebike-management) ----------------

func BusinessGetUserByPin(ctx context.Context, cmdCtx CommandContext, pin string) (*UserDO, error) {
	url := fmt.Sprintf("%s/user/getUserByPin", getManagementBaseURL())
	req := UserCmd{CommandContext: cmdCtx, Pin: pin}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "getUserByPin", req, resp.Body())

	var result ManagementResult[UserCO]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("getUserByPin RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromCO(result.Data), nil
}

func BusinessGetUserByPhone(ctx context.Context, cmdCtx CommandContext, phone string) (*UserDO, error) {
	url := fmt.Sprintf("%s/user/getUserByPhone", getManagementBaseURL())
	req := UserCmd{CommandContext: cmdCtx, Phone: phone}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "getUserByPhone", req, resp.Body())

	var result ManagementResult[UserCO]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("getUserByPhone RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromCO(result.Data), nil
}

func BusinessGetUserByEmail(ctx context.Context, cmdCtx CommandContext, email string) (*UserDO, error) {
	url := fmt.Sprintf("%s/user/getUserByEmail", getManagementBaseURL())
	req := UserCmd{CommandContext: cmdCtx, Email: email}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "getUserByEmail", req, resp.Body())

	var result ManagementResult[UserCO]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("getUserByEmail RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromCO(result.Data), nil
}

func BusinessUpdateLastLogin(ctx context.Context, cmdCtx CommandContext, pin string) {
	url := fmt.Sprintf("%s/user/updateLastLogin", getManagementBaseURL())
	req := UserCmd{CommandContext: cmdCtx, Pin: pin}
	resp, err := Client.R().SetHeader("Content-Type", "application/json").SetBody(req).Post(url)
	if err == nil {
		logRPC(ctx, "updateUserLastLogin", req, resp.Body())
	}
}

func toUserDOFromCO(co UserCO) *UserDO {
	if co.Pin == "" {
		return nil
	}
	statusVal := 0
	if co.Status != nil && *co.Status {
		statusVal = 1
	}
	return &UserDO{
		Pin:         co.Pin,
		Nickname:    co.Name,
		Phone:       co.Phone,
		Email:       co.Email,
		Password:    co.Password,
		Authorities: co.Authorities,
		Status:      statusVal,
		Codes:       co.SaasCode,
	}
}

// ---------------- CLIENT AUTH RPCs (using ebike-user) ----------------

func ClientGetUserByPin(ctx context.Context, cmdCtx CommandContext, pin string) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/selectUserByPin", getUserBaseURL())
	req := UserByPinQuery{CommandContext: cmdCtx, Pin: pin}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "selectUserByPin", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("selectUserByPin RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientGetUserByPhone(ctx context.Context, cmdCtx CommandContext, phone string) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/selectUserByPhone", getUserBaseURL())
	req := UserByPhoneQuery{CommandContext: cmdCtx, Phone: phone}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "selectUserByPhone", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("selectUserByPhone RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientGetUserByEmail(ctx context.Context, cmdCtx CommandContext, email string) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/selectUserByEmail", getUserBaseURL())
	req := UserByEmailQuery{CommandContext: cmdCtx, Email: email}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "selectUserByEmail", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("selectUserByEmail RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientGetUserBySocial(ctx context.Context, cmdCtx CommandContext, appId, openId string, thirdType int) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/selectUserBySocial", getUserBaseURL())
	req := UserBySocialQuery{
		CommandContext: cmdCtx,
		AppId:          appId,
		OpenId:         openId,
		ThirdType:      thirdType,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "selectUserBySocial", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("selectUserBySocial RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientSocialUserRegister(ctx context.Context, cmdCtx CommandContext, appId, openId, nickname, avatar, email, phone, profile string, thirdType int) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/socialUserRegister", getUserBaseURL())
	req := SocialUserInfoCmd{
		CommandContext: cmdCtx,
		AppId:          appId,
		OpenId:         openId,
		ThirdType:      thirdType,
		Nickname:       nickname,
		Avatar:         avatar,
		Email:          email,
		Phone:          phone,
		Profile:        profile,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "socialUserRegister", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("socialUserRegister RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientAddSocialUser(ctx context.Context, cmdCtx CommandContext, appId, openId, nickname, avatar, email, phone, profile string, thirdType int) (*ThirdUserInfoCo, error) {
	url := fmt.Sprintf("%s/userAuthInfo/addThirdUserInfo", getUserBaseURL())
	req := SocialUserInfoCmd{
		CommandContext: cmdCtx,
		AppId:          appId,
		OpenId:         openId,
		ThirdType:      thirdType,
		Nickname:       nickname,
		Avatar:         avatar,
		Email:          email,
		Phone:          phone,
		Profile:        profile,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "addThirdUserInfo", req, resp.Body())

	var result ManagementResult[ThirdUserInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("addThirdUserInfo RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return &result.Data, nil
}

func ClientSelectThirdUserInfoById(ctx context.Context, cmdCtx CommandContext, id int64) (*ThirdUserInfoCo, error) {
	url := fmt.Sprintf("%s/userAuthInfo/selectThirdUserInfoById", getUserBaseURL())
	req := UserByThirdIdCmd{CommandContext: cmdCtx, Id: id}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "selectThirdUserInfoById", req, resp.Body())

	var result ManagementResult[ThirdUserInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("selectThirdUserInfoById RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return &result.Data, nil
}

func ClientUpdateLastLogin(ctx context.Context, cmdCtx CommandContext, pin string) {
	url := fmt.Sprintf("%s/user/updateLastLogin", getUserBaseURL())
	req := UserByPinQuery{CommandContext: cmdCtx, Pin: pin}
	resp, err := Client.R().SetHeader("Content-Type", "application/json").SetBody(req).Post(url)
	if err == nil {
		logRPC(ctx, "updateUserLastLogin", req, resp.Body())
	}
}

func ClientSyncUserData(ctx context.Context, cmdCtx CommandContext, pin string) error {
	url := fmt.Sprintf("%s/gray/syncUserData", getUserBaseURL())
	req := GrayOrderQuery{
		CommandContext: cmdCtx,
		TenantId:       cmdCtx.TenantId,
		Pin:            pin,
	}
	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return err
	}
	logRPC(ctx, "syncUserData", req, resp.Body())

	var result ManagementResult[int]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return fmt.Errorf("syncUserData RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return nil
}

func ClientAddAuthInfo(ctx context.Context, cmdCtx CommandContext, authName, authNo, pin string) error {
	url := fmt.Sprintf("%s/user/addAuthInfo", getUserBaseURL())
	req := UserAuthInfoCmd{
		CommandContext: cmdCtx,
		AuthName:       authName,
		AuthNo:         authNo,
		Pin:            pin,
	}
	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return err
	}
	logRPC(ctx, "addAuthInfo", req, resp.Body())

	var result ManagementResult[bool]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return fmt.Errorf("addAuthInfo RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return nil
}

func toUserDOFromAuthCo(authCo UserAuthInfoCo) *UserDO {
	if authCo.Pin == "" {
		return nil
	}
	return &UserDO{
		Pin:         authCo.Pin,
		Nickname:    authCo.Nickname,
		Phone:       authCo.Phone,
		Email:       authCo.Email,
		Password:    authCo.Password,
		Authorities: authCo.Authorities,
		Status:      authCo.Status,
		Avatar:      authCo.Avatar,
	}
}

// ---------------- VERIFY CODE RPCs (using ebike-management) ----------------

func VerifyPhoneMessageCode(ctx context.Context, cmdCtx CommandContext, scene, phone, messageCode string) (bool, error) {
	if !config.AppConfig.EnablePhoneVerify {
		return true, nil
	}

	url := fmt.Sprintf("%s/code/verify", getManagementBaseURL())
	req := CodeVerifyCmd{
		CommandContext: cmdCtx,
		Phone:          phone,
		Code:           messageCode,
		Scene:          scene,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return false, err
	}
	logRPC(ctx, "verifyPhoneMessageCode", req, resp.Body())

	var result ManagementResult[bool]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return false, err
	}

	return result.Data, nil
}

func VerifyEmailMessageCode(ctx context.Context, cmdCtx CommandContext, scene, email, messageCode string) (bool, error) {
	if !config.AppConfig.EnableEmailVerify {
		return true, nil
	}

	url := fmt.Sprintf("%s/email/code/verify", getManagementBaseURL())
	req := EmailCodeVerifyCmd{
		CommandContext: cmdCtx,
		Email:          email,
		Code:           messageCode,
		Scene:          scene,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return false, err
	}
	logRPC(ctx, "verifyEmailMessageCode", req, resp.Body())

	var result ManagementResult[bool]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return false, err
	}

	return result.Data, nil
}

func ClientUserPhoneRegister(ctx context.Context, cmdCtx CommandContext, phone string) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/phoneRegister", getUserBaseURL())
	req := UserPhoneCmd{
		CommandContext: cmdCtx,
		Phone:          phone,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "phoneRegister", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("phoneRegister RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}

func ClientUserEmailRegister(ctx context.Context, cmdCtx CommandContext, email string) (*UserDO, error) {
	url := fmt.Sprintf("%s/userAuthInfo/emailRegister", getUserBaseURL())
	req := UserEmailCmd{
		CommandContext: cmdCtx,
		Email:          email,
	}

	resp, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	logRPC(ctx, "emailRegister", req, resp.Body())

	var result ManagementResult[UserAuthInfoCo]
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "000000" && result.Code != "SUCCESS" {
		return nil, fmt.Errorf("emailRegister RPC error: code=%s, msg=%s", result.Code, result.Msg)
	}

	return toUserDOFromAuthCo(result.Data), nil
}
