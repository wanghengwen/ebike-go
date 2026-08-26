package api

import (
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerEcuJobQueryRoutes attaches the EcuJobQuery endpoints (EcuJobQueryApi).
//
// Each endpoint reads an async device-command job result from the openapi
// gateway, so (like the other EcuQuery forwards) they are NOT shadow-compared.
func registerEcuJobQueryRoutes(r *gin.Engine) {
	r.POST("/device/paas/jobQuery/lock", lockCommandJobState)
	r.POST("/device/paas/jobQuery/defend", defendCommandJobState)
	r.POST("/device/paas/jobQuery/setInnerParam", setInnerParamJobState)
	r.POST("/device/paas/jobQuery/updateInnerFence", updateInnerFenceJobState)
	r.POST("/device/paas/jobQuery/voiceUpgrade", voiceUpgradeJobState)
	r.POST("/device/paas/jobQuery/deviceUpgrade", deviceUpgradeJobState)
	r.POST("/device/paas/jobQuery/voice", voiceCommandJobState)
	r.POST("/device/paas/jobQuery/batteryCompartment", batteryCompartmentCommandJobState)
	r.POST("/device/paas/jobQuery/bluetooth", bluetoothJobState)
	r.POST("/device/paas/jobQuery/innerParam", getInnerParamJobState)
	r.POST("/device/paas/jobQuery/deviceInfo", getDeviceInfoJobState)
	r.POST("/device/paas/jobQuery/blueTBeacon", getBlueTBeaconJobState)
	r.POST("/device/paas/jobQuery/getJobSuc", getJobSuc)
}

func lockCommandJobState(c *gin.Context) { jobStateHandler(c, service.LockCommandJobState) }

func defendCommandJobState(c *gin.Context) { jobStateHandler(c, service.DefendCommandJobState) }

func setInnerParamJobState(c *gin.Context) { jobStateHandler(c, service.SetInnerParamJobState) }

func updateInnerFenceJobState(c *gin.Context) { jobStateHandler(c, service.UpdateInnerFenceJobState) }

func voiceUpgradeJobState(c *gin.Context) { jobStateHandler(c, service.VoiceUpgradeJobState) }

func deviceUpgradeJobState(c *gin.Context) { jobStateHandler(c, service.DeviceUpgradeJobState) }

func voiceCommandJobState(c *gin.Context) { jobStateHandler(c, service.VoiceCommandJobState) }

func batteryCompartmentCommandJobState(c *gin.Context) {
	jobStateHandler(c, service.BatteryCompartmentCommandJobState)
}

func bluetoothJobState(c *gin.Context) { jobStateHandler(c, service.BluetoothJobState) }

func getInnerParamJobState(c *gin.Context) { jobStateHandler(c, service.GetInnerParamJobState) }

func getDeviceInfoJobState(c *gin.Context) { jobStateHandler(c, service.GetDeviceInfoJobState) }

func getBlueTBeaconJobState(c *gin.Context) { jobStateHandler(c, service.GetBlueTBeaconJobState) }

// jobStateHandler binds the JobStateQry, runs the job-state lookup, and writes
// the result via the DeviceResultHelper.toDeviceResult mapping.
func jobStateHandler(c *gin.Context, fn func(dto.JobStateQry) (dto.CommandResult, error)) {
	var q dto.JobStateQry
	if !bindBody(c, &q) {
		return
	}
	if q.JobId == "" {
		c.JSON(http.StatusOK, Fail(CodeParamError, ""))
		return
	}
	backfillTenant(c, &q.CommandContext)
	cr, err := fn(q)
	writeDeviceResult(c, cr, err)
}

// getJobSuc ports EcuJobQueryController.getJobSuc: returns a JobSucCo reflecting
// whether the job's ecuCode indicated success.
func getJobSuc(c *gin.Context) {
	var q dto.JobStateQry
	if !bindBody(c, &q) {
		return
	}
	if q.JobId == "" {
		c.JSON(http.StatusOK, Fail(CodeParamError, ""))
		return
	}
	backfillTenant(c, &q.CommandContext)
	suc, err := service.GetJobSuc(q)
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, dto.JobSucCo{JobSuc: false}))
		return
	}
	if suc {
		c.JSON(http.StatusOK, Ok(dto.JobSucCo{JobSuc: true}))
		return
	}
	c.JSON(http.StatusOK, FailData(CodeEcuExeFailed, "ECU execution failed", dto.JobSucCo{JobSuc: false}))
}

// writeDeviceResult mirrors DeviceResultHelper.toDeviceResult: ecuCode
// null/"0" -> success; "132"/"138"/other -> 17002 failure (distinct messages)
// that still carries the CommandResult. Shared by EcuJobQuery and EcuCommand.
func writeDeviceResult(c *gin.Context, cr dto.CommandResult, err error) {
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, cr))
		return
	}
	switch cr.EcuCodeValue() {
	case "", "0":
		c.JSON(http.StatusOK, Ok(cr))
	case "132":
		c.JSON(http.StatusOK, FailData(CodeEcuExeFailed, "车辆移动中，不能还车", cr))
	case "138":
		c.JSON(http.StatusOK, FailData(CodeEcuExeFailed, "操作过于频繁", cr))
	default:
		c.JSON(http.StatusOK, FailData(CodeEcuExeFailed, "ECU execution failed", cr))
	}
}
