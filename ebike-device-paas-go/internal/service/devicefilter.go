package service

import (
	"errors"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// ErrParam mirrors the Java MsgCodeEnum.PARAM_EXCEPTION path.
var ErrParam = errors.New("param exception")

// GetDeviceListByCarIdList ports DeviceInfoServiceImpl.getDeviceListByCarIdList:
// carId -> imei -> device-info -> DeviceListCo. Empty carIdList yields [].
func GetDeviceListByCarIdList(tenantID string, carIDs []string) []dto.DeviceListCo {
	if len(carIDs) == 0 {
		return []dto.DeviceListCo{}
	}
	imeis := repository.GetImeiListByCarId(tenantID, carIDs)
	devices := repository.GetDeviceInfoList(tenantID, imeis)
	out := make([]dto.DeviceListCo, 0, len(devices))
	for _, d := range devices {
		out = append(out, *mapDeviceListItem(d))
	}
	return out
}

// QueryDeviceCacheGps ports DeviceInfoServiceImpl.queryDeviceCacheGps:
// imei -> device-info -> DeviceGpsCo (lng/lat default 0.0).
func QueryDeviceCacheGps(tenantID string, imeiList []string) []dto.DeviceGpsCo {
	devices := repository.GetDeviceInfoList(tenantID, imeiList)
	out := make([]dto.DeviceGpsCo, 0, len(devices))
	for _, d := range devices {
		out = append(out, dto.DeviceGpsCo{
			Imei:  strPtr(d, "imei"),
			CarId: strPtr(d, "carId"),
			Lng:   f64PtrDefault(d, "lng", 0),
			Lat:   f64PtrDefault(d, "lat", 0),
		})
	}
	return out
}

// QueryCarImeiBind ports DeviceInfoServiceImpl.queryDeviceListByImeiList:
// carIdList -> positional imei pairs, else imeiList -> positional carId pairs.
func QueryCarImeiBind(tenantID string, qry dto.ImeiListQry) ([]dto.CarImeiCo, error) {
	if len(qry.CarIdList) > 0 {
		imeis := repository.MGetImeiByCarIds(tenantID, qry.CarIdList)
		out := make([]dto.CarImeiCo, 0, len(qry.CarIdList))
		for i, carID := range qry.CarIdList {
			c := carID
			out = append(out, dto.CarImeiCo{Imei: strOrNil(imeis[i]), CarId: &c})
		}
		return out, nil
	}
	if len(qry.ImeiList) > 0 {
		carIDs := repository.MGetCarIdByImeis(tenantID, qry.ImeiList)
		out := make([]dto.CarImeiCo, 0, len(qry.ImeiList))
		for i, imei := range qry.ImeiList {
			im := imei
			out = append(out, dto.CarImeiCo{Imei: &im, CarId: strOrNil(carIDs[i])})
		}
		return out, nil
	}
	return nil, ErrParam
}

// GetDeviceFilterList ports DeviceInfoServiceImpl.getDeviceFilterList: read the
// service-area devices, apply the platform filter, and plain-copy to DevicePageCo
// (carHelmetState / address are left null, matching the non-amap copyProperties).
func GetDeviceFilterList(tenantID string, qry dto.DevicePageQry) []dto.DevicePageCo {
	devices := repository.GetDeviceByServiceId(tenantID, qry.ServiceId.Int64())
	filtered := filterDevices(devices, qry)
	out := make([]dto.DevicePageCo, 0, len(filtered))
	for _, d := range filtered {
		out = append(out, mapDevicePageCo(d))
	}
	return out
}

func mapDevicePageCo(d map[string]interface{}) dto.DevicePageCo {
	return dto.DevicePageCo{
		Imei:           strPtr(d, "imei"),
		CarId:          strPtr(d, "carId"),
		Timestamp:      i64Ptr(d, "timestamp"),
		IsOnline:       intPtr(d, "isOnline"),
		RestBattery:    restBattery(d),
		RidingState:    intPtr(d, "ridingState"),
		OperationState: intList(d, "operationState"),
		AlarmState:     intList(d, "alarmState"),
		Lng:            f64PtrDefault(d, "lng", 0),
		Lat:            f64PtrDefault(d, "lat", 0),
		ScanLng:        f64Ptr(d, "scanLng"),
		ScanLat:        f64Ptr(d, "scanLat"),
		HelmetSOC:      helmetSOC(d),
		CarTagTypeIds:  intList(d, "carTagTypeIds"),
	}
}

// filterDevices ports the single-arg getFilterDevice (platform-end) predicate chain.
func filterDevices(devices []map[string]interface{}, q dto.DevicePageQry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(devices))
	for _, d := range devices {
		if q.RidingState != nil {
			if rs, ok := asInt(d["ridingState"]); !ok || rs != *q.RidingState {
				continue
			}
		}
		if !operationStateMatch(d, q.OperationState) {
			continue
		}
		if q.AlarmState != nil && !containsInt(intList(d, "alarmState"), *q.AlarmState) {
			continue
		}
		if !filterNoOrderTime(q.NoOrderTime, d) {
			continue
		}
		if !filterOrderTime(q.OrderTime, d) {
			continue
		}
		if q.RestBattery != nil {
			rb := 0
			if p := restBattery(d); p != nil {
				rb = *p
			}
			if rb > *q.RestBattery {
				continue
			}
		}
		if v := q.BatteryId.Ptr(); v != nil {
			bid, ok := asInt64(d["batteryId"])
			if !ok || bid != *v {
				continue
			}
		}
		if q.CarHelmetState != nil && carHelmetState(d) != *q.CarHelmetState {
			continue
		}
		if q.HelmetSOC != nil {
			hs := 0
			if p := helmetSOC(d); p != nil {
				hs = *p
			}
			if hs > *q.HelmetSOC {
				continue
			}
		}
		if len(q.CarTagTypeIds) > 0 && !intersectsInt(q.CarTagTypeIds, intList(d, "carTagTypeIds")) {
			continue
		}
		out = append(out, d)
	}
	return out
}

// operationStateMatch ports the operationState branch of getFilterDevice.
func operationStateMatch(d map[string]interface{}, op *int) bool {
	states := intList(d, "operationState")
	if op == nil {
		return !containsInt(states, 1)
	}
	if *op == 1 && containsInt(states, 1) {
		return true
	}
	return containsInt(states, *op) && !containsInt(states, 1)
}

// filterOrderTime ports DeviceInfoServiceImpl.filterOrderTime.
func filterOrderTime(orderTime *float64, d map[string]interface{}) bool {
	if orderTime == nil {
		return true
	}
	rs, _ := asInt(d["ridingState"])
	if rs != 2 && rs != 3 {
		return false
	}
	now := time.Now().UnixMilli()
	lock, _ := asInt64(d["lockTime"])
	unlock := now
	if v, ok := asInt64(d["unlockTime"]); ok {
		unlock = v
	}
	if unlock > lock {
		lock = now
	}
	return float64(lock-unlock) >= *orderTime*60*60*1000
}

// filterNoOrderTime ports DeviceInfoServiceImpl.filterNoOrderTime.
func filterNoOrderTime(noOrderTime *float64, d map[string]interface{}) bool {
	if noOrderTime == nil {
		return true
	}
	if containsInt(intList(d, "operationState"), 2) {
		return false
	}
	lock, _ := asInt64(d["lockTime"])
	unlock, _ := asInt64(d["unlockTime"])
	if unlock > lock {
		return false
	}
	return float64(time.Now().UnixMilli()-lock) >= *noOrderTime*60*60*1000
}

func containsInt(list []int, v int) bool {
	for _, e := range list {
		if e == v {
			return true
		}
	}
	return false
}

func intersectsInt(a, b []int) bool {
	set := make(map[int]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
