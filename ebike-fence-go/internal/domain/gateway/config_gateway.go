package gateway

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type ConfigGateway struct {
	localCache *expirable.LRU[string, *config.ConfigBackcarCO]
}

var (
	configGatewayOnce sync.Once
	configGatewayInst *ConfigGateway
)

// NewConfigGateway returns a process-wide singleton. Each expirable.LRU spawns a
// background goroutine; callers must not create a new gateway per request.
func NewConfigGateway() *ConfigGateway {
	configGatewayOnce.Do(func() {
		cache := expirable.NewLRU[string, *config.ConfigBackcarCO](1000, nil, time.Second*15)
		configGatewayInst = &ConfigGateway{localCache: cache}
	})
	return configGatewayInst
}

// InvalidateConfigByServiceId evicts the locally cached backcar config for a
// service so admin writes (BackcarService.Create/Update) are visible to
// GetConfigByServiceId immediately instead of waiting out the TTL. Needed
// because the process-wide singleton cache (see NewConfigGateway) actually
// holds entries now; before that fix each call got a throwaway cache, so
// staleness was never observable.
func (g *ConfigGateway) InvalidateConfigByServiceId(tenantId string, serviceId int64) {
	g.localCache.Remove(rediskeys.BackCarConfig(tenantId, serviceId))
}

func (g *ConfigGateway) GetConfigByServiceId(ctx context.Context, tenantId string, serviceId int64) (*config.ConfigBackcarCO, error) {
	key := rediskeys.BackCarConfig(tenantId, serviceId)
	if conf, ok := g.localCache.Get(key); ok {
		return conf, nil
	}

	rdb := pkgredis.GetClient()
	var conf *config.ConfigBackcarCO
	if rdb == nil {
		defaultBuffer := 10.0
		conf = &config.ConfigBackcarCO{ServiceId: &serviceId, BufferDistance: &defaultBuffer}
		g.localCache.Add(key, conf)
		return conf, nil
	}

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || val == "null" || val == "" {
		defaultBuffer := 10.0
		conf = &config.ConfigBackcarCO{ServiceId: &serviceId, BufferDistance: &defaultBuffer}
	} else if err != nil {
		return nil, err
	} else {
		conf = &config.ConfigBackcarCO{}
		if err := json.Unmarshal([]byte(val), conf); err != nil {
			return nil, err
		}
	}
	g.localCache.Add(key, conf)
	return conf, nil
}

// UseCarConfigLoader loads use-car config when Redis cache is incomplete.
type UseCarConfigLoader func(ctx context.Context, tenantID string, serviceID int64) (*config.ConfigUseCarCO, error)

var useCarConfigLoader UseCarConfigLoader

// SetUseCarConfigLoader registers a DB fallback for GetUseCarConfig.
func SetUseCarConfigLoader(fn UseCarConfigLoader) {
	useCarConfigLoader = fn
}

func (g *ConfigGateway) GetUseCarConfig(ctx context.Context, tenantId string, serviceId int64) (*config.ConfigUseCarCO, error) {
	conf := g.readUseCarFromRedis(ctx, rediskeys.UseCarConfig(tenantId, serviceId))
	if conf != nil && conf.NearLine != nil {
		return conf, nil
	}
	if useCarConfigLoader != nil {
		if loaded, err := useCarConfigLoader(ctx, tenantId, serviceId); err == nil && loaded != nil {
			if conf == nil {
				return loaded, nil
			}
			if loaded.NearLine != nil {
				conf.NearLine = loaded.NearLine
			}
			if conf.ServiceId == nil && loaded.ServiceId != nil {
				conf.ServiceId = loaded.ServiceId
			}
			if conf.IzBeacon == nil && loaded.IzBeacon != nil {
				conf.IzBeacon = loaded.IzBeacon
			}
			return conf, nil
		}
	}
	if conf != nil {
		return conf, nil
	}
	sid := serviceId
	return &config.ConfigUseCarCO{ServiceId: &sid}, nil
}

func (g *ConfigGateway) readUseCarFromRedis(ctx context.Context, key string) *config.ConfigUseCarCO {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	val, err := rdb.Get(ctx, key).Result()
	if err != nil || rediskeys.IsCacheMiss(val) {
		return nil
	}
	var fields useCarRedisFields
	if json.Unmarshal([]byte(val), &fields) != nil {
		return nil
	}
	if fields.ServiceId == nil && fields.NearLine == nil && fields.IzBeacon == nil {
		return nil
	}
	conf := &config.ConfigUseCarCO{
		ServiceId: fields.ServiceId,
		NearLine:  fields.NearLine,
		IzBeacon:  fields.IzBeacon,
	}
	if len(fields.HelmetConfig) > 0 && fields.HelmetConfig[0] == '{' {
		var helmet config.HelmetConfig
		if json.Unmarshal(fields.HelmetConfig, &helmet) == nil {
			conf.HelmetConfig = &helmet
		}
	}
	return conf
}

type useCarRedisFields struct {
	ServiceId    *int64          `json:"serviceId"`
	NearLine     *int            `json:"nearLine"`
	IzBeacon     *bool           `json:"izBeacon"`
	HelmetConfig json.RawMessage `json:"helmetConfig"`
}

func (g *ConfigGateway) GetPayConfigRaw(ctx context.Context, tenantId string, serviceId int64) ([]byte, error) {
	return g.getRaw(ctx, rediskeys.PayConfig(tenantId, serviceId))
}

func (g *ConfigGateway) GetBaseItemConfigRaw(ctx context.Context, tenantId string, serviceId int64) ([]byte, error) {
	return g.getRaw(ctx, rediskeys.BaseItemConfig(tenantId, serviceId))
}

func (g *ConfigGateway) GetRidingPermissionRaw(ctx context.Context, tenantId string, serviceId int64) ([]byte, error) {
	return g.getRaw(ctx, rediskeys.RidingPermission(tenantId, serviceId))
}

func (g *ConfigGateway) getRaw(ctx context.Context, key string) ([]byte, error) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil, nil
	}
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || rediskeys.IsCacheMiss(val) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (g *ConfigGateway) GetBaseItemConfig(ctx context.Context, tenantId string, serviceId int64) (*config.ConfigBaseItemCO, error) {
	key := rediskeys.BaseItemConfig(tenantId, serviceId)
	rdb := pkgredis.GetClient()
	sid := serviceId
	if rdb == nil {
		return &config.ConfigBaseItemCO{ServiceId: &sid}, nil
	}
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || val == "null" || val == "" {
		return &config.ConfigBaseItemCO{ServiceId: &sid}, nil
	}
	if err != nil {
		return nil, err
	}
	conf := &config.ConfigBaseItemCO{}
	if err := json.Unmarshal([]byte(val), conf); err != nil {
		return nil, err
	}
	return conf, nil
}
