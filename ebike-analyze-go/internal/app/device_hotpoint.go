package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/common/bizerror"
	"ebike-analyze-go/internal/common/geohash"
	"ebike-analyze-go/internal/common/protocol"
	"ebike-analyze-go/internal/infrastructure/redisgw"
	"ebike-analyze-go/internal/pkg/rpc"
)

type DeviceInfoService struct{}

type serviceAreaIDCO struct {
	ID       int64  `json:"id"`
	TenantID string `json:"tenantId"`
}

func (s *DeviceInfoService) GetDeviceList(cmd *dto.DeviceListQry, tenantID string) []dto.DeviceListCo {
	imeiList := cmd.ImeiList
	if len(cmd.CarIDList) > 0 {
		imeiList = redisgw.GetImeiListByCarIDs(tenantID, cmd.CarIDList)
	}
	redisImeiList := redisgw.GetImeiListFromService(tenantID, cmd.ServiceIDList, cmd.ReportTime)
	if len(imeiList) == 0 {
		imeiList = redisImeiList
	} else {
		imeiList = redisgw.RetainImeiList(imeiList, redisImeiList)
	}
	if len(imeiList) == 0 {
		return []dto.DeviceListCo{}
	}
	devices := redisgw.GetDeviceInfoList(tenantID, imeiList)
	collectedImeis := make([]string, len(devices))
	for i, d := range devices {
		collectedImeis[i] = d.Imei
	}
	tags := redisgw.GetMoveAlarmTags(tenantID, collectedImeis)
	out := make([]dto.DeviceListCo, 0, len(devices))
	for i, d := range devices {
		var moveTag *int
		if i < len(tags) && tags[i] != nil {
			if n, ok := tags[i].(int); ok {
				moveTag = &n
			}
		}
		out = append(out, mapDeviceToListCo(d, moveTag))
	}
	return out
}

func mapDeviceToListCo(d *protocol.DeviceInfo, moveTag *int) dto.DeviceListCo {
	return dto.DeviceListCo{
		Imei:            d.Imei,
		CarID:           d.CarID,
		ServiceID:       d.ServiceID,
		Acc:             d.Acc,
		Lat:             d.Lat,
		Lng:             d.Lng,
		Timestamp:       d.Timestamp,
		IsOnline:        d.IsOnline,
		Voltage:         d.Voltage,
		ReportTime:      d.ReportTime,
		IsOutofServAera: d.IsOutofServAera,
		NoParkID:        d.NoParkID,
		ForParkID:       d.ForParkID,
		NoRideParkID:    d.NoRideParkID,
		BatteryID:       d.BatteryID,
		BatteryLock:     d.BatteryLock,
		MoveAlarmTag:    moveTag,
		RestBattery:     d.RestBattery,
		LockTime:        d.LockTime,
		UnlockTime:      d.UnlockTime,
		RidingState:     d.RidingState,
		OperationState:  d.OperationState,
		AlarmState:      d.AlarmState,
	}
}

func (s *DeviceInfoService) CarStatistics(ctx context.Context) ([]dto.DeviceCarStatisticsCo, error) {
	allPairs, err := s.fetchAllServiceTenant(ctx)
	if err != nil {
		return nil, bizerror.New("00001", bizerror.FormatException(err))
	}
	if len(allPairs) == 0 {
		return []dto.DeviceCarStatisticsCo{}, nil
	}
	statMap := make(map[int64]*dto.DeviceCarStatisticsCo, len(allPairs))
	for _, p := range allPairs {
		statMap[p.ServiceID] = newDeviceCarStatisticsCo(p.TenantID, p.ServiceID)
	}
	flatImeis := redisgw.FlattenTenantServiceImeis(allPairs)
	for i := 0; i < len(flatImeis); i += 100 {
		end := i + 100
		if end > len(flatImeis) {
			end = len(flatImeis)
		}
		batch := flatImeis[i:end]
		devices := redisgw.GetRackDevicesByTenantServiceImei(batch)
		byService := groupDevicesByService(devices)
		now := time.Now().UnixMilli()
		for serviceID, list := range byService {
			st := statMap[serviceID]
			if st == nil {
				if len(list) == 0 {
					continue
				}
				st = newDeviceCarStatisticsCo(list[0].TenantID, serviceID)
				statMap[serviceID] = st
			}
			for _, d := range list {
				countRidingState(st, d.RidingState)
				countFreeTime(st, d.LockTime, now)
				countVoltage(st, d.RestBattery)
			}
		}
	}
	out := make([]dto.DeviceCarStatisticsCo, 0, len(statMap))
	for _, v := range statMap {
		out = append(out, *v)
	}
	return out, nil
}

func newDeviceCarStatisticsCo(tenantID string, serviceID int64) *dto.DeviceCarStatisticsCo {
	return &dto.DeviceCarStatisticsCo{TenantID: tenantID, ServiceID: serviceID}
}

func groupDevicesByService(devices []*protocol.DeviceInfo) map[int64][]*protocol.DeviceInfo {
	out := map[int64][]*protocol.DeviceInfo{}
	for _, d := range devices {
		if d.OperationState != nil && protocol.ContainsOpState(d.OperationState, 1) {
			continue
		}
		out[d.ServiceID] = append(out[d.ServiceID], d)
	}
	return out
}

func (s *DeviceInfoService) fetchAllServiceTenant(ctx context.Context) ([]redisgw.ServiceTenant, error) {
	type baseCmd struct {
		CommandContext *dto.CommandContext `json:"commandContext"`
	}
	type result struct {
		Success bool              `json:"success"`
		Data    []serviceAreaIDCO `json:"data"`
	}
	cmd := baseCmd{CommandContext: &dto.CommandContext{TenantId: "1", TraceId: fmt.Sprintf("%d", time.Now().UnixNano())}}
	var res result
	err := rpc.PostToService(ctx, rpc.ServiceFence, "/serviceArea/getAllService", cmd, &res)
	if err != nil {
		return nil, err
	}
	// Java ResultHelper.getResultData: success with empty data is not an error.
	if !res.Success {
		return nil, fmt.Errorf("fence getAllService failed")
	}
	if len(res.Data) == 0 {
		return []redisgw.ServiceTenant{}, nil
	}
	out := make([]redisgw.ServiceTenant, 0, len(res.Data))
	for _, item := range res.Data {
		out = append(out, redisgw.ServiceTenant{ServiceID: item.ID, TenantID: item.TenantID})
	}
	return out, nil
}

func countRidingState(st *dto.DeviceCarStatisticsCo, state *int) {
	if state == nil {
		return
	}
	switch *state {
	case 1:
		st.CanRent++
	case 2:
		st.Riding++
	case 3:
		st.Parking++
	case 4:
		st.Booking++
	case 5:
		st.Operation++
	}
}

// countFreeTime mirrors Java's `if (deviceInfoDO.getLockTime() != null)` check:
// any non-null lockTime (including 0) participates in the idle-time bucketing.
func countFreeTime(st *dto.DeviceCarStatisticsCo, lockTime *int64, now int64) {
	if lockTime == nil {
		return
	}
	dff := now - *lockTime
	switch {
	case dff > 1*3600*1000 && dff <= 3*3600*1000:
		st.FreeTimeOneToThree++
	case dff > 3*3600*1000 && dff <= 6*3600*1000:
		st.FreeTimeThreeToSix++
	case dff > 6*3600*1000 && dff <= 12*3600*1000:
		st.FreeTimeSixToTwelve++
	case dff > 12*3600*1000 && dff <= 24*3600*1000:
		st.FreeTimeHalfOrOneDay++
	case dff > 24*3600*1000 && dff <= 48*3600*1000:
		st.FreeTimeOneOrTowDay++
	case dff > 48*3600*1000:
		st.FreeTimeTowDayMore++
	}
}

func countVoltage(st *dto.DeviceCarStatisticsCo, restBattery *int) {
	v := 0
	if restBattery != nil {
		v = *restBattery
	}
	if v == 0 {
		st.VoltageZero++
	}
	if v > 0 && v <= 20 {
		st.VoltageZeroToTwenty++
	}
	if v > 20 && v <= 35 {
		st.VoltageTwentyToThirtyFive++
	}
	if v > 35 {
		st.VoltageThirtyFiveMore++
	}
}

type HotPointService struct{}

func (s *HotPointService) StartList(cmd *dto.HotPointCmd, tenantID string) ([]dto.HotPointCo, error) {
	prefix := fmt.Sprintf("hot_point_start:%s:%d:", tenantID, cmd.ServiceID)
	return s.list(cmd, prefix)
}

func (s *HotPointService) EndList(cmd *dto.HotPointCmd, tenantID string) ([]dto.HotPointCo, error) {
	prefix := fmt.Sprintf("hot_point_end:%s:%d:", tenantID, cmd.ServiceID)
	return s.list(cmd, prefix)
}

func (s *HotPointService) list(cmd *dto.HotPointCmd, keyPrefix string) ([]dto.HotPointCo, error) {
	start := cmd.Start.AsTime()
	end := cmd.End.AsTime()
	if end.Before(start) {
		return nil, bizerror.New("00001", "结束时间不能大于开始时间")
	}
	keys := buildHotPointKeysFromPrefix(keyPrefix, start, end)
	merged := redisgw.HGetAllHotPoint(keys)
	if len(merged) == 0 {
		return []dto.HotPointCo{}, nil
	}
	out := make([]dto.HotPointCo, 0, len(merged))
	for code, count := range merged {
		if strings.TrimSpace(code) == "" {
			continue
		}
		coord := geohash.GetSpaceCoordinate(code)
		out = append(out, dto.HotPointCo{Lat: coord[0], Lng: coord[1], Count: count})
	}
	return out, nil
}

func buildHotPointKeysFromPrefix(prefix string, start, end time.Time) []string {
	var keys []string
	for d := start; d.Before(end); d = d.Add(24 * time.Hour) {
		keys = append(keys, prefix+d.Format("20060102"))
	}
	return keys
}
