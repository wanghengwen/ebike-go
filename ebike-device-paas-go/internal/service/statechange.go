package service

import (
	"errors"
	"log"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// StateChange ports StateChangeServiceImpl: device-state writes (Redis device_info
// SETRANGE + service-gfence zAdd) and the C34 Kafka location upload. Write
// endpoints -> NOT shadow-compared.

// ErrDeviceNull mirrors DevicePassMsgCode.DEVICE_NULL_HAVE (对应设备不存在).
var ErrDeviceNull = errors.New("device null")

// ErrDeviceNullRack mirrors DevicePassMsgCode.DEVICE_NULL_RACK (车辆未上架).
var ErrDeviceNullRack = errors.New("device null rack")

// RidingStateChange ports StateChangeServiceImpl.ridingStateChange ->
// EbikeDataImpl.ridingStateChange (update ridingState + operationState).
func RidingStateChange(cmd *dto.RidingStateChangeCmd) {
	if cmd.Imei == "" {
		return
	}
	device := map[string]interface{}{
		"imei":       cmd.Imei,
		"reportTime": time.Now().UnixMilli(),
	}
	if cmd.RidingState != nil {
		device["ridingState"] = *cmd.RidingState
	}
	if cmd.OperationState != nil {
		device["operationState"] = cmd.OperationState
	}
	repository.UpdateDeviceState(dto.TenantOf(cmd.CommandContext), device, flexInt64Val(cmd.ServiceID))
}

// AlarmStateChange ports StateChangeServiceImpl.alarmStateChange ->
// EbikeDataImpl.alarmStateChange (update alarmState).
func AlarmStateChange(cmd *dto.AlarmStateChangeCmd) {
	if cmd.Imei == "" {
		return
	}
	device := map[string]interface{}{
		"imei":       cmd.Imei,
		"reportTime": time.Now().UnixMilli(),
	}
	if cmd.AlarmType != nil {
		device["alarmState"] = cmd.AlarmType
	}
	repository.UpdateDeviceState(dto.TenantOf(cmd.CommandContext), device, flexInt64Val(cmd.ServiceID))
}

// LocationChange ports StateChangeServiceImpl.locationChange: emit a C34 gps-only
// message per imei.
func LocationChange(cmd *dto.LocationChangeCmd) {
	tenantID := dto.TenantOf(cmd.CommandContext)
	lng, lat := f64(cmd.Lng), f64(cmd.Lat)
	for _, imei := range cmd.ImeiList {
		sendC34(tenantID, imei, c34Location(lng, lat))
	}
}

// ScanLocationChange ports StateChangeServiceImpl.scanLocationChange: resolve
// carId -> imei, require an on-shelf serviceId, then update scanLng/scanLat.
func ScanLocationChange(cmd *dto.ScanLocationChangeCmd) {
	tenantID := dto.TenantOf(cmd.CommandContext)
	imei := repository.GetImeiByCarId(tenantID, cmd.CarID)
	if imei == "" {
		log.Printf("[scanLocationChange] carId:%s get imei is null", cmd.CarID)
		return
	}
	device := repository.GetDeviceByImei(tenantID, imei)
	sid, ok := serviceIDOf(device)
	if !ok {
		return
	}
	upd := map[string]interface{}{
		"imei":       imei,
		"scanLng":    f64(cmd.Lng),
		"scanLat":    f64(cmd.Lat),
		"reportTime": time.Now().UnixMilli(),
	}
	repository.UpdateDeviceState(tenantID, upd, sid)
}

func flexInt64Val(f dto.FlexInt64) int64 {
	if f.V == nil {
		return 0
	}
	return *f.V
}

func int64Val(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// serviceIDOf extracts the decoded serviceId; ok=false when the device record is
// missing or has no serviceId (Java Assert.notNull(device)/notNull(serviceId)).
func serviceIDOf(device map[string]interface{}) (int64, bool) {
	if device == nil {
		return 0, false
	}
	sid, ok := device["serviceId"].(int64)
	if !ok {
		return 0, false
	}
	return sid, true
}
