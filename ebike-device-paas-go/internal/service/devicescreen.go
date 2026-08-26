package service

import (
	"sort"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

const hourMs = int64(60 * 60 * 1000)

// onShelf reports whether the device is off-shelf (operationState contains 1).
func onShelf(d map[string]interface{}) bool {
	return containsInt(intList(d, "operationState"), 1)
}

// ridingCounts tallies the 5 riding states, mirroring the shared switch.
func ridingCounts(d map[string]interface{}, canRent, booking, riding, parking, operation *int) {
	rs, _ := asInt(d["ridingState"])
	switch rs {
	case 1:
		*canRent++
	case 2:
		*riding++
	case 3:
		*parking++
	case 4:
		*booking++
	case 5:
		*operation++
	}
}

// idleMs returns now-lockTime when lockTime>unlockTime (idle), else -1.
func idleMs(d map[string]interface{}, now int64) int64 {
	lock, _ := asInt64(d["lockTime"])
	unlock, _ := asInt64(d["unlockTime"])
	if lock > unlock {
		return now - lock
	}
	return -1
}

func restBatteryOrZero(d map[string]interface{}) int {
	if p := restBattery(d); p != nil {
		return *p
	}
	return 0
}

func gpsList(devices []map[string]interface{}, ridingState *int) []dto.DeviceGpsCo {
	out := make([]dto.DeviceGpsCo, 0, len(devices))
	for _, d := range devices {
		if ridingState != nil {
			if rs, ok := asInt(d["ridingState"]); !ok || rs != *ridingState {
				continue
			}
		}
		out = append(out, dto.DeviceGpsCo{
			Imei:  strPtr(d, "imei"),
			CarId: strPtr(d, "carId"),
			Lng:   f64PtrDefault(d, "lng", 0),
			Lat:   f64PtrDefault(d, "lat", 0),
		})
	}
	return out
}

// QueryDeviceScreen ports DeviceInfoServiceImpl.queryDeviceScreen (V1 dashboard).
func QueryDeviceScreen(tenantID string, qry dto.DeviceScreenQry) dto.DeviceScreenCo {
	all := repository.GetDeviceInfoListByServiceList(tenantID, []int64(qry.ServiceIdList))
	online := make([]map[string]interface{}, 0, len(all))
	for _, d := range all {
		if !onShelf(d) {
			online = append(online, d)
		}
	}

	var co dto.DeviceScreenCo
	now := time.Now().UnixMilli()
	for _, d := range online {
		ridingCounts(d, &co.CanRent, &co.Booking, &co.Riding, &co.Parking, &co.Operation)
		if diff := idleMs(d, now); diff >= 0 {
			switch {
			case diff > 1*hourMs && diff <= 3*hourMs:
				co.FreeTimeOneToThree++
			case diff > 3*hourMs && diff <= 6*hourMs:
				co.FreeTimeThreeToSix++
			case diff > 6*hourMs && diff <= 12*hourMs:
				co.FreeTimeSixToTwelve++
			case diff > 12*hourMs && diff <= 24*hourMs:
				co.FreeTimeHalfOrOneDay++
			case diff > 24*hourMs && diff <= 48*hourMs:
				co.FreeTimeOneOrTowDay++
			case diff > 48*hourMs:
				co.FreeTimeTowDayMore++
			}
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
	co.Gps = gpsList(online, qry.RidingState)
	return co
}

// QueryDeviceScreenV2 ports DeviceInfoServiceImpl.queryDeviceScreenV2.
func QueryDeviceScreenV2(tenantID string, qry dto.DeviceScreenQry) dto.DeviceScreenCoV2 {
	all := repository.GetDeviceInfoListByServiceList(tenantID, []int64(qry.ServiceIdList))
	online := make([]map[string]interface{}, 0, len(all))
	for _, d := range all {
		if !onShelf(d) {
			online = append(online, d)
		}
	}

	co := dto.DeviceScreenCoV2{
		Total:   len(all),
		Offline: len(all) - len(online),
	}
	now := time.Now().UnixMilli()
	for _, d := range online {
		ridingCounts(d, &co.CanRent, &co.Booking, &co.Riding, &co.Parking, &co.Operation)
		if diff := idleMs(d, now); diff >= 0 {
			switch {
			case diff >= 0 && diff <= 3*hourMs:
				co.FreeTimeZeroToThree++
			case diff > 3*hourMs && diff <= 6*hourMs:
				co.FreeTimeThreeToSix++
			case diff > 6*hourMs && diff <= 12*hourMs:
				co.FreeTimeSixToTwelve++
			case diff > 12*hourMs && diff <= 24*hourMs:
				co.FreeTimeHalfOrOneDay++
			case diff > 24*hourMs && diff <= 48*hourMs:
				co.FreeTimeOneOrTowDay++
			case diff > 48*hourMs && diff <= 72*hourMs:
				co.FreeTimeTowDayOrThreeDay++
			case diff > 72*hourMs:
				co.FreeTimeThreeDayMore++
			}
		}
		rb := restBatteryOrZero(d)
		switch {
		case rb >= 0 && rb <= 20:
			co.VoltageZeroToTwenty++
		case rb > 20 && rb <= 40:
			co.VoltageTwentyToForty++
		case rb > 40 && rb <= 60:
			co.VoltageFortyToSixty++
		case rb > 60 && rb <= 80:
			co.VoltageSixtyToEighty++
		case rb > 80:
			co.VoltageEightyMore++
		}
	}
	co.Gps = gpsList(online, qry.RidingState)
	return co
}

// CarCount ports DeviceInfoServiceImpl.carCount: per-service totals, sorted by
// total descending (mirroring the Comparable contract).
func CarCount(tenantID string, qry dto.DeviceScreenQry) []dto.DeviceScreenCoV2 {
	all := repository.GetDeviceInfoListByServiceList(tenantID, []int64(qry.ServiceIdList))
	totals := map[int64]int{}
	for _, d := range all {
		if sid, ok := asInt64(d["serviceId"]); ok {
			totals[sid]++
		}
	}
	out := make([]dto.DeviceScreenCoV2, 0, len(qry.ServiceIdList))
	for _, sid := range qry.ServiceIdList {
		s := sid
		out = append(out, dto.DeviceScreenCoV2{ServiceId: &s, Total: totals[sid]})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Total > out[j].Total })
	return out
}
