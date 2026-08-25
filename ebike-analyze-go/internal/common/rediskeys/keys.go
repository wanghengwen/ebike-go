package rediskeys

import "fmt"

func DeviceInfo(tenantID, imei string) string {
	return fmt.Sprintf("device_info_%s_%s", tenantID, imei)
}

func ServiceGfenceCarZSet(tenantID string, serviceID int64, hash int) string {
	return fmt.Sprintf("service_gfence_%s_%d_%d", tenantID, serviceID, hash)
}

func CarImeiBind(tenantID, carID string) string {
	return fmt.Sprintf("car_imei_%s_%s", tenantID, carID)
}

func MoveAlarmTag(tenantID, imei string) string {
	return fmt.Sprintf("move_alarm_tag_%s_%s", tenantID, imei)
}

func HotPointStart(tenantID string, serviceID int64, yyyyMMdd string) string {
	return fmt.Sprintf("hot_point_start:%s:%d:%s", tenantID, serviceID, yyyyMMdd)
}

func HotPointEnd(tenantID string, serviceID int64, yyyyMMdd string) string {
	return fmt.Sprintf("hot_point_end:%s:%d:%s", tenantID, serviceID, yyyyMMdd)
}

const DeviceHashCount = 10
