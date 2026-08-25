package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type FacebookVerifyAccessTokenData struct {
	Data FacebookVerifyAccessToken `json:"data"`
}

type FacebookVerifyAccessToken struct {
	AppId   string `json:"app_id"`
	UserId  string `json:"user_id"`
	IsValid bool   `json:"is_valid"`
}

type FacebookProfile struct {
	Name    string                  `json:"name"`
	Email   string                  `json:"email"`
	Picture FacebookProfilePicture  `json:"picture"`
}

type FacebookProfilePicture struct {
	Data FacebookProfilePictureData `json:"data"`
}

type FacebookProfilePictureData struct {
	Url string `json:"url"`
}

type FacebookUserInfo struct {
	AppId   string
	UserID  string
	Email   string
	Name    string
	Avatar  string
	ThirdType int
}

func VerifyFacebookAccessToken(accessToken string) (*FacebookUserInfo, error) {
	validURL := fmt.Sprintf(
		"https://graph.facebook.com/debug_token?access_token=%s&input_token=%s",
		url.QueryEscape(accessToken),
		url.QueryEscape(accessToken),
	)
	resp, err := restyClient.R().Get(validURL)
	if err != nil {
		return nil, fmt.Errorf("getFacebookVerifyAccessToken failed: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("facebook debug_token status %d", resp.StatusCode())
	}

	var verifyData FacebookVerifyAccessTokenData
	if err := json.Unmarshal(resp.Body(), &verifyData); err != nil {
		return nil, err
	}
	verify := verifyData.Data
	if verify.UserId == "" {
		return nil, fmt.Errorf("accessToken verify failed. getFacebookVerifyAccessToken is null")
	}
	if !verify.IsValid {
		return nil, fmt.Errorf("accessToken verify failed. is_valid=false")
	}

	profileURL := fmt.Sprintf(
		"https://graph.facebook.com/%s?fields=name,picture,email&access_token=%s",
		url.QueryEscape(verify.UserId),
		url.QueryEscape(accessToken),
	)
	profileResp, err := restyClient.R().Get(profileURL)
	if err != nil {
		return nil, fmt.Errorf("getFacebookProfile failed: %w", err)
	}
	if profileResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("facebook profile status %d", profileResp.StatusCode())
	}

	var profile FacebookProfile
	if err := json.Unmarshal(profileResp.Body(), &profile); err != nil {
		return nil, err
	}
	if profile.Email == "" {
		return nil, fmt.Errorf("FacebookProfile.email must not be empty")
	}

	avatar := ""
	if profile.Picture.Data.Url != "" {
		avatar = profile.Picture.Data.Url
	}

	return &FacebookUserInfo{
		AppId:     verify.AppId,
		UserID:    verify.UserId,
		Email:     profile.Email,
		Name:      profile.Name,
		Avatar:    avatar,
		ThirdType: 3,
	}, nil
}

func FetchWechatComponentAccessToken(componentAppId, componentAppSecret, componentVerifyTicket string) (string, error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	form := url.Values{}
	form.Set("component_appid", componentAppId)
	form.Set("component_appsecret", componentAppSecret)
	form.Set("component_verify_ticket", componentVerifyTicket)

	resp, err := restyClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(form.Encode()).
		Post(apiURL)
	if err != nil {
		return "", err
	}

	var result componentAccessTokenResp
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}
	if result.ComponentAccessToken == "" {
		return "", fmt.Errorf("component token error: code=%d, msg=%s", result.Errcode, result.Errmsg)
	}
	return result.ComponentAccessToken, nil
}

type componentAccessTokenResp struct {
	ComponentAccessToken string `json:"component_access_token"`
	Errcode              int    `json:"errcode"`
	Errmsg               string `json:"errmsg"`
}

func WechatComponentJscode2Session(appId, jsCode, componentAppId, componentAccessToken string) (*WechatSession, error) {
	apiURL := "https://api.weixin.qq.com/sns/component/jscode2session"
	resp, err := restyClient.R().
		SetQueryParams(map[string]string{
			"appid":                  appId,
			"js_code":                jsCode,
			"grant_type":             "authorization_code",
			"component_appid":        componentAppId,
			"component_access_token": componentAccessToken,
		}).
		Get(apiURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("wechat component api status %d", resp.StatusCode())
	}

	var session WechatSession
	if err := json.Unmarshal(resp.Body(), &session); err != nil {
		return nil, err
	}
	if session.Errcode != 0 {
		return nil, fmt.Errorf("wechat component error: code=%d, msg=%s", session.Errcode, session.Errmsg)
	}
	return &session, nil
}
