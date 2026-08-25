package rpc

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
)

// DeviceHelmetAdapter adapts DeviceRPC to gateway.DeviceHelmetClient.
type DeviceHelmetAdapter struct {
	*DeviceRPC
}

func NewDeviceHelmetAdapter(r *DeviceRPC) *DeviceHelmetAdapter {
	return &DeviceHelmetAdapter{DeviceRPC: r}
}

func (a *DeviceHelmetAdapter) GetDeviceDetail(ctx context.Context, cmdCtx *dto.CommandContext, imei string) (*gateway.HelmetDeviceDetail, error) {
	d, err := a.DeviceRPC.GetDeviceDetail(ctx, cmdCtx, imei)
	if err != nil || d == nil {
		return nil, err
	}
	return &gateway.HelmetDeviceDetail{HelmetReact: d.HelmetReact}, nil
}

func (a *DeviceHelmetAdapter) HelmetCommand(ctx context.Context, cmdCtx *dto.CommandContext, imei, carId string, sw int) (*gateway.HelmetCommandResult, error) {
	r, err := a.DeviceRPC.HelmetCommand(ctx, cmdCtx, imei, carId, sw)
	if err != nil || r == nil {
		return nil, err
	}
	return &gateway.HelmetCommandResult{EcuCode: r.EcuCode}, nil
}

var _ gateway.DeviceHelmetClient = (*DeviceHelmetAdapter)(nil)
