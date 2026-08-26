package service

import (
	"encoding/json"
	"sort"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/repository"
)

const (
	defaultNearbyLimit  = 20
	defaultNearbyRadius = 200.0
)

// GetDeviceList ports DeviceInfoServiceImpl.getDeviceList: resolve the imei set
// (carIdList wins), intersect with the service-area sets, read devices, apply
// the optional role filter, and project to DeviceListCo.
func GetDeviceList(tenantID string, qry dto.DeviceListQry) []dto.DeviceListCo {
	imeiList := qry.ImeiList
	if len(qry.CarIdList) > 0 {
		imeiList = repository.GetImeiListByCarId(tenantID, qry.CarIdList)
	}

	var reportTime int64
	if v := qry.ReportTime.Ptr(); v != nil {
		reportTime = *v
	}
	redisImei := repository.GetImeiListByService(tenantID, []int64(qry.ServiceIdList), reportTime)

	if len(imeiList) == 0 {
		imeiList = redisImei
	} else {
		imeiList = retainAll(imeiList, redisImei)
	}
	if len(imeiList) == 0 {
		return []dto.DeviceListCo{}
	}

	devices := repository.GetDeviceInfoList(tenantID, imeiList)

	// Role filter: a non-empty carInfos set restricts visible cars (empty/nil = all).
	if carInfos := client.UserCarInfos(qry.CommandContext); len(carInfos) > 0 {
		set := make(map[string]struct{}, len(carInfos))
		for _, c := range carInfos {
			set[c] = struct{}{}
		}
		filtered := devices[:0]
		for _, d := range devices {
			if carID, _ := d["carId"].(string); carID != "" {
				if _, ok := set[carID]; ok {
					filtered = append(filtered, d)
				}
			}
		}
		devices = filtered
	}

	out := make([]dto.DeviceListCo, 0, len(devices))
	for _, d := range devices {
		out = append(out, *mapDeviceListItem(d))
	}
	return out
}

// GetUseableEbikeLocation ports DeviceInfoServiceImpl.getUseableEbikeLocation:
// GEO-radius nearby cars (optionally across all services), filter to available
// online in-area bikes, sort by distance, and cap the count. Returns nil when
// no candidates (matches Java returning null).
func GetUseableEbikeLocation(tenantID string, qry dto.DeviceLocationQry) []dto.DeviceLocationCo {
	cc := qry.CommandContext
	canRideOther := client.FenceIzCanRideOtherService(qry.ServiceId.Int64(), cc)

	radius := defaultNearbyRadius
	if qry.Radius != nil {
		radius = *qry.Radius
	}
	if cfg := client.FenceHideCarConfig(qry.ServiceId.Int64(), cc); cfg != "" {
		var j struct {
			HideCarSwitch int     `json:"hideCarSwitch"`
			Distance      float64 `json:"distance,string"`
		}
		if json.Unmarshal([]byte(cfg), &j) == nil && j.HideCarSwitch == 1 {
			radius = j.Distance
		}
	}

	locations := map[string]float64{}
	if canRideOther {
		for _, sid := range client.FenceAllServiceIds(cc) {
			for imei, d := range repository.GetLocations(tenantID, sid, qry.Lat, qry.Lng, radius) {
				locations[imei] = d
			}
		}
	} else {
		locations = repository.GetLocations(tenantID, qry.ServiceId.Int64(), qry.Lat, qry.Lng, radius)
	}
	if len(locations) == 0 {
		return nil
	}

	imeis := make([]string, 0, len(locations))
	for imei := range locations {
		imeis = append(imeis, imei)
	}
	devices := repository.GetDeviceInfoList(tenantID, imeis)

	limit := defaultNearbyLimit
	if qry.Limit != nil {
		limit = *qry.Limit
	}
	return filterNearby(devices, locations, limit)
}

// filterNearby is the pure transform: keep available online in-area bikes, set
// distance, sort ascending, and cap to limit. Mirrors the Java stream pipeline.
func filterNearby(devices []map[string]interface{}, locations map[string]float64, limit int) []dto.DeviceLocationCo {
	out := make([]dto.DeviceLocationCo, 0, len(devices))
	for _, d := range devices {
		if rs, ok := asInt(d["ridingState"]); !ok || rs != 1 {
			continue
		}
		if _, hasNoRide := asInt64(d["noRideParkId"]); hasNoRide {
			continue
		}
		if on, ok := asInt(d["isOnline"]); !ok || on != 1 {
			continue
		}
		if oos, ok := asInt(d["isOutofServAera"]); ok && oos == 1 {
			continue
		}
		imei, _ := d["imei"].(string)
		dist := locations[imei]
		out = append(out, dto.DeviceLocationCo{
			CarId:       strPtr(d, "carId"),
			Imei:        strPtr(d, "imei"),
			Lat:         f64PtrDefault(d, "lat", 0),
			Lng:         f64PtrDefault(d, "lng", 0),
			Distance:    &dist,
			RestBattery: restBattery(d),
			RestMileage: intPtr(d, "restMileage"),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return derefF(out[i].Distance) < derefF(out[j].Distance)
	})

	if limit >= 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// retainAll keeps elements of base that are present in keep (preserving order),
// mirroring Java List.retainAll(set).
func retainAll(base, keep []string) []string {
	set := make(map[string]struct{}, len(keep))
	for _, k := range keep {
		set[k] = struct{}{}
	}
	out := base[:0]
	for _, b := range base {
		if _, ok := set[b]; ok {
			out = append(out, b)
		}
	}
	return out
}

func derefF(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
