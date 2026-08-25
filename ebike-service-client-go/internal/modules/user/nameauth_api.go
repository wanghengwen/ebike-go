package user

// Ports Java UserNameAuthController + UserNameAuthServiceImpl + UserNameAuthGatewayImpl.

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

type authCo struct {
	ResultCode  *int    `json:"resultCode"`
	HandleUser  *string `json:"handleUser"`
	HandlePhone *string `json:"handlePhone"`
	IzAuth      *bool   `json:"izAuth"`
	Msg         *string `json:"msg"`
}

type izNeedAuthCo struct {
	IzNeedAuth *bool `json:"izNeedAuth"`
}

type authStateCo struct {
	AuthState *int  `json:"authState"`
	IzAuth    *bool `json:"izAuth"`
}

func convertAuthCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co authCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertIzNeedAuthCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co izNeedAuthCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertAuthStateCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co authStateCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertBoolean(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var b bool
	if json.Unmarshal(raw, &b) != nil {
		return raw
	}
	out, _ := json.Marshal(b)
	return out
}


// ---------------------------------------------------------------------------
// POST /client/user/auth
// ---------------------------------------------------------------------------

type clientNameAuthCmd struct {
	dto.ClientDTO
	AuthNo   string  `json:"authNo" binding:"required"`   // @NotBlank + @Pattern
	AuthName string  `json:"authName" binding:"required"` // @NotBlank
	Type     *int    `json:"type" binding:"required"`     // @NotNull
	Image    *string `json:"image"`
}

// authNoPattern is the ID-card regex from ClientNameAuthCmd (RE2 compatible).
var authNoPattern = regexp.MustCompile(`^\d{6}((((((19|20)\d{2})(0[13-9]|1[012])(0[1-9]|[12]\d|30))|(((19|20)\d{2})(0[13578]|1[02])31)|((19|20)\d{2})02(0[1-9]|1\d|2[0-8])|((((19|20)([13579][26]|[2468][048]|0[48]))|(2000))0229))\d{3})|((((\d{2})(0[13-9]|1[012])(0[1-9]|[12]\d|30))|((\d{2})(0[13578]|1[02])31)|((\d{2})02(0[1-9]|1\d|2[0-8]))|(([13579][26]|[2468][048]|0[048])0229))\d{2}))(\d|X|x)$`)

// authNoPatternMsg resolves the i18n key
// {UserNameAuthController.auth.ClientNameAuthCmd.authNo} like the Java
// MessageSource (i18n/msg bundle, locale from Accept-Language; zh_CN default).
func authNoPatternMsg(c *gin.Context) string {
	lang := strings.ToLower(c.GetHeader("Accept-Language"))
	if strings.HasPrefix(lang, "en") {
		return "illegal ID number"
	}
	return "身份证号不合法"
}

func nameAuth(c *gin.Context) {
	var req clientNameAuthCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authNo":   "must not be blank",
		"authName": "must not be blank",
		"type":     "must not be null",
	})) {
		return
	}
	if !web.NotBlank(c, "authNo", req.AuthNo) || !web.NotBlank(c, "authName", req.AuthName) {
		return
	}
	if !authNoPattern.MatchString(req.AuthNo) {
		web.WriteParamError(c, "authNo "+authNoPatternMsg(c))
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// AuthCmd copy {authNo, authName, type, image} + pin
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/auth", map[string]interface{}{
		"pin":      cmdCtx.Pin,
		"authNo":   req.AuthNo,
		"authName": req.AuthName,
		"type":     req.Type,
		"image":    req.Image,
	}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAuthCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/rentCheck
// ---------------------------------------------------------------------------

type clientRentCheckCmd struct {
	dto.ClientDTO
	Image string `json:"image" binding:"required"` // @NotBlank
}

func rentCheck(c *gin.Context) {
	var req clientRentCheckCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{"image": "must not be blank"})) {
		return
	}
	if !web.NotBlank(c, "image", req.Image) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// RentCheckCmd{image, pin} (no property copy)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/rentCheck",
		map[string]interface{}{"image": req.Image, "pin": cmdCtx.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/auth/upload
// ---------------------------------------------------------------------------

type clientUploadNameAuthCmd struct {
	dto.ClientDTO
	AuthNo    string `json:"authNo" binding:"required"`    // @NotBlank(message = "")
	AuthName  string `json:"authName" binding:"required"`  // @NotBlank(message = "")
	FrontCard string `json:"frontCard" binding:"required"` // @NotBlank(message = "")
	BackCard  string `json:"backCard" binding:"required"`  // @NotBlank(message = "")
}

func uploadAuth(c *gin.Context) {
	var req clientUploadNameAuthCmd
	// The Java annotations declare message="" so the response is "<field> " (trailing space).
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"authNo": "", "authName": "", "frontCard": "", "backCard": "",
	})) {
		return
	}
	if !notBlankMsg(c, "authNo", req.AuthNo, "") || !notBlankMsg(c, "authName", req.AuthName, "") ||
		!notBlankMsg(c, "frontCard", req.FrontCard, "") || !notBlankMsg(c, "backCard", req.BackCard, "") {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/auth/upload", map[string]interface{}{
		"pin":       cmdCtx.Pin,
		"authNo":    req.AuthNo,
		"authName":  req.AuthName,
		"frontCard": req.FrontCard,
		"backCard":  req.BackCard,
	}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAuthCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/auth/cancel
// ---------------------------------------------------------------------------

func cancelAuth(c *gin.Context) {
	// ClientCancelAuthApplyCmd adds no fields to ClientDTO
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	if _, ok := callData(c, rpc.ServiceUser, "/auth/cancel",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/auth/izNeed
// ---------------------------------------------------------------------------

type clientIzNeedAuthCmd struct {
	dto.ClientDTO
	Pin       *string `json:"pin" binding:"required"` // @NotNull (String)
	ServiceId *int64  `json:"serviceId"`
}

func izNeedAuth(c *gin.Context) {
	var req clientIzNeedAuthCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{"pin": "must not be null"})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// Java builds IzNeedAuthCmd with genParam(Class) (NO property copy) and only
	// sets userPin = cmd.getPin() — the request serviceId is intentionally dropped.
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/auth/izNeed",
		map[string]interface{}{"userPin": req.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertIzNeedAuthCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/auth/state
// ---------------------------------------------------------------------------

func authState(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	// NameAuthApi.authState: @PostMapping("/auth/State") — capital S in Java.
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/auth/State",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAuthStateCo(result.Data)
	}
	web.RespondResult(c, result, err)
}
