package service

import (
	"errors"
	"strconv"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// ErrCarImeiBind is returned when a carId cannot be resolved to an imei.
var ErrCarImeiBind = errors.New("carId not bound to imei")

// Helmet business-state codes (Java HelmetStateEnum).
const (
	helmetNotBind       = 1
	helmetReactState    = 2
	helmetRidingNotWear = 3
	helmetWear          = 4
	helmetLost          = 5
	helmetLockWarn      = 6
)

// GetDeviceDetail ports DeviceInfoServiceImpl.getDeviceDetail: resolve imei
// (carId wins), read+decode the device, copy to DeviceDetailCo (DO getter
// defaults), then set version + carHelmetState.
func GetDeviceDetail(tenantID string, qry dto.DeviceDetailQry) (*dto.DeviceDetailCo, error) {
	imei := qry.Imei
	if qry.CarId != "" {
		imei = repository.GetImeiByCarId(tenantID, qry.CarId)
		if imei == "" {
			return nil, ErrCarImeiBind
		}
	}
	dev := repository.GetDeviceByImei(tenantID, imei)
	if dev == nil {
		return nil, ErrDeviceNotFound
	}
	return mapDeviceDetail(dev), nil
}

// mapDeviceDetail copies the decoded device map into a DeviceDetailCo applying
// the DeviceInfoDO getter defaults and the two computed fields.
func mapDeviceDetail(d map[string]interface{}) *dto.DeviceDetailCo {
	co := &dto.DeviceDetailCo{
		Imei:      strPtr(d, "imei"),
		Imsi:      strPtr(d, "imsi"),
		CarId:     strPtr(d, "carId"),
		ServiceId: i64Ptr(d, "serviceId"),
		TenantId:  strPtr(d, "tenantId"),

		ReportTime: i64Ptr(d, "reportTime"),
		GsmSignal:  intPtr(d, "gsmSignal"),
		Defend:     intPtrDefault(d, "defend", 1),
		Acc:        intPtrDefault(d, "acc", 0),
		Wgs84Lat:   f64Ptr(d, "wgs84Lat"),
		Wgs84Lng:   f64Ptr(d, "wgs84Lng"),
		Lat:        f64PtrDefault(d, "lat", 0),
		Lng:        f64PtrDefault(d, "lng", 0),
		Timestamp:  i64Ptr(d, "timestamp"),
		Speed:      f64Ptr(d, "speed"),
		Course:     f64Ptr(d, "course"),
		Hdop:       f64Ptr(d, "hdop"),

		BatteryLock:     intPtrDefault(d, "batteryLock", 1),
		BatteryConnect:  intPtr(d, "batteryConnect"),
		BackWheelLock:   intPtrDefault(d, "backWheelLock", 1),
		IsAutoLock:      intPtr(d, "isAutoLock"),
		Voltage:         intPtrDefault(d, "voltage", 0),
		IsMoving:        intPtr(d, "isMoving"),
		TotalMiles:      f64Ptr(d, "totalMiles"),
		MoveAlarmOn:     intPtr(d, "moveAlarmOn"),
		OverSpeedOn:     intPtr(d, "overSpeedOn"),
		RfidCarId:       strPtr(d, "rfidCarId"),
		HeadingAngle:    intPtr(d, "headingAngle"),
		HelmetType:      intPtr(d, "helmetType"),
		HelmetLock:      intPtrDefault(d, "helmetLock", 0),
		HelmetReact:     intPtrDefault(d, "helmetReact", 0),
		IsWheelSpan:     intPtr(d, "isWheelSpan"),
		IsFenceEnable:   intPtr(d, "isFenceEnable"),
		IsOutofServAera: intPtr(d, "isOutofServAera"),
		NoParkId:        i64Ptr(d, "noParkId"),
		ForParkId:       i64Ptr(d, "forParkId"),
		NoRideParkId:    i64Ptr(d, "noRideParkId"),
		BatteryId:       i64Ptr(d, "batteryId"),
		IsPowerExist:    intPtr(d, "isPowerExist"),
		IsOnline:        intPtr(d, "isOnline"),

		BmsSN:        strPtr(d, "bmsSN"),
		Soc:          intPtr(d, "soc"),
		BmsTimeStamp: i64Ptr(d, "bmsTimeStamp"),

		RestMileage: intPtr(d, "restMileage"),
		LockTime:    i64Ptr(d, "lockTime"),
		UnlockTime:  i64Ptr(d, "unlockTime"),

		RidingState:    intPtr(d, "ridingState"),
		OperationState: intList(d, "operationState"),
		AlarmState:     intList(d, "alarmState"),

		HelmetBind:  intPtrDefault(d, "helmetBind", 0),
		HelmetSOC:   helmetSOC(d),
		HelmetState: intPtrDefault(d, "helmetState", 0),

		HelmetAngleFault:       intPtr(d, "helmetAngleFault"),
		HelmetCapacitanceFault: intPtr(d, "helmetCapacitanceFault"),
		HelmetTinfraredFault:   intPtr(d, "helmetTinfraredFault"),
		HelmetPressureFault:    intPtr(d, "helmetPressureFault"),

		IsDisconnect:  intPtr(d, "isDisconnect"),
		RfidAck:       intPtr(d, "rfidAck"),
		RfidTimestamp: i64Ptr(d, "rfidTimestamp"),

		MaintainAreaId: i64Ptr(d, "maintainAreaId"),
		CarTagTypeIds:  intList(d, "carTagTypeIds"),

		Overload:          intPtr(d, "overload"),
		OverloadThreshold: intPtr(d, "overloadThreshold"),
		IsSupportOverload: intPtr(d, "isSupportOverload"),
	}
	co.RestBattery = restBattery(d)
	if v := numberToVersion(d["version"]); v != nil {
		co.Version = v
	}
	hs := carHelmetState(d)
	co.CarHelmetState = &hs
	return co
}

// GetDeviceListByImeiList ports DeviceInfoServiceImpl.getDeviceListByImeiList:
// read device-info for each imei (carId non-blank) and project to DeviceListCo.
func GetDeviceListByImeiList(tenantID string, imeiList []string) []dto.DeviceListCo {
	devices := repository.GetDeviceInfoList(tenantID, imeiList)
	out := make([]dto.DeviceListCo, 0, len(devices))
	for _, d := range devices {
		out = append(out, *mapDeviceListItem(d))
	}
	return out
}

func mapDeviceListItem(d map[string]interface{}) *dto.DeviceListCo {
	co := &dto.DeviceListCo{
		Imei:      strPtr(d, "imei"),
		CarId:     strPtr(d, "carId"),
		ServiceId: i64Ptr(d, "serviceId"),

		MaintainAreaId: i64Ptr(d, "maintainAreaId"),

		Acc:             intPtrDefault(d, "acc", 0),
		Defend:          intPtrDefault(d, "defend", 1),
		Lat:             f64PtrDefault(d, "lat", 0),
		Lng:             f64PtrDefault(d, "lng", 0),
		Timestamp:       i64Ptr(d, "timestamp"),
		IsOnline:        intPtr(d, "isOnline"),
		Voltage:         intPtrDefault(d, "voltage", 0),
		ReportTime:      i64Ptr(d, "reportTime"),
		IsOutofServAera: intPtr(d, "isOutofServAera"),
		NoParkId:        i64Ptr(d, "noParkId"),
		ForParkId:       i64Ptr(d, "forParkId"),
		NoRideParkId:    i64Ptr(d, "noRideParkId"),
		BatteryId:       i64Ptr(d, "batteryId"),
		BatteryLock:     intPtrDefault(d, "batteryLock", 1),

		LockTime:   i64Ptr(d, "lockTime"),
		UnlockTime: i64Ptr(d, "unlockTime"),

		RidingState:    intPtr(d, "ridingState"),
		OperationState: intList(d, "operationState"),
		AlarmState:     intList(d, "alarmState"),
		CarTagTypeIds:  intList(d, "carTagTypeIds"),
	}
	co.RestBattery = restBattery(d)
	return co
}

// restBattery mirrors DeviceInfoDO.getRestBattery: soc when fresh (non-zero and
// bmsTimeStamp within 30min), else the raw restBattery field.
func restBattery(d map[string]interface{}) *int {
	soc, hasSoc := asInt(d["soc"])
	bms, _ := asInt64(d["bmsTimeStamp"])
	if hasSoc && soc != 0 && (bms+30*60)*1000 > time.Now().UnixMilli() {
		s := soc
		return &s
	}
	return intPtr(d, "restBattery")
}

// helmetSOC mirrors DeviceInfoDO.getHelmetSOC: cap at 100, null when absent.
func helmetSOC(d map[string]interface{}) *int {
	v, ok := asInt(d["helmetSOC"])
	if !ok {
		return nil
	}
	if v > 100 {
		v = 100
	}
	return &v
}

// numberToVersion mirrors DeviceUtils.number2Version(Long): major.minor.patch
// from the packed integer version. The raw stored value is a numeric string.
func numberToVersion(v interface{}) *string {
	var n int64
	switch t := v.(type) {
	case string:
		parsed, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil
		}
		n = parsed
	case int64:
		n = t
	case int:
		n = int64(t)
	case float64:
		n = int64(t)
	default:
		return nil
	}
	s := strconv.FormatInt((n&0xff0000)>>16, 10) + "." +
		strconv.FormatInt((n&0xff00)>>8, 10) + "." +
		strconv.FormatInt(n&0xff, 10)
	return &s
}

// carHelmetState mirrors DeviceInfoServiceImpl.getCarHelmetState using the DO
// getter defaults (absent helmet fields -> 0, isCarBindHelmet -> 0).
func carHelmetState(d map[string]interface{}) int {
	isBind := intDefault(d, "isCarBindHelmet", 0)
	if isBind == 0 {
		return helmetNotBind
	}
	lock := intDefault(d, "helmetLock", 0)
	react := intDefault(d, "helmetReact", 0)
	if (lock ^ react) == 1 {
		return helmetLockWarn
	}
	riding, _ := asInt(d["ridingState"])
	if riding == 2 || riding == 3 {
		if intDefault(d, "helmetBind", 0) == 1 {
			if intDefault(d, "helmetState", 0) == 1 {
				return helmetWear
			}
			return helmetRidingNotWear
		}
		if lock == 0 && react == 0 {
			return helmetWear
		}
	} else {
		if (lock | react) == 0 {
			return helmetLost
		}
	}
	return helmetReactState
}

// ---- decoded-map accessors ----

func strPtr(d map[string]interface{}, k string) *string {
	if s, ok := d[k].(string); ok {
		return &s
	}
	return nil
}

func intPtr(d map[string]interface{}, k string) *int {
	if v, ok := asInt(d[k]); ok {
		return &v
	}
	return nil
}

func intPtrDefault(d map[string]interface{}, k string, def int) *int {
	if v, ok := asInt(d[k]); ok {
		return &v
	}
	return &def
}

func intDefault(d map[string]interface{}, k string, def int) int {
	if v, ok := asInt(d[k]); ok {
		return v
	}
	return def
}

func i64Ptr(d map[string]interface{}, k string) *int64 {
	if v, ok := asInt64(d[k]); ok {
		return &v
	}
	return nil
}

func f64Ptr(d map[string]interface{}, k string) *float64 {
	if v, ok := asFloat64(d[k]); ok {
		return &v
	}
	return nil
}

func f64PtrDefault(d map[string]interface{}, k string, def float64) *float64 {
	if v, ok := asFloat64(d[k]); ok {
		return &v
	}
	return &def
}

func intList(d map[string]interface{}, k string) []int {
	switch t := d[k].(type) {
	case []int:
		if t == nil {
			return []int{}
		}
		return t
	case []interface{}:
		out := make([]int, 0, len(t))
		for _, e := range t {
			if n, ok := asInt(e); ok {
				out = append(out, n)
			}
		}
		return out
	default:
		return []int{}
	}
}
