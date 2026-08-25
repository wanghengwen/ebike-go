package gateway

import (
	"context"
	"fmt"
	"strconv"

	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

// FenceRfidGateway resolves RFID card -> fence id (Java FenceRfidGatewayImpl.queryFenceIdCache).
type FenceRfidGateway struct{}

func NewFenceRfidGateway() *FenceRfidGateway {
	return &FenceRfidGateway{}
}

func (g *FenceRfidGateway) GetFenceIdByRfid(ctx context.Context, tenantId, rfidCode string) (int64, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil || rfidCode == "" {
		return 0, nil
	}
	key := fmt.Sprintf("rfid_fence_%s_%s", tenantId, rfidCode)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}
