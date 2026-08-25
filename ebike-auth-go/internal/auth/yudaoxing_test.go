package auth

import "testing"

func TestYudaoUserInfo_PhoneForSocialRegister(t *testing.T) {
	tests := []struct {
		name         string
		yuhuanCustom bool
		phone        string
		userCode     string
		want         string
	}{
		// 默认行为：保持原有实现，返回 phone 字段原值，不加工。
		{name: "default empty", yuhuanCustom: false, phone: "", userCode: "13800138000", want: ""},
		{name: "default use phone", yuhuanCustom: false, phone: "13800138000", userCode: "x", want: "13800138000"},
		// 玉环定制：对齐 ydx getPhone()，取 userCode 补 +86-。
		{name: "yuhuan empty", yuhuanCustom: true, phone: "", userCode: "", want: ""},
		{name: "yuhuan raw userCode", yuhuanCustom: true, phone: "", userCode: "13800138000", want: "+86-13800138000"},
		{name: "yuhuan already prefixed", yuhuanCustom: true, phone: "", userCode: "+86-13800138000", want: "+86-13800138000"},
		{name: "yuhuan fallback to phone", yuhuanCustom: true, phone: "13800138000", userCode: "", want: "+86-13800138000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &YudaoUserInfo{Phone: tt.phone, UserCode: tt.userCode}
			if got := u.PhoneForSocialRegister(tt.yuhuanCustom); got != tt.want {
				t.Fatalf("PhoneForSocialRegister(%v) = %q, want %q", tt.yuhuanCustom, got, tt.want)
			}
		})
	}
}

func TestYudaoUserInfo_YuhuanYudaoxingLoginBlockMsg(t *testing.T) {
	// realFlag=0 时，仅玉环定制开启才拦截。
	u := &YudaoUserInfo{RealFlag: "0"}
	if got := u.YuhuanYudaoxingLoginBlockMsg(false); got != "" {
		t.Fatalf("disabled custom should not block, got %q", got)
	}
	if got := u.YuhuanYudaoxingLoginBlockMsg(true); got != "未实名，请进行实名认证" {
		t.Fatalf("enabled custom block msg = %q", got)
	}

	// 已实名（realFlag 非 "0"）即使开启也不拦截。
	authed := &YudaoUserInfo{RealFlag: "1"}
	if got := authed.YuhuanYudaoxingLoginBlockMsg(true); got != "" {
		t.Fatalf("authed user should not block, got %q", got)
	}
}
