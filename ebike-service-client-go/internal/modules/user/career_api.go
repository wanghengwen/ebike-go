package user

// Ports Java UserCareerController + UserCareerServiceImpl +
// UserCareerGatewayImpl / RidingPermissionGatewayImpl / UserGatewayImpl.getCareerTags.

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// POST /client/user/career/upload
// ---------------------------------------------------------------------------

type clientCareerCmd struct {
	dto.ClientDTO
	Company   string  `json:"company" binding:"required"` // @NotBlank("单位名不能为空")
	TagId     *int64  `json:"tagId"`
	CertNo    string  `json:"certNo" binding:"required"`    // @NotBlank("证件号不能为空")
	FrontCard string  `json:"frontCard" binding:"required"` // @NotBlank("证件照不能为空")
	BackCard  *string `json:"backCard"`
}

func uploadCareer(c *gin.Context) {
	var req clientCareerCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"company":   "单位名不能为空",
		"certNo":    "证件号不能为空",
		"frontCard": "证件照不能为空",
	})) {
		return
	}
	if !notBlankMsg(c, "company", req.Company, "单位名不能为空") ||
		!notBlankMsg(c, "certNo", req.CertNo, "证件号不能为空") ||
		!notBlankMsg(c, "frontCard", req.FrontCard, "证件照不能为空") {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// CareerCmd copy {tagId, company, certNo, frontCard, backCard} + pin
	if _, ok := callData(c, rpc.ServiceUser, "/career/upload", map[string]interface{}{
		"pin":       cmdCtx.Pin,
		"tagId":     req.TagId,
		"company":   req.Company,
		"certNo":    req.CertNo,
		"frontCard": req.FrontCard,
		"backCard":  req.BackCard,
	}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/career/cancel
// ---------------------------------------------------------------------------

func cancelCareer(c *gin.Context) {
	// ClientCancelCareerCmd adds no fields to ClientDTO
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	if _, ok := callData(c, rpc.ServiceUser, "/career/cancel",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/config/enable
// ---------------------------------------------------------------------------

type serviceCmd struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"` // @NotNull
}

type userConfigCO struct {
	CareerEnable *int `json:"careerEnable"`
}

func configEnable(c *gin.Context) {
	var req serviceCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{"serviceId": "must not be null"})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	// RidingPermissionGatewayImpl.get uses getResultData -> failures bubble up.
	raw, ok := callData(c, rpc.ServiceFence, "/ridingPermission/get",
		map[string]interface{}{"serviceId": req.ServiceId}, cmdCtx)
	if !ok {
		return
	}
	if !isNullJSON(raw) {
		var rp ridingPermissionCO
		if err := json.Unmarshal(raw, &rp); err != nil {
			web.WriteException(c, err.Error())
			return
		}
		// Java: `ridingPermissionCO != null && ridingPermissionCO.getIzDefaultPermit()`
		// unboxes the Boolean -> NPE when izDefaultPermit is null.
		if rp.IzDefaultPermit == nil {
			writeNPE(c)
			return
		}
		if *rp.IzDefaultPermit {
			zero := 0
			c.JSON(http.StatusOK, dto.NewSuccessResult(userConfigCO{CareerEnable: &zero}))
			return
		}
	}

	// UserGatewayImpl.getCareerTags: CareerTagCmd{serviceIds:[String.valueOf(serviceId)]}
	rawT, ok := callData(c, rpc.ServiceUser, "/tag/info",
		map[string]interface{}{"serviceIds": []string{strconv.FormatInt(*req.ServiceId, 10)}}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(rawT) {
		// Java: careerTagServiceCo.getEnableTags().size() on null -> NPE
		writeNPE(c)
		return
	}
	var tagSvc struct {
		EnableTags []json.RawMessage `json:"enableTags"`
	}
	if err := json.Unmarshal(rawT, &tagSvc); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if tagSvc.EnableTags == nil {
		// Java: null list .size() -> NPE
		writeNPE(c)
		return
	}
	enable := 0
	if len(tagSvc.EnableTags) > 0 {
		enable = 1 // 1 可以职业认证
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(userConfigCO{CareerEnable: &enable}))
}
