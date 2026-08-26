package service

import (
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// EcuBleCommandReport ports EcuBleCommandReportServiceImpl: bluetooth control
// result callbacks that either emit a C34 Kafka snapshot or write the device-info
// cache. Write/report endpoints -> NOT shadow-compared.

// requireDeviceRack mirrors the Assert.notNull(device)/notNull(serviceId) guard
// shared by most BLE reports, returning the serviceId for the report writes.
func requireDeviceRack(tenantID, imei string) (int64, error) {
	device := repository.GetDeviceByImei(tenantID, imei)
	if device == nil {
		return 0, ErrDeviceNull
	}
	sid, ok := serviceIDOf(device)
	if !ok {
		return 0, ErrDeviceNullRack
	}
	return sid, nil
}

// BleLockCommandReport ports bleLockCommandReport: manage the no-risk-control
// flag, then emit a C34 acc snapshot.
func BleLockCommandReport(cmd *dto.BleLockReportCmd) error {
	tenantID := dto.TenantOf(cmd.CommandContext)
	if cmd.Acc != nil {
		accAuditLog(cmd.CommandContext, cmd.Imei, cmd.CarID, *cmd.Acc, jsonStr(cmd))
	}
	if _, err := requireDeviceRack(tenantID, cmd.Imei); err != nil {
		return err
	}
	izRisk := cmd.IzRiskControl == nil || *cmd.IzRiskControl // defaults true
	if cmd.Acc != nil && *cmd.Acc == 1 {
		if !izRisk {
			repository.SetNoRiskControlEx(tenantID, cmd.Imei)
		} else {
			repository.DeleteNoRiskControlEx(tenantID, cmd.Imei)
		}
	}
	sendC34(tenantID, cmd.Imei, c34Lock(cmd.Acc, f64(cmd.Lng), f64(cmd.Lat)))
	return nil
}

// BleDefendCommandReport ports bleDefendCommandReport: clear no-risk-control,
// then emit a C34 defend snapshot.
func BleDefendCommandReport(cmd *dto.BleDefendReportCmd) error {
	tenantID := dto.TenantOf(cmd.CommandContext)
	if cmd.Defend != nil {
		defendAuditLog(cmd.CommandContext, cmd.Imei, cmd.CarID, *cmd.Defend, jsonStr(cmd))
	}
	if _, err := requireDeviceRack(tenantID, cmd.Imei); err != nil {
		return err
	}
	repository.DeleteNoRiskControlEx(tenantID, cmd.Imei)
	sendC34(tenantID, cmd.Imei, c34Defend(cmd.Defend, f64(cmd.Lng), f64(cmd.Lat)))
	return nil
}

// BleHelmetCommandReport ports bleHelmetCommandReport -> changeHelmetLockState.
func BleHelmetCommandReport(cmd *dto.BleHelmetLockReportCmd) error {
	return updateSwitchState(dto.TenantOf(cmd.CommandContext), cmd.Imei, "helmetLock", cmd.Sw)
}

// BleRearWheelLockReport ports bleRearWheelLockReport -> changeRearWheelLockState.
func BleRearWheelLockReport(cmd *dto.BleRearWheelLockReportCmd) error {
	if cmd.Sw != nil {
		rearWheelLockAuditLog(cmd.CommandContext, cmd.Imei, cmd.CarID, *cmd.Sw, jsonStr(cmd))
	}
	return updateSwitchState(dto.TenantOf(cmd.CommandContext), cmd.Imei, "backWheelLock", cmd.Sw)
}

// BleBatteryCompartmentCommandReport ports bleBatteryCompartmentCommandReport ->
// changeBatteryLockState.
func BleBatteryCompartmentCommandReport(cmd *dto.BleBatteryCompartmentReportCmd) error {
	if cmd.Sw != nil {
		batteryCompartmentAuditLog(cmd.CommandContext, cmd.Imei, cmd.CarID, *cmd.Sw, jsonStr(cmd))
	}
	return updateSwitchState(dto.TenantOf(cmd.CommandContext), cmd.Imei, "batteryLock", cmd.Sw)
}

// updateSwitchState writes a single switch field (+reportTime) after the
// device/serviceId asserts, mirroring the changeXxxLockState helpers.
func updateSwitchState(tenantID, imei, field string, sw *int) error {
	sid, err := requireDeviceRack(tenantID, imei)
	if err != nil {
		return err
	}
	device := map[string]interface{}{
		"imei":       imei,
		"reportTime": time.Now().UnixMilli(),
	}
	if sw != nil {
		device[field] = *sw
	}
	repository.UpdateDeviceState(tenantID, device, sid)
	return nil
}

// BleDeviceInfoReport ports bleDeviceInfoReport: emit a full C34 snapshot (no
// device assert).
func BleDeviceInfoReport(cmd *dto.BleDeviceInfoReportCmd) error {
	tenantID := dto.TenantOf(cmd.CommandContext)
	result := c34DeviceInfo(&dtoBleDeviceInfo{
		Acc:          cmd.Acc,
		Defend:       cmd.Defend,
		Gsm:          cmd.Gsm,
		Voltage:      cmd.Voltage,
		Helmet6Lock:  cmd.Helmet6Lock,
		Helmet6React: cmd.Helmet6React,
		TotalMiles:   cmd.TotalMiles,
		Speed:        cmd.Speed,
		Course:       cmd.Course,
		Timestamp:    cmd.Timestamp,
		Lng:          cmd.Lng,
		Lat:          cmd.Lat,
	})
	sendC34(tenantID, cmd.Imei, result)
	return nil
}

// RfidInfoReport ports rfidInfoReport -> rfidInfoReport (update rfidAck/carId/ts).
func RfidInfoReport(cmd *dto.BleRfidInfoReportCmd) error {
	tenantID := dto.TenantOf(cmd.CommandContext)
	sid, err := requireDeviceRack(tenantID, cmd.Imei)
	if err != nil {
		return err
	}
	device := map[string]interface{}{
		"imei":       cmd.Imei,
		"reportTime": time.Now().UnixMilli(),
	}
	if cmd.RfidAck != nil {
		device["rfidAck"] = *cmd.RfidAck
	}
	if cmd.RfidCarId != "" {
		device["rfidCarId"] = cmd.RfidCarId
	}
	if cmd.RfidTimestamp != nil {
		device["rfidTimestamp"] = *cmd.RfidTimestamp
	}
	repository.UpdateDeviceState(tenantID, device, sid)
	return nil
}
