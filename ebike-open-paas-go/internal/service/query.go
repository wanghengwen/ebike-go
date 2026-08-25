package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"ebike-open-paas-go/internal/client"
	"ebike-open-paas-go/internal/contract"
	"ebike-open-paas-go/internal/pkg/config"
)

// AllDevicesResult is Xiaoan's allDevices data payload.
type AllDevicesResult struct {
	Total   int             `json:"total"`
	Devices []AllDeviceItem `json:"devices"`
}

// AllDeviceItem is one IMEI entry.
type AllDeviceItem struct {
	Imei string `json:"imei"`
}

// AllDevices pages registered devices for a tenant from anvelink-console
// (anvelink_console_device). pageNumber is 0-based (Xiaoan); console pageNo is 1-based.
// No fence / management / carId binding — same ownership source as EnsureDeviceBelongs.
func AllDevices(tenantID string, pageSize, pageNumber int) (*AllDevicesResult, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNumber < 0 {
		pageNumber = 0
	}

	var page struct {
		Total   int64 `json:"total"`
		Records []struct {
			Imei string `json:"imei"`
		} `json:"records"`
	}
	q := url.Values{}
	q.Set("pageNo", strconv.Itoa(pageNumber+1))
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("tenantId", tenantID)
	q.Set("deviceType", "ebike")
	if err := client.GetConsole("/device/getList", q, &page); err != nil {
		return nil, fmt.Errorf("%s: %v", contract.ErrBizError, err)
	}

	seen := map[string]struct{}{}
	items := make([]AllDeviceItem, 0, len(page.Records))
	for _, d := range page.Records {
		if d.Imei == "" {
			continue
		}
		if _, ok := seen[d.Imei]; ok {
			continue
		}
		seen[d.Imei] = struct{}{}
		items = append(items, AllDeviceItem{Imei: d.Imei})
	}
	return &AllDevicesResult{Total: int(page.Total), Devices: items}, nil
}

// RealtimeDeviceInfo runs C34 via paas deviceInfo.
//
// payload is left unset on purpose: paas types it as a JSON *string* and
// defaults it to `{"type":"c34"}`, so sending an object here fails to bind
// upstream and no realtime query is issued at all.
func RealtimeDeviceInfo(tenantID, imei string) (map[string]interface{}, error) {
	env, err := client.PostPaas("/device/paas/deviceInfo", map[string]interface{}{
		"imei":           imei,
		"async":          false,
		"commandContext": client.CommandContext(tenantID, ""),
	})
	if err != nil {
		return nil, err
	}
	var cr client.CommandResult
	if len(env.Data) > 0 {
		_ = json.Unmarshal(env.Data, &cr)
	}
	code := contract.MapEcuCode(cr.EcuCodeString())
	if !env.Success && cr.EcuCode == nil {
		return nil, &client.GatewayError{Code: env.Code, Msg: env.Msg}
	}
	return map[string]interface{}{
		"code":   code,
		"result": projectRealtimeResult(c34Result(cr.Result)),
	}, nil
}

// c34Result normalises the CommandResult payload, which paas returns either as an
// object or as a JSON string depending on the upstream path.
func c34Result(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return t
	case string:
		if t == "" {
			return nil
		}
		var m map[string]interface{}
		if json.Unmarshal([]byte(t), &m) == nil {
			return m
		}
	}
	return nil
}

// projectRealtimeResult maps the internal c34 result onto the field names the
// Xiaoan realtime contract defines, dropping everything else.
//
// Forwarding the internal object verbatim was both under- and over-inclusive: it
// leaked fields that only mean something inside our platform (helmet6Lock,
// bmsComm, etcSpeed, kickStand, rfid …) while never producing the `battery`
// object or `GSMSignal` the spec documents. Absent fields are omitted, matching
// the spec's "字段是否存在与设备类型相关".
func projectRealtimeResult(m map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	if len(m) == 0 {
		return out
	}
	// Same name and meaning on both sides.
	for _, key := range []string{
		"acc", "defend", "voltageMv", "gsm", "audioType",
		"wheelLock", "seatLock", "isWheelSpan", "powerState",
	} {
		if v, ok := present(m, key); ok {
			out[key] = v
		}
	}
	// The spec publishes signal strength under both names.
	if v, ok := present(m, "gsm"); ok {
		out["GSMSignal"] = v
	}
	if battery := projectBattery(m); len(battery) > 0 {
		out["battery"] = battery
	}
	if gps := projectRealtimeGPS(m["gps"]); len(gps) > 0 {
		out["gps"] = gps
	}
	// cell is the documented fallback when the module has no GPS fix.
	if v, ok := present(m, "cell"); ok {
		out["cell"] = v
	}
	return out
}

// projectBattery assembles the spec's battery object out of the BMS fields c34
// carries. `type` has no counterpart in the c34 result and is left out.
func projectBattery(m map[string]interface{}) map[string]interface{} {
	battery := map[string]interface{}{}
	if v, ok := present(m, "bmsSoc"); ok {
		battery["percent"] = v
	}
	if v, ok := present(m, "bmsVoltage"); ok {
		battery["voltage"] = v
	}
	return battery
}

// projectRealtimeGPS emits the five documented GPS fields in WGS84, which is the
// coordinate system this endpoint specifies — the internal result carries both
// pairs, and passing the GCJ02 one through would offset every point.
func projectRealtimeGPS(v interface{}) map[string]interface{} {
	gps, _ := v.(map[string]interface{})
	if len(gps) == 0 {
		return nil
	}
	out := map[string]interface{}{}
	if ts, ok := present(gps, "timestamp"); ok {
		out["timestamp"] = ts
	}
	lat, hasLat := present(gps, "wgs84Lat")
	lng, hasLng := present(gps, "wgs84Lng")
	if !hasLat || !hasLng {
		return nil
	}
	out["lat"] = lat
	out["lng"] = lng
	for _, key := range []string{"speed", "course"} {
		if val, ok := present(gps, key); ok {
			out[key] = val
		}
	}
	return out
}

// present reports whether the key carries a value; the upstream serializes
// absent fields as null rather than omitting them.
func present(m map[string]interface{}, key string) (interface{}, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, false
	}
	return v, true
}

// MaxTrajectoryRange caps a single trajectory query.
//
// The spec reserves error 105 ("请求范围过大") for this but does not name the
// limit. 7 days is picked to match the trajectory retention the worker keeps;
// without a cap, one request spanning months walks the whole ZSET and the cost
// lands on the shared worker rather than on the caller.
const MaxTrajectoryRange = 7 * 24 * time.Hour

// ErrRangeTooLarge reports a trajectory window wider than MaxTrajectoryRange.
var ErrRangeTooLarge = errors.New("requested time range is too large")

// ErrRangeInvalid reports a window that does not describe a period at all.
var ErrRangeInvalid = errors.New("endTime must be later than startTime")

// GPSPoints fetches trajectory as WGS84 points. Xiaoan times are seconds.
func GPSPoints(tenantID, imei string, startSec, endSec int64) ([]map[string]interface{}, error) {
	if endSec <= startSec {
		return nil, ErrRangeInvalid
	}
	if time.Duration(endSec-startSec)*time.Second > MaxTrajectoryRange {
		return nil, ErrRangeTooLarge
	}
	var points []struct {
		Lng       float64 `json:"lng"`
		Lat       float64 `json:"lat"`
		Timestamp int64   `json:"timestamp"`
		Speed     float64 `json:"speed"`
		Course    float64 `json:"course"`
	}
	typeZero := 0 // WGS84
	err := client.PostWorker("/ebike/gps/getTrajectory", map[string]interface{}{
		"imei":      imei,
		"startTime": startSec * 1000,
		"endTime":   endSec * 1000,
		"type":      typeZero,
	}, tenantID, &points)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		ts := p.Timestamp
		if ts > 1e12 {
			ts = ts / 1000
		}
		out = append(out, map[string]interface{}{
			"lat":       p.Lat,
			"lon":       p.Lng,
			"timestamp": ts,
			"speed":     p.Speed,
			"course":    p.Course,
		})
	}
	return out, nil
}

// Address returns the latest reverse-geocoded address for the device.
func Address(tenantID, imei string) (map[string]interface{}, error) {
	detail, err := GetDeviceInfo(tenantID, imei)
	if err != nil {
		return nil, err
	}
	lng, lat, ok := detail.Coordinates()
	if !ok {
		// Reverse-geocoding (0,0) returns a point in the Gulf of Guinea, which
		// looks like a real answer.
		return nil, fmt.Errorf("%s: device has no cached position", contract.ErrDeviceInfoIsNull)
	}
	apiName := config.GlobalConfig().Xyy.MapServiceConfig.API
	if apiName == "" {
		apiName = "aMap"
	}
	var addr struct {
		Name   string `json:"name"`
		Area   string `json:"area"`
		Adcode string `json:"adcode"`
	}
	traceID := client.NewTraceID()
	err = client.PostMap("/map/regeo", map[string]interface{}{
		"api":       apiName,
		"traceId":   traceID,
		"tenantId":  tenantID,
		"longitude": fmt.Sprintf("%f", lng),
		"latitude":  fmt.Sprintf("%f", lat),
	}, &addr)
	if err != nil {
		return nil, err
	}
	var ts int64
	if detail.LastGPSTimestamp != nil {
		ts = *detail.LastGPSTimestamp
	}
	return map[string]interface{}{
		"timestamp":  ts,
		"address":    addr.Name,
		"originInfo": addr,
	}, nil
}

// BatteryInfo maps available shadow fields into Xiaoan's batteryInfo shape.
//
// The spec lists capacity, cycle, remaining (mAH), temperature, version and
// eight BMS fault bits. The device shadow carries none of them — only bmsSN,
// soc (a percentage) and bmsTimeStamp — so they are omitted, which the spec
// allows ("字段是否存在与设备类型相关"). In particular `remaining` is not filled
// from soc: the spec defines it in mAH, and a 59 that means "59%" read as
// "59 mAH" is worse than no field at all. Serving the full shape needs the BMS
// capacity, which only arrives on a Bin66 report (see the bms callback).
func BatteryInfo(tenantID, imei string) (map[string]interface{}, error) {
	raw, err := deviceShadow(tenantID, imei)
	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{"imei": imei}
	if v := OptInt(raw, "isOnline"); v != nil {
		out["isOnline"] = *v
	}
	if sn := OptString(raw, "bmsSN"); sn != "" {
		out["SN"] = sn
	}
	return out, nil
}

// BmsInfo returns the cached BMS summary.
//
// capacity and cycle are absent for the same reason as in BatteryInfo: the
// shadow stores only the SN.
func BmsInfo(tenantID, imei string) (map[string]interface{}, error) {
	raw, err := deviceShadow(tenantID, imei)
	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{}
	if sn := OptString(raw, "bmsSN"); sn != "" {
		out["SN"] = sn
	}
	return out, nil
}

// CurrentBmsInfo queries the device for its battery state.
//
// It goes through the c34 realtime query rather than /device/paas/transmission:
// transmission is a raw passthrough that paas answers with success regardless of
// what the device says, and its `c` values are device opcodes, not the decoder
// ids they resemble — so the old c=41 call could not have returned BMS data
// under any circumstances (cmd 41 decodes as a GPS frame).
//
// c34 carries bmsSoc and bmsVoltage. current, temperature and remain are not in
// it and are therefore omitted rather than approximated.
func CurrentBmsInfo(tenantID, imei string) (map[string]interface{}, error) {
	env, err := client.PostPaas("/device/paas/deviceInfo", map[string]interface{}{
		"imei":           imei,
		"async":          false,
		"commandContext": client.CommandContext(tenantID, ""),
	})
	if err != nil {
		return nil, err
	}
	var cr client.CommandResult
	if len(env.Data) > 0 {
		_ = json.Unmarshal(env.Data, &cr)
	}
	if !env.Success && cr.EcuCode == nil {
		return nil, &client.GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if code := contract.MapEcuCode(cr.EcuCodeString()); code != contract.CodeOK {
		return nil, fmt.Errorf("device did not answer (code %d)", code)
	}
	m := c34Result(cr.Result)
	out := map[string]interface{}{}
	if v, ok := present(m, "bmsSoc"); ok {
		out["SOC"] = v
	}
	if v, ok := present(m, "bmsVoltage"); ok {
		out["voltage"] = v
	}
	return out, nil
}

// deviceShadow fetches the cached device record shared by the battery endpoints.
func deviceShadow(tenantID, imei string) (map[string]interface{}, error) {
	env, err := client.PostPaas("/device/paas/device/detail", map[string]interface{}{
		"imei":           imei,
		"commandContext": client.CommandContext(tenantID, ""),
	})
	if err != nil {
		return nil, err
	}
	if !env.Success || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, fmt.Errorf("%s", contract.ErrDeviceInfoIsNull)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(env.Data, &raw); err != nil {
		return nil, fmt.Errorf("%s", contract.ErrDeviceInfoIsNull)
	}
	return raw, nil
}
