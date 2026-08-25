package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/cache"
	"ebike-device-worker-go/internal/pkg/redis"
	"ebike-device-worker-go/internal/pkg/web"

	goredis "github.com/redis/go-redis/v9"
)

const (
	orderTrajectoryCachePrefix  = "order_traj:"
	orderTrajectoryMissSentinel = "__miss__"

	orderTrajectoryNegativeTTL = 5 * time.Second
	orderTrajectoryPositiveTTL = 15 * time.Minute
)

type orderTrajectoryCacheState int

const (
	orderTrajectoryCacheMiss orderTrajectoryCacheState = iota
	orderTrajectoryCacheHitNegative
	orderTrajectoryCacheHitPositive
)

var (
	orderTrajNegL1 = cache.NewTTLCache(10000, int(orderTrajectoryNegativeTTL.Seconds()))
	orderTrajPosL1 = cache.NewTTLCache(10000, int(orderTrajectoryPositiveTTL.Seconds()))

	orderTrajGen   sync.Map // orderId -> uint64
	orderTrajLoads sync.Map // cacheKey -> *orderTrajInflight
)

type orderTrajInflight struct {
	done chan struct{}
	data []dto.TrajectoryCo
	err  error
}

func orderTrajectoryCacheGen(orderID string) uint64 {
	v, _ := orderTrajGen.LoadOrStore(orderID, uint64(0))
	return v.(uint64)
}

func bumpOrderTrajectoryCacheGen(orderID string) {
	v, _ := orderTrajGen.LoadOrStore(orderID, uint64(0))
	orderTrajGen.Store(orderID, v.(uint64)+1)
}

func orderTrajectoryCacheKey(orderID string, coordType int, imei string) string {
	if imei == "" {
		imei = "_"
	}
	gen := orderTrajectoryCacheGen(orderID)
	return redis.GetKey(fmt.Sprintf("%s%s:g%d:t%d:i%s", orderTrajectoryCachePrefix, orderID, gen, coordType, imei))
}

func orderTrajectoryNotExistErr() error {
	return &web.BizError{Code: dto.CodeOrderTrajectoryNotExist, Msg: "轨迹不存在"}
}

func fetchOrderTrajectory(orderID string, coordType int, imei string) ([]dto.TrajectoryCo, error) {
	key := orderTrajectoryCacheKey(orderID, coordType, imei)

	if _, ok := orderTrajNegL1.Get(key, nil); ok {
		log.Printf("[itinerary-cache] L1 negative hit key=%s", key)
		return nil, orderTrajectoryNotExistErr()
	}
	if v, ok := orderTrajPosL1.Get(key, nil); ok {
		if traj, ok := v.([]dto.TrajectoryCo); ok {
			log.Printf("[itinerary-cache] L1 positive hit key=%s points=%d", key, len(traj))
			return traj, nil
		}
	}

	if traj, state := getOrderTrajectoryFromRedis(key); state == orderTrajectoryCacheHitNegative {
		orderTrajNegL1.Put(key, true)
		log.Printf("[itinerary-cache] redis negative hit key=%s", key)
		return nil, orderTrajectoryNotExistErr()
	} else if state == orderTrajectoryCacheHitPositive {
		orderTrajPosL1.Put(key, traj)
		log.Printf("[itinerary-cache] redis positive hit key=%s points=%d", key, len(traj))
		return traj, nil
	}

	inflight := &orderTrajInflight{done: make(chan struct{})}
	actual, loaded := orderTrajLoads.LoadOrStore(key, inflight)
	if loaded {
		wait := actual.(*orderTrajInflight)
		<-wait.done
		if wait.err != nil {
			return nil, wait.err
		}
		log.Printf("[itinerary-cache] coalesced hit key=%s points=%d", key, len(wait.data))
		return wait.data, nil
	}
	defer orderTrajLoads.Delete(key)

	traj, err := loadOrderTrajectoryFromDB(orderID, coordType, imei, key)
	inflight.data = traj
	inflight.err = err
	close(inflight.done)
	return traj, err
}

func loadOrderTrajectoryFromDB(orderID string, coordType int, imei, key string) ([]dto.TrajectoryCo, error) {
	itinerary, err := queryItinerary(orderID, imei)
	if err != nil {
		if biz, ok := err.(*web.BizError); ok && biz.Code == dto.CodeOrderTrajectoryNotExist {
			storeOrderTrajectoryNegative(key)
			log.Printf("[itinerary-cache] db miss key=%s", key)
		}
		return nil, err
	}
	trajectories, err := parseItineraryTrajectory(itinerary, coordType)
	if err != nil {
		return nil, err
	}
	storeOrderTrajectoryPositive(key, trajectories)
	log.Printf("[itinerary-cache] db hit key=%s points=%d", key, len(trajectories))
	return trajectories, nil
}

func getOrderTrajectoryFromRedis(key string) ([]dto.TrajectoryCo, orderTrajectoryCacheState) {
	if redis.Client == nil {
		return nil, orderTrajectoryCacheMiss
	}
	raw, err := redis.Client.Get(redis.Ctx(), key).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, orderTrajectoryCacheMiss
	}
	if err != nil {
		log.Printf("[itinerary-cache] redis get error key=%s: %v", key, err)
		return nil, orderTrajectoryCacheMiss
	}
	if raw == orderTrajectoryMissSentinel {
		return nil, orderTrajectoryCacheHitNegative
	}
	var trajectories []dto.TrajectoryCo
	if err := json.Unmarshal([]byte(raw), &trajectories); err != nil {
		log.Printf("[itinerary-cache] redis unmarshal error key=%s: %v", key, err)
		return nil, orderTrajectoryCacheMiss
	}
	return trajectories, orderTrajectoryCacheHitPositive
}

func storeOrderTrajectoryNegative(key string) {
	orderTrajNegL1.Put(key, true)
	if redis.Client == nil {
		return
	}
	if err := redis.Client.Set(redis.Ctx(), key, orderTrajectoryMissSentinel, orderTrajectoryNegativeTTL).Err(); err != nil {
		log.Printf("[itinerary-cache] redis set negative error key=%s: %v", key, err)
	}
}

func storeOrderTrajectoryPositive(key string, trajectories []dto.TrajectoryCo) {
	orderTrajPosL1.Put(key, trajectories)
	if redis.Client == nil {
		return
	}
	payload, err := json.Marshal(trajectories)
	if err != nil {
		return
	}
	if err := redis.Client.Set(redis.Ctx(), key, payload, orderTrajectoryPositiveTTL).Err(); err != nil {
		log.Printf("[itinerary-cache] redis set positive error key=%s: %v", key, err)
	}
}

func invalidateOrderTrajectoryCache(orderID string) {
	if orderID == "" {
		return
	}
	bumpOrderTrajectoryCacheGen(orderID)
	if redis.Client == nil {
		return
	}
	pattern := redis.GetKey(orderTrajectoryCachePrefix + orderID + ":*")
	var cursor uint64
	for {
		keys, next, err := redis.Client.Scan(redis.Ctx(), cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("[itinerary-cache] redis invalidate scan error order=%s: %v", orderID, err)
			return
		}
		if len(keys) > 0 {
			if err := redis.Client.Del(redis.Ctx(), keys...).Err(); err != nil {
				log.Printf("[itinerary-cache] redis invalidate del error order=%s: %v", orderID, err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
}

func itineraryCacheBumpGen(orderID string) {
	bumpOrderTrajectoryCacheGen(orderID)
}
