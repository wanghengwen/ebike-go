package validator

import "ebike-device-openapi-go/internal/api/dto"

// ValidateIMEI mirrors Java @Length(min=15, max=15) on EBikeRequest.imei.
func ValidateIMEI(imei string) *dto.Result {
	if imei == "" {
		return &dto.Result{Success: false, Code: "10002", Msg: "imei不能为空"}
	}
	if len(imei) != 15 {
		return &dto.Result{Success: false, Code: "10002", Msg: "imei应为15位数字"}
	}
	for _, ch := range imei {
		if ch < '0' || ch > '9' {
			return &dto.Result{Success: false, Code: "10002", Msg: "imei应为15位数字"}
		}
	}
	return nil
}
