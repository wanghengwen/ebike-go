package protocol

import (
	"strconv"
	"strings"
)

type fieldSpec struct {
	name   string
	offset int
	length int
}

var deviceFields []fieldSpec

func init() {
	specs := []struct {
		name   string
		length int
	}{
		{"imei", 15}, {"imsi", 15}, {"version", 11}, {"deviceType", 3},
		{"carId", 15}, {"serviceId", 19}, {"tenantId", 20}, {"reportTime", 13},
		{"gsmSignal", 3}, {"defend", 1}, {"acc", 1}, {"wgs84Lat", 10}, {"wgs84Lng", 11},
		{"lat", 10}, {"lng", 11}, {"timestamp", 13}, {"speed", 6}, {"course", 6}, {"hdop", 7},
		{"batteryLock", 1}, {"batteryConnect", 1}, {"backWheelLock", 1}, {"isAutoLock", 1},
		{"voltage", 6}, {"isMoving", 1}, {"totalMiles", 11}, {"fenceVersion", 11},
		{"moveAlarmOn", 1}, {"overSpeedOn", 1}, {"rfidAck", 16}, {"headingAngle", 6},
		{"helmetType", 1}, {"helmetLock", 1}, {"helmetReact", 1}, {"isWheelSpan", 1},
		{"isFenceEnable", 1}, {"isOutofServAera", 1}, {"noParkId", 19}, {"forParkId", 19},
		{"batteryId", 19}, {"maxMileage", 3}, {"restMileage", 3}, {"isPowerExist", 1},
		{"isOnline", 1}, {"bmsSN", 20}, {"soc", 3}, {"bmsTimeStamp", 13},
		{"restBattery", 3}, {"lockTime", 13}, {"unlockTime", 13},
		{"ridingState", 2}, {"operationState", 16}, {"alarmState", 16},
		{"helmetBind", 1}, {"helmetSOC", 3}, {"helmetState", 1},
		{"helmetAngleFault", 1}, {"helmetCapacitanceFault", 1}, {"helmetTinfraredFault", 1},
		{"helmetPressureFault", 1}, {"noRideParkId", 19}, {"isDisconnect", 1},
	}
	offset := 0
	for _, s := range specs {
		deviceFields = append(deviceFields, fieldSpec{name: s.name, offset: offset, length: s.length})
		offset += s.length
	}
}

// DeviceInfo mirrors Java DeviceInfoDO fields used by analyze service.
type DeviceInfo struct {
	Imei            string
	CarID           string
	ServiceID       int64
	TenantID        string
	Acc             *int
	Lat             float64
	Lng             float64
	Timestamp       int64
	IsOnline        *int
	Voltage         *int
	ReportTime      int64
	IsOutofServAera *int
	NoParkID        *int64
	ForParkID       *int64
	NoRideParkID    *int64
	BatteryID       *int64
	BatteryLock     *int
	RestBattery     *int
	LockTime        *int64
	UnlockTime      *int64
	RidingState     *int
	OperationState  []int
	AlarmState      []int
}

// DecodeDevice parses the fixed-length Redis device protocol string.
func DecodeDevice(raw string) *DeviceInfo {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	d := &DeviceInfo{}
	for _, f := range deviceFields {
		end := f.offset + f.length
		if len(raw) < end {
			break
		}
		val := strings.TrimSpace(raw[f.offset:end])
		if val == "" {
			continue
		}
		switch f.name {
		case "imei":
			d.Imei = val
		case "carId":
			d.CarID = val
		case "serviceId":
			d.ServiceID, _ = strconv.ParseInt(val, 10, 64)
		case "tenantId":
			d.TenantID = val
		case "acc":
			d.Acc = parseIntPtr(val)
		case "lat":
			d.Lat, _ = strconv.ParseFloat(val, 64)
		case "lng":
			d.Lng, _ = strconv.ParseFloat(val, 64)
		case "timestamp":
			d.Timestamp, _ = strconv.ParseInt(val, 10, 64)
		case "isOnline":
			d.IsOnline = parseIntPtr(val)
		case "voltage":
			d.Voltage = parseIntPtr(val)
		case "reportTime":
			d.ReportTime, _ = strconv.ParseInt(val, 10, 64)
		case "isOutofServAera":
			d.IsOutofServAera = parseIntPtr(val)
		case "noParkId":
			d.NoParkID = parseInt64Ptr(val)
		case "forParkId":
			d.ForParkID = parseInt64Ptr(val)
		case "noRideParkId":
			d.NoRideParkID = parseInt64Ptr(val)
		case "batteryId":
			d.BatteryID = parseInt64Ptr(val)
		case "batteryLock":
			d.BatteryLock = parseIntPtr(val)
		case "restBattery":
			d.RestBattery = parseIntPtr(val)
		case "lockTime":
			d.LockTime = parseInt64Ptr(val)
		case "unlockTime":
			d.UnlockTime = parseInt64Ptr(val)
		case "ridingState":
			d.RidingState = parseIntPtr(val)
		case "operationState":
			d.OperationState = getOneIndexes(val)
		case "alarmState":
			d.AlarmState = getOneIndexes(val)
		}
	}
	return d
}

func parseIntPtr(val string) *int {
	n, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &n
}

func parseInt64Ptr(val string) *int64 {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func getOneIndexes(hexStr string) []int {
	if strings.TrimSpace(hexStr) == "" {
		return nil
	}
	value, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil || value == 0 {
		return nil
	}
	var indexes []int
	idx := 0
	for value != 0 {
		if value&1 == 1 {
			indexes = append(indexes, idx)
		}
		value >>= 1
		idx++
	}
	return indexes
}

func ContainsOpState(states []int, target int) bool {
	for _, s := range states {
		if s == target {
			return true
		}
	}
	return false
}

func IntVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func Int64Val(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
