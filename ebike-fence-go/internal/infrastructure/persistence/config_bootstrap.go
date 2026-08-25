package persistence

import (
	"context"

	domainconfig "ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

// InitConfigGatewayHelpers wires DB fallbacks for config reads outside config API handlers.
func InitConfigGatewayHelpers() {
	if Config == nil {
		return
	}
	repo := Config
	gateway.SetUseCarConfigLoader(func(ctx context.Context, tenantID string, serviceID int64) (*domainconfig.ConfigUseCarCO, error) {
		var row model.ConfigUseCar
		if err := repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return nil, err
		}
		sid := serviceID
		if row.ID == 0 {
			return &domainconfig.ConfigUseCarCO{ServiceId: &sid}, nil
		}
		co := &domainconfig.ConfigUseCarCO{ServiceId: &sid}
		if row.NearLine.Valid {
			n := int(row.NearLine.Int32)
			co.NearLine = &n
		}
		if row.IzBeacon.Valid {
			b := row.IzBeacon.Bool
			co.IzBeacon = &b
		}
		return co, nil
	})
}
