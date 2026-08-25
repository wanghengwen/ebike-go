package gateway

import (
	"context"
	"log"
	"strconv"
	"time"

	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"
	"ebike-fence-go/internal/pkg/shadow"

	"github.com/go-redis/redis/v8"
)

// PartGateway reads part audit flags from Redis, matching Java PartGatewayImpl.
type PartGateway struct{}

func NewPartGateway() *PartGateway {
	return &PartGateway{}
}

func (g *PartGateway) getAudit(ctx context.Context, key string) (bool, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return false, nil
	}
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return val != "", nil
	}
	return parsed, nil
}

func (g *PartGateway) GetHelmetAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.HelmetAudit(tenantId, carId, *orderId))
}

func (g *PartGateway) GetPointAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.PointAudit(tenantId, carId, *orderId))
}

func (g *PartGateway) GetCameraAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.CameraAudit(tenantId, carId, *orderId))
}

func (g *PartGateway) GetKickstandAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.KickstandAudit(tenantId, carId, *orderId))
}

func (g *PartGateway) GetLocationAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.LocationAudit(tenantId, carId, *orderId))
}

// GetDirectionAudit mirrors Java PartGatewayImpl.getDirectionAudit.
func (g *PartGateway) GetDirectionAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	if orderId == nil {
		return false, nil
	}
	return g.getAudit(ctx, rediskeys.DirectionAudit(tenantId, carId, *orderId))
}

// GetTBeaconAudit mirrors Java PartGatewayImpl.getTBeaconAudit (same key as point audit).
func (g *PartGateway) GetTBeaconAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	return g.GetPointAudit(ctx, tenantId, carId, orderId)
}

// IncCameraErrorCount mirrors Java PartGatewayImpl.incCameraErrorCount.
func (g *PartGateway) IncCameraErrorCount(ctx context.Context, tenantId, carId string) (int64, error) {
	if shadow.IsShadowTest(ctx) {
		log.Printf("[SHADOW TEST] Skipped IncCameraErrorCount: tenantId=%s, carId=%s", tenantId, carId)
		return 1, nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return 0, nil
	}
	key := rediskeys.CameraErrorCount(tenantId, carId)
	return rdb.Incr(ctx, key).Result()
}

// DelCameraErrorCount mirrors Java PartGatewayImpl.delCameraErrorCount.
func (g *PartGateway) DelCameraErrorCount(ctx context.Context, tenantId, carId string) error {
	if shadow.IsShadowTest(ctx) {
		log.Printf("[SHADOW TEST] Skipped DelCameraErrorCount: tenantId=%s, carId=%s", tenantId, carId)
		return nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, rediskeys.CameraErrorCount(tenantId, carId)).Err()
}

// GetCameraFailCount mirrors Java PartGatewayImpl.getCameraFailCount.
func (g *PartGateway) GetCameraFailCount(ctx context.Context, tenantId, carId string, orderId *int64) (int64, error) {
	if orderId == nil {
		return 0, nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return 0, nil
	}
	val, err := rdb.Get(ctx, rediskeys.CameraFailCount(tenantId, carId, *orderId)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, nil
	}
	return n, nil
}

func (g *PartGateway) setAudit(ctx context.Context, key string, ttl time.Duration) error {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	return rdb.Set(ctx, key, "true", ttl).Err()
}

func (g *PartGateway) SetHelmetAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.HelmetAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetPointAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.PointAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetTBeaconAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.PointAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetDirectionAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.DirectionAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetKickstandAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.KickstandAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetCameraAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.CameraAudit(tenantId, carId, orderId), 24*time.Hour)
}

func (g *PartGateway) SetLocationAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return g.setAudit(ctx, rediskeys.LocationAudit(tenantId, carId, orderId), 7*24*time.Hour)
}

// IncCameraFailCount mirrors Java PartGatewayImpl.incCameraFailCount.
func (g *PartGateway) IncCameraFailCount(ctx context.Context, tenantId, carId string, orderId *int64) (int64, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil || orderId == nil {
		return 0, nil
	}
	key := rediskeys.CameraFailCount(tenantId, carId, *orderId)
	n, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = rdb.Expire(ctx, key, time.Hour).Err()
	}
	return n, nil
}
