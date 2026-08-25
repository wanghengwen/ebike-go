package tps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// BaiDuAuthParam holds config for BaiDu provider.
type BaiDuAuthParam struct {
	AppId      string  `json:"appId"`
	ApiKey     string  `json:"apiKey"`
	SecretKey  string  `json:"secretKey"`
	MatchScore float64 `json:"matchScore"`
}

// baiDuTokenResponse is the OAuth2 token response.
type baiDuTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// baiDuMatchRequest is a single image entry in the match request array.
type baiDuMatchRequest struct {
	Image     string `json:"image"`
	ImageType string `json:"image_type"`
}

// baiDuMatchResponse is the face match API response.
type baiDuMatchResponse struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	LogID     int64  `json:"log_id"`
	Result    struct {
		Score float64 `json:"score"`
	} `json:"result"`
}

// BaiDuAuth implements TpsIdAuth for BaiDu provider.
type BaiDuAuth struct{}

// Authenticate is not supported for BaiDu (face comparison only).
// Mirrors Java's AbstractTpsIdAuth.authenticate default (returns false); we
// additionally tag the response so upstream returns a clean failure instead of
// dereferencing a nil authResponse.
func (b *BaiDuAuth) Authenticate(param CommonTpsIdAuthParam, responseData map[string]interface{}) bool {
	log.Printf("[%s] BaiDu: Authenticate not supported, use MatchFaceImage", param.TraceId)
	responseData["status"] = 3
	responseData["authResponse"] = AuthResponse{IzSame: false}
	return false
}

// MatchFaceImage performs face comparison via BaiDu AI REST API.
func (b *BaiDuAuth) MatchFaceImage(param CommonTpsFaceMatchParam, responseData map[string]interface{}) bool {
	var config BaiDuAuthParam
	if err := parseConfig(param.Config, &config); err != nil {
		log.Printf("[%s] BaiDu: failed to parse config: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}

	// Step 1: Get access token
	accessToken, err := b.getAccessToken(param.TraceId, config)
	if err != nil {
		log.Printf("[%s] BaiDu: failed to get access token: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}

	// Step 2: Call face match API
	matchURL := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/face/v3/match?access_token=%s", accessToken)

	reqBody := []baiDuMatchRequest{
		{Image: param.SourceImg, ImageType: "BASE64"},
		{Image: param.TargetImg, ImageType: "BASE64"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[%s] BaiDu: failed to marshal match request: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}

	log.Printf("[%s] BaiDu face match request: url=%s", param.TraceId, matchURL)

	client := baiduHTTPClient()
	req, err := http.NewRequest(http.MethodPost, matchURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[%s] BaiDu: failed to create match request: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] BaiDu: match request failed: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] BaiDu: failed to read match response: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}

	log.Printf("[%s] BaiDu face match response: %s", param.TraceId, string(respBody))

	// Store raw response as parsed JSON (matches Java storing the API JSON).
	var rawResponse interface{}
	if err := json.Unmarshal(respBody, &rawResponse); err != nil {
		log.Printf("[%s] BaiDu: failed to parse raw response: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}
	responseData["response"] = rawResponse

	var matchResp baiDuMatchResponse
	if err := json.Unmarshal(respBody, &matchResp); err != nil {
		log.Printf("[%s] BaiDu: failed to parse match response: %v", param.TraceId, err)
		return baiduException(responseData, err.Error())
	}

	responseData["uniqueId"] = fmt.Sprintf("%d", matchResp.LogID)

	// Java records score=0 when error_code != 0, otherwise the returned score.
	score := 0.0
	if matchResp.ErrorCode == 0 {
		score = matchResp.Result.Score
	}
	scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
	responseData["score"] = score

	// Java: matchScore.compareTo(score) < 0  =>  matchScore < score
	if config.MatchScore < score {
		responseData["status"] = 1
		responseData["authResponse"] = AuthResponse{IzSame: true, Score: scoreStr, Info: "认证成功"}
		return true
	}

	responseData["status"] = 2
	responseData["authResponse"] = AuthResponse{IzSame: false, Score: scoreStr, Info: "人脸校验未通过"}
	return false
}

// getAccessToken retrieves an OAuth2 access token from BaiDu.
func (b *BaiDuAuth) getAccessToken(traceId string, config BaiDuAuthParam) (string, error) {
	tokenURL := fmt.Sprintf(
		"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		config.ApiKey, config.SecretKey,
	)

	log.Printf("[%s] BaiDu: requesting access token", traceId)

	client := baiduHTTPClient()
	resp, err := client.Post(tokenURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	log.Printf("[%s] BaiDu: token response: %s", traceId, string(body))

	var tokenResp baiDuTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	return tokenResp.AccessToken, nil
}

// String returns a readable description for logging.
func (b *BaiDuAuth) String() string {
	return fmt.Sprintf("BaiDuAuth{channel=%d}", ChannelBaiDu)
}

// baiduException mirrors Java's catch block: status=3, score=0,
// response={exception}, authResponse{izSame:false, info:msg}.
func baiduException(responseData map[string]interface{}, msg string) bool {
	responseData["status"] = 3
	responseData["score"] = 0.0
	responseData["response"] = map[string]interface{}{"exception": msg}
	responseData["authResponse"] = AuthResponse{IzSame: false, Info: msg}
	return false
}

// baiduHTTPClient mirrors Java AipFace: connect 2s, read 3s.
func baiduHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 3 * time.Second,
		},
	}
}
