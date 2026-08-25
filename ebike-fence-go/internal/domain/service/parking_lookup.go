package service

import (
	"context"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/rpc"
)

// GetParkingByCarId resolves parking: full ParkingCO (native) → legacy RPC fallback.
func GetParkingByCarId(ctx context.Context, tenantId, carId string, detailGw *gateway.ParkingDetailGateway) (map[string]interface{}, error) {
	if parkingCOByCarID != nil {
		co, err := parkingCOByCarID(ctx, tenantId, carId, detailGw)
		if err != nil {
			return nil, err
		}
		if co != nil {
			return ParkingCOToClientMap(co), nil
		}
	}
	return rpc.GetParkingByCarId(ctx, tenantId, carId)
}
