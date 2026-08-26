package api

import (
	"errors"
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerDeviceTrajectoryRoutes attaches the DeviceTrajectory endpoints.
//
// Each forwards to the ebike-device-worker gateway (trajectory / metric), so
// (like the other forwards) they are NOT shadow-compared.
func registerDeviceTrajectoryRoutes(r *gin.Engine) {
	r.POST("/device/trajectory/realTime", trajectoryRealTime)
	r.POST("/device/trajectory/history/batch", trajectoryHistoryBatch)
	r.POST("/device/trajectory/history", trajectoryHistory)
	r.POST("/device/trajectory/saveDb", trajectorySaveDb)
	r.POST("/device/trajectory/distance", getTrajectoryDistance)
	r.POST("/device/trajectory/getMetric", getMetric)
}

func traceOf(cc *dto.CommandContext) string {
	if cc == nil {
		return ""
	}
	return cc.TraceID
}

// trajectoryRealTime ports DeviceTrajectoryController.trajectoryRealTime.
func trajectoryRealTime(c *gin.Context) {
	var q dto.TrajectoryRealTimeQry
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	list, err := service.TrajectoryRealTime(traceOf(q.CommandContext), q)
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(list))
}

// trajectoryHistory ports DeviceTrajectoryController.trajectoryHistory.
func trajectoryHistory(c *gin.Context) {
	var q dto.TrajectoryHistoryQry
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	list, err := service.TrajectoryHistory(traceOf(q.CommandContext), q)
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(list))
}

// trajectoryHistoryBatch ports DeviceTrajectoryController.trajectoryHistoryBatch.
func trajectoryHistoryBatch(c *gin.Context) {
	var q dto.TrajectoryHistoryBatchQry
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	m, err := service.TrajectoryHistoryBatch(traceOf(q.CommandContext), q)
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(m))
}

// getTrajectoryDistance ports DeviceTrajectoryController.getTrajectoryDistance
// (errors degrade to distance 0 inside the service).
func getTrajectoryDistance(c *gin.Context) {
	var q dto.TrajectoryRealTimeQry
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	c.JSON(http.StatusOK, Ok(service.GetTrajectoryDistance(traceOf(q.CommandContext), q)))
}

// getMetric ports DeviceTrajectoryController.getMetric.
func getMetric(c *gin.Context) {
	var q dto.MetricQry
	if !bindBody(c, &q) {
		return
	}
	tenantID := resolveTenant(c, q.CommandContext)
	m, err := service.GetMetric(tenantID, traceOf(q.CommandContext), q)
	if err != nil {
		if errors.Is(err, service.ErrCarImeiBind) {
			c.JSON(http.StatusOK, Fail(CodeImeiCarIdBind, "请输入正确的车辆号"))
			return
		}
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(m))
}

// trajectorySaveDb ports DeviceTrajectoryController.trajectorySaveDb: a worker
// write (save trajectory to db). NOT shadow-compared.
func trajectorySaveDb(c *gin.Context) {
	var cmd dto.TrajectoryDbCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	if err := service.TrajectorySaveDb(traceOf(cmd.CommandContext), cmd); err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(true))
}
