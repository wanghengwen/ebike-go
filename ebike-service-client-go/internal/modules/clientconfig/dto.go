package clientconfig

import (
	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

var getConfigMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "不能为null",
}

// serviceDTOBindMsgs mirrors @Validated ServiceDTO (ClientDTO @NotEmpty + serviceId @NotNull).
var serviceDTOBindMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "不能为null",
}

// bindGetConfigDTO mirrors @Validated GetConfigDTO (serviceId @NotNull).
func bindGetConfigDTO(c *gin.Context, req *getConfigDTO) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, getConfigMsgs)
}

func bindConfigPayDTO(c *gin.Context, req *configPayDTO) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, getConfigMsgs)
}

func bindConfigBaseItemDTO(c *gin.Context, req *configBaseItemDTO) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, getConfigMsgs)
}

// getConfigDTO mirrors GetConfigDTO (@NotNull serviceId).
type getConfigDTO struct {
	dto.ClientDTO
	ServiceId *int64  `json:"serviceId" binding:"required"`
	Id        *int64  `json:"id"`
	UserPin   *string `json:"userPin"`
}

// homeNavDTO mirrors HomeNavDTO.
type homeNavDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId"`
	Id        *int64 `json:"id"`
}

// configBackCarDTO mirrors ConfigBackCarDTO.
type configBackCarDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// carIdDTO mirrors CarIdDTO (@NotNull id).
type carIdDTO struct {
	dto.ClientDTO
	Id *int64 `json:"id" binding:"required"`
}

// configPayDTO / configBaseItemDTO mirror fence client DTOs (forwarded).
type configPayDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

type configBaseItemDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// configParkApplyDTO mirrors ConfigParkApplyDTO.
type configParkApplyDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// siteApplicationDTO mirrors SiteApplicationDTO (@Validated).
type siteApplicationDTO struct {
	dto.ClientDTO
	ServiceId       *int64  `json:"serviceId" binding:"required"`
	Location        string  `json:"location" binding:"required"`
	Lat             float64 `json:"lat" binding:"required"`
	Lng             float64 `json:"lng" binding:"required"`
	PhotoUrl        string  `json:"photoUrl" binding:"required"`
	ApplicantRemark string  `json:"applicantRemark" binding:"omitempty,max=50"`
}

// helmetCmd mirrors HelmetCmd (no validation).
type helmetCmd struct {
	dto.ClientDTO
	CarId string `json:"carId"`
}

// adsConfigDTO mirrors AdsConfigDTO (@Validated serviceId).
type adsConfigDTO struct {
	dto.ServiceDTO
	Ads                json.RawMessage `json:"ads"`
	HalfMiniProgramAds json.RawMessage `json:"halfMiniProgramAds"`
}

func convertBoolean(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var b bool
	if json.Unmarshal(raw, &b) != nil {
		return raw
	}
	out, _ := json.Marshal(b)
	return out
}

type adConfig struct {
	Type   *string `json:"type"`
	Id     *string `json:"id"`
	Source *string `json:"source"`
	Enable *bool   `json:"enable"`
	AppId  *string `json:"appId"`
	Name   *string `json:"name"`
	Path   *string `json:"path"`
}

type ads struct {
	Home       *adConfig `json:"home"`
	Pay        *adConfig `json:"pay"`
	UserInfo   *adConfig `json:"userInfo"`
	Charge     *adConfig `json:"charge"`
	ChargeList *adConfig `json:"chargeList"`
	Order      *adConfig `json:"order"`
	Precycling *adConfig `json:"precycling"`
	Wallet     *adConfig `json:"wallet"`
	Withdraw   *adConfig `json:"withdraw"`
}

type halfMiniProgramAdsConfig struct {
	Precycling *adConfig `json:"precycling"`
	Pay        *adConfig `json:"pay"`
	AfterPay   *adConfig `json:"afterPay"`
}

type adsConfigCO struct {
	ServiceId          *javacompat.LongStr       `json:"serviceId"`
	YgId               *string                   `json:"ygId"`
	YgEmpId            *string                   `json:"ygEmpId"`
	CoralAdId          *string                   `json:"coralAdId"`
	WanjuyuanEnable    *bool                     `json:"wanjuyuanEnable"`
	Ads                *ads                      `json:"ads"`
	HalfMiniProgramAds *halfMiniProgramAdsConfig `json:"halfMiniProgramAds"`
}

func convertAdsConfigCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co adsConfigCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type adConfigDTOIn struct {
	BannerAds   *string         `json:"bannerAds"`
	MotivateAds *string         `json:"motivateAds"`
	PlaqueAds   *string         `json:"plaqueAds"`
	VideoAds    *string         `json:"videoAds"`
	IzOn        json.RawMessage `json:"izOn"`
	ServiceId   json.RawMessage `json:"serviceId"`
}

type adConfigDTOOut struct {
	BannerAds   *string             `json:"bannerAds"`
	MotivateAds *string             `json:"motivateAds"`
	PlaqueAds   *string             `json:"plaqueAds"`
	VideoAds    *string             `json:"videoAds"`
	IzOn        json.RawMessage     `json:"izOn"`
	ServiceId   *javacompat.LongStr `json:"serviceId"`
}

func convertAdConfigDTO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var in adConfigDTOIn
	if json.Unmarshal(raw, &in) != nil {
		return raw
	}
	out := adConfigDTOOut{
		BannerAds:   in.BannerAds,
		MotivateAds: in.MotivateAds,
		PlaqueAds:   in.PlaqueAds,
		VideoAds:    in.VideoAds,
		IzOn:        javacompat.CoalesceJSONArray(in.IzOn),
	}
	if !javacompat.IsNullJSON(in.ServiceId) {
		var id javacompat.LongStr
		if json.Unmarshal(in.ServiceId, &id) == nil {
			out.ServiceId = &id
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return raw
	}
	return b
}

type configParkApplyCO struct {
	Id          *javacompat.LongStr `json:"id"`
	ServiceId   *javacompat.LongStr `json:"serviceId"`
	IzApplyPark *int                `json:"izApplyPark"`
}

func convertConfigParkApplyCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co configParkApplyCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}
