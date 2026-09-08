package order

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

func bindUserBillConfigQuery(c *gin.Context, req *userBillConfigQuery) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, billConfigMsgs)
}

func getBillConfig(c *gin.Context) {
	var req userBillConfigQuery
	if !bindUserBillConfigQuery(c, &req) {
		return
	}
	if !web.NotBlank(c, "userPin", req.UserPin) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	raw, ok := callForData(c, rpc.ServiceOrder, "/order/config/getUser", &req, cmdCtx)
	if !ok {
		return
	}
	backBody := map[string]*int64{"id": req.ServiceId}
	backResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/backcar/getConfigByServiceId", backBody, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !backResult.Success {
		web.WriteBizResult(c, backResult)
		return
	}
	if isNullRaw(backResult.Data) {
		javacompat.WriteNPE(c)
		return
	}
	out := map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &out)
	merged := javacompat.ClientDTONullMap()
	for k, v := range out {
		if k == "commandContext" {
			continue
		}
		switch k {
		case "id", "serviceId":
			if p := longStr(v); p != nil {
				merged[k] = *p
			} else {
				merged[k] = nil
			}
		case "ladderItem":
			if parsed := javacompat.ParseJSONArrayField(v); parsed != nil {
				var arr interface{}
				_ = json.Unmarshal(parsed, &arr)
				merged[k] = arr
			} else {
				merged[k] = nil
			}
		default:
			switch k {
			case "discount":
				var val float64
				if json.Unmarshal(v, &val) == nil {
					merged[k] = val
				}
			default:
				var val interface{}
				if json.Unmarshal(v, &val) == nil {
					merged[k] = val
				}
			}
		}
	}
	var back map[string]json.RawMessage
	if json.Unmarshal(backResult.Data, &back) == nil {
		for _, k := range []string{"allowOutofService", "allowInNostop", "allowOutofParking",
			"penaltyInNostop", "penaltyOutofService", "dispatchCost", "allowInBanRiding", "penaltyInBanRiding"} {
			if v, ok := back[k]; ok {
				var val interface{}
				if json.Unmarshal(v, &val) == nil {
					merged[k] = val
				}
			}
		}
	}
	ensureCBillingConfigNullFields(merged)
	merged["izPopup"] = resolveBillConfigIzPopup(c, req, out, cmdCtx)
	mb, _ := json.Marshal(merged)
	// Marshal the merged map directly so ClientDTO parent fields stay null
	// (embedding dto.ClientDTO would turn null into "" / omitempty drops).
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(mb)))
}

// ensureCBillingConfigNullFields adds CBillingConfigCmd fields that Java always
// serializes but order getUserBillConfig omits (fields commented out on BillingConfigCmd).
func ensureCBillingConfigNullFields(m map[string]interface{}) {
	if _, ok := m["maxPenaltyOutofParking"]; !ok {
		m["maxPenaltyOutofParking"] = nil
	}
}

// resolveBillConfigIzPopup mirrors BillConfigServiceImpl.getBillConfig popup logic.
// SHADOW: /popup/bill mutates t_popup; Java runs before Go mirror, so re-calling
// popup would return a different boolean. Reuse Java's izPopup from X-Shadow-Java-Result.
func resolveBillConfigIzPopup(c *gin.Context, req userBillConfigQuery, out map[string]json.RawMessage, cmdCtx *dto.CommandContext) bool {
	const defaultPopup = true
	if req.UserPin == "" || req.Scene != 1 {
		return defaultPopup
	}
	if v, ok := javacompat.DataFieldBoolFromShadowResult(c.Request.Header.Get("X-Shadow-Java-Result"), "izPopup"); ok {
		return v
	}
	if idRaw, ok := out["id"]; ok {
		popBody := popupCmdBody{UserPin: req.UserPin, BillConfigId: parseBillConfigID(longStr(idRaw))}
		if popResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOrder, "/popup/bill", popBody, cmdCtx); err == nil && popResult != nil && popResult.Success {
			var b bool
			if json.Unmarshal(popResult.Data, &b) == nil {
				return b
			}
		}
	}
	return defaultPopup
}

func stringifyAny(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}

func parseBillConfigID(v *string) int64 {
	if v == nil {
		return 0
	}
	n, _ := strconv.ParseInt(*v, 10, 64)
	return n
}

func invoiceCreate(c *gin.Context) {
	var req invoiceDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if !validateInvoiceCreate(c, &req) {
		return
	}
	orderIds := stringifyOrderIds(req.OrderIds)
	body := invoiceCmdBody{
		ServiceId: req.ServiceId, Type: req.Type, Title: req.Title, CompanyEin: req.CompanyEin,
		Email: req.Email, Content: req.Content, Bank: req.Bank, CompanyAddress: req.CompanyAddress,
		BankAccount: req.BankAccount, CompanyPhone: req.CompanyPhone, Phone: req.Phone, OrderIds: orderIds,
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	if _, ok := callForData(c, rpc.ServiceOrder, "/invoice/create", body, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

func validateInvoiceCreate(c *gin.Context, req *invoiceDTO) bool {
	// @NotNull allows empty string; only null is rejected (Java Bean Validation).
	checks := []struct {
		field string
		ok    bool
	}{
		{"serviceId", req.ServiceId != nil}, {"type", req.Type != nil}, {"title", req.Title != nil},
		{"companyEin", req.CompanyEin != nil}, {"email", req.Email != nil},
		{"content", req.Content != nil},
	}
	for _, ch := range checks {
		if !ch.ok {
			web.WriteParamError(c, ch.field+" must not be null")
			return false
		}
	}
	return validateInvoiceOrderIds(c, req.OrderIds)
}

func validateInvoiceOrderIds(c *gin.Context, raw json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		web.WriteParamError(c, "orderIds must not be blank")
		return false
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		if len(arr) == 0 {
			web.WriteParamError(c, "orderIds must not be blank")
			return false
		}
		return true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if strings.TrimSpace(s) == "" {
			web.WriteParamError(c, "orderIds must not be blank")
			return false
		}
		return true
	}
	web.WriteParamError(c, "orderIds must not be blank")
	return false
}

func stringifyOrderIds(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		s := ""
		return &s
	}
	return &[]string{string(raw)}[0]
}

func invoicePage(c *gin.Context) {
	req := dto.PageClientDTO{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, pageMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	searchCount := true
	if req.SearchCount != nil {
		searchCount = *req.SearchCount
	}
	body := invoiceQueryBody{
		PageNum: req.PageNum, PageSize: req.PageSize, Orders: req.Orders,
		SearchCount: searchCount, LastRecordId: req.LastRecordId, UserPin: cmdCtx.Pin,
	}
	raw, ok := callForData(c, rpc.ServiceOrder, "/invoice/pageInvoice", body, cmdCtx)
	if !ok {
		return
	}
	var page pageDTOIn
	if json.Unmarshal(raw, &page) != nil {
		web.WriteException(c, "failed to decode invoice page")
		return
	}
	out := pageOutFrom(&page)
	var list []cInvoiceCO
	for _, item := range page.List {
		if co, ok := convertCInvoiceCO(item); ok {
			list = append(list, co)
		}
	}
	if len(page.List) > 0 && len(list) == 0 {
		web.WriteException(c, "failed to decode invoice page list")
		return
	}
	if list == nil {
		list = []cInvoiceCO{}
	}
	out.List = list
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

func createReturnBikeAudit(c *gin.Context) {
	var req returnBikeAuditDTO
	if !web.BindJSON(c, &req, returnBikeAuditMsgs) {
		return
	}
	if !web.NotBlank(c, "photoUrl", req.PhotoUrl) {
		return
	}
	if utf16Len(req.UserReason) > 50 {
		web.WriteParamError(c, "userReason size must be between 0 and 50")
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, ok := callForData(c, rpc.ServiceOrder, "/returnBikeAudit/createReturnBikeAudit", &req, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

func izCanCameraAudit(c *gin.Context) {
	var req cameraAuditDTO
	if !web.BindJSON(c, &req, cameraAuditMsgs) {
		return
	}
	if !web.NotBlank(c, "carId", req.CarId) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/part/izCanCameraAudit", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

func izCapable(c *gin.Context) {
	var req returnBikeAuditNoValid
	if !web.BindJSON(c, &req, nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOrder, "/returnBikeAudit/izCapable", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

func createUserTicket(c *gin.Context) {
	var req userTicketDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if req.OrderId == nil {
		web.WriteParamError(c, "orderId must not be null")
		return
	}
	if req.Initiator == nil {
		web.WriteParamError(c, "initiator must not be null")
		return
	}
	photo := joinPhotoUrls(req.PhotoUrl)
	if strings.TrimSpace(photo) == "" {
		web.WriteParamError(c, "photoUrl must not be blank")
		return
	}
	if req.UserReason != nil && utf16Len(*req.UserReason) > 200 {
		web.WriteParamError(c, "userReason size must be between 0 and 200")
		return
	}
	body := userTicketCmdBody{OrderId: req.OrderId, Initiator: req.Initiator, PhotoUrl: &photo, UserReason: req.UserReason}
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	if _, ok := callForData(c, rpc.ServiceOrder, "/userTicket/createUserTicket", body, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

func joinPhotoUrls(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var arr []string
	if json.Unmarshal(raw, &arr) == nil {
		if len(arr) > 0 && len(arr) < 4 {
			return strings.Join(arr, ",")
		}
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return ""
}

func getUserTicket(c *gin.Context) {
	req := pageClientNoValid{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, nil) {
		return
	}
	req.normalizeSearchCount()
	cmdCtx := middleware.CompleteCommandContext(c, req.clientDTOPtr())
	phone := fetchUserPhone(c, cmdCtx)
	body := userTicketListQueryBody{
		PageNum: req.PageNum, PageSize: req.PageSize, Orders: req.Orders,
		SearchCount: *req.SearchCount, LastRecordId: req.LastRecordId,
	}
	if phone != nil {
		b, _ := json.Marshal(phone)
		body.Phone = b
	}
	raw, ok := callForData(c, rpc.ServiceOrder, "/userTicket/pageUserTickets", body, cmdCtx)
	if !ok {
		return
	}
	var page pageDTOIn
	if json.Unmarshal(raw, &page) != nil {
		web.WriteException(c, "failed to decode user ticket page")
		return
	}
	for i, item := range page.List {
		var row map[string]interface{}
		if json.Unmarshal(item, &row) == nil {
			if iz, ok := row["izPaid"].(float64); ok {
				if int(iz) == 3 {
					row["izPaid"] = 0
				} else {
					row["izPaid"] = 1
				}
				b, _ := json.Marshal(row)
				page.List[i] = b
			}
		}
	}
	out := pageOutFrom(&page)
	var list []userTicketListCO
	for _, item := range page.List {
		var co userTicketListCO
		if json.Unmarshal(item, &co) == nil {
			list = append(list, co)
		}
	}
	if len(page.List) > 0 && len(list) == 0 {
		web.WriteException(c, "failed to decode user ticket page list")
		return
	}
	if list == nil {
		list = []userTicketListCO{}
	}
	out.List = list
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

func fetchUserPhone(c *gin.Context, cmdCtx *dto.CommandContext) *string {
	if cmdCtx == nil || cmdCtx.Pin == "" {
		return nil
	}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/user/detail", map[string]string{"pin": cmdCtx.Pin}, cmdCtx)
	if err != nil || result == nil || !result.Success || isNullRaw(result.Data) {
		return nil
	}
	var u struct {
		Phone *string `json:"phone"`
	}
	if json.Unmarshal(result.Data, &u) == nil {
		return u.Phone
	}
	return nil
}

func eBikeLocation(c *gin.Context) {
	var req clientDeviceLocationQry
	if !web.BindJSON(c, &req, deviceLocationMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmdCtx.TenantId = req.TenantId
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceDevicePaas, "/device/paas/device/eBikeLocation", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertDeviceLocationCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

func getBlueToothToken(c *gin.Context) {
	var req clientImeiDTO
	if !web.BindJSON(c, &req, imeiMsgs) {
		return
	}
	if !web.NotBlank(c, "imei", req.Imei) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceDevicePaas, "/device/paas/device/getBlueToothToken", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBlueToothTokenCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

func carSearchVoice(c *gin.Context) {
	var req clientImeiDTO
	if !web.BindJSON(c, &req, imeiMsgs) {
		return
	}
	if !web.NotBlank(c, "imei", req.Imei) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := voiceCmdBody{Async: false, Imei: req.Imei, Idx: 9}
	if _, ok := callForData(c, rpc.ServiceDevicePaas, "/device/paas/voice", body, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(true))
}
