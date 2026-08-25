package rpc

import (
	"context"
	"encoding/json"

	"ebike-fence-go/internal/api/dto"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
)

type idCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Id             int64               `json:"id"`
}

type carIdCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Id             string              `json:"id"`
}

type parkingCountResult struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Data    int64  `json:"data"`
}

type parkingByCarResult struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

// ParkingCarCount calls /parking/parkingCarCount on a legacy ebike-fence instance
// until the Go service implements DB-backed parking counts directly.
func ParkingCarCount(ctx context.Context, tenantId string, parkingId int64) (int64, error) {
	cmd := idCmd{
		CommandContext: &dto.CommandContext{TenantId: tenantId},
		Id:             parkingId,
	}
	var res parkingCountResult
	if err := pkg_rpc.PostToService(ctx, pkg_rpc.ServiceFence, "/parking/parkingCarCount", cmd, &res); err != nil {
		return 0, err
	}
	if !res.Success {
		return 0, nil
	}
	return res.Data, nil
}

// GetParkingByCarId calls /parking/getByCarId until DB read path is ported.
func GetParkingByCarId(ctx context.Context, tenantId, carId string) (map[string]interface{}, error) {
	cmd := carIdCmd{
		CommandContext: &dto.CommandContext{TenantId: tenantId},
		Id:             carId,
	}
	var res parkingByCarResult
	if err := pkg_rpc.PostToService(ctx, pkg_rpc.ServiceFence, "/parking/getByCarId", cmd, &res); err != nil {
		return nil, err
	}
	if !res.Success || len(res.Data) == 0 || string(res.Data) == "null" {
		return nil, nil
	}
	var co map[string]interface{}
	if err := json.Unmarshal(res.Data, &co); err != nil {
		return nil, err
	}
	return co, nil
}
