package model

import (
	"encoding/json"
	"strconv"
	"strings"
)

// R is the unified response structure, matching Java's R<T> class.
// Java uses: { "success": true/false, "msg": "...", "data": ... }
type R struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// AuthRequest matches Java's AuthRequest (extends ClientDTO).
// Field names must match Java's JSON serialization exactly.
type AuthRequest struct {
	TraceId   string `json:"traceId" binding:"required"`
	TenantId  string `json:"tenantId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	IdCardNum string `json:"idCardNum" binding:"required"` // Java uses idCardNum, not identityNo
	Image     string `json:"image"`                        // Java uses image, not imageUrl (base64 for 三要素)
	Type      int    `json:"type" binding:"required"`      // 1=二要素, 2=三要素
}

// UnmarshalJSON mirrors Java fastjson leniency: tenantId/type may arrive as
// either JSON strings or numbers.
func (r *AuthRequest) UnmarshalJSON(data []byte) error {
	var aux struct {
		TraceId   string          `json:"traceId"`
		TenantId  json.RawMessage `json:"tenantId"`
		Name      string          `json:"name"`
		IdCardNum string          `json:"idCardNum"`
		Image     string          `json:"image"`
		Type      json.RawMessage `json:"type"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.TraceId = aux.TraceId
	r.Name = aux.Name
	r.IdCardNum = aux.IdCardNum
	r.Image = aux.Image
	r.TenantId = rawToString(aux.TenantId)
	r.Type = rawToInt(aux.Type)
	return nil
}

// ChargeRequest matches Java's ChargeRequest (extends ClientDTO).
type ChargeRequest struct {
	TraceId  string `json:"traceId" binding:"required"`
	TenantId string `json:"tenantId" binding:"required"`
	Amount   string `json:"amount" binding:"required"` // Java BigDecimal: accept JSON number or string
	Quantity int64  `json:"quantity" binding:"required"`
	Type     int    `json:"type" binding:"required"` // 1=二要素充值, 2=三要素充值
}

// UnmarshalJSON mirrors Java fastjson leniency: amount (BigDecimal) accepts
// both JSON numbers and strings; tenantId/type/quantity accept either form.
func (r *ChargeRequest) UnmarshalJSON(data []byte) error {
	var aux struct {
		TraceId  string          `json:"traceId"`
		TenantId json.RawMessage `json:"tenantId"`
		Amount   json.RawMessage `json:"amount"`
		Quantity json.RawMessage `json:"quantity"`
		Type     json.RawMessage `json:"type"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.TraceId = aux.TraceId
	r.TenantId = rawToString(aux.TenantId)
	r.Amount = rawToString(aux.Amount)
	r.Quantity = rawToInt64(aux.Quantity)
	r.Type = rawToInt(aux.Type)
	return nil
}

// AuthRecordRequest matches Java's AuthRecordRequest.
// Used by /countAuthTimes, /queryAuthRecord, /exportAuthRecord.
type AuthRecordRequest struct {
	Type     int    `json:"type" binding:"required"`   // 1=二要素, 2=三要素
	Date     string `json:"date" binding:"required"`   // yyyy_MM format
	TenantId string `json:"tenantId"`                  // Required when action=count
	Action   string `json:"action" binding:"required"` // "query" or "count"
	TraceId  string `json:"traceId" binding:"required"`
}

// UnmarshalJSON mirrors Java fastjson leniency: tenantId/type accept either
// JSON strings or numbers.
func (r *AuthRecordRequest) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type     json.RawMessage `json:"type"`
		Date     string          `json:"date"`
		TenantId json.RawMessage `json:"tenantId"`
		Action   string          `json:"action"`
		TraceId  string          `json:"traceId"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Date = aux.Date
	r.Action = aux.Action
	r.TraceId = aux.TraceId
	r.TenantId = rawToString(aux.TenantId)
	r.Type = rawToInt(aux.Type)
	return nil
}

// rawToString extracts a string from a JSON value that may be a string or a
// number (matching fastjson's coercion of any scalar into a String field).
func rawToString(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return ""
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str); err == nil {
			return str
		}
		return ""
	}
	return s
}

// rawToInt parses an int from a JSON value that may be a number or a string.
func rawToInt(raw json.RawMessage) int {
	return int(rawToInt64(raw))
}

// rawToInt64 parses an int64 from a JSON value that may be a number or a string.
func rawToInt64(raw json.RawMessage) int64 {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return 0
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str); err != nil {
			return 0
		}
		s = strings.TrimSpace(str)
		if s == "" {
			return 0
		}
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	return 0
}
