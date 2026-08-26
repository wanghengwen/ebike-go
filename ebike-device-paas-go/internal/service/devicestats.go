package service

import (
	"sort"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// freeTimeBucketV1 classifies an idle duration (ms) into the V1 buckets used by
// carStatistics / carStatisticsByService: 0=(1,3]h 1=(3,6] 2=(6,12] 3=(12,24]
// 4=(24,48] 5=(48,)h, -1 = none (<=1h).
func freeTimeBucketV1(diff int64) int {
	switch {
	case diff > 1*hourMs && diff <= 3*hourMs:
		return 0
	case diff > 3*hourMs && diff <= 6*hourMs:
		return 1
	case diff > 6*hourMs && diff <= 12*hourMs:
		return 2
	case diff > 12*hourMs && diff <= 24*hourMs:
		return 3
	case diff > 24*hourMs && diff <= 48*hourMs:
		return 4
	case diff > 48*hourMs:
		return 5
	default:
		return -1
	}
}

// QueryServiceStatisticsList ports DeviceInfoServiceImpl.queryServiceStatisticsList:
// per service (in input order), tally riding / operation / alarm states over the
// on-shelf devices.
func QueryServiceStatisticsList(tenantID string, ids []int64) []dto.CarStateServiceStatisticsCo {
	out := make([]dto.CarStateServiceStatisticsCo, 0, len(ids))
	for _, sid := range ids {
		s := sid
		all := repository.GetDeviceByServiceId(tenantID, sid)
		co := dto.CarStateServiceStatisticsCo{ServiceId: &s}
		for _, d := range all {
			if onShelf(d) {
				continue
			}
			co.Operating++
			ridingCounts(d, &co.CanRent, &co.Booking, &co.Riding, &co.Parking, &co.Operation)
			op := intList(d, "operationState")
			if containsInt(op, 2) {
				co.MoveCar++
			}
			if containsInt(op, 3) {
				co.ChangeBattery++
			}
			if containsInt(op, 4) {
				co.LowBattery++
			}
			if containsInt(op, 5) {
				co.Fixing++
			}
			if containsInt(op, 6) {
				co.DragBack++
			}
			if containsInt(intList(d, "alarmState"), 5) {
				co.OutParkingZone++
			}
		}
		co.UnSheleves = len(all) - co.Operating
		out = append(out, co)
	}
	return out
}

// GetCarNumByServiceId ports DeviceInfoServiceImpl.getCarNumByServiceId: the
// service-area on-shelf devices grouped by serviceId -> count. Java returns the
// groups in HashMap order; we sort by serviceId for determinism.
func GetCarNumByServiceId(tenantID string, serviceIDs []int64) []dto.ServiceCarNumCo {
	imeis := repository.GetImeiListByService(tenantID, serviceIDs, 0)
	devices := repository.GetDeviceInfoList(tenantID, imeis)
	counts := map[int64]int{}
	for _, d := range devices {
		if onShelf(d) {
			continue
		}
		if sid, ok := asInt64(d["serviceId"]); ok {
			counts[sid]++
		}
	}
	out := make([]dto.ServiceCarNumCo, 0, len(counts))
	for sid, cnt := range counts {
		s := sid
		out = append(out, dto.ServiceCarNumCo{ServiceId: &s, Count: cnt})
	}
	sort.SliceStable(out, func(i, j int) bool { return *out[i].ServiceId < *out[j].ServiceId })
	return out
}

func svcFreeTimeAdd(co *dto.CarServiceStatisticsCo, bucket int) {
	switch bucket {
	case 0:
		co.FreeTimeOneToThree++
	case 1:
		co.FreeTimeThreeToSix++
	case 2:
		co.FreeTimeSixToTwelve++
	case 3:
		co.FreeTimeHalfOrOneDay++
	case 4:
		co.FreeTimeOneOrTowDay++
	case 5:
		co.FreeTimeTowDayMore++
	}
}

func parkFreeTimeAdd(co *dto.CarParkingStatisticsCo, bucket int) {
	switch bucket {
	case 0:
		co.FreeTimeOneToThree++
	case 1:
		co.FreeTimeThreeToSix++
	case 2:
		co.FreeTimeSixToTwelve++
	case 3:
		co.FreeTimeHalfOrOneDay++
	case 4:
		co.FreeTimeOneOrTowDay++
	case 5:
		co.FreeTimeTowDayMore++
	}
}

// CarStatisticsByService ports DeviceInfoServiceImpl.carStatisticsByService:
// over the on-shelf devices of each service, accumulate per-service stats (keyed
// by serviceId) and per-parking stats (keyed by forParkId; null parking dropped).
// The two result lists follow Java HashMap order, so shadow compares them
// order-insensitively (serviceStatistics / parkingStatistics unordered keys).
func CarStatisticsByService(tenantID string, serviceIDs []int64) dto.CarStatisticsByServiceCo {
	svcMap := map[int64]*dto.CarServiceStatisticsCo{}
	parkMap := map[int64]*dto.CarParkingStatisticsCo{}
	var svcOrder, parkOrder []int64

	for _, sid := range serviceIDs {
		devices := repository.GetRackDeviceByServiceId(tenantID, sid)
		for _, d := range devices {
			devServiceID, hasService := asInt64(d["serviceId"])
			if !hasService {
				continue
			}
			tid := strPtr(d, "tenantId")

			svc := svcMap[devServiceID]
			if svc == nil {
				ds := devServiceID
				svc = &dto.CarServiceStatisticsCo{TenantId: tid, ServiceId: &ds}
				svcMap[devServiceID] = svc
				svcOrder = append(svcOrder, devServiceID)
			}

			parkID, hasPark := asInt64(d["forParkId"])
			var park *dto.CarParkingStatisticsCo
			if hasPark {
				park = parkMap[parkID]
				if park == nil {
					pid := parkID
					ds := devServiceID
					park = &dto.CarParkingStatisticsCo{TenantId: tid, ServiceId: &ds, ParkingId: &pid}
					parkMap[parkID] = park
					parkOrder = append(parkOrder, parkID)
				}
			}

			rs, _ := asInt(d["ridingState"])
			switch rs {
			case 1:
				svc.CanRent++
				if park != nil {
					park.CanRent++
				}
			case 2:
				svc.Riding++
			case 3:
				svc.Parking++
			case 4:
				svc.Booking++
				if park != nil {
					park.Booking++
				}
			case 5:
				svc.Operation++
				if park != nil {
					park.Operation++
				}
			}

			op := intList(d, "operationState")
			if containsInt(op, 4) {
				svc.LowBattery++
			}
			if park != nil {
				if len(intList(d, "alarmState")) > 0 {
					park.Alarm++
				}
				if containsInt(op, 5) {
					park.Fault++
				}
			}

			lock, _ := asInt64(d["lockTime"])
			unlock, _ := asInt64(d["unlockTime"])
			if lock > unlock {
				bucket := freeTimeBucketV1(time.Now().UnixMilli() - lock)
				svcFreeTimeAdd(svc, bucket)
				if park != nil {
					parkFreeTimeAdd(park, bucket)
				}
			}

			rb := restBatteryOrZero(d)
			switch {
			case rb == 0:
				svc.VoltageZero++
			case rb > 0 && rb <= 20:
				svc.VoltageZeroToTwenty++
			case rb > 20 && rb <= 35:
				svc.VoltageTwentyToThirtyFive++
			case rb > 35:
				svc.VoltageThirtyFiveMore++
			}
		}
	}

	res := dto.CarStatisticsByServiceCo{
		ServiceStatistics: make([]dto.CarServiceStatisticsCo, 0, len(svcOrder)),
		ParkingStatistics: make([]dto.CarParkingStatisticsCo, 0, len(parkOrder)),
	}
	for _, sid := range svcOrder {
		res.ServiceStatistics = append(res.ServiceStatistics, *svcMap[sid])
	}
	for _, pid := range parkOrder {
		p := parkMap[pid]
		p.Idle = p.FreeTimeOneOrTowDay + p.FreeTimeTowDayMore
		res.ParkingStatistics = append(res.ParkingStatistics, *p)
	}
	return res
}
