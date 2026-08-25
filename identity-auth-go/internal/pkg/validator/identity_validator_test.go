package validator

import "testing"

func TestValidateAuthRequest(t *testing.T) {
	if msg := ValidateAuthRequest("", "1", "name", "id", "", 1); msg != "traceId不能为空" {
		t.Fatalf("unexpected msg: %s", msg)
	}
	if msg := ValidateAuthRequest("t1", "abc", "name", "id", "", 1); msg != "tenantId只能为数字" {
		t.Fatalf("unexpected msg: %s", msg)
	}
	if msg := ValidateAuthRequest("t1", "1", "name", "id", "not-base64!!!", 2); msg != "人脸照非base64格式" {
		t.Fatalf("unexpected msg: %s", msg)
	}
}

func TestValidateAuthRecordRequest(t *testing.T) {
	if msg := ValidateAuthRecordRequest("count", 1, "2024_01", "", "t1"); msg != "tenantId不能为空" {
		t.Fatalf("unexpected msg: %s", msg)
	}
	if msg := ValidateAuthRecordRequest("query", 1, "bad-date", "1", "t1"); msg != "日期格式应为yyyy_MM" {
		t.Fatalf("unexpected msg: %s", msg)
	}
}
