package controller

import (
	"ebike-device-worker-go/internal/api"
	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/service"

	"github.com/gin-gonic/gin"
)

// InitNativeHandlers registers all native route handlers (for tests and router setup).
func InitNativeHandlers() {
	registerNativeHandlers()
}

func registerNativeHandlers() {
	api.RegisterHandlers(map[string]gin.HandlerFunc{
		"/ebike/gps/getTrajectory":                   getTrajectory,
		"/ebike/gps/getTrajectoryDistance":           getTrajectoryDistance,
		"/ebike/gps/getMetric":                       getMetric,
		"/ebike/gps/getOrderTrajectory":              getOrderTrajectory,
		"/ebike/gps/getOrderTrajectoryDistance":      getOrderTrajectoryDistance,
		"/ebike/gps/getMovingEBikeTrajectory":        getMovingEBikeTrajectory,
		"/ebike/gps/getBatchOrderTrajectory":         getBatchOrderTrajectory,
		"/ebike/gps/getBatchOrderTrajectoryDistance": getBatchOrderTrajectoryDistance,
		"/cmd/getEbikeCmd":                           getEbikeCmd,
		"/cmd/saveEbikeCmd":                          saveEbikeCmd,
		"/ebike/gps/saveOrderTrajectory":             saveOrderTrajectory,
		"/ebike/gps/saveMovingEBikeTrajectory":       saveMovingEBikeTrajectory,
	})
}

func getTrajectory(c *gin.Context) {
	var req dto.TrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetTrajectory(req)
	})
}

func getTrajectoryDistance(c *gin.Context) {
	var req dto.TrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetTrajectoryDistance(req)
	})
}

func getMetric(c *gin.Context) {
	var req dto.MetricCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetMetric(req)
	})
}

func getOrderTrajectory(c *gin.Context) {
	var req dto.OrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetOrderTrajectory(req)
	})
}

func getOrderTrajectoryDistance(c *gin.Context) {
	var req dto.OrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetOrderTrajectoryDistance(req)
	})
}

func getMovingEBikeTrajectory(c *gin.Context) {
	var req dto.OrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetOrderTrajectory(req)
	})
}

func getBatchOrderTrajectory(c *gin.Context) {
	var req dto.BatchOrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetBatchOrderTrajectory(req)
	})
}

func getBatchOrderTrajectoryDistance(c *gin.Context) {
	var req dto.BatchOrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return service.GetBatchOrderTrajectoryDistance(req)
	})
}

func getEbikeCmd(c *gin.Context) {
	var req dto.CmdQueryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return nil, nil
	})
}

func saveEbikeCmd(c *gin.Context) {
	var req dto.DeviceCommandCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		return nil, nil
	})
}

func saveOrderTrajectory(c *gin.Context) {
	var req dto.SaveOrderTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		if err := service.SaveOrderTrajectory(req); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func saveMovingEBikeTrajectory(c *gin.Context) {
	var req dto.SaveMovingTrajectoryCmd
	api.BindAndServe(c, &req, nil, func() (interface{}, error) {
		if err := service.SaveMovingEBikeTrajectory(req); err != nil {
			return nil, err
		}
		return nil, nil
	})
}
