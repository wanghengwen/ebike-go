package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"identity-auth-go/internal/model"
	"identity-auth-go/internal/pkg/shadow"
	idvalidator "identity-auth-go/internal/pkg/validator"
	"identity-auth-go/internal/service"

	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	svc *service.QueryService
}

func NewQueryHandler(svc *service.QueryService) *QueryHandler {
	return &QueryHandler{svc: svc}
}

// queryRecordJSON mirrors the JSON shape produced by Java's
// CallRecordServiceImpl.query() -> R.data (a list of CallRecord objects):
// camelCase keys, curated field set, callAt as epoch millis, responseData as a
// nested JSON object, and a constant score of 0 (Java's query() never sets it).
type queryRecordJSON struct {
	TenantId     string          `json:"tenantId"`
	IdentityNo   string          `json:"identityNo"`
	Name         string          `json:"name"`
	CallAt       int64           `json:"callAt"`
	Status       int             `json:"status"`
	Score        int             `json:"score"`
	ResponseData json.RawMessage `json:"responseData,omitempty"`
}

func (h *QueryHandler) validateAuthRecordRequest(req model.AuthRecordRequest) string {
	return idvalidator.ValidateAuthRecordRequest(req.Action, req.Type, req.Date, req.TenantId, req.TraceId)
}

func toQueryJSON(rows []service.AuthRecordRow) []queryRecordJSON {
	list := make([]queryRecordJSON, 0, len(rows))
	for _, row := range rows {
		item := queryRecordJSON{
			TenantId:   row.TenantId,
			IdentityNo: row.IdentityNo,
			Name:       row.Name,
			CallAt:     callAtMillis(row),
			Status:     row.Status,
			Score:      0,
		}
		if rd := strings.TrimSpace(row.ResponseData); rd != "" && json.Valid([]byte(rd)) {
			item.ResponseData = json.RawMessage(rd)
		}
		list = append(list, item)
	}
	return list
}

// callAtMillis mirrors Java: new Timestamp(localDateTime.toEpochSecond()*1000),
// i.e. seconds precision multiplied to milliseconds.
func callAtMillis(row service.AuthRecordRow) int64 {
	if row.CallAt.IsZero() {
		return 0
	}
	return row.CallAt.Unix() * 1000
}

// CountAuthTimes handles POST /countAuthTimes
// Java expects AuthRecordRequest with action="count"
func (h *QueryHandler) CountAuthTimes(c *gin.Context) {
	reqBody, _ := c.GetRawData()

	var req model.AuthRecordRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}
	if msg := h.validateAuthRecordRequest(req); msg != "" {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: msg})
		return
	}

	result, err := h.svc.CountAuthTimes(req)
	if err != nil {
		log.Printf("[CountAuthTimes] DB error: %v, traceId=%s", err, req.TraceId)
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}

	resp := model.R{
		Success: true,
		Data:    result,
	}

	goRespBytes, _ := json.Marshal(resp)
	shadow.CompareWithJava("POST", c.Request.URL.Path, reqBody, goRespBytes)

	c.Data(http.StatusOK, "application/json", goRespBytes)
}

// QueryAuthRecord handles POST /queryAuthRecord
// Java expects AuthRecordRequest with action="query"
func (h *QueryHandler) QueryAuthRecord(c *gin.Context) {
	reqBody, _ := c.GetRawData()

	var req model.AuthRecordRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}
	if msg := h.validateAuthRecordRequest(req); msg != "" {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: msg})
		return
	}

	records, err := h.svc.QueryAuthRecord(req)
	if err != nil {
		log.Printf("[QueryAuthRecord] DB error: %v, traceId=%s", err, req.TraceId)
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}

	resp := model.R{
		Success: true,
		Data:    toQueryJSON(records),
	}

	goRespBytes, _ := json.Marshal(resp)
	shadow.CompareWithJava("POST", c.Request.URL.Path, reqBody, goRespBytes)

	c.Data(http.StatusOK, "application/json", goRespBytes)
}

// ExportAuthRecord handles GET /exportAuthRecord.
// Mirrors Java CountAuthHandler.exportAuthRecord + CountAuthServiceImpl.exportAuthRecord:
// a header buffer followed by per-record buffers concatenated WITHOUT separators
// (Java emits no line breaks between records).
func (h *QueryHandler) ExportAuthRecord(c *gin.Context) {
	reqBody, _ := c.GetRawData()

	var req model.AuthRecordRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}
	if msg := h.validateAuthRecordRequest(req); msg != "" {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: msg})
		return
	}

	records, err := h.svc.QueryAuthRecord(req)
	if err != nil {
		log.Printf("[ExportAuthRecord] DB error: %v, traceId=%s", err, req.TraceId)
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}

	var sb strings.Builder
	sb.WriteString("租户ID,姓名,身份证号,认证时间,认证结果,认证明细")
	for _, row := range records {
		statusText := "失败"
		if row.Status == 1 {
			statusText = "成功"
		}
		callAt := ""
		if !row.CallAt.IsZero() {
			callAt = row.CallAt.Format("2006-01-02 15:04:05")
		}
		sb.WriteString(row.TenantId + "," + row.Name + "," + row.IdentityNo + "," + callAt + "," + statusText + "," + row.ResponseData)
	}

	c.Header("Content-Disposition", "attachment; filename=AuthRecord.csv")
	c.Header("Accept-Ranges", "bytes")
	c.Data(http.StatusOK, "application/octet-stream", []byte(sb.String()))
}
