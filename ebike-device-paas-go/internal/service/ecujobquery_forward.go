package service

import (
	"encoding/json"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
)

// EcuJobQuery endpoints port EcuJobQueryServiceImpl: each reads an async job
// result from the openapi gateway (GET get_job_result) and maps it via the
// RpcResultToCo String-variant convertors.
//
// Reading a job result is idempotent, but it still forwards live to the upstream
// gateway, so (like the other EcuQuery forwards) it is NOT shadow-compared.

// jobState fetches the job result and maps the inner result object with mapResult
// (nil for the result-less variants). It mirrors RpcResultToCo.toNullResultCommandResult
// for String: async/jobId stay empty, ecuCode comes from the result.code field.
func jobState(q dto.JobStateQry, mapResult func(json.RawMessage) interface{}) (dto.CommandResult, error) {
	result, err := client.GetJobState(q.JobId, dto.TenantOf(q.CommandContext))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildJobCommandResult(result, mapResult), nil
}

func buildJobCommandResult(result string, mapResult func(json.RawMessage) interface{}) dto.CommandResult {
	var cr dto.CommandResult
	if result == "" {
		return cr
	}
	var outer struct {
		Code   string          `json:"code"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(result), &outer); err != nil {
		return cr
	}
	cr.EcuCode = &outer.Code
	if len(outer.Result) > 0 && string(outer.Result) != "null" && mapResult != nil {
		cr.Result = mapResult(outer.Result)
	}
	return cr
}

// LockCommandJobState ports EcuJobQueryServiceImpl.lockCommandJobQuery.
func LockCommandJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// SetInnerParamJobState ports setInnerParamJobQuery.
func SetInnerParamJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// UpdateInnerFenceJobState ports updateInnerFenceJobQuery.
func UpdateInnerFenceJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// VoiceUpgradeJobState ports voiceUpgradeJobState.
func VoiceUpgradeJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// DeviceUpgradeJobState ports deviceUpgradeJobState.
func DeviceUpgradeJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// VoiceCommandJobState ports voiceCommandJobQuery.
func VoiceCommandJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// BatteryCompartmentCommandJobState ports batteryCompartmentCommandJobQuery.
func BatteryCompartmentCommandJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, nil)
}

// DefendCommandJobState ports defendCommandJobQuery (RpcResultToCo::toDefendCo).
func DefendCommandJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, mapDefend)
}

// GetInnerParamJobState ports getInnerParamJobState (RpcResultToCo::toInnerParamQryCo).
func GetInnerParamJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, mapInnerParam)
}

// BluetoothJobState ports bluetoothJobQuery (toDefaultCommandResult BluetoothCo).
func BluetoothJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, mapBluetooth)
}

// GetDeviceInfoJobState ports getDeviceInfoJobState (toDefaultCommandResult
// DeviceInfoCo). Unlike EcuQuery.getDeviceInfo, the job-query path does NOT apply
// the coordinate.type / ecu.debug transforms, so the result object is passed through.
func GetDeviceInfoJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, mapJobResultObject)
}

// GetBlueTBeaconJobState ports getBlueTBeaconJobState (toDefaultCommandResult BlueTBeaconInfoCo).
func GetBlueTBeaconJobState(q dto.JobStateQry) (dto.CommandResult, error) {
	return jobState(q, mapBlueTBeacon)
}

// GetJobSuc ports getJobSuc: success iff ecuCode is empty/"0" (mirrors
// DeviceResultHelper.toDeviceResult applied in the controller). Returns the bool
// alongside the gateway error (if any) so the handler can build the JobSucCo.
func GetJobSuc(q dto.JobStateQry) (bool, error) {
	cr, err := jobState(q, nil)
	if err != nil {
		return false, err
	}
	ec := cr.EcuCodeValue()
	return ec == "" || ec == "0", nil
}

func mapBluetooth(raw json.RawMessage) interface{} {
	var co dto.BluetoothCo
	_ = json.Unmarshal(raw, &co)
	return co
}

// mapDefend mirrors RpcResultToCo.toDefendCo: parse DefendCo then copy "lon" into lng.
func mapDefend(raw json.RawMessage) interface{} {
	var co dto.DefendCo
	_ = json.Unmarshal(raw, &co)
	var m struct {
		Lon *float64 `json:"lon"`
	}
	if json.Unmarshal(raw, &m) == nil && m.Lon != nil {
		co.Lng = m.Lon
	}
	return co
}

// mapJobResultObject returns the gateway result object as-is (passthrough),
// mirroring RpcResultToCo.toCo without typed coercion or transforms.
func mapJobResultObject(raw json.RawMessage) interface{} {
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	return m
}
