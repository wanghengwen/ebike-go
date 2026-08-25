package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ebike-open-paas-go/internal/contract"
	"ebike-open-paas-go/internal/dispatcher"
	"ebike-open-paas-go/internal/event"
	"ebike-open-paas-go/internal/middleware"
	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/repository"
	"ebike-open-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

func readJSON(c *gin.Context) (map[string]interface{}, bool) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		contract.Fail(c, contract.ErrInputParamsMiss, "request body unreadable")
		return nil, false
	}
	if len(body) == 0 {
		return map[string]interface{}{}, true
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		contract.Fail(c, contract.ErrInputParamsMiss, "request body invalid json")
		return nil, false
	}
	return raw, true
}

// requireImei validates the imei and that it belongs to the caller's tenant.
//
// notBelongAsCode selects the shape of the 1001 answer. Command endpoints
// document their result as data.code, so 1001 belongs there; the query endpoints
// return a domain object and have no code field, so for them 1001 has to travel
// in the error envelope or a caller sees `{"code":1001}` where a device record
// was promised.
func requireImei(c *gin.Context, v interface{}, notBelongAsCode bool) (string, bool) {
	imei, err := service.NormalizeImei(v)
	if err != nil {
		contract.Fail(c, contract.ErrImeiIllegal, err.Error())
		return "", false
	}
	tenantID := middleware.TenantIDFrom(c)
	agentID := middleware.AgentIDFrom(c)
	if err := service.EnsureDeviceBelongs(agentID, tenantID, imei); err != nil {
		if strings.Contains(err.Error(), strconv.Itoa(contract.CodeNotBelongAgent)) {
			if notBelongAsCode {
				contract.OKDeviceCode(c, contract.CodeNotBelongAgent)
			} else {
				contract.Fail(c, contract.ErrBizError, "device does not belong to this agent (1001)")
			}
			return "", false
		}
		contract.Fail(c, contract.ErrImeiIllegal, "imei illegal")
		return "", false
	}
	return imei, true
}

// requireImeiForQuery is requireImei for the GET query endpoints.
func requireImeiForQuery(c *gin.Context, v interface{}) (string, bool) {
	return requireImei(c, v, false)
}

// requireImeiForCommand is requireImei for the device-command endpoints.
func requireImeiForCommand(c *gin.Context, v interface{}) (string, bool) {
	return requireImei(c, v, true)
}

// upstreamFailure answers a failed upstream call without quoting it.
//
// An upstream error text carries ClusterIP addresses, internal paths and raw
// response bodies, and this API is published on a public ingress — so the detail
// goes to the log and the caller gets the error type the spec defines.
func upstreamFailure(c *gin.Context, errorType, what string, err error) {
	log.Printf("[api] %s %s failed: %v", c.Request.URL.Path, what, err)
	contract.Fail(c, errorType, "")
}

func optsFrom(raw map[string]interface{}) service.CommandOpts {
	var opts service.CommandOpts
	if _, ok := raw["idx"]; ok {
		v := service.AsInt(raw["idx"], 0)
		opts.Idx = &v
	}
	if _, ok := raw["volume"]; ok {
		v := service.AsInt(raw["volume"], 0)
		opts.Volume = &v
	}
	if _, ok := raw["isTBeacon"]; ok {
		v := service.AsInt(raw["isTBeacon"], 0)
		opts.IsTBeacon = &v
	}
	return opts
}

func writeCmd(c *gin.Context, code int, err error) {
	if err != nil && code == contract.CodeServerInternal {
		upstreamFailure(c, contract.ErrSendCmdError, "device command", err)
		return
	}
	contract.OKDeviceCode(c, code)
}

// ---- query handlers ----

func allDevices(c *gin.Context) {
	pageSize := service.AsInt(c.Query("pageSize"), 10)
	pageNumber := service.AsInt(c.Query("pageNumber"), 0)
	if q := c.Query("pageSize"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			pageSize = n
		}
	}
	if q := c.Query("pageNumber"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			pageNumber = n
		}
	}
	data, err := service.AllDevices(middleware.TenantIDFrom(c), pageSize, pageNumber)
	if err != nil {
		upstreamFailure(c, contract.ErrBizError, "device list", err)
		return
	}
	contract.OK(c, data)
}

func deviceInfo(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	data, err := service.GetDeviceInfo(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrDeviceInfoIsNull, "device shadow", err)
		return
	}
	contract.OK(c, data)
}

func realtimeDevice(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForQuery(c, raw["imei"])
	if !ok {
		return
	}
	data, err := service.RealtimeDeviceInfo(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrSendCmdError, "realtime device query", err)
		return
	}
	contract.OK(c, data)
}

func gpsPoints(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	start, _ := strconv.ParseInt(c.Query("startTime"), 10, 64)
	end, _ := strconv.ParseInt(c.Query("endTime"), 10, 64)
	if start <= 0 || end <= 0 {
		contract.Fail(c, contract.ErrInputParamsMiss, "startTime/endTime missing")
		return
	}
	data, err := service.GPSPoints(middleware.TenantIDFrom(c), imei, start, end)
	switch {
	case errors.Is(err, service.ErrRangeTooLarge):
		contract.Fail(c, contract.ErrRangeTooLarge,
			"request range too large (105): at most "+service.MaxTrajectoryRange.String())
		return
	case errors.Is(err, service.ErrRangeInvalid):
		contract.Fail(c, contract.ErrInputParamsMiss, "endTime must be later than startTime")
		return
	case err != nil:
		upstreamFailure(c, contract.ErrBizError, "trajectory query", err)
		return
	}
	contract.OK(c, data)
}

func address(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	data, err := service.Address(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrBizError, "reverse geocode", err)
		return
	}
	contract.OK(c, data)
}

func batteryInfo(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	data, err := service.BatteryInfo(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrDeviceInfoIsNull, "cached battery info", err)
		return
	}
	contract.OK(c, data)
}

func bmsInfo(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	data, err := service.BmsInfo(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrDeviceInfoIsNull, "cached bms info", err)
		return
	}
	contract.OK(c, data)
}

func currentBmsInfo(c *gin.Context) {
	imei, ok := requireImeiForQuery(c, c.Query("imei"))
	if !ok {
		return
	}
	data, err := service.CurrentBmsInfo(middleware.TenantIDFrom(c), imei)
	if err != nil {
		upstreamFailure(c, contract.ErrBizError, "realtime bms query", err)
		return
	}
	contract.OK(c, data)
}

// ---- command handlers ----

func lock(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	locked := service.AsInt(raw["locked"], -1)
	if locked != 0 && locked != 1 {
		contract.Fail(c, contract.ErrInputParamsMiss, "locked missing")
		return
	}
	code, err := service.LockControl(middleware.TenantIDFrom(c), imei, locked, optsFrom(raw))
	writeCmd(c, code, err)
}

func acc(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	v := service.AsInt(raw["acc"], -1)
	if v != 0 && v != 1 {
		contract.Fail(c, contract.ErrInputParamsMiss, "acc missing")
		return
	}
	code, err := service.AccControl(middleware.TenantIDFrom(c), imei, v, optsFrom(raw))
	writeCmd(c, code, err)
}

func defend(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	v := service.AsInt(raw["defend"], -1)
	if v != 0 && v != 1 {
		contract.Fail(c, contract.ErrInputParamsMiss, "defend missing")
		return
	}
	code, err := service.DefendControl(middleware.TenantIDFrom(c), imei, v, optsFrom(raw))
	writeCmd(c, code, err)
}

func backWheel(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	locked := service.AsInt(raw["locked"], -1)
	if locked != 0 && locked != 1 {
		contract.Fail(c, contract.ErrInputParamsMiss, "locked missing")
		return
	}
	code, err := service.BackWheelControl(middleware.TenantIDFrom(c), imei, locked, optsFrom(raw))
	writeCmd(c, code, err)
}

func batteryCompartment(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	locked := service.AsInt(raw["locked"], -1)
	if locked != 0 && locked != 1 {
		contract.Fail(c, contract.ErrInputParamsMiss, "locked missing")
		return
	}
	code, err := service.BatteryCompartmentControl(middleware.TenantIDFrom(c), imei, locked, optsFrom(raw))
	writeCmd(c, code, err)
}

func deviceVoice(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	index := service.AsInt(raw["index"], 0)
	volume := service.AsInt(raw["volume"], 0)
	code, err := service.DeviceVoice(middleware.TenantIDFrom(c), imei, index, volume)
	writeCmd(c, code, err)
}

func bluetooth(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	var token *int64
	if _, ok := raw["token"]; ok {
		v := service.AsInt64(raw["token"], 0)
		token = &v
	}
	name := service.AsString(raw["name"])
	code, err := service.Bluetooth(middleware.TenantIDFrom(c), imei, token, name)
	if err != nil && code == contract.CodeServerInternal {
		upstreamFailure(c, contract.ErrSendCmdError, "bluetooth config", err)
		return
	}
	// tm is our response time: paas returns no timestamp for this command, and
	// the spec's example value is a second-precision one taken at reply.
	contract.OK(c, map[string]interface{}{"code": code, "tm": time.Now().Unix()})
}

func reboot(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	code, err := service.Reboot(middleware.TenantIDFrom(c), imei)
	writeCmd(c, code, err)
}

func mc(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	imei, ok := requireImeiForCommand(c, raw["imei"])
	if !ok {
		return
	}
	if _, present := raw["speed"]; !present {
		contract.OKDeviceCode(c, contract.CodeParamMissing)
		return
	}
	// The spec defines speed as a percentage with no lower bound, so 0 is a
	// legitimate request (limit the controller to a standstill) and rejecting it
	// blocked a documented value. An out-of-range one is a parameter error, which
	// this endpoint reports as data.code 114 like every other device result.
	speed := service.AsInt(raw["speed"], -1)
	if speed < 0 || speed > 100 {
		contract.OKDeviceCode(c, contract.CodeParamInvalid)
		return
	}
	code, err := service.SetLimitSpeed(middleware.TenantIDFrom(c), imei, speed)
	writeCmd(c, code, err)
}

func batteryPowerSwitch(c *gin.Context) {
	contract.Fail(c, contract.ErrBizError, "batteryPowerSwitch not supported")
}

func lbs2gps(c *gin.Context) {
	contract.Fail(c, contract.ErrBizError, "lbs2gps not supported")
}

func sms(c *gin.Context) {
	contract.Fail(c, contract.ErrBizError, "sms device control not supported")
}

// ---- callback CRUD ----

func registerCallback(c *gin.Context) {
	raw, ok := readJSON(c)
	if !ok {
		return
	}
	event := service.AsInt(raw["event"], 0)
	url := service.AsString(raw["url"])
	if !repository.IsValidEvent(event) || url == "" {
		contract.Fail(c, contract.ErrInputParamsMiss, "event/url missing")
		return
	}
	// Accepting a subscription we can never deliver is worse than refusing it:
	// the caller would wait indefinitely for events that have no source. UART
	// frames are not published to saas_0, so event 5 has nothing behind it.
	if !repository.IsDeliverableEvent(event) {
		contract.Fail(c, contract.ErrBizError,
			"event 5 (UART) is not available on this platform; no UART frames are published")
		return
	}
	n, err := repository.RegisterCallback(middleware.AgentIDFrom(c), event, url)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCallbackURL) {
			contract.Fail(c, contract.ErrInputParamsMiss, err.Error())
			return
		}
		upstreamFailure(c, contract.ErrBizError, "callback registration", err)
		return
	}
	// Re-registration is the customer's signal that the endpoint is healthy
	// again: clear any temporary mute without touching the Redis membership
	// beyond what RegisterCallback already wrote.
	dispatcher.ResetCircuit(url)
	contract.OK(c, n)
}

func getCallback(c *gin.Context) {
	event, _ := strconv.Atoi(c.Query("event"))
	if !repository.IsValidEvent(event) {
		contract.Fail(c, contract.ErrInputParamsMiss, "event missing")
		return
	}
	urls, err := repository.ListCallbacks(middleware.AgentIDFrom(c), event)
	if err != nil {
		upstreamFailure(c, contract.ErrBizError, "callback list", err)
		return
	}
	if urls == nil {
		urls = []string{}
	}
	contract.OK(c, urls)
}

func deleteCallback(c *gin.Context) {
	event, _ := strconv.Atoi(c.Query("event"))
	url := c.Query("url")
	if !repository.IsValidEvent(event) || url == "" {
		contract.Fail(c, contract.ErrInputParamsMiss, "event/url missing")
		return
	}
	n, err := repository.UnregisterCallback(middleware.AgentIDFrom(c), event, url)
	if err != nil {
		upstreamFailure(c, contract.ErrBizError, "callback removal", err)
		return
	}
	contract.OK(c, n)
}

// ---- internal ----

// debugDeviceBelong returns ownership diagnostics (internal only).
//
// Query: imei (required), and either agentId or tenantId. When agentId is given,
// tenantId is taken from Nacos open.agents.
func debugDeviceBelong(c *gin.Context) {
	imei, err := service.NormalizeImei(c.Query("imei"))
	if err != nil {
		contract.Fail(c, contract.ErrImeiIllegal, err.Error())
		return
	}
	tenantID := strings.TrimSpace(c.Query("tenantId"))
	agentID := strings.TrimSpace(c.Query("agentId"))
	if agentID != "" {
		if a, ok := config.Agent(agentID); ok {
			tenantID = a.TenantID
		} else {
			contract.Fail(c, contract.ErrBizError, "agentId not in open.agents registry")
			return
		}
	}
	if tenantID == "" {
		contract.Fail(c, contract.ErrInputParamsMiss, "tenantId or agentId required")
		return
	}
	d := service.DiagnoseDeviceBelongs(tenantID, imei)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"diagnostics": d,
			"agentId":     agentID,
		},
	})
}

// callbackStats reports the saas_0 consumer and dispatcher counters.
//
// "unmatched" dominating "emitted" is expected and healthy: saas_0 carries every
// tenant's device traffic and only tenants with a registered callback produce
// events. "dropped" climbing means a third-party URL cannot keep up.
func callbackStats(c *gin.Context) {
	seen, unmatched, malformed, emitted, panicked := event.Stats()
	dropped, delivered, failed := dispatcher.Stats()
	suppressed, tripped := dispatcher.CircuitStats()
	c.JSON(http.StatusOK, gin.H{"success": true, "code": "0", "msg": "成功", "data": gin.H{
		"seen":       seen,
		"unmatched":  unmatched,
		"malformed":  malformed,
		"emitted":    emitted,
		"panicked":   panicked,
		"dropped":    dropped,
		"delivered":  delivered,
		"failed":     failed,
		"suppressed": suppressed,
		"tripped":    tripped,
	}})
}
