package persistence

import (
	"context"
	"fmt"

	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"

	"go.uber.org/zap"
)

type bizError struct {
	code string
	msg  string
}

func (e *bizError) Error() string { return e.msg }
func (e *bizError) Code() string  { return e.code }

// QueryFenceByID loads fence detail from MySQL only (native_only; no Java ebike-fence fallback).
func QueryFenceByID(ctx context.Context, tenantId string, id int64) (*gateway.FenceE, error) {
	if Fence == nil {
		return nil, fmt.Errorf("fence repository not configured")
	}
	fe, err := Fence.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if fe == nil {
		return nil, &bizError{code: "13002", msg: "Data query failed, no data found"}
	}
	return fe, nil
}

// QueryFenceByIDs batch-fetches fences from MySQL only.
func QueryFenceByIDs(ctx context.Context, tenantId string, ids []int64) (map[int64]gateway.FenceE, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if Fence == nil {
		return nil, fmt.Errorf("fence repository not configured")
	}
	return Fence.GetByIDs(ctx, ids)
}

func fenceTypeFromPrefix(prefix string) (int, bool) {
	switch prefix {
	case "fence_serviceArea":
		return fence.TypeServiceArea, true
	case "fence_parking":
		return fence.TypeParking, true
	case "fence_noParking":
		return fence.TypeNoParking, true
	case "fence_banRiding":
		return fence.TypeBanRiding, true
	case "fence_maintainArea":
		return fence.TypeMaintainArea, true
	case "fence_custom":
		return fence.TypeCustom, true
	default:
		return 0, false
	}
}

// RegisterFenceMissFetcher wires DB-only fallback into the gateway layer.
func RegisterFenceMissFetcher() {
	gateway.SetFenceMissFetcher(func(ctx context.Context, tenantId string, id int64) (*gateway.FenceE, error) {
		fe, err := QueryFenceByID(ctx, tenantId, id)
		if err != nil {
			zap.L().Warn("fence cache miss DB lookup failed",
				zap.String("tenantId", tenantId),
				zap.Int64("fenceId", id),
				zap.Error(err),
			)
		}
		return fe, err
	})

	gateway.SetFenceGeoListFetcher(func(ctx context.Context, tenantId, prefix string, serviceID int64, lng, lat, radius float64, count int) ([]gateway.FenceE, error) {
		if Fence == nil {
			return nil, nil
		}
		fenceType, ok := fenceTypeFromPrefix(prefix)
		if !ok {
			return nil, nil
		}
		var list []gateway.FenceE
		var err error
		if serviceID > 0 {
			list, err = Fence.ListByTypeAndServiceID(ctx, tenantId, fenceType, serviceID)
		} else {
			list, err = Fence.ListByType(ctx, tenantId, fenceType)
		}
		if err != nil {
			zap.L().Warn("fence GEO fallback DB lookup failed",
				zap.String("tenantId", tenantId),
				zap.String("prefix", prefix),
				zap.Error(err),
			)
			return nil, err
		}
		// Batch-write ALL fences to Redis GEO/detail; sets empty marker if list is empty.
		gateway.PopulateGeoCache(ctx, prefix, tenantId, list)
		// Return distance-filtered results (original contract preserved).
		return gateway.FencesNearPoint(list, lng, lat, radius, count), nil
	})
}
