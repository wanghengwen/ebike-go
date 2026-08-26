// Package repository reads device data from Redis and decodes the fixed-length
// device-info strings via the protocol codec, mirroring the Java DeviceInfoQuery.
package repository

import (
	"strconv"
	"time"

	"ebike-device-paas-go/internal/pkg/redis"
	"ebike-device-paas-go/internal/pkg/rediskey"
	"ebike-device-paas-go/internal/protocol"
)

// GetDeviceInfoList mirrors DeviceInfoQueryImpl.getDeviceInfoDOList: MGET the
// device_info keys for the given imeis, decode each non-empty string, and keep
// only entries with a non-blank carId.
func GetDeviceInfoList(tenantID string, imeiList []string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(imeiList))
	if len(imeiList) == 0 {
		return out
	}
	keys := make([]string, len(imeiList))
	for i, imei := range imeiList {
		keys[i] = rediskey.DeviceInfo(tenantID, imei)
	}
	for _, raw := range redis.MGet(keys...) {
		if raw == "" {
			continue
		}
		dev := protocol.Decode(raw)
		if carID, _ := dev["carId"].(string); carID != "" {
			out = append(out, dev)
		}
	}
	return out
}

// GetDeviceByImei mirrors DeviceInfoQueryImpl.getDeviceByImei: GET + decode.
// Returns nil when the key is missing.
func GetDeviceByImei(tenantID, imei string) map[string]interface{} {
	raw := redis.Get(rediskey.DeviceInfo(tenantID, imei))
	if raw == "" {
		return nil
	}
	return protocol.Decode(raw)
}

// GetImeiByCarId mirrors DeviceInfoQueryImpl.getImeiByCarId: GET car_imei_* key.
// Returns "" when the binding is missing.
func GetImeiByCarId(tenantID, carID string) string {
	return redis.Get(rediskey.CarImeiBind(tenantID, carID))
}

// deviceHashCount mirrors DeviceInfoQueryImpl.DEVICE_HASH_COUNT.
const deviceHashCount = 3

// GetImeiListByService mirrors DeviceInfoQueryImpl.getImeiList: for each service
// and hash bucket, ZRANGEBYSCORE the service_gfence zset (score >= reportTime).
func GetImeiListByService(tenantID string, serviceIDs []int64, reportTime int64) []string {
	var out []string
	for _, sid := range serviceIDs {
		for h := 0; h < deviceHashCount; h++ {
			key := rediskey.ServiceGfenceCarZset(tenantID, sid, h)
			out = append(out, redis.ZRangeByScoreMin(key, reportTime)...)
		}
	}
	return out
}

// GetImeiListByCarId mirrors DeviceInfoQueryImpl.getImeiListByCarId: MGET the
// car_imei bindings, dropping empties.
func GetImeiListByCarId(tenantID string, carIDs []string) []string {
	if len(carIDs) == 0 {
		return nil
	}
	keys := make([]string, len(carIDs))
	for i, c := range carIDs {
		keys[i] = rediskey.CarImeiBind(tenantID, c)
	}
	var out []string
	for _, v := range redis.MGet(keys...) {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// GetImeiByServiceId mirrors DeviceInfoQueryImpl.getImeiByServiceId: ZRANGE the
// full service_gfence zset across all hash buckets.
func GetImeiByServiceId(tenantID string, serviceID int64) []string {
	var out []string
	for h := 0; h < deviceHashCount; h++ {
		out = append(out, redis.ZRange(rediskey.ServiceGfenceCarZset(tenantID, serviceID, h), 0, -1)...)
	}
	return out
}

// GetDeviceByServiceId mirrors DeviceInfoQueryImpl.getDeviceByServiceId: the
// service-area imei set, then their device-info records.
func GetDeviceByServiceId(tenantID string, serviceID int64) []map[string]interface{} {
	return GetDeviceInfoList(tenantID, GetImeiByServiceId(tenantID, serviceID))
}

// GetBluetoothToken mirrors DeviceInfoQueryImpl.getBlueToothToken: GET the
// bluetooth_token key; nil when absent/blank.
func GetBluetoothToken(tenantID, imei string) *int {
	raw := redis.Get(rediskey.BluetoothToken(tenantID, imei))
	if raw == "" {
		return nil
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return &v
	}
	return nil
}

// GetSaddleOverloadContact mirrors DeviceInfoQueryImpl.getSaddleOverloadContact:
// GET saddle_overload_contact; default 0.
func GetSaddleOverloadContact(tenantID, imei string) int {
	raw := redis.Get(rediskey.SaddleOverloadContact(tenantID, imei))
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return 0
}

// GetCameraStateRaw mirrors DeviceInfoQueryImpl.queryCameraState: GET camera_info.
func GetCameraStateRaw(tenantID, imei string) string {
	return redis.Get(rediskey.CameraInfo(tenantID, imei))
}

// GetOneClickReturnFailNotify mirrors DeviceInfoQueryImpl.getOneClickReturnFailNotify.
func GetOneClickReturnFailNotify(tenantID, carID string) string {
	return redis.Get(rediskey.OneClickReturnBikeNotify(tenantID, carID))
}

// QueryImeiByBattery mirrors DeviceInfoQueryImpl.queryImeiByBattery: ZRANGEBYSCORE
// the service_battery zset across hash buckets.
func QueryImeiByBattery(tenantID string, serviceID int64, minBattery, maxBattery int) []string {
	var out []string
	for h := 0; h < deviceHashCount; h++ {
		key := rediskey.ServiceBatteryZset(tenantID, serviceID, h)
		out = append(out, redis.ZRangeByScoreRange(key, float64(minBattery), float64(maxBattery))...)
	}
	return out
}

func queryScoreMap(tenantID string, serviceID int64, min, max float64, key func(string, int64, int) string) map[string]float64 {
	out := map[string]float64{}
	for h := 0; h < deviceHashCount; h++ {
		for _, sm := range redis.ZRangeByScoreWithScores(key(tenantID, serviceID, h), min, max) {
			out[sm.Member] = sm.Score
		}
	}
	return out
}

// QueryTotalMilesMap mirrors queryImeiAndTotalMilesByTotalMiles (imei -> totalMiles).
func QueryTotalMilesMap(tenantID string, serviceID int64, min, max float64) map[string]float64 {
	return queryScoreMap(tenantID, serviceID, min, max, rediskey.ServiceTotalMilesZset)
}

// QueryLockTimeMap mirrors queryDeviceByOrderTime (imei -> lockTime).
func QueryLockTimeMap(tenantID string, serviceID int64, min, max float64) map[string]float64 {
	return queryScoreMap(tenantID, serviceID, min, max, rediskey.ServiceLockTimeZset)
}

// QueryStaticTimeMap mirrors queryDeviceByStaticTime (imei -> endTime).
func QueryStaticTimeMap(tenantID string, serviceID int64, min, max float64) map[string]float64 {
	return queryScoreMap(tenantID, serviceID, min, max, rediskey.ServiceCarStaticTimeZset)
}

// RemoveDeviceTotalMiles mirrors DeviceInfoQueryImpl.removeDeviceTotalMiles:
// ZREM the imei from service_total_miles at hash = lastDigit(imei) % 3.
func RemoveDeviceTotalMiles(tenantID string, serviceID int64, imei string) {
	if imei == "" {
		return
	}
	last := imei[len(imei)-1]
	if last < '0' || last > '9' {
		return
	}
	hash := int(last-'0') % 3
	redis.ZRem(rediskey.ServiceTotalMilesZset(tenantID, serviceID, hash), imei)
}

// DeviceMapFakeNx mirrors DeviceInfoQueryImpl.deviceMapFakeNx: SETNX a 3s lock.
func DeviceMapFakeNx(tenantID string, serviceID int64) bool {
	return redis.SetNX(rediskey.DeviceMapFakeNx(tenantID, serviceID), "1", 3*time.Second)
}

// SetDeviceMapFake mirrors DeviceInfoQueryImpl.setDeviceMapFake: store amount +
// the location JSON array.
func SetDeviceMapFake(tenantID string, serviceID int64, locationsJSON string, amount int) {
	redis.Set(rediskey.DeviceMapFakeAmount(tenantID, serviceID), strconv.Itoa(amount))
	redis.Set(rediskey.DeviceMapFakeLocation(tenantID, serviceID), locationsJSON)
}

// GetDeviceMapAmount mirrors DeviceInfoQueryImpl.getDeviceMapAmount; default 0.
func GetDeviceMapAmount(tenantID string, serviceID int64) int {
	raw := redis.Get(rediskey.DeviceMapFakeAmount(tenantID, serviceID))
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return 0
}

// GetDeviceMapLocationRaw mirrors DeviceInfoQueryImpl.getDeviceMapLocation: the
// raw JSON array string of fake locations ("" when absent).
func GetDeviceMapLocationRaw(tenantID string, serviceID int64) string {
	return redis.Get(rediskey.DeviceMapFakeLocation(tenantID, serviceID))
}

// TenantImei is a (tenantId, imei) pair for cross-tenant device-info reads.
type TenantImei struct {
	TenantID string
	Imei     string
}

// GetRackDevicesByTenantImei mirrors DeviceInfoQueryImpl.getRackDeviceByTenantServiceImei:
// MGET device_info across tenants, decode, keep non-blank carId and on-shelf
// (operationState not containing 1).
func GetRackDevicesByTenantImei(items []TenantImei) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	if len(items) == 0 {
		return out
	}
	keys := make([]string, len(items))
	for i, it := range items {
		keys[i] = rediskey.DeviceInfo(it.TenantID, it.Imei)
	}
	for _, raw := range redis.MGet(keys...) {
		if raw == "" {
			continue
		}
		dev := protocol.Decode(raw)
		if carID, _ := dev["carId"].(string); carID == "" {
			continue
		}
		if op, ok := dev["operationState"].([]int); ok && containsIntSlice(op, 1) {
			continue
		}
		out = append(out, dev)
	}
	return out
}

// GetRackDeviceByServiceId mirrors DeviceInfoQueryImpl.getRackDeviceByServiceId:
// the service-area devices filtered to on-shelf (operationState not containing 1).
func GetRackDeviceByServiceId(tenantID string, serviceID int64) []map[string]interface{} {
	all := GetDeviceByServiceId(tenantID, serviceID)
	out := make([]map[string]interface{}, 0, len(all))
	for _, d := range all {
		if op, ok := d["operationState"].([]int); ok && containsIntSlice(op, 1) {
			continue
		}
		out = append(out, d)
	}
	return out
}

func containsIntSlice(list []int, v int) bool {
	for _, e := range list {
		if e == v {
			return true
		}
	}
	return false
}

// GetDeviceInfoListByServiceList mirrors DeviceInfoQueryImpl.getDeviceInfoListByServiceList:
// the imei sets of all given services (ZRANGE full, 3 buckets), then device-info.
func GetDeviceInfoListByServiceList(tenantID string, serviceIDs []int64) []map[string]interface{} {
	var imeis []string
	for _, sid := range serviceIDs {
		imeis = append(imeis, GetImeiByServiceId(tenantID, sid)...)
	}
	return GetDeviceInfoList(tenantID, imeis)
}

// MGetImeiByCarIds returns imei per carId positionally ("" when unbound),
// mirroring Java redisMapper.mGet(car_imei keys) (nulls preserved).
func MGetImeiByCarIds(tenantID string, carIDs []string) []string {
	keys := make([]string, len(carIDs))
	for i, c := range carIDs {
		keys[i] = rediskey.CarImeiBind(tenantID, c)
	}
	return redis.MGet(keys...)
}

// MGetCarIdByImeis returns carId per imei positionally ("" when unbound),
// mirroring Java redisMapper.mGet(imei_car keys).
func MGetCarIdByImeis(tenantID string, imeis []string) []string {
	keys := make([]string, len(imeis))
	for i, im := range imeis {
		keys[i] = rediskey.ImeiCarBind(tenantID, im)
	}
	return redis.MGet(keys...)
}

// SetNoRiskControlEx mirrors DeviceInfoQueryImpl.setNoRiskControlEx: mark the
// imei as not risk-controlled for 4 hours.
func SetNoRiskControlEx(tenantID, imei string) {
	redis.SetEx(rediskey.NoRiskControlEx(tenantID, imei), "1", 4*time.Hour)
}

// DeleteNoRiskControlEx mirrors DeviceInfoQueryImpl.deleteNoRiskControlEx.
func DeleteNoRiskControlEx(tenantID, imei string) {
	redis.Del(rediskey.NoRiskControlEx(tenantID, imei))
}

// SetTempUnLockTag mirrors DeviceInfoQueryImpl.setTempUnLockTag: temporary
// power-on tag with a seconds TTL.
func SetTempUnLockTag(tenantID, imei string, seconds int64) {
	redis.SetEx(rediskey.TempRidingTag(tenantID, imei), "1", time.Duration(seconds)*time.Second)
}

// ClearSaddleOverloadContact mirrors DeviceInfoQueryImpl.clearSaddleOverloadContact.
func ClearSaddleOverloadContact(tenantID, imei string) {
	redis.Del(rediskey.SaddleOverloadContact(tenantID, imei))
}

// UpdateDeviceState mirrors DeviceRepository.update: encode the non-null fields
// of the partial device record into SETRANGE ops on the device_info string, then
// refresh the service-gfence zset report time. device must carry "imei". A blank
// tenantID or zero serviceID skips the zAdd (matches updateReportTime's guard).
func UpdateDeviceState(tenantID string, device map[string]interface{}, serviceID int64) {
	imei, _ := device["imei"].(string)
	if imei == "" {
		return
	}
	key := rediskey.DeviceInfo(tenantID, imei)
	for _, sr := range protocol.EncodeNotNull(device) {
		redis.SetRange(key, sr.Value, sr.Offset)
	}
	updateReportTime(tenantID, serviceID, imei)
}

// updateReportTime mirrors DeviceRepository.updateReportTime: zAdd imei -> now in
// the service_gfence_{tenant}_{serviceId}_{hash} zset (hash = lastDigit % 3).
func updateReportTime(tenantID string, serviceID int64, imei string) {
	if tenantID == "" || serviceID == 0 || imei == "" {
		return
	}
	last := imei[len(imei)-1]
	if last < '0' || last > '9' {
		return
	}
	hash := int(last-'0') % deviceHashCount
	redis.ZAdd(rediskey.ServiceGfenceCarZset(tenantID, serviceID, hash), imei, float64(time.Now().UnixMilli()))
}

// GetLocations mirrors DeviceInfoQueryImpl.getLocations: GEORADIUS each hash
// bucket of the device_location set, merging imei->distance(meters).
func GetLocations(tenantID string, serviceID int64, lat, lng, radius float64) map[string]float64 {
	out := map[string]float64{}
	for h := 0; h < deviceHashCount; h++ {
		key := rediskey.DeviceLocation(tenantID, serviceID, h)
		for imei, dist := range redis.GeoRadius(key, lng, lat, radius) {
			out[imei] = dist
		}
	}
	return out
}
