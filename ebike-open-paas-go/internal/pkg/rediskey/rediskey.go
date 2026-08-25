// Package rediskey centralises the Redis key layout owned by this service.
//
// Naming follows the platform convention `{semantic}_{scope}_{id}` used by the
// Java DevicePassRedisKey, with an `openPaas` prefix so the open platform's keys
// are trivially separable from device-shadow keys in the shared instance.
package rediskey

import "fmt"

// Callback returns the set key holding the registered callback URLs for one
// agent + event pair. A set matches Xiaoan's semantics exactly: registering the
// same URL twice reports 0 added, and multiple URLs per event are supported.
//
//	openPaas_callback_{agentId}_{event}  ->  SET of url
func Callback(agentID string, event int) string {
	return fmt.Sprintf("openPaas_callback_%s_%d", agentID, event)
}

// NotifyState returns the hash holding the last-known state used to derive the
// notify codes saas_0 does not report directly (fence crossings, SOC steps).
//
//	openPaas_notifyState_{tenantId}_{imei}  ->  HASH field -> last value
func NotifyState(tenantID, imei string) string {
	return fmt.Sprintf("openPaas_notifyState_%s_%s", tenantID, imei)
}

// RateLimit returns the sliding-window counter hash key for an agent.
//
//	openPaas_rateLimit_{agentId}  ->  HASH bucket(yyyyMMddHHmm) -> count
func RateLimit(agentID string) string {
	return fmt.Sprintf("openPaas_rateLimit_%s", agentID)
}

// DeviceInfo returns the device shadow key read for cached queries, matching the
// Java DevicePassRedisKey DEVICE_INFO format written by ebike-device-consume.
//
//	device_info_{tenantId}_{imei}
func DeviceInfo(tenantID, imei string) string {
	return fmt.Sprintf("device_info_%s_%s", tenantID, imei)
}

// DeviceEbike returns the IOT device-registry hash written by anvelink-console
// (saveBatchTenantDevice) and read by openapi's verifyTenantId.
//
//	device_ebike_{imei}  ->  HASH tenantId, deviceProtocol, …
func DeviceEbike(imei string) string {
	return fmt.Sprintf("device_ebike_%s", imei)
}

// ImeiCar returns the imei -> carId binding key (rental-domain; not used for
// open-paas device ownership).
//
//	imei_car_{tenantId}_{imei}
func ImeiCar(tenantID, imei string) string {
	return fmt.Sprintf("imei_car_%s_%s", tenantID, imei)
}
