package tps

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encoding/json"
)

// ChuangLanAuthParam holds config for ChuangLan provider.
// Parsed from the supplier's JSON config.
type ChuangLanAuthParam struct {
	AppId         string `json:"appId"`
	AppKey        string `json:"appKey"`
	IdCardAuthUrl string `json:"idCardAuthUrl"`
	FaceMatchUrl  string `json:"faceMatchUrl"`
}

// ChuangLanAuth implements TpsIdAuth for ChuangLan provider.
type ChuangLanAuth struct{}

// Authenticate performs identity verification via ChuangLan.
// Supports AuthTypeIdentityCard (二要素) and AuthTypeIdentityCardFace (三要素).
func (c *ChuangLanAuth) Authenticate(param CommonTpsIdAuthParam, responseData map[string]interface{}) bool {
	var config ChuangLanAuthParam
	if err := parseConfig(param.Config, &config); err != nil {
		log.Printf("[%s] ChuangLan: failed to parse config: %v", param.TraceId, err)
		return chuangLanException(responseData, err.Error())
	}

	switch param.AuthType {
	case AuthTypeIdentityCard:
		return c.authenticateIdCard(param, config, responseData)
	case AuthTypeIdentityCardFace:
		return c.authenticateFace(param, config, responseData)
	default:
		log.Printf("[%s] ChuangLan: unsupported auth type: %d", param.TraceId, param.AuthType)
		return false
	}
}

// authenticateIdCard performs 二要素 (name + ID card) verification.
func (c *ChuangLanAuth) authenticateIdCard(param CommonTpsIdAuthParam, config ChuangLanAuthParam, responseData map[string]interface{}) bool {
	body, err := chuangLanPost(config.IdCardAuthUrl, url.Values{
		"appId":  {config.AppId},
		"appKey": {config.AppKey},
		"name":   {param.Name},
		"idNum":  {param.IdCardNum},
	})
	if err != nil {
		log.Printf("[%s] ChuangLan idCard auth error: %v", param.TraceId, err)
		return chuangLanException(responseData, err.Error())
	}

	bodyJSON := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJSON); err != nil {
		log.Printf("[%s] ChuangLan idCard auth: failed to parse response: %v", param.TraceId, err)
		return chuangLanException(responseData, err.Error())
	}
	data := asObject(bodyJSON["data"])
	log.Printf("[%s] ChuangLan idCard code=%s result=%s orderNo=%s remark=%s",
		param.TraceId, asString(bodyJSON["code"]), asString(data["result"]), asString(data["orderNo"]), asString(data["remark"]))

	if asString(bodyJSON["code"]) == "200000" {
		if asString(data["result"]) == "01" {
			responseData["status"] = 1
			responseData["uniqueId"] = asString(data["orderNo"])
			responseData["response"] = data
			responseData["authResponse"] = AuthResponse{IzSame: true, Info: "身份校验通过"}
			return true
		}
		responseData["status"] = 2
		responseData["uniqueId"] = asString(data["orderNo"])
		responseData["response"] = data
		responseData["authResponse"] = AuthResponse{IzSame: false, Info: "身份校验未通过"}
		return false
	}

	responseData["status"] = 2
	responseData["response"] = bodyJSON
	responseData["authResponse"] = AuthResponse{IzSame: false, Info: "身份校验未通过"}
	return false
}

// authenticateFace performs 三要素 (name + ID card + face) verification.
func (c *ChuangLanAuth) authenticateFace(param CommonTpsIdAuthParam, config ChuangLanAuthParam, responseData map[string]interface{}) bool {
	body, err := chuangLanPost(config.FaceMatchUrl, url.Values{
		"appId":  {config.AppId},
		"appKey": {config.AppKey},
		"image":  {param.Image},
		"name":   {param.Name},
		"idNum":  {param.IdCardNum},
	})
	if err != nil {
		log.Printf("[%s] ChuangLan face auth error: %v", param.TraceId, err)
		return chuangLanFaceException(responseData, err.Error())
	}

	bodyJSON := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJSON); err != nil {
		log.Printf("[%s] ChuangLan face auth: failed to parse response: %v", param.TraceId, err)
		return chuangLanFaceException(responseData, err.Error())
	}
	data := asObject(bodyJSON["data"])
	log.Printf("[%s] ChuangLan face code=%s idcardResult=%s photoResult=%s orderNo=%s score=%s",
		param.TraceId, asString(bodyJSON["code"]), asString(data["idcardResult"]), asString(data["photoResult"]),
		asString(data["orderNo"]), asString(data["photoScore"]))
	photoScore := asString(data["photoScore"])
	orderNo := asString(data["orderNo"])

	if asString(bodyJSON["code"]) != "200000" {
		responseData["status"] = 2
		responseData["score"] = "0"
		responseData["response"] = bodyJSON
		responseData["authResponse"] = AuthResponse{IzSame: false, Score: "0", Info: asString(bodyJSON["message"])}
		return false
	}

	idcardResult := asString(data["idcardResult"]) == "01"
	photoResult := asString(data["photoResult"]) == "01"

	if !idcardResult {
		responseData["status"] = 2
		responseData["score"] = "0"
		responseData["uniqueId"] = orderNo
		responseData["response"] = data
		responseData["authResponse"] = AuthResponse{IzSame: false, Score: photoScore, Info: "身份校验未通过"}
		return false
	}

	if photoResult {
		responseData["status"] = 1
		responseData["score"] = photoScore
		responseData["uniqueId"] = orderNo
		responseData["response"] = data
		responseData["authResponse"] = AuthResponse{IzSame: true, Score: photoScore, Info: "认证成功"}
		return true
	}

	responseData["status"] = 2
	responseData["score"] = photoScore
	responseData["uniqueId"] = orderNo
	responseData["response"] = data
	responseData["authResponse"] = AuthResponse{IzSame: false, Score: photoScore, Info: "人脸校验未通过"}
	return false
}

// MatchFaceImage delegates to Authenticate with AuthTypeIdentityCardFace.
// Uses param.SourceImg as the face image for comparison.
func (c *ChuangLanAuth) MatchFaceImage(param CommonTpsFaceMatchParam, responseData map[string]interface{}) bool {
	authParam := CommonTpsIdAuthParam{
		TraceId:   param.TraceId,
		AuthType:  AuthTypeIdentityCardFace,
		Config:    param.Config,
		Name:      param.Name,
		IdCardNum: param.IdCardNum,
		Image:     param.SourceImg,
	}
	return c.Authenticate(authParam, responseData)
}

// String returns a readable description for logging.
func (c *ChuangLanAuth) String() string {
	return fmt.Sprintf("ChuangLanAuth{channel=%d}", ChannelChuangLan)
}

// chuangLanPost sends an x-www-form-urlencoded POST and returns the raw body.
func chuangLanPost(targetURL string, form url.Values) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, targetURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// chuangLanException mirrors Java's catch block for the 二要素 path:
// status=3, response={exception}, authResponse{izSame:false, info:msg}.
func chuangLanException(responseData map[string]interface{}, msg string) bool {
	responseData["status"] = 3
	responseData["response"] = map[string]interface{}{"exception": msg}
	responseData["authResponse"] = AuthResponse{IzSame: false, Info: msg}
	return false
}

// chuangLanFaceException mirrors Java's catch block for the 三要素 path:
// status=3, score=0, response={exception}, authResponse{izSame:false, score:0, info:msg}.
func chuangLanFaceException(responseData map[string]interface{}, msg string) bool {
	responseData["status"] = 3
	responseData["score"] = "0"
	responseData["response"] = map[string]interface{}{"exception": msg}
	responseData["authResponse"] = AuthResponse{IzSame: false, Score: "0", Info: msg}
	return false
}
