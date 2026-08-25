package marketing

import (
	"encoding/json"

	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// userAddVoucher ports VoucherController.addUserVoucher.
// Java gateway: UserAddVoucherCmd = genParam(UserAddVoucherCmd.class, dto)
// (copies serviceId + voucherCode) -> voucherApi.addUserVoucher.
func userAddVoucher(c *gin.Context) {
	var req userAddVoucherDTO
	if !web.BindJSON(c, &req, voucherMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userAddVoucherCmd{ServiceId: req.ServiceId, VoucherCode: req.VoucherCode}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/voucher/user_add_voucher", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAddUserVoucherCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type addUserVoucherCO struct {
	RewardType *int    `json:"rewardType"`
	RewardInfo *string `json:"rewardInfo"`
}

func convertAddUserVoucherCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co addUserVoucherCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}
