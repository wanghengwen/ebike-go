package rediskeys

import (
	"fmt"
	"strings"
)

// Redis key templates aligned with Java FenceRedisKey (preserve typos for compatibility).

func ServiceArea(tenantID string, id int64) string {
	return fmt.Sprintf("fence_serviceArea_%s_%d", tenantID, id)
}

func Parking(tenantID string, id int64) string {
	return fmt.Sprintf("fence_parking_%s_%d", tenantID, id)
}

func NoParking(tenantID string, id int64) string {
	return fmt.Sprintf("fence_noParking_%s_%d", tenantID, id)
}

func BanRiding(tenantID string, id int64) string {
	return fmt.Sprintf("fence_banRiding_%s_%d", tenantID, id)
}

func MaintainArea(tenantID string, id int64) string {
	return fmt.Sprintf("fence_maintainArea_%s_%d", tenantID, id)
}

func FenceCustom(tenantID string, id int64) string {
	return fmt.Sprintf("fence_custom_%s_%d", tenantID, id)
}

func ServiceAreaGeo(tenantID string) string { return fmt.Sprintf("fence_serviceArea_geo_%s", tenantID) }
func ParkingGeo(tenantID string) string   { return fmt.Sprintf("fence_parking_geo_%s", tenantID) }
func NoParkingGeo(tenantID string) string   { return fmt.Sprintf("fence_noParking_geo_%s", tenantID) }
func BanRidingGeo(tenantID string) string   { return fmt.Sprintf("fence_banRiding_geo_%s", tenantID) }
func MaintainAreaGeo(tenantID string) string {
	return fmt.Sprintf("fence_maintainArea_geo_%s", tenantID)
}
func FenceCustomGeo(tenantID, id string) string {
	return fmt.Sprintf("fence_custom_geo_%s_%s", tenantID, id)
}

// FenceCustomGeoSearchKey is the Redis GEO key used by Java when only tenantId is passed to
// FenceRedisKey.FENCE_CUSTOM_GEO.format(tenantId) — the trailing "{id}" placeholder is left literal.
func FenceCustomGeoSearchKey(tenantID string) string {
	return fmt.Sprintf("fence_custom_geo_%s_{id}", tenantID)
}

func HomeScrollMsg(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_home_scroll_msg_%s_%d", tenantID, serviceID)
}

func HomeActivityEntrance(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_home_activity_entrance_%s_%d", tenantID, serviceID)
}

func HomeNav(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_home_nav_%s_%d", tenantID, serviceID)
}

func FAQ(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_faq_%s_%d", tenantID, serviceID)
}

func GuidePage(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_guide_page_%s_%d", tenantID, serviceID)
}

func SpecialTips(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_special_tips_%s_%d", tenantID, serviceID)
}

func CustomerService(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_customer_service_%s_%d", tenantID, serviceID)
}

func BaseItemConfig(tenantID string, serviceID int64) string {
	// Java FenceRedisKey.BASE_ITEM_CONFIG = config_base_item_config_{tenantId}_{serviceId}
	return fmt.Sprintf("config_base_item_config_%s_%d", tenantID, serviceID)
}

func PayConfig(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_pay_config_%s_%d", tenantID, serviceID)
}

func CreditScoreConfig(tenantID string) string {
	return fmt.Sprintf("config_credit_score_config_%s", tenantID)
}

func ParkApplyConfig(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_park_apply_config_%s_%d", tenantID, serviceID)
}

func PushRidingCardConfig(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_push_riding_card_config_v2_%s_%d", tenantID, serviceID)
}

func UseCarConfig(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_use_car_config_v2_%s_%d", tenantID, serviceID)
}

func BackCarConfig(tenantID string, serviceID int64) string {
	return fmt.Sprintf("config_back_car_config_%s_%d", tenantID, serviceID)
}

func AlarmContact(tenantID string, serviceID int64) string {
	return fmt.Sprintf("alarm_contact_%s_%d", tenantID, serviceID)
}

func AdConfig(tenantID string, serviceID int64) string {
	// Java FenceRedisKey.AD_CONFIG = "ad_config_{tenantId}_{serviceId}" (no _v2 suffix),
	// must match so C-end consumers reading ad_config_* find the same value.
	return fmt.Sprintf("ad_config_%s_%d", tenantID, serviceID)
}

func RidingPermission(tenantID string, serviceID int64) string {
	return fmt.Sprintf("riding_permission_v2_%s_%d", tenantID, serviceID)
}

func ParkingDetail(tenantID, carID string) string {
	return fmt.Sprintf("parking_deletail_%s_%s", tenantID, carID)
}

func HelmetLock(tenantID, carID string) string {
	return fmt.Sprintf("helmet_lock_%s_%s", tenantID, carID)
}

func HelmetReact(tenantID, carID string) string {
	return fmt.Sprintf("helmet_react%s_%s", tenantID, carID)
}

func HelmetAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("helmet_repair%s_%s_%d", tenantID, carID, orderID)
}

func PointAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("point_%s_%s_%d", tenantID, carID, orderID)
}

func DirectionAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("direction_%s_%s_%d", tenantID, carID, orderID)
}

func KickstandAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("kickstand_%s_%s_%d", tenantID, carID, orderID)
}

func CameraAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("camera_%s_%s_%d", tenantID, carID, orderID)
}

func CameraFailCount(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("camera_fail_count_%s_%s_%d", tenantID, carID, orderID)
}

func CameraErrorCount(tenantID, carID string) string {
	return fmt.Sprintf("camera_fail_count_%s_%s", tenantID, carID)
}

func LocationAudit(tenantID, carID string, orderID int64) string {
	return fmt.Sprintf("location_%s_%s_%d", tenantID, carID, orderID)
}

func BluetoothBeaconInfo(tenantID, imei string) string {
	return fmt.Sprintf("bluetooth_beacon_info_%s_%s", tenantID, imei)
}

func ResourceClickPerson(pin string, resourceID int64) string {
	return fmt.Sprintf("resource_click_person_%s_%d", pin, resourceID)
}

func AreaTagRefCache(fenceType string, fenceID int64) string {
	return fmt.Sprintf("fence_tag_ref_cache_%s_%d", fenceType, fenceID)
}

func FenceTagListCache(tenantID string) string {
	return fmt.Sprintf("fence_tag_list_cache_%s", tenantID)
}

func ProtocolConfigDefaultList() string { return "protocol_config_default_list" }

func ProtocolConfigByType(tenantID string, serviceID int64, typ string) string {
	return fmt.Sprintf("protocol_config_%s_%d_%s", tenantID, serviceID, typ)
}

func FenceRfidBind(tenantID string, fenceID int64) string {
	return fmt.Sprintf("fence_rfid_%s_%d", tenantID, fenceID)
}

func RfidFenceBind(tenantID, rfid string) string {
	return fmt.Sprintf("rfid_fence_%s_%s", tenantID, rfid)
}

// IsCacheMiss reports whether a Redis value should be treated as cache miss (Java semantics).
func IsCacheMiss(val string) bool {
	val = strings.TrimSpace(val)
	return val == "" || val == "null" || val == "[]"
}
