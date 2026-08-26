package service

import (
	"strconv"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// QueryDeviceOpeMap ports DeviceInfoServiceImpl.queryDeviceOpeMap: an operations
// dashboard aggregate over the service list (riding / operation / alarm / battery
// tallies + battery-scheme map), plus filtered GPS markers.
func QueryDeviceOpeMap(tenantID string, qry dto.DeviceOpeMapQry) dto.DeviceOpeMapCo {
	deviceInfoList := repository.GetDeviceInfoListByServiceList(tenantID, []int64(qry.ServiceIdList))

	online := make([]map[string]interface{}, 0, len(deviceInfoList))
	for _, d := range deviceInfoList {
		if !onShelf(d) {
			online = append(online, d)
		}
	}

	co := dto.DeviceOpeMapCo{
		Operating:  len(online),
		UnSheleves: len(deviceInfoList) - len(online),
		BatteryId:  map[string]int{},
		Gps:        []dto.DeviceGpsOpeMapCo{},
	}

	for _, d := range online {
		ridingCounts(d, &co.CanRent, &co.Booking, &co.Riding, &co.Parking, &co.Operation)

		ops := intList(d, "operationState")
		if containsInt(ops, 2) {
			co.MoveCar++
		}
		if containsInt(ops, 3) {
			co.ChangeBattery++
		}
		if containsInt(ops, 4) {
			co.LowBattery++
		}
		if containsInt(ops, 5) {
			co.Fixing++
		}
		if containsInt(ops, 6) {
			co.DragBack++
		}

		al := intList(d, "alarmState")
		if containsInt(al, 2) {
			co.MoveAlarm++
		}
		if containsInt(al, 3) {
			co.OutGfence++
		}
		if containsInt(al, 4) {
			co.NoParkingZone++
		}
		if containsInt(al, 5) {
			co.OutParkingZone++
		}
		if containsInt(al, 6) {
			co.PowerCut++
		}
		if containsInt(al, 7) {
			co.Offline++
		}
		if containsInt(al, 8) {
			co.OrderWithoutGps++
		}
		if containsInt(al, 9) {
			co.Lost++
		}
		if containsInt(al, 10) {
			co.TooLongOrder++
		}
		if containsInt(al, 11) {
			co.TooShortOrder++
		}
		if containsInt(al, 12) {
			co.UnlockAbnormal++
		}
		if containsInt(al, 13) {
			co.HelmetLost++
		}
		if containsInt(al, 14) {
			co.HelmetFault++
		}

		rb := restBatteryOrZero(d)
		if rb <= 10 {
			co.RestBatteryLess10++
		}
		if rb <= 20 {
			co.RestBatteryLess20++
		}
		if rb <= 30 {
			co.RestBatteryLess30++
		}
		if rb <= 35 {
			co.RestBatteryLess35++
		}
		if rb <= 40 {
			co.RestBatteryLess40++
		}
		if qry.RestBattery != nil && rb <= *qry.RestBattery {
			co.RestBatteryLessCustom++
		}

		if bid, ok := asInt64(d["batteryId"]); ok {
			co.BatteryId[strconv.FormatInt(bid, 10)]++
		}
	}

	// GPS markers reuse the platform getFilterDevice predicate chain over the full
	// (incl. off-rack) device list.
	gpsFilter := dto.DevicePageQry{
		RidingState:    qry.RidingState,
		OperationState: qry.OperationState,
		AlarmState:     qry.AlarmState,
		NoOrderTime:    qry.NoOrderTime,
		OrderTime:      qry.OrderTime,
		RestBattery:    qry.RestBattery,
		BatteryId:      qry.BatteryId,
	}
	for _, d := range filterDevices(deviceInfoList, gpsFilter) {
		co.Gps = append(co.Gps, dto.DeviceGpsOpeMapCo{
			Imei:           strPtr(d, "imei"),
			CarId:          strPtr(d, "carId"),
			Lng:            f64PtrDefault(d, "lng", 0),
			Lat:            f64PtrDefault(d, "lat", 0),
			RidingState:    intPtr(d, "ridingState"),
			RestBattery:    restBattery(d),
			OperationState: intList(d, "operationState"),
			AlarmState:     intList(d, "alarmState"),
			CarTagTypeIds:  intList(d, "carTagTypeIds"),
		})
	}
	return co
}
