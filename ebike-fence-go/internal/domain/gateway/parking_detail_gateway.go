package gateway

import (
	"context"
	"encoding/json"

	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"
	"ebike-fence-go/internal/pkg/shadow"

	"github.com/go-redis/redis/v8"
)

// ParkingDetailE mirrors Java ParkingDetailDO fields used by bind/unbind flows.
type ParkingDetailE struct {
	Id             int64  `json:"id"`
	Imei           string `json:"imei"`
	ServiceId      int64  `json:"serviceId"`
	CarId          string `json:"carId"`
	ParkingId      int64  `json:"parkingId"`
	NoParkingId    int64  `json:"noParkingId"`
	BanRidingId    int64  `json:"banRidingId"`
	MaintainAreaId int64  `json:"maintainAreaId"`
	NearParkingId  int64  `json:"nearParkingId"`
	FenceCustomId  int64  `json:"fenceCustomId"`
}

// ParkingDetailGateway reads vehicle parking bindings from Redis (Java PARKING_DETAIL key).
type ParkingDetailGateway struct{}

func NewParkingDetailGateway() *ParkingDetailGateway {
	return &ParkingDetailGateway{}
}

func (g *ParkingDetailGateway) GetByCarIdFromRedis(ctx context.Context, tenantId, carId string) (*ParkingDetailE, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil || carId == "" {
		return nil, nil
	}
	key := rediskeys.ParkingDetail(tenantId, carId)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || val == "" || val == "null" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var detail ParkingDetailE
	if err := json.Unmarshal([]byte(val), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// SetDetailCache writes parking detail JSON to Redis (Java saveOrUpdateV1 side effect).
func SetDetailCache(ctx context.Context, tenantId string, detail *ParkingDetailE) error {
	if shadow.IsShadowTest(ctx) {
		return nil
	}
	if detail == nil || detail.CarId == "" {
		return nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	raw, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	key := rediskeys.ParkingDetail(tenantId, detail.CarId)
	return rdb.Set(ctx, key, raw, 0).Err()
}
