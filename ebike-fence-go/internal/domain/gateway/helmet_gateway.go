package gateway

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

// HelmetCommandResult mirrors device RPC helmet command response.
type HelmetCommandResult struct {
	EcuCode string `json:"ecuCode"`
}

// HelmetDeviceDetail minimal device detail for helmet lock recording.
type HelmetDeviceDetail struct {
	HelmetReact *int `json:"helmetReact"`
}

// DeviceHelmetClient abstracts device RPC to avoid gateway->rpc import cycle.
type DeviceHelmetClient interface {
	GetDeviceDetail(ctx context.Context, cmdCtx *dto.CommandContext, imei string) (*HelmetDeviceDetail, error)
	HelmetCommand(ctx context.Context, cmdCtx *dto.CommandContext, imei, carId string, sw int) (*HelmetCommandResult, error)
}

type HelmetGateway struct {
	deviceRpc DeviceHelmetClient
}

func NewHelmetGateway(deviceRpc DeviceHelmetClient) *HelmetGateway {
	return &HelmetGateway{deviceRpc: deviceRpc}
}

// GetRecordHelmetLock directly checks Redis for helmet lock status
func (g *HelmetGateway) GetRecordHelmetLock(ctx context.Context, tenantId, carId string) (string, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return "", nil
	}
	key := rediskeys.HelmetLock(tenantId, carId)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (g *HelmetGateway) GetRecordHelmetReact(ctx context.Context, tenantId, carId string) (string, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return "", nil
	}
	key := rediskeys.HelmetReact(tenantId, carId)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (g *HelmetGateway) DelHelmet(ctx context.Context, tenantId, carId string) error {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	_ = rdb.Del(ctx, rediskeys.HelmetLock(tenantId, carId)).Err()
	return rdb.Del(ctx, rediskeys.HelmetReact(tenantId, carId)).Err()
}

func (g *HelmetGateway) Unlock(ctx context.Context, tenantId string, cmdCtx *dto.CommandContext, imei, carId string) (bool, error) {
	if g.deviceRpc == nil {
		return false, nil
	}
	cmdCtx = dto.EnsureCommandContext(cmdCtx, tenantId)
	detail, _ := g.deviceRpc.GetDeviceDetail(ctx, cmdCtx, imei)
	result, err := g.deviceRpc.HelmetCommand(ctx, cmdCtx, imei, carId, 0)
	if err != nil {
		return false, err
	}
	g.recordHelmetLock(ctx, tenantId, carId, result.EcuCode, detail)
	return true, nil
}

func (g *HelmetGateway) Lock(ctx context.Context, tenantId string, cmdCtx *dto.CommandContext, imei, carId string) (bool, error) {
	if g.deviceRpc == nil {
		return false, nil
	}
	cmdCtx = dto.EnsureCommandContext(cmdCtx, tenantId)
	_, err := g.deviceRpc.HelmetCommand(ctx, cmdCtx, imei, carId, 1)
	if err != nil {
		return false, err
	}
	_ = g.DelHelmet(ctx, tenantId, carId)
	return true, nil
}

func (g *HelmetGateway) recordHelmetLock(ctx context.Context, tenantId, carId, value string, detail *HelmetDeviceDetail) {
	if value == "" {
		return
	}
	if detail != nil && detail.HelmetReact != nil && *detail.HelmetReact == 0 {
		return
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	_ = rdb.Set(ctx, rediskeys.HelmetLock(tenantId, carId), value, 0).Err()
}
