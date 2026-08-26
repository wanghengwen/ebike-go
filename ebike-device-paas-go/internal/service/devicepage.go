package service

import (
	"sort"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/repository"
)

// GetDevicePage ports DeviceInfoServiceImpl.getDevicePage (platform-end pagination).
//
// The amap reverse-geocode (address/scanAddress) IS ported: after the page list is
// built it calls map-service /map/batchRegeo (see fillPageAddresses). The page
// cache is not ported — every page is computed from the real path (the Java cache
// is only an optimisation, so results match). Filtering / sort (reportTime desc) /
// pagination / carHelmetState match.
func GetDevicePage(tenantID string, qry dto.DevicePageQry) dto.DevicePageDTO {
	devices := repository.GetDeviceByServiceId(tenantID, qry.ServiceId.Int64())
	filtered := filterDevices(devices, qry)
	// Platform page always carries carHelmetState: Java sets it on both the
	// page-1 real path and the page>=2 cache path (getDevicePage).
	page := paginateDevicePage(filtered, qry.PageNum, qry.PageSize, true)
	fillPageAddresses(tenantID, traceOf(qry.CommandContext), page.List)
	return page
}

// GetDevicePageBus ports DeviceInfoServiceImpl.getDevicePageBus (merchant-end).
// Java getDevicePageBus does NOT call amap, so address/scanAddress stay null (no cache).
func GetDevicePageBus(tenantID string, qry dto.DevicePageBusQry) dto.DevicePageDTO {
	imeis := repository.GetImeiByServiceId(tenantID, qry.ServiceId.Int64())
	devices := repository.GetDeviceInfoList(tenantID, imeis)

	// Role filter: a non-empty carInfos set restricts visible cars (empty/nil = all).
	if carInfos := client.UserCarInfos(qry.CommandContext); len(carInfos) > 0 {
		set := make(map[string]struct{}, len(carInfos))
		for _, c := range carInfos {
			set[c] = struct{}{}
		}
		kept := devices[:0]
		for _, d := range devices {
			if carID, _ := d["carId"].(string); carID != "" {
				if _, ok := set[carID]; ok {
					kept = append(kept, d)
				}
			}
		}
		devices = kept
	}

	filtered := businessFilterDevices(devices, qry)
	// Java getDevicePageBus only sets carHelmetState on the page-1 real path; the
	// page>=2 cache path does a plain DevicePageCo copy, leaving carHelmetState
	// null. We mirror that: page-1 carries it, page>=2 leaves it nil. (The Java
	// behaviour is cache-state dependent; this matches the warm-cache prod path.)
	return paginateDevicePage(filtered, qry.PageNum, qry.PageSize, qry.PageNum == 1)
}

// paginateDevicePage sorts by reportTime desc, slices the requested page, and maps
// to DevicePageCo (with carHelmetState; empty address). count is the full size.
func paginateDevicePage(devices []map[string]interface{}, pageNum, pageSize int, withHelmetState bool) dto.DevicePageDTO {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	sort.SliceStable(devices, func(i, j int) bool {
		ri, _ := asInt64(devices[i]["reportTime"])
		rj, _ := asInt64(devices[j]["reportTime"])
		return ri > rj
	})

	result := dto.DevicePageDTO{
		Count:       int64(len(devices)),
		PageNum:     pageNum,
		PageSize:    pageSize,
		SearchCount: true,
		List:        []dto.DevicePageCo{},
	}

	start := (pageNum - 1) * pageSize
	if start < 0 || start >= len(devices) {
		return result
	}
	end := start + pageSize
	if end > len(devices) {
		end = len(devices)
	}
	for _, d := range devices[start:end] {
		result.List = append(result.List, mapDevicePageCoFull(d, withHelmetState))
	}
	return result
}

// mapDevicePageCoFull builds a DevicePageCo, optionally setting carHelmetState
// (Java sets it on the real path / platform cache path, but not the merchant
// page>=2 cache path). address / scanAddress are left nil here; getDevicePage
// fills them via fillPageAddresses, getDevicePageBus leaves them null.
func mapDevicePageCoFull(d map[string]interface{}, withHelmetState bool) dto.DevicePageCo {
	co := mapDevicePageCo(d)
	if withHelmetState {
		chs := carHelmetState(d)
		co.CarHelmetState = &chs
	}
	return co
}

// fillPageAddresses ports the amap reverse-geocode block in getDevicePage: build
// the locations list (every car's lng/lat in order, then scanLng/scanLat for cars
// that have both), call map-service, then assign address/scanAddress positionally
// (scan addresses live at offset len(list)+scanPoint). On RPC failure (empty
// result) every car's address/scanAddress is set to "" — mirroring Java exactly.
func fillPageAddresses(tenantID, traceID string, list []dto.DevicePageCo) {
	if len(list) == 0 {
		return
	}
	locations := make([]client.GeoPoint, 0, len(list)*2)
	for _, co := range list {
		locations = append(locations, client.GeoPoint{Longitude: f64OrZero(co.Lng), Latitude: f64OrZero(co.Lat)})
	}
	for _, co := range list {
		if co.ScanLng != nil && co.ScanLat != nil {
			locations = append(locations, client.GeoPoint{Longitude: *co.ScanLng, Latitude: *co.ScanLat})
		}
	}
	addressList := client.MapBatchRegeo(tenantID, traceID, locations)
	scanPoint := 0
	for i := range list {
		if len(addressList) == 0 {
			empty := ""
			list[i].Address = &empty
			list[i].ScanAddress = &empty
			continue
		}
		if i < len(addressList) {
			addr := addressList[i]
			list[i].Address = &addr
		}
		if list[i].ScanLng != nil && list[i].ScanLat != nil {
			if idx := scanPoint + len(list); idx < len(addressList) {
				sa := addressList[idx]
				list[i].ScanAddress = &sa
			}
			scanPoint++
		}
	}
}

// f64OrZero dereferences a *float64, defaulting to 0 (matches DeviceInfoDO.getLat/
// getLng null->0.0 already applied upstream).
func f64OrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// businessFilterDevices ports the list-based getFilterDevice (merchant-end):
// minCarId/maxCarId range, izFilterRiding, union/intersection state filters, then
// the shared noOrderTime/orderTime/restBattery/batteryIds/carTagTypeIds chain.
func businessFilterDevices(devices []map[string]interface{}, q dto.DevicePageBusQry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(devices))
	for _, d := range devices {
		carID, _ := d["carId"].(string)
		if q.MinCarId != "" && carID < q.MinCarId {
			continue
		}
		if q.MaxCarId != "" && carID > q.MaxCarId {
			continue
		}
		if q.IzFilterRiding != nil && *q.IzFilterRiding {
			if rs, _ := asInt(d["ridingState"]); rs == 2 || rs == 3 {
				continue
			}
		}

		ops := intList(d, "operationState")
		al := intList(d, "alarmState")
		rs, _ := asInt(d["ridingState"])

		noStates := len(q.RidingStates) == 0 && len(q.OperationStates) == 0 && len(q.AlarmStates) == 0
		if noStates {
			if containsInt(ops, 1) {
				continue
			}
		} else if !q.IzStateAnd {
			ridingFilter := len(q.RidingStates) > 0 && containsInt(q.RidingStates, rs)
			opeFilter := len(q.OperationStates) > 0 && intersectsInt(q.OperationStates, ops)
			alarmFilter := len(q.AlarmStates) > 0 && intersectsInt(q.AlarmStates, al)
			stateFilter := ridingFilter || opeFilter || alarmFilter
			if len(q.OperationStates) > 0 && containsInt(q.OperationStates, 1) {
				if !stateFilter {
					continue
				}
			} else {
				if containsInt(ops, 1) || !stateFilter {
					continue
				}
			}
		} else {
			ridingFilter := len(q.RidingStates) == 0 || (len(q.RidingStates) == 1 && containsInt(q.RidingStates, rs))
			opeFilter := len(q.OperationStates) == 0 || containsAllInt(ops, q.OperationStates)
			alarmFilter := len(q.AlarmStates) == 0 || containsAllInt(al, q.AlarmStates)
			stateFilter := ridingFilter && opeFilter && alarmFilter
			if len(q.OperationStates) > 0 && containsInt(q.OperationStates, 1) {
				if !stateFilter {
					continue
				}
			} else {
				if containsInt(ops, 1) || !stateFilter {
					continue
				}
			}
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
		if len(q.BatteryIds) > 0 {
			bid, ok := asInt64(d["batteryId"])
			if !ok || !containsInt64(q.BatteryIds, bid) {
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

// containsAllInt reports whether every element of want is present in have.
func containsAllInt(have, want []int) bool {
	for _, w := range want {
		if !containsInt(have, w) {
			return false
		}
	}
	return true
}

func containsInt64(list []int64, v int64) bool {
	for _, e := range list {
		if e == v {
			return true
		}
	}
	return false
}
