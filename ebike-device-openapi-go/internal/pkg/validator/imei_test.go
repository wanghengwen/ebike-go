package validator

import "testing"

func TestValidateIMEI(t *testing.T) {
	if err := ValidateIMEI(""); err == nil || err.Msg != "imei不能为空" {
		t.Fatalf("expected empty imei error, got %+v", err)
	}
	if err := ValidateIMEI("123"); err == nil || err.Msg != "imei应为15位数字" {
		t.Fatalf("expected length error, got %+v", err)
	}
	if err := ValidateIMEI("86255105985840X"); err == nil {
		t.Fatal("expected non-digit error")
	}
	if err := ValidateIMEI("862551059858406"); err != nil {
		t.Fatalf("expected valid imei, got %+v", err)
	}
}
