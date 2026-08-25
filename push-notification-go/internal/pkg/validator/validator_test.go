package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidateChinaPhone(t *testing.T) {
	v := validator.New()
	_ = v.RegisterValidation("china_phone", validateChinaPhone)

	type req struct {
		Phone string `validate:"china_phone"`
	}

	cases := []struct {
		phone string
		valid bool
	}{
		{"13800138000", true},
		{"19912345678", true},
		{"12000000000", false},
		{"23800138000", false},
		{"1380013800", false},
		{"", false},
	}

	for _, tc := range cases {
		err := v.Struct(req{Phone: tc.phone})
		got := err == nil
		if got != tc.valid {
			t.Errorf("phone=%q valid=%v want=%v err=%v", tc.phone, got, tc.valid, err)
		}
	}
}

func TestFormatBindErrorPhonePattern(t *testing.T) {
	v := validator.New()
	_ = v.RegisterValidation("china_phone", validateChinaPhone)

	type msgReq struct {
		Phone string `validate:"required,china_phone"`
	}

	err := v.Struct(msgReq{Phone: "12000000000"})
	bizErr := FormatBindError(err, false)
	if bizErr.Code != "4000" {
		t.Fatalf("code=%s want 4000", bizErr.Code)
	}
	if bizErr.Msg != "phone 手机号格式错误" {
		t.Fatalf("msg=%q want %q", bizErr.Msg, "phone 手机号格式错误")
	}
}

func TestFormatBindErrorPhoneRequired(t *testing.T) {
	v := validator.New()
	_ = v.RegisterValidation("china_phone", validateChinaPhone)

	type msgReq struct {
		Phone string `validate:"required,china_phone"`
	}

	err := v.Struct(msgReq{Phone: ""})
	bizErr := FormatBindError(err, false)
	if bizErr.Msg != "phone 手机号不能为空" {
		t.Fatalf("msg=%q want %q", bizErr.Msg, "phone 手机号不能为空")
	}
}
