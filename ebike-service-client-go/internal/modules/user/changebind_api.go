package user

// Ports Java ChangeBindController + ChangeBindServiceImpl + ChangeBindGatewayImpl.
// Every endpoint returns ResultHelper.success(1) after a getResultData call.

import (
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// POST /client/user/changBind/add   (Java path typo preserved)
// ---------------------------------------------------------------------------

type clientChangeBindCmd struct {
	dto.ClientDTO
	AuthName  string `json:"authName" binding:"required"`  // @NotBlank
	AuthNo    string `json:"authNo" binding:"required"`    // @NotBlank
	FrontCard string `json:"frontCard" binding:"required"` // @NotBlank
	BackCard  string `json:"backCard" binding:"required"`  // @NotBlank
	ApplyType *int   `json:"applyType" binding:"required"` // @NotNull
}

func addChangeBind(c *gin.Context) {
	var req clientChangeBindCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authName":  "must not be blank",
		"authNo":    "must not be blank",
		"frontCard": "must not be blank",
		"backCard":  "must not be blank",
		"applyType": "must not be null",
	})) {
		return
	}
	if !web.NotBlank(c, "authName", req.AuthName) || !web.NotBlank(c, "authNo", req.AuthNo) ||
		!web.NotBlank(c, "frontCard", req.FrontCard) || !web.NotBlank(c, "backCard", req.BackCard) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, ok := callData(c, rpc.ServiceUser, "/changeBind/add", map[string]interface{}{
		"pin":       cmdCtx.Pin,
		"authName":  req.AuthName,
		"authNo":    req.AuthNo,
		"frontCard": req.FrontCard,
		"backCard":  req.BackCard,
		"applyType": req.ApplyType,
	}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/changeBind/checkPreviousPhone
// ---------------------------------------------------------------------------

type clientCheckPhoneCmd struct {
	dto.ClientDTO
	AuthNo     *string `json:"authNo" binding:"required"`     // @NotNull("身份证号不能为空")
	AuthName   *string `json:"authName" binding:"required"`   // @NotNull("身份证姓名不能为空")
	FourNumber *string `json:"fourNumber" binding:"required"` // @NotNull("参数不能为空")
	// Java field default: izCheckComplete = false
	IzCheckComplete *bool `json:"izCheckComplete"`
}

func checkPreviousPhone(c *gin.Context) {
	izCheckDefault := false
	req := clientCheckPhoneCmd{IzCheckComplete: &izCheckDefault}
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authNo":     "身份证号不能为空",
		"authName":   "身份证姓名不能为空",
		"fourNumber": "参数不能为空",
	})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// CheckPhoneCmd copy {authNo, authName, fourNumber, izCheckComplete};
	// the Java gateway does NOT set pin for this call.
	if _, ok := callData(c, rpc.ServiceUser, "/changeBind/checkPreviousPhone", map[string]interface{}{
		"authNo":          req.AuthNo,
		"authName":        req.AuthName,
		"fourNumber":      req.FourNumber,
		"izCheckComplete": req.IzCheckComplete,
	}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/changeBind/withFace
// ---------------------------------------------------------------------------

type clientChangeBindWithFaceCmd struct {
	dto.ClientDTO
	AuthName   *string `json:"authName" binding:"required"`  // @NotNull("姓名不能为空")
	AuthNo     *string `json:"authNo" binding:"required"`    // @NotNull("身份证号不能为空")
	ApplyType  *int    `json:"applyType" binding:"required"` // @NotNull("换绑类型不能为空")
	FourNumber *string `json:"fourNumber"`
}

func changeBindWithFace(c *gin.Context) {
	var req clientChangeBindWithFaceCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authName":  "姓名不能为空",
		"authNo":    "身份证号不能为空",
		"applyType": "换绑类型不能为空",
	})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, ok := callData(c, rpc.ServiceUser, "/changeBind/withFace", map[string]interface{}{
		"pin":        cmdCtx.Pin,
		"authName":   req.AuthName,
		"authNo":     req.AuthNo,
		"applyType":  req.ApplyType,
		"fourNumber": req.FourNumber,
	}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/changeBind/withPhoneAviodAudit
// ---------------------------------------------------------------------------

type clientChangeBindAviodAuditCmd struct {
	dto.ClientDTO
	AuthName  string `json:"authName" binding:"required"`  // @NotBlank
	AuthNo    string `json:"authNo" binding:"required"`    // @NotBlank
	Phone     string `json:"phone" binding:"required"`     // @NotBlank
	ApplyType *int   `json:"applyType" binding:"required"` // @NotNull
}

func changeBindPhoneAviodAudit(c *gin.Context) {
	var req clientChangeBindAviodAuditCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authName":  "must not be blank",
		"authNo":    "must not be blank",
		"phone":     "must not be blank",
		"applyType": "must not be null",
	})) {
		return
	}
	if !web.NotBlank(c, "authName", req.AuthName) || !web.NotBlank(c, "authNo", req.AuthNo) ||
		!web.NotBlank(c, "phone", req.Phone) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// ChangeBindAviodAuditCmd has no applyType field, so the copy drops it.
	if _, ok := callData(c, rpc.ServiceUser, "/changeBind/withPhoneAviodAudit", map[string]interface{}{
		"pin":      cmdCtx.Pin,
		"authName": req.AuthName,
		"authNo":   req.AuthNo,
		"phone":    req.Phone,
	}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/changeBind/cancel
// ---------------------------------------------------------------------------

func cancelChangeBind(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	if _, ok := callData(c, rpc.ServiceUser, "/changeBind/cancel",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}
