package redisgw

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ebike-analyze-go/internal/common/protocol"
	"ebike-analyze-go/internal/common/rediskeys"
	pkgredis "ebike-analyze-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

type ServiceTenant struct {
	ServiceID int64
	TenantID  string
}

// TenantServiceImei pairs tenant, service and imei for cross-tenant rack queries.
type TenantServiceImei struct {
	TenantID  string
	ServiceID int64
	Imei      string
}

func GetImeiListByCarIDs(tenantID string, carIDs []string) []string {
	client := pkgredis.GetClient()
	if client == nil || len(carIDs) == 0 {
		return nil
	}
	keys := make([]string, len(carIDs))
	for i, carID := range carIDs {
		keys[i] = rediskeys.CarImeiBind(tenantID, carID)
	}
	vals, err := client.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil
	}
	var out []string
	for _, v := range vals {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

// GetImeiListFromService mirrors Java DeviceInfoQueryImpl.getImeiList.
// Java uses `imeiList.addAll(imeiSet)` across services/hash-buckets WITHOUT
// deduplication, so the same imei can legitimately appear multiple times if
// present in more than one bucket/service - Go intentionally does not dedup
// either, to keep downstream counts (e.g. carStatistics grouping) identical.
func GetImeiListFromService(tenantID string, serviceIDs []int64, reportTime *int64) []string {
	client := pkgredis.GetClient()
	if client == nil || len(serviceIDs) == 0 {
		return nil
	}
	min := int64(0)
	if reportTime != nil {
		min = *reportTime
	}
	ctx := context.Background()
	minStr := strconv.FormatInt(min, 10)
	var out []string
	for _, sid := range serviceIDs {
		for h := 0; h < rediskeys.DeviceHashCount; h++ {
			key := rediskeys.ServiceGfenceCarZSet(tenantID, sid, h)
			vals, err := client.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: minStr, Max: "+inf"}).Result()
			if err != nil {
				continue
			}
			out = append(out, vals...)
		}
	}
	return out
}

// RetainImeiList mirrors Java List.retainAll.
func RetainImeiList(imeis, filter []string) []string {
	if len(filter) == 0 {
		return nil
	}
	set := map[string]struct{}{}
	for _, v := range filter {
		set[v] = struct{}{}
	}
	var out []string
	for _, v := range imeis {
		if _, ok := set[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

func GetDeviceInfoList(tenantID string, imeis []string) []*protocol.DeviceInfo {
	return mgetDevices(tenantID, imeis)
}

func mgetDevices(tenantID string, imeis []string) []*protocol.DeviceInfo {
	client := pkgredis.GetClient()
	if client == nil || len(imeis) == 0 {
		return nil
	}
	keys := make([]string, len(imeis))
	for i, imei := range imeis {
		keys[i] = rediskeys.DeviceInfo(tenantID, imei)
	}
	vals, _ := client.MGet(context.Background(), keys...).Result()
	var out []*protocol.DeviceInfo
	for _, v := range vals {
		if s, ok := v.(string); ok {
			d := protocol.DecodeDevice(s)
			if d != nil && strings.TrimSpace(d.CarID) != "" {
				out = append(out, d)
			}
		}
	}
	return out
}

// GetRackDevicesByTenantServiceImei mirrors Java getRackDeviceByTenantServiceImei.
func GetRackDevicesByTenantServiceImei(items []TenantServiceImei) []*protocol.DeviceInfo {
	client := pkgredis.GetClient()
	if client == nil || len(items) == 0 {
		return nil
	}
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = rediskeys.DeviceInfo(item.TenantID, item.Imei)
	}
	vals, _ := client.MGet(context.Background(), keys...).Result()
	var out []*protocol.DeviceInfo
	for _, v := range vals {
		if s, ok := v.(string); ok {
			d := protocol.DecodeDevice(s)
			if d != nil && strings.TrimSpace(d.CarID) != "" {
				out = append(out, d)
			}
		}
	}
	return out
}

func GetMoveAlarmTags(tenantID string, imeis []string) []interface{} {
	client := pkgredis.GetClient()
	if client == nil || len(imeis) == 0 {
		return nil
	}
	keys := make([]string, len(imeis))
	for i, imei := range imeis {
		keys[i] = rediskeys.MoveAlarmTag(tenantID, imei)
	}
	vals, _ := client.MGet(context.Background(), keys...).Result()
	out := make([]interface{}, len(imeis))
	for i, v := range vals {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
				out[i] = n
				continue
			}
		}
		out[i] = nil
	}
	return out
}

// FlattenTenantServiceImeis mirrors Java getTenantServiceImei. Java's
// `tenantServiceImeiDtos.addAll(collect)` across the 10 hash buckets does NOT
// dedup, so a car whose imei is (incorrectly) present in more than one bucket
// gets counted once per occurrence downstream in carStatistics - Go preserves
// that behavior rather than silently fixing it.
func FlattenTenantServiceImeis(pairs []ServiceTenant) []TenantServiceImei {
	client := pkgredis.GetClient()
	if client == nil {
		return nil
	}
	ctx := context.Background()
	var out []TenantServiceImei
	for _, p := range pairs {
		for h := 0; h < rediskeys.DeviceHashCount; h++ {
			zkey := rediskeys.ServiceGfenceCarZSet(p.TenantID, p.ServiceID, h)
			vals, err := client.ZRange(ctx, zkey, 0, -1).Result()
			if err != nil {
				continue
			}
			for _, imei := range vals {
				out = append(out, TenantServiceImei{TenantID: p.TenantID, ServiceID: p.ServiceID, Imei: imei})
			}
		}
	}
	return out
}

func HGetAllHotPoint(keys []string) map[string]int {
	client := pkgredis.GetClient()
	if client == nil || len(keys) == 0 {
		return nil
	}
	ctx := context.Background()
	pipe := client.Pipeline()
	cmds := make([]*redis.StringStringMapCmd, len(keys))
	for i, k := range keys {
		cmds[i] = pipe.HGetAll(ctx, k)
	}
	_, _ = pipe.Exec(ctx)
	merged := map[string]int{}
	for _, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil || len(m) == 0 {
			continue
		}
		for field, countStr := range m {
			n, _ := strconv.Atoi(countStr)
			merged[field] += n
		}
	}
	return merged
}

func buildHotPointKeys(prefix string, tenantID string, serviceID int64, start, end time.Time) []string {
	fullPrefix := fmt.Sprintf("%s%s:%d:", prefix, tenantID, serviceID)
	return buildHotPointKeysFromPrefix(fullPrefix, start, end)
}

func buildHotPointKeysFromPrefix(prefix string, start, end time.Time) []string {
	var keys []string
	for d := start; d.Before(end); d = d.Add(24 * time.Hour) {
		keys = append(keys, prefix+d.Format("20060102"))
	}
	return keys
}
