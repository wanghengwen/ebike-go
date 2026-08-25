package configsvc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

const defaultHideCarConfigJSON = `{"hideCarSwitch":0,"distance":0}`

func applyUseCarQueryDefaults(co *dto.ConfigUseCarCO, fromRedis bool) {
	if co == nil {
		return
	}
	if co.HideCarConfig == nil {
		s := defaultHideCarConfigJSON
		co.HideCarConfig = &s
	}
	if !fromRedis {
		return
	}
	if co.OutServiceAreaAutoLock == nil {
		v := 10
		co.OutServiceAreaAutoLock = &v
	}
	if co.IzAuth == nil {
		t := true
		co.IzAuth = &t
	}
}

func ensureUseCarCONearLine(co *dto.ConfigUseCarCO) {
	if co == nil || co.NearLine != nil {
		return
	}
	v := 200
	co.NearLine = &v
}

func applyUseCarStopServiceRecovery(co *dto.ConfigUseCarCO) {
	if co == nil || co.IzStopService == nil || !*co.IzStopService {
		return
	}
	if co.IzAutoRecovery == nil || !*co.IzAutoRecovery {
		return
	}
	if co.RecoveryData == nil || *co.RecoveryData == "" {
		return
	}
	t, err := time.Parse("2006-01-02T15:04:05", *co.RecoveryData)
	if err != nil {
		return
	}
	if t.Before(time.Now()) {
		f := false
		co.IzStopService = &f
	}
}

func finalizeUseCarCO(tenantID string, co *dto.ConfigUseCarCO, fromRedis bool) *dto.ConfigUseCarCO {
	if co == nil {
		return nil
	}
	ensureUseCarCONearLine(co)
	applyUseCarStopServiceRecovery(co)
	normalizeUseCarTimeFields(co)
	applyUseCarQueryDefaults(co, fromRedis)
	applyUseCarNeedAuth(tenantID, co)
	return co
}

func (s *UseCarService) loadUseCarCO(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigUseCarCO, bool, error) {
	key := rediskeys.UseCarConfig(tenantID, serviceID)
	rdb := pkgredis.GetClient()
	if rdb != nil {
		val, err := rdb.Get(ctx, key).Result()
		if err == nil && !rediskeys.IsCacheMiss(val) {
			co, err := unmarshalUseCarCO(val)
			if err != nil {
				return nil, true, err
			}
			return co, true, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, false, err
		}
	}
	raw, err := s.marshalUseCarFromDB(ctx, tenantID, serviceID)
	if err != nil {
		return nil, false, err
	}
	if raw == "" {
		return nil, false, nil
	}
	if rdb != nil {
		_ = rdb.Set(ctx, key, raw, 0).Err()
	}
	co, err := unmarshalUseCarCO(raw)
	if err != nil {
		return nil, false, err
	}
	return co, false, nil
}

func (s *UseCarService) marshalUseCarFromDB(ctx context.Context, tenantID string, serviceID int64) (string, error) {
	var row model.ConfigUseCar
	if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
		return "", err
	}
	if row.ID == 0 {
		return "", nil
	}
	return convert.MarshalRow(convert.UseCarToCO(&row))
}

func unmarshalUseCarCO(raw string) (*dto.ConfigUseCarCO, error) {
	var co dto.ConfigUseCarCO
	if err := json.Unmarshal([]byte(raw), &co); err != nil {
		return nil, err
	}
	return &co, nil
}
