package api

import (
	"errors"
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerEcuBleCommandReportRoutes attaches the EcuBleCommandReportApi
// endpoints. Bluetooth control-result callbacks (Redis write / C34 Kafka) ->
// NOT shadow-compared.
func registerEcuBleCommandReportRoutes(r *gin.Engine) {
	r.POST("/device/paas/ble/lock/report", bleLockCommandReport)
	r.POST("/device/paas/ble/defend/report", bleDefendCommandReport)
	r.POST("/device/paas/ble/helmetLock/report", bleHelmetCommandReport)
	r.POST("/device/paas/ble/rearWheelLock/report", bleRearWheelLockReport)
	r.POST("/device/paas/ble/batteryCompartment/report", bleBatteryCompartmentCommandReport)
	r.POST("/device/paas/ble/deviceInfo/report", bleDeviceInfoReport)
	r.POST("/device/paas/ble/rfid/report", rfidInfoReport)
}

// writeReportResult maps the report service result to the Result envelope:
// success -> ResultHelper.success(); the device-asserts surface as 17006/170018.
func writeReportResult(c *gin.Context, err error) {
	switch {
	case err == nil:
		c.JSON(http.StatusOK, Ok(nil))
	case errors.Is(err, service.ErrDeviceNull):
		c.JSON(http.StatusOK, Fail(CodeDeviceNullHave, ""))
	case errors.Is(err, service.ErrDeviceNullRack):
		c.JSON(http.StatusOK, Fail(CodeDeviceNullRack, ""))
	default:
		c.JSON(http.StatusOK, Fail(CodeParamError, err.Error()))
	}
}

func bleLockCommandReport(c *gin.Context) {
	var cmd dto.BleLockReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleLockCommandReport(&cmd))
}

func bleDefendCommandReport(c *gin.Context) {
	var cmd dto.BleDefendReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleDefendCommandReport(&cmd))
}

func bleHelmetCommandReport(c *gin.Context) {
	var cmd dto.BleHelmetLockReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleHelmetCommandReport(&cmd))
}

func bleRearWheelLockReport(c *gin.Context) {
	var cmd dto.BleRearWheelLockReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleRearWheelLockReport(&cmd))
}

func bleBatteryCompartmentCommandReport(c *gin.Context) {
	var cmd dto.BleBatteryCompartmentReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleBatteryCompartmentCommandReport(&cmd))
}

func bleDeviceInfoReport(c *gin.Context) {
	var cmd dto.BleDeviceInfoReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.BleDeviceInfoReport(&cmd))
}

func rfidInfoReport(c *gin.Context) {
	var cmd dto.BleRfidInfoReportCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	writeReportResult(c, service.RfidInfoReport(&cmd))
}
