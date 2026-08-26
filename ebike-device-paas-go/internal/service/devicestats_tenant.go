package service

import (
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/repository"
)

// OnlineNumByTenantId ports DeviceInfoServiceImpl.onlineNumByTenantId: count the
// current tenant's on-shelf (operationState not containing 1) devices across all
// of its service areas. Requires fence (service-area list); 0 when unconfigured.
func OnlineNumByTenantId(tenantID string, cc *dto.CommandContext) int {
	sids := client.FenceAllServiceIds(cc)
	var imeis []string
	for _, sid := range sids {
		imeis = append(imeis, repository.GetImeiByServiceId(tenantID, sid)...)
	}
	devices := repository.GetDeviceInfoList(tenantID, imeis)
	count := 0
	for _, d := range devices {
		if !onShelf(d) {
			count++
		}
	}
	return count
}

// collectTenantImei gathers deduped (tenant, imei) pairs over all service areas,
// mirroring getTenantServiceImei's TreeSet de-dup by tenantId+imei.
func collectTenantImei(serviceTenants []client.ServiceTenant) []repository.TenantImei {
	seen := map[string]struct{}{}
	var items []repository.TenantImei
	for _, st := range serviceTenants {
		for _, imei := range repository.GetImeiByServiceId(st.TenantID, st.ServiceID) {
			key := st.TenantID + "|" + imei
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, repository.TenantImei{TenantID: st.TenantID, Imei: imei})
		}
	}
	return items
}

// CarStatistics ports DeviceInfoServiceImpl.carStatistics: cross-tenant per-service
// state aggregation seeded by all service areas. Grouping uses the decoded device
// serviceId. Result follows HashMap order -> shadow compares unordered ("data").
func CarStatistics(cc *dto.CommandContext) []dto.CarServiceStatisticsCo {
	serviceTenants := client.FenceAllServiceTenant(cc)

	stats := map[int64]*dto.CarServiceStatisticsCo{}
	var order []int64
	ensure := func(sid int64, tenantID string) *dto.CarServiceStatisticsCo {
		co := stats[sid]
		if co == nil {
			s := sid
			tid := tenantID
			co = &dto.CarServiceStatisticsCo{TenantId: &tid, ServiceId: &s}
			stats[sid] = co
			order = append(order, sid)
		}
		return co
	}
	for _, st := range serviceTenants {
		ensure(st.ServiceID, st.TenantID)
	}

	devices := repository.GetRackDevicesByTenantImei(collectTenantImei(serviceTenants))
	now := time.Now().UnixMilli()
	for _, d := range devices {
		sid, ok := asInt64(d["serviceId"])
		if !ok {
			continue
		}
		tid := ""
		if p := strPtr(d, "tenantId"); p != nil {
			tid = *p
		}
		co := ensure(sid, tid)
		ridingCounts(d, &co.CanRent, &co.Booking, &co.Riding, &co.Parking, &co.Operation)
		if containsInt(intList(d, "operationState"), 4) {
			co.LowBattery++
		}
		// carStatistics uses lockTime != null (no lock>unlock guard).
		if lock, has := asInt64(d["lockTime"]); has {
			svcFreeTimeAdd(co, freeTimeBucketV1(now-lock))
		}
		rb := restBatteryOrZero(d)
		switch {
		case rb == 0:
			co.VoltageZero++
		case rb > 0 && rb <= 20:
			co.VoltageZeroToTwenty++
		case rb > 20 && rb <= 35:
			co.VoltageTwentyToThirtyFive++
		case rb > 35:
			co.VoltageThirtyFiveMore++
		}
	}

	out := make([]dto.CarServiceStatisticsCo, 0, len(order))
	for _, sid := range order {
		out = append(out, *stats[sid])
	}
	return out
}

// CarParkingStatistics ports DeviceInfoServiceImpl.carParkingStatistics:
// cross-tenant per-parking aggregation over all service areas. Only devices with
// a non-null forParkId are counted; entries are keyed by parkingId and created
// lazily from the first device seen. Result follows HashMap order -> shadow
// compares unordered ("data"). freeTime fields stay 0 here (idle derived = 0).
func CarParkingStatistics(cc *dto.CommandContext) []dto.CarParkingStatisticsCo {
	serviceTenants := client.FenceAllServiceTenant(cc)
	devices := repository.GetRackDevicesByTenantImei(collectTenantImei(serviceTenants))

	parks := map[int64]*dto.CarParkingStatisticsCo{}
	var order []int64
	for _, d := range devices {
		parkID, hasPark := asInt64(d["forParkId"])
		if !hasPark {
			continue
		}
		co := parks[parkID]
		if co == nil {
			pid := parkID
			co = &dto.CarParkingStatisticsCo{
				TenantId:  strPtr(d, "tenantId"),
				ServiceId: i64Ptr(d, "serviceId"),
				ParkingId: &pid,
			}
			parks[parkID] = co
			order = append(order, parkID)
		}
		switch rs, _ := asInt(d["ridingState"]); rs {
		case 1:
			co.CanRent++
		case 4:
			co.Booking++
		case 5:
			co.Operation++
		}
		if len(intList(d, "alarmState")) > 0 {
			co.Alarm++
		}
		if containsInt(intList(d, "operationState"), 5) {
			co.Fault++
		}
	}

	out := make([]dto.CarParkingStatisticsCo, 0, len(order))
	for _, pid := range order {
		out = append(out, *parks[pid])
	}
	return out
}

// GetRackCarNumAll ports DeviceInfoServiceImpl.getRackCarNumAll: cross-tenant
// per-service on-shelf counts seeded by all service areas. Result follows
// HashMap order -> shadow compares unordered ("data").
func GetRackCarNumAll(cc *dto.CommandContext) []dto.RackCarNumCo {
	serviceTenants := client.FenceAllServiceTenant(cc)

	counts := map[int64]*dto.RackCarNumCo{}
	var order []int64
	ensure := func(sid int64, tenantID string) *dto.RackCarNumCo {
		co := counts[sid]
		if co == nil {
			s := sid
			tid := tenantID
			co = &dto.RackCarNumCo{TenantId: &tid, ServiceId: &s, Count: 0}
			counts[sid] = co
			order = append(order, sid)
		}
		return co
	}
	for _, st := range serviceTenants {
		ensure(st.ServiceID, st.TenantID)
	}

	devices := repository.GetRackDevicesByTenantImei(collectTenantImei(serviceTenants))
	for _, d := range devices {
		sid, ok := asInt64(d["serviceId"])
		if !ok {
			continue
		}
		tid := ""
		if p := strPtr(d, "tenantId"); p != nil {
			tid = *p
		}
		ensure(sid, tid).Count++
	}

	out := make([]dto.RackCarNumCo, 0, len(order))
	for _, sid := range order {
		out = append(out, *counts[sid])
	}
	return out
}
