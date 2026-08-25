package gateway

import (
	"context"
	"map-service-go/internal/model"
)

type MapApiGateway interface {
	GetNavigate(ctx context.Context, origin, destination string, navType int, tenantId string) (*model.Navigate, error)
	GetLocationByAddress(ctx context.Context, address, city, tenantId string) (*model.Location, error)
	GetAddressByLocation(ctx context.Context, longitude, latitude, tenantId string) (*model.AddressCo, error)
	GetBatchAddressByLocation(ctx context.Context, locations []model.Location, tenantId string) ([]model.AddressCo, error)
}
