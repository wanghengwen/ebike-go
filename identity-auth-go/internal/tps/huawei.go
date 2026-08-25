package tps

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	frs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/frs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/frs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/frs/v2/region"
)

// HuaWeiAuthParam holds config for HuaWei provider.
type HuaWeiAuthParam struct {
	RegionId   string  `json:"regionId"`
	ApiKey     string  `json:"apiKey"`
	SecretKey  string  `json:"secretKey"`
	MatchScore float64 `json:"matchScore"`
}

// HuaWeiAuth implements TpsIdAuth for HuaWei provider.
type HuaWeiAuth struct{}

// Authenticate is not supported for HuaWei (face comparison only).
func (h *HuaWeiAuth) Authenticate(param CommonTpsIdAuthParam, responseData map[string]interface{}) bool {
	log.Printf("[%s] HuaWei: Authenticate not supported, use MatchFaceImage", param.TraceId)
	responseData["status"] = 3
	responseData["authResponse"] = AuthResponse{IzSame: false}
	return false
}

// MatchFaceImage performs face comparison via Huawei Cloud FRS SDK (AK/SK).
func (h *HuaWeiAuth) MatchFaceImage(param CommonTpsFaceMatchParam, responseData map[string]interface{}) bool {
	var config HuaWeiAuthParam
	if err := parseConfig(param.Config, &config); err != nil {
		log.Printf("[%s] HuaWei: failed to parse config: %v", param.TraceId, err)
		responseData["status"] = 3
		responseData["authResponse"] = AuthResponse{IzSame: false}
		return false
	}

	client, err := newFrsClient(config.RegionId, config.ApiKey, config.SecretKey)
	if err != nil {
		log.Printf("[%s] HuaWei: failed to create client: %v", param.TraceId, err)
		responseData["status"] = 3
		responseData["score"] = 0
		responseData["response"] = map[string]interface{}{"exception": err.Error()}
		responseData["authResponse"] = AuthResponse{IzSame: false, Info: err.Error()}
		return false
	}

	request := &model.CompareFaceByBase64Request{}
	request.Body = &model.FaceCompareBase64Req{
		Image1Base64: param.SourceImg,
		Image2Base64: param.TargetImg,
	}

	compareResponse, err := client.CompareFaceByBase64(request)
	log.Printf("[%s] HuaWei face compare response: %+v, err=%v", param.TraceId, compareResponse, err)
	if err != nil {
		log.Printf("[%s] HuaWei: compare request failed: %v", param.TraceId, err)
		responseData["status"] = 3
		responseData["score"] = 0
		responseData["response"] = map[string]interface{}{"exception": err.Error()}
		responseData["authResponse"] = AuthResponse{IzSame: false, Info: err.Error()}
		return false
	}

	responseJSON, _ := json.Marshal(compareResponse)
	responseData["response"] = string(responseJSON)

	if compareResponse.HttpStatusCode == 200 && compareResponse.Similarity != nil {
		score := *compareResponse.Similarity * 100
		scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
		responseData["score"] = score

		if config.MatchScore < score {
			responseData["status"] = 1
			responseData["authResponse"] = AuthResponse{IzSame: true, Score: scoreStr, Info: "认证成功"}
			return true
		}

		responseData["status"] = 2
		responseData["authResponse"] = AuthResponse{IzSame: false, Score: scoreStr, Info: "人脸校验未通过"}
		return false
	}

	responseData["status"] = 2
	responseData["score"] = 0
	responseData["authResponse"] = AuthResponse{IzSame: false, Score: "0", Info: "人脸校验未通过"}
	return false
}

func newFrsClient(regionId, ak, sk string) (*frs.FrsClient, error) {
	auth, err := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	hcClient, err := frs.FrsClientBuilder().
		WithRegion(region.ValueOf(regionId)).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	return frs.NewFrsClient(hcClient), nil
}
