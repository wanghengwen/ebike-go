package rpc

import (
	"context"
	"fmt"

	"ebike-fence-go/internal/api/dto"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
)

type OrderRPC struct{}

func NewOrderRPC() *OrderRPC { return &OrderRPC{} }

// HasOrder mirrors Java OrderApi.hasOrder — true when user has orders (old user).
func (r *OrderRPC) HasOrder(ctx context.Context, pin string, cmdCtx *dto.CommandContext) (bool, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"userPin":        pin,
	}
	var res Result[bool]
	if err := pkg_rpc.PostToService(ctx, "ebike-order", "/order/hasOrder", body, &res); err != nil {
		return false, err
	}
	if !res.OK() {
		return false, fmt.Errorf("order api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}
