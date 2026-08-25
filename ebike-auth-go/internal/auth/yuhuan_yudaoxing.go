package auth

import "strings"

// 玉环公共电单车（客户定制）玉岛行登录扩展。
//
// 对应 ebike-auth-client 的 ydx 分支行为，仅在该客户场景启用：
//   - 玉岛行侧 realFlag=0 时拒绝登录
//   - 社交注册手机号由 userCode 规范为 +86- 前缀
//
// 开启方式（任选其一，运行时环境变量优先级最高）：
//   - 编译：go build -tags=yuhuan_yudaoxing
//   - 环境变量：YUHUAN_YUDAOXING_CUSTOM=true
//   - Nacos：xyy.yuhuanYudaoxingCustom: true

// YuhuanYudaoxingLoginBlockMsg 返回玉环定制下的登录拦截文案；空字符串表示允许继续登录。
func (u *YudaoUserInfo) YuhuanYudaoxingLoginBlockMsg(yuhuanCustom bool) string {
	if !yuhuanCustom || u == nil {
		return ""
	}
	if u.RealFlag == "0" {
		return "未实名，请进行实名认证"
	}
	return ""
}

// PhoneForSocialRegister 返回社交注册使用的手机号。
//   - 默认（未开启玉环定制）：保持原有行为，返回玉岛行响应的 phone 字段，不做任何加工。
//   - 玉环定制开启：对齐 ydx 分支 getPhone()，取 userCode 并补全 +86- 前缀
//     （userCode 为空时回退到 phone 字段，保证不误伤）。
func (u *YudaoUserInfo) PhoneForSocialRegister(yuhuanCustom bool) string {
	if u == nil {
		return ""
	}
	if !yuhuanCustom {
		return u.Phone
	}
	code := u.UserCode
	if code == "" {
		code = u.Phone
	}
	if code == "" {
		return ""
	}
	if strings.HasPrefix(code, "+86-") {
		return code
	}
	return "+86-" + code
}
