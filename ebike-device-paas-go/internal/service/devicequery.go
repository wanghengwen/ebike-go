package service

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/geo"
	"ebike-device-paas-go/internal/repository"
)

// genDeviceMapFake business errors (mirroring DevicePassMsgCode).
var (
	// ErrGenFakeOutOfLimit -> 17016 (fakeAmount over the 5*10^5 cap).
	ErrGenFakeOutOfLimit = errors.New("gen fake device out of limit")
	// ErrGenFakeNx -> 17017 (a generation is already running for this service).
	ErrGenFakeNx = errors.New("gen fake device in progress")
)

// fake-location defaults mirror DeviceMapLocationDo's incr-ctor constants.
const (
	defaultIncrImei  = "881230040020233"
	defaultIncrCarId = "110600000"
)

// GenDeviceMapFake ports DeviceInfoServiceImpl.genDeviceMapFake: generate a fake
// device map (real on-rack devices + random in-polygon points) and cache it.
// This is a WRITE (lock + 2 SETs); it is not shadow-compared.
func GenDeviceMapFake(tenantID string, cmd dto.GenDeviceMapFakeCmd) error {
	if float64(cmd.FakeAmount) > 5*math.Pow(10, 5) {
		return ErrGenFakeOutOfLimit
	}
	if !repository.DeviceMapFakeNx(tenantID, cmd.ServiceId.Int64()) {
		return ErrGenFakeNx
	}

	rack := repository.GetRackDeviceByServiceId(tenantID, cmd.ServiceId.Int64())
	locations := make([]dto.DeviceMapLocationCo, 0, cmd.FakeAmount)
	realCount := cmd.FakeAmount
	if realCount > len(rack) {
		realCount = len(rack)
	}
	for _, d := range rack[:realCount] {
		locations = append(locations, dto.DeviceMapLocationCo{
			Lng:   f64PtrDefault(d, "lng", 0),
			Lat:   f64PtrDefault(d, "lat", 0),
			Imei:  strPtr(d, "imei"),
			CarId: strPtr(d, "carId"),
		})
	}
	if cmd.FakeAmount > len(rack) {
		pointStr := client.FenceServicePoint(cmd.ServiceId.Int64(), cmd.CommandContext)
		for _, p := range geo.RandomPointsInPolygon(pointStr, cmd.FakeAmount-len(rack)) {
			lng, lat := p.Lng, p.Lat
			imei, carID := defaultIncrImei, defaultIncrCarId
			locations = append(locations, dto.DeviceMapLocationCo{
				Lng: &lng, Lat: &lat, Imei: &imei, CarId: &carID,
			})
		}
	}

	payload, err := json.Marshal(locations)
	if err != nil {
		return err
	}
	repository.SetDeviceMapFake(tenantID, cmd.ServiceId.Int64(), string(payload), cmd.Amount)
	return nil
}

// GetBlueToothToken ports DeviceInfoServiceImpl.getBlueToothToken: validates the
// device exists and belongs to the tenant, then reads the bluetooth token
// (default DEFAULT_BLUETOOTH_TOKEN when absent).
const defaultBluetoothToken = 168428805

// GetBlueToothToken returns the cached bluetooth token, or ErrDeviceNotFound when
// the device is missing / belongs to another tenant.
func GetBlueToothToken(tenantID, imei string) (dto.BlueToothTokenCo, error) {
	dev := repository.GetDeviceByImei(tenantID, imei)
	if dev == nil {
		return dto.BlueToothTokenCo{}, ErrDeviceNotFound
	}
	if t, _ := dev["tenantId"].(string); t != "" && t != tenantID {
		return dto.BlueToothTokenCo{}, ErrDeviceNotFound
	}
	tok := repository.GetBluetoothToken(tenantID, imei)
	if tok == nil {
		return dto.BlueToothTokenCo{Token: defaultBluetoothToken}, nil
	}
	return dto.BlueToothTokenCo{Token: *tok}, nil
}

// QuerySaddleOverloadContact ports DeviceInfoServiceImpl.querySaddleOverloadContact:
// decode the three saddle-contact bits from the cached payload (default 0).
func QuerySaddleOverloadContact(tenantID, imei string) dto.SaddleOverloadContactCo {
	payload := repository.GetSaddleOverloadContact(tenantID, imei)
	return dto.SaddleOverloadContactCo{
		FrontSaddleContact: payload & 1,
		CentSaddleContact:  (payload >> 1) & 1,
		BackSaddleContact:  (payload >> 2) & 1,
	}
}

// QueryCameraState ports DeviceInfoServiceImpl.queryCameraState: returns nil when
// the cache is empty, otherwise the parsed camera state.
func QueryCameraState(tenantID, imei string) *dto.CameraCacheCo {
	raw := repository.GetCameraStateRaw(tenantID, imei)
	if raw == "" {
		return nil
	}
	var co dto.CameraCacheCo
	if err := json.Unmarshal([]byte(raw), &co); err != nil {
		return nil
	}
	return &co
}

// QueryDeviceMapFake ports DeviceInfoServiceImpl.queryDeviceMapFake: cached amount
// + fake location list for a service.
func QueryDeviceMapFake(tenantID string, serviceID int64) dto.DeviceMapFakeCo {
	co := dto.DeviceMapFakeCo{
		Amount: repository.GetDeviceMapAmount(tenantID, serviceID),
		Result: []dto.DeviceMapLocationCo{},
	}
	raw := repository.GetDeviceMapLocationRaw(tenantID, serviceID)
	if raw != "" {
		var locs []dto.DeviceMapLocationCo
		if err := json.Unmarshal([]byte(raw), &locs); err == nil && locs != nil {
			co.Result = locs
		}
	}
	return co
}

// OneClickReturnNotify ports DeviceInfoServiceImpl.oneClickReturnNotify: map the
// cached fail-notify code to the part-analysis result. canUse defaults to true.
func OneClickReturnNotify(tenantID, carID string) dto.PartAnalysisResultCO {
	canUse := true
	res := dto.PartAnalysisResultCO{CanUse: &canUse}

	raw := repository.GetOneClickReturnFailNotify(tenantID, carID)
	if raw == "" {
		return res
	}
	code, err := strconv.Atoi(raw)
	if err != nil {
		return res
	}
	izExist := true
	failed := false
	switch code {
	case 2:
		name := "rfid"
		res.Name, res.IzExist, res.Result, res.State = &name, &izExist, &failed, []string{"DEFAULT_0"}
	case 3:
		name := "direction"
		res.Name, res.IzExist, res.Result, res.State = &name, &izExist, &failed, []string{"DEFAULT_0"}
	case 4:
		name := "beacon"
		res.Name, res.IzExist, res.Result, res.State = &name, &izExist, &failed, []string{"DEFAULT_0"}
	case 6:
		name := "helmet"
		res.Name, res.IzExist, res.Result, res.State = &name, &izExist, &failed, []string{"HELMET_6LOCK_0", "HELMET_6REACT_0"}
	}
	return res
}

// QueryDeviceByBattery ports DeviceInfoServiceImpl.queryDeviceByBattery.
func QueryDeviceByBattery(tenantID string, qry dto.DeviceByBatteryQry) []dto.DeviceByBatteryCo {
	min := 0
	if qry.MinRestBattery != nil {
		min = *qry.MinRestBattery
	}
	max := 100
	if qry.MaxRestBattery != nil {
		max = *qry.MaxRestBattery
	}
	imeis := repository.QueryImeiByBattery(tenantID, qry.ServiceId.Int64(), min, max)
	devices := repository.GetDeviceInfoList(tenantID, imeis)
	out := []dto.DeviceByBatteryCo{}
	for _, d := range devices {
		if !serviceMatch(d, qry.ServiceId.Int64()) {
			continue
		}
		out = append(out, dto.DeviceByBatteryCo{
			Imei:           strPtr(d, "imei"),
			CarId:          strPtr(d, "carId"),
			RestBattery:    restBattery(d),
			ServiceId:      i64Ptr(d, "serviceId"),
			MaintainAreaId: i64Ptr(d, "maintainAreaId"),
			RidingState:    intPtr(d, "ridingState"),
			OperationState: intList(d, "operationState"),
			AlarmState:     intList(d, "alarmState"),
		})
	}
	return out
}

// QueryDeviceByTotalMiles ports DeviceInfoServiceImpl.queryDeviceByTotalMiles.
func QueryDeviceByTotalMiles(tenantID string, qry dto.DeviceByTotalMilesQry) []dto.DeviceByTotalMilesCo {
	min := 0.0
	if qry.MinTotalMiles != nil {
		min = *qry.MinTotalMiles
	}
	max := math.MaxFloat64
	if qry.MaxTotalMiles != nil {
		max = *qry.MaxTotalMiles
	}
	scoreMap := repository.QueryTotalMilesMap(tenantID, qry.ServiceId.Int64(), min, max)
	devices := repository.GetDeviceInfoList(tenantID, keysOf(scoreMap))
	out := []dto.DeviceByTotalMilesCo{}
	for _, d := range devices {
		if !serviceMatch(d, qry.ServiceId.Int64()) {
			continue
		}
		imei, _ := d["imei"].(string)
		miles := scoreMap[imei]
		out = append(out, dto.DeviceByTotalMilesCo{
			Imei:           strPtr(d, "imei"),
			CarId:          strPtr(d, "carId"),
			TotalMiles:     &miles,
			ServiceId:      i64Ptr(d, "serviceId"),
			MaintainAreaId: i64Ptr(d, "maintainAreaId"),
			RidingState:    intPtr(d, "ridingState"),
			OperationState: intList(d, "operationState"),
			AlarmState:     intList(d, "alarmState"),
		})
	}
	return out
}

// QueryDeviceByNoOrderTime ports DeviceInfoServiceImpl.queryDeviceByNoOrderTime.
// The lockTime window is derived from now-min/now-max; noOrderTime = now-lockTime.
func QueryDeviceByNoOrderTime(tenantID string, qry dto.DeviceByNoOrderTimeQry) []dto.DeviceByNoOrderTimeCo {
	now := time.Now().UnixMilli()
	minNo := int64(0)
	if v := qry.MinNoOrderTime.Ptr(); v != nil {
		minNo = *v
	}
	maxNo := int64(math.MaxInt64)
	if v := qry.MaxNoOrderTime.Ptr(); v != nil {
		maxNo = *v
	}
	maxLock := now - minNo
	minLock := now - maxNo
	scoreMap := repository.QueryLockTimeMap(tenantID, qry.ServiceId.Int64(), float64(minLock), float64(maxLock))
	devices := repository.GetDeviceInfoList(tenantID, keysOf(scoreMap))
	out := []dto.DeviceByNoOrderTimeCo{}
	for _, d := range devices {
		if !serviceMatch(d, qry.ServiceId.Int64()) {
			continue
		}
		imei, _ := d["imei"].(string)
		noOrder := time.Now().UnixMilli() - int64(scoreMap[imei])
		out = append(out, dto.DeviceByNoOrderTimeCo{
			Imei:           strPtr(d, "imei"),
			CarId:          strPtr(d, "carId"),
			NoOrderTime:    &noOrder,
			ServiceId:      i64Ptr(d, "serviceId"),
			MaintainAreaId: i64Ptr(d, "maintainAreaId"),
			RidingState:    intPtr(d, "ridingState"),
			OperationState: intList(d, "operationState"),
			AlarmState:     intList(d, "alarmState"),
		})
	}
	return out
}

// QueryDeviceByStaticTime ports DeviceInfoServiceImpl.queryDeviceByStaticTime.
func QueryDeviceByStaticTime(tenantID string, qry dto.DeviceByStaticTimeQry) []dto.DeviceByStaticTimeCo {
	now := time.Now().UnixMilli()
	minStatic := int64(0)
	if v := qry.MinStaticTime.Ptr(); v != nil {
		minStatic = *v
	}
	maxStatic := int64(math.MaxInt64)
	if v := qry.MaxStaticTime.Ptr(); v != nil {
		maxStatic = *v
	}
	maxEnd := now - minStatic
	minEnd := now - maxStatic
	scoreMap := repository.QueryStaticTimeMap(tenantID, qry.ServiceId.Int64(), float64(minEnd), float64(maxEnd))
	devices := repository.GetDeviceInfoList(tenantID, keysOf(scoreMap))
	out := []dto.DeviceByStaticTimeCo{}
	for _, d := range devices {
		if !serviceMatch(d, qry.ServiceId.Int64()) {
			continue
		}
		imei, _ := d["imei"].(string)
		static := time.Now().UnixMilli() - int64(scoreMap[imei])
		out = append(out, dto.DeviceByStaticTimeCo{
			Imei:           strPtr(d, "imei"),
			CarId:          strPtr(d, "carId"),
			StaticTime:     &static,
			ServiceId:      i64Ptr(d, "serviceId"),
			MaintainAreaId: i64Ptr(d, "maintainAreaId"),
			RidingState:    intPtr(d, "ridingState"),
			OperationState: intList(d, "operationState"),
			AlarmState:     intList(d, "alarmState"),
		})
	}
	return out
}

// RemoveDeviceTotalMiles ports DeviceInfoServiceImpl.removeDeviceTotalMiles:
// look up the device's serviceId then ZREM the imei from its total-miles zset.
func RemoveDeviceTotalMiles(tenantID, imei string) {
	dev := repository.GetDeviceByImei(tenantID, imei)
	if dev == nil {
		return
	}
	sid, ok := asInt64(dev["serviceId"])
	if !ok {
		return
	}
	repository.RemoveDeviceTotalMiles(tenantID, sid, imei)
}

// serviceMatch keeps only on-rack devices (operationState not containing 1) whose
// serviceId equals the requested one, mirroring the shared Java stream filters.
func serviceMatch(d map[string]interface{}, serviceID int64) bool {
	if onShelf(d) {
		return false
	}
	sid, ok := asInt64(d["serviceId"])
	return ok && sid == serviceID
}

func keysOf(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
