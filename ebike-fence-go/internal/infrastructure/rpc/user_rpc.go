package rpc

import (
	"context"
	"fmt"

	"ebike-fence-go/internal/api/dto"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
)

type UserRPC struct{}

func NewUserRPC() *UserRPC { return &UserRPC{} }

type AppUser struct {
	Pin   string `json:"pin"`
	Phone string `json:"phone"`
}

// BatchGetByPins mirrors Java UserApiRpc.batchGet (ebike-user /user/getList).
func (r *UserRPC) BatchGetByPins(ctx context.Context, cmdCtx *dto.CommandContext, pins []string) (map[string]AppUser, error) {
	if len(pins) == 0 {
		return map[string]AppUser{}, nil
	}
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"pins":           pins,
	}
	var res Result[[]AppUser]
	if err := pkg_rpc.PostToService(ctx, "ebike-user", "/user/getList", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("user api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	out := make(map[string]AppUser, len(res.Data))
	for _, u := range res.Data {
		if u.Pin != "" {
			out[u.Pin] = u
		}
	}
	return out, nil
}

// UserDetail mirrors Java UserApiRpc.getUserByPin -> UserApi.getUserDetail result.
type UserDetail struct {
	Pin      string `json:"pin"`
	AuthName string `json:"authName"`
	Phone    string `json:"phone"`
}

// GetUserDetailByPin mirrors Java UserApiRpc.getUserByPin (ebike-user /user/detail).
func (r *UserRPC) GetUserDetailByPin(ctx context.Context, cmdCtx *dto.CommandContext, pin string) (*UserDetail, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"pin":            pin,
	}
	var res Result[*UserDetail]
	if err := pkg_rpc.PostToService(ctx, "ebike-user", "/user/detail", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("user api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// CreditScoreInit mirrors Java CreditScoreApi.init (ebike-user /creditScore/init).
// Java propagates downstream failure (ResultHelper.getResultData throws), so the
// caller must abort the credit-score enable on error.
func (r *UserRPC) CreditScoreInit(ctx context.Context, cmdCtx *dto.CommandContext, score *float64, pin string) error {
	body := map[string]interface{}{"commandContext": cmdCtx}
	if score != nil {
		body["score"] = *score
	}
	if pin != "" {
		body["pin"] = pin
	}
	var res Result[bool]
	if err := pkg_rpc.PostToService(ctx, "ebike-user", "/creditScore/init", body, &res); err != nil {
		return err
	}
	if !res.OK() {
		return fmt.Errorf("creditScore init error: code=%d msg=%s", res.Code, res.Msg)
	}
	return nil
}
