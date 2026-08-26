// Package rediskey builds Redis keys identical to the Java DevicePassRedisKey
// formats (placeholder substitution), so Go and Java read the same entries.
package rediskey

import "fmt"

// DeviceInfo: device_info_{tenantId}_{imei} — fixed-length device-info string.
func DeviceInfo(tenantID, imei string) string {
	return fmt.Sprintf("device_info_%s_%s", tenantID, imei)
}

// CarImeiBind: car_imei_{tenantId}_{carId} — carId -> imei.
func CarImeiBind(tenantID, carID string) string {
	return fmt.Sprintf("car_imei_%s_%s", tenantID, carID)
}

// ImeiCarBind: imei_car_{tenantId}_{imei} — imei -> carId.
func ImeiCarBind(tenantID, imei string) string {
	return fmt.Sprintf("imei_car_%s_%s", tenantID, imei)
}

// ServiceGfenceCarZset: service_gfence_{tenantId}_{serviceId}_{hash}.
func ServiceGfenceCarZset(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_gfence_%s_%d_%d", tenantID, serviceID, hash)
}

// DeviceLocation: device_location_{tenantId}_{serviceId}_{hash} — GEO set.
func DeviceLocation(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("device_location_%s_%d_%d", tenantID, serviceID, hash)
}

// BluetoothToken: bluetooth_token_{tenantId}_{imei}.
func BluetoothToken(tenantID, imei string) string {
	return fmt.Sprintf("bluetooth_token_%s_%s", tenantID, imei)
}

// SaddleOverloadContact: saddle_overload_contact_{tenantId}_{imei}.
func SaddleOverloadContact(tenantID, imei string) string {
	return fmt.Sprintf("saddle_overload_contact_%s_%s", tenantID, imei)
}

// CameraInfo: camera_info_{tenantId}_{imei} — JSON camera state.
func CameraInfo(tenantID, imei string) string {
	return fmt.Sprintf("camera_info_%s_%s", tenantID, imei)
}

// OneClickReturnBikeNotify: one_click_return_bike_notify_{tenantId}_{carId}.
func OneClickReturnBikeNotify(tenantID, carID string) string {
	return fmt.Sprintf("one_click_return_bike_notify_%s_%s", tenantID, carID)
}

// ServiceBatteryZset: service_battery_{tenantId}_{serviceId}_{hash} — score=restBattery.
func ServiceBatteryZset(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_battery_%s_%d_%d", tenantID, serviceID, hash)
}

// ServiceTotalMilesZset: service_total_miles_{tenantId}_{serviceId}_{hash} — score=totalMiles.
func ServiceTotalMilesZset(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_total_miles_%s_%d_%d", tenantID, serviceID, hash)
}

// ServiceLockTimeZset: service_lock_time_{tenantId}_{serviceId}_{hash} — score=lockTime.
func ServiceLockTimeZset(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_lock_time_%s_%d_%d", tenantID, serviceID, hash)
}

// ServiceCarStaticTimeZset: service_car_static_time_{tenantId}_{serviceId}_{hash} — score=endTime.
func ServiceCarStaticTimeZset(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_car_static_time_%s_%d_%d", tenantID, serviceID, hash)
}

// DeviceMapFakeNx: device_map_fake_nx_{tenantId}_{serviceId} — gen lock.
func DeviceMapFakeNx(tenantID string, serviceID int64) string {
	return fmt.Sprintf("device_map_fake_nx_%s_%d", tenantID, serviceID)
}

// DeviceMapFakeAmount: device_map_fake_amount_{tenantId}_{serviceId}.
func DeviceMapFakeAmount(tenantID string, serviceID int64) string {
	return fmt.Sprintf("device_map_fake_amount_%s_%d", tenantID, serviceID)
}

// DeviceMapFakeLocation: device_map_fake_location_{tenantId}_{serviceId} — JSON list.
func DeviceMapFakeLocation(tenantID string, serviceID int64) string {
	return fmt.Sprintf("device_map_fake_location_%s_%d", tenantID, serviceID)
}

// NoRiskControlEx: no_risk_control_ex_{tenantId}_{imei} — "not risk-controlled" flag.
func NoRiskControlEx(tenantID, imei string) string {
	return fmt.Sprintf("no_risk_control_ex_%s_%s", tenantID, imei)
}

// TempRidingTag: temp_riding_tag_{tenantId}_{imei} — temporary power-on tag.
func TempRidingTag(tenantID, imei string) string {
	return fmt.Sprintf("temp_riding_tag_%s_%s", tenantID, imei)
}
