package rpc

import (
	"context"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence"

	"go.uber.org/zap"
)

// QueryFenceByID loads fence detail from MySQL.
func QueryFenceByID(ctx context.Context, tenantId string, id int64) (*gateway.FenceE, error) {
	_ = tenantId
	if persistence.Fence == nil {
		return nil, nil
	}
	return persistence.Fence.GetByID(ctx, id)
}

// QueryFenceByIDs batch-fetches fences from MySQL.
func QueryFenceByIDs(ctx context.Context, tenantId string, ids []int64) (map[int64]gateway.FenceE, error) {
	_ = tenantId
	if persistence.Fence == nil {
		return nil, nil
	}
	return persistence.Fence.GetByIDs(ctx, ids)
}

// RegisterFenceMissFetcher wires DB lookup into the gateway cache-miss path.
func RegisterFenceMissFetcher() {
	gateway.SetFenceMissFetcher(func(ctx context.Context, tenantId string, id int64) (*gateway.FenceE, error) {
		fe, err := QueryFenceByID(ctx, tenantId, id)
		if err != nil {
			zap.L().Warn("fence cache miss fallback failed",
				zap.String("tenantId", tenantId),
				zap.Int64("fenceId", id),
				zap.Error(err),
			)
		}
		return fe, err
	})
}
