package validator

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// blank reports whether s is empty or whitespace-only, matching Spring's
// ValidationUtils.rejectIfEmptyOrWhitespace.
func blank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ValidateAuthRequest returns the first validation error message, or "" if valid.
// Matches Java AuthRequestValidator.
func ValidateAuthRequest(traceId, tenantId, name, idCardNum, image string, authType int) string {
	if blank(traceId) {
		return "traceId不能为空"
	}
	if blank(tenantId) {
		return "tenantId不能为空"
	}
	if blank(idCardNum) {
		return "idCardNum不能为空"
	}
	if blank(name) {
		return "name不能为空"
	}
	if authType == 0 {
		return "type不能为空"
	}
	if authType != 1 && authType != 2 {
		return "type参数错误"
	}
	if len(traceId) > 64 {
		return "traceId不能超过64位"
	}
	if msg := validateTenantIDNumeric(tenantId); msg != "" {
		return msg
	}
	if authType == 2 {
		if image == "" {
			return "image不能为空"
		}
		if _, err := base64.StdEncoding.DecodeString(image); err != nil {
			return "人脸照非base64格式"
		}
	}
	return ""
}

// ValidateChargeRequest returns the first validation error message, or "" if valid.
// Matches Java ChargeRequestValidator.
func ValidateChargeRequest(traceId, tenantId, amount string, quantity int64, authType int) string {
	if blank(traceId) {
		return "traceId不能为空"
	}
	if blank(tenantId) {
		return "tenantId不能为空"
	}
	if blank(amount) {
		return "amount不能为空"
	}
	if quantity == 0 {
		return "quantity不能为空"
	}
	if authType == 0 {
		return "type不能为空"
	}
	if authType != 1 && authType != 2 {
		return "type参数错误"
	}
	if msg := validateTenantIDNumeric(tenantId); msg != "" {
		return msg
	}
	if len(traceId) > 64 {
		return "traceId不能超过64位"
	}
	return ""
}

// ValidateAuthRecordRequest returns the first validation error message, or "" if valid.
// Matches Java CountAuthRequestValidator.
func ValidateAuthRecordRequest(action string, authType int, date, tenantId, traceId string) string {
	if action == "" {
		return "action不能为空"
	}
	if authType == 0 {
		return "type不能为空"
	}
	if blank(date) {
		return "date不能为空"
	}
	if blank(traceId) {
		return "traceId不能为空"
	}
	if authType != 1 && authType != 2 {
		return "type参数错误"
	}
	if action != "query" && action != "count" {
		return "action参数错误"
	}
	if _, err := time.Parse("2006_01", date); err != nil {
		return "日期格式应为yyyy_MM"
	}
	if action == "count" && tenantId == "" {
		return "tenantId不能为空"
	}
	if tenantId != "" {
		if msg := validateTenantIDNumeric(tenantId); msg != "" {
			return msg
		}
	}
	return ""
}

func validateTenantIDNumeric(tenantId string) string {
	if _, err := strconv.Atoi(tenantId); err != nil {
		return "tenantId只能为数字"
	}
	return ""
}
