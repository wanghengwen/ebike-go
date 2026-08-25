package service

import (
	"context"
	"encoding/json"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
)

// ParkingCOByCarIDResolver loads a full ParkingCO for a vehicle (Java getParkingByCarId).
type ParkingCOByCarIDResolver func(ctx context.Context, tenantID, carID string, detailGw *gateway.ParkingDetailGateway) (*dto.ParkingCO, error)

var parkingCOByCarID ParkingCOByCarIDResolver

// SetParkingCOByCarIDResolver wires fenceadmin.GetByCarID from controller init (avoids import cycle).
func SetParkingCOByCarIDResolver(fn ParkingCOByCarIDResolver) {
	parkingCOByCarID = fn
}

// ParkingCOToClientMap serializes ParkingCO like Java client responses (audit fields empty/null).
func ParkingCOToClientMap(co *dto.ParkingCO) map[string]interface{} {
	if co == nil {
		return nil
	}
	client := *co
	stripClientParkingCO(&client)
	raw, err := json.Marshal(client)
	if err != nil {
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func stripClientParkingCO(co *dto.ParkingCO) {
	co.Pics = ""
	co.TenantId = ""
	co.CreatedPin = ""
	co.CreatedAt = ""
	co.UpdatedPin = ""
	co.UpdatedAt = ""
	co.Distance = 0
	co.RefTags = nil
}
