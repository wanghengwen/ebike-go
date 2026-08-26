package client

import (
	"encoding/json"
	"log"

	"ebike-device-paas-go/internal/pkg/config"
)

// These cross-service reads degrade gracefully: when the upstream URL is not
// configured or the call fails, they return the same defaults the Java logic
// falls back to, so device/list and eBikeLocation still serve the dominant
// (single-service, no role-restriction) path.

// UserCarInfos mirrors EbikeManageRpcImpl.userCarInfos (management user/carInfos):
// the set of carIds the caller may see. Empty/nil means "all cars" (no filter).
func UserCarInfos(commandContext interface{}) []string {
	base := config.GlobalConfig.Xyy.ManagementURL
	if base == "" {
		return nil
	}
	env, err := postJSON(base+"/user/carInfos", map[string]interface{}{"commandContext": commandContext})
	if err != nil || !env.Success {
		log.Printf("[client] user/carInfos failed (treating as no-filter): %v", err)
		return nil
	}
	var carIDs []string
	if len(env.Data) > 0 {
		_ = json.Unmarshal(env.Data, &carIDs)
	}
	return carIDs
}

// FenceIzCanRideOtherService mirrors FenceRpcImpl.getIzCanRideOtherService
// (fence /config/backcar/getConfigByServiceId -> izUseOtherParking). Default false.
func FenceIzCanRideOtherService(serviceID int64, commandContext interface{}) bool {
	base := config.GlobalConfig.Xyy.FenceURL
	if base == "" {
		return false
	}
	env, err := postJSON(base+"/config/backcar/getConfigByServiceId", map[string]interface{}{
		"id":             serviceID,
		"commandContext": commandContext,
	})
	if err != nil || !env.Success || len(env.Data) == 0 {
		return false
	}
	var co struct {
		IzUseOtherParking *bool `json:"izUseOtherParking"`
	}
	if json.Unmarshal(env.Data, &co) != nil || co.IzUseOtherParking == nil {
		return false
	}
	return *co.IzUseOtherParking
}

// FenceHideCarConfig mirrors FenceRpcImpl.getHideCarConfig
// (fence /config/usecar/getConfigByServiceId -> hideCarConfig). Default "".
func FenceHideCarConfig(serviceID int64, commandContext interface{}) string {
	base := config.GlobalConfig.Xyy.FenceURL
	if base == "" {
		return ""
	}
	env, err := postJSON(base+"/config/usecar/getConfigByServiceId", map[string]interface{}{
		"serviceId":      serviceID,
		"commandContext": commandContext,
	})
	if err != nil || !env.Success || len(env.Data) == 0 {
		return ""
	}
	var co struct {
		HideCarConfig string `json:"hideCarConfig"`
	}
	if json.Unmarshal(env.Data, &co) != nil {
		return ""
	}
	return co.HideCarConfig
}

// ServiceTenant pairs a service area id with its owning tenant (fence getAllService).
type ServiceTenant struct {
	TenantID  string
	ServiceID int64
}

// FenceAllServiceTenant mirrors DeviceInfoQueryImpl.getAllServiceTenant
// (fence /serviceArea/getAllService with tenant "1" -> [{id, tenantId}]).
// Cross-tenant; default nil when fence is not configured.
func FenceAllServiceTenant(commandContext interface{}) []ServiceTenant {
	base := config.GlobalConfig.Xyy.FenceURL
	if base == "" {
		return nil
	}
	cc := commandContext
	if cc == nil {
		cc = map[string]interface{}{"tenantId": "1"}
	}
	env, err := postJSON(base+"/serviceArea/getAllService", map[string]interface{}{"commandContext": cc})
	if err != nil || !env.Success || len(env.Data) == 0 {
		log.Printf("[client] serviceArea/getAllService failed: %v", err)
		return nil
	}
	var areas []struct {
		ID       *int64 `json:"id"`
		TenantID string `json:"tenantId"`
	}
	if json.Unmarshal(env.Data, &areas) != nil {
		return nil
	}
	out := make([]ServiceTenant, 0, len(areas))
	for _, a := range areas {
		if a.ID != nil {
			out = append(out, ServiceTenant{TenantID: a.TenantID, ServiceID: *a.ID})
		}
	}
	return out
}

// FenceServicePoint mirrors DeviceInfoQueryImpl.getServicePoint
// (fence /serviceArea/getById -> pointList JSON polygon). Default "".
func FenceServicePoint(serviceID int64, commandContext interface{}) string {
	base := config.GlobalConfig.Xyy.FenceURL
	if base == "" {
		return ""
	}
	env, err := postJSON(base+"/serviceArea/getById", map[string]interface{}{
		"id":             serviceID,
		"commandContext": commandContext,
	})
	if err != nil || !env.Success || len(env.Data) == 0 {
		log.Printf("[client] serviceArea/getById failed: %v", err)
		return ""
	}
	var co struct {
		PointList string `json:"pointList"`
	}
	if json.Unmarshal(env.Data, &co) != nil {
		return ""
	}
	return co.PointList
}

// FenceAllServiceIds mirrors FenceRpcImpl.getAllServiceIdByTenantTd
// (fence /serviceArea/getList -> [].id). Default nil.
func FenceAllServiceIds(commandContext interface{}) []int64 {
	base := config.GlobalConfig.Xyy.FenceURL
	if base == "" {
		return nil
	}
	env, err := postJSON(base+"/serviceArea/getList", map[string]interface{}{"commandContext": commandContext})
	if err != nil || !env.Success || len(env.Data) == 0 {
		return nil
	}
	var areas []struct {
		ID *int64 `json:"id"`
	}
	if json.Unmarshal(env.Data, &areas) != nil {
		return nil
	}
	out := make([]int64, 0, len(areas))
	for _, a := range areas {
		if a.ID != nil {
			out = append(out, *a.ID)
		}
	}
	return out
}
