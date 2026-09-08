package controller

import (
	"context"
	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/web"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	globalHelmetGateway        *gateway.HelmetGateway
	globalConfigGateway        *gateway.ConfigGateway
	globalPartGateway          *gateway.PartGateway
	globalFenceRfidGateway     *gateway.FenceRfidGateway
	globalParkingDetailGateway *gateway.ParkingDetailGateway
	globalDeviceRpc            *rpc.DeviceRPC
	globalManagementRpc        *rpc.ManagementRPC
	globalTicketRpc            *rpc.TicketRPC
	globalPartService          *service.PartService
	globalHelmetService        *service.HelmetService
	globalReturnCarService     *service.ReturnCarService
	globalRidingCarService     *service.RidingCarService
	initOnce                   sync.Once
)

func initGlobals() {
	initOnce.Do(func() {
		logger, _ := zap.NewProduction()
		zap.ReplaceGlobals(logger)

		globalDeviceRpc = rpc.NewDeviceRPC()
		globalManagementRpc = rpc.NewManagementRPC()
		globalTicketRpc = rpc.NewTicketRPC()

		globalHelmetGateway = gateway.NewHelmetGateway(rpc.NewDeviceHelmetAdapter(globalDeviceRpc))
		globalConfigGateway = gateway.NewConfigGateway()
		globalPartGateway = gateway.NewPartGateway()
		globalFenceRfidGateway = gateway.NewFenceRfidGateway()
		globalParkingDetailGateway = gateway.NewParkingDetailGateway()

		initFenceAdmin()
		service.SetParkingCOByCarIDResolver(func(ctx context.Context, tenantID, carID string, detailGw *gateway.ParkingDetailGateway) (*dto.ParkingCO, error) {
			if detailGw == nil {
				detailGw = globalParkingDetailGateway
			}
			return parkingAdmin.GetByCarID(ctx, tenantID, carID, detailGw)
		})

		globalPartService = service.NewPartService(globalHelmetGateway, globalPartGateway, globalFenceRfidGateway, globalConfigGateway, globalDeviceRpc, globalManagementRpc, globalTicketRpc)
		globalHelmetService = service.NewHelmetService(globalHelmetGateway, globalManagementRpc, globalConfigGateway, globalDeviceRpc)
		globalReturnCarService = service.NewReturnCarService(globalConfigGateway, globalPartService, globalDeviceRpc, globalFenceRfidGateway)
		globalRidingCarService = service.NewRidingCarService(globalConfigGateway, globalPartService, globalParkingDetailGateway)
	})
}

func bindReturnCarCmd(c *gin.Context) (dto.ReturnCarCmd, string, bool) {
	var req dto.ReturnCarCmd
	if !web.BindJSON(c, &req, map[string]string{
		"commandContext": "must not be null",
		"serviceAreaId":  "must not be null",
		"carCmd":         "must not be null",
	}) {
		return req, "", false
	}
	if req.CarCmd == nil {
		web.WriteParamError(c, "carCmd must not be null")
		return req, "", false
	}

	cmdCtx, ok := middleware.ApplyCommand(c, &req.Command)
	if !ok {
		return req, "", false
	}
	return req, cmdCtx.TenantId, true
}

func respondReturnCar(c *gin.Context, tenantId string, req dto.ReturnCarCmd, fullReturn bool) {
	var res dto.ReturnCarCO
	var err error
	if fullReturn {
		res, err = globalReturnCarService.ReturnCar(c.Request.Context(), tenantId, req)
	} else {
		res, err = globalReturnCarService.FenceRelation(c.Request.Context(), tenantId, req)
	}
	if err != nil {
		if biz, ok := err.(*service.BizError); ok {
			web.WriteBizError(c, biz.Code, biz.Msg)
			return
		}
		zap.L().Error("return car failed", zap.Error(err), zap.String("trace_id", middleware.GetTraceIdFromCtx(c.Request.Context())))
		web.WriteException(c, "Internal Server Error")
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func logReturnCarRequest(c *gin.Context, req dto.ReturnCarCmd) {
	var orderId int64
	if req.OrderId != nil {
		orderId = *req.OrderId
	}
	traceID := middleware.GetTraceIdFromCtx(c.Request.Context())
	if req.CarCmd.ReportTime != nil {
		reportTime := *req.CarCmd.ReportTime
		staleSecs := time.Now().Unix() - reportTime
		zap.L().Info("ReturnCar request",
			zap.Int64("orderId", orderId),
			zap.String("carImei", req.CarCmd.Imei),
			zap.Int64("reportTime", reportTime),
			zap.Int64("staleSecs", staleSecs),
			zap.String("trace_id", traceID))
	} else {
		zap.L().Info("ReturnCar request missing reportTime",
			zap.Int64("orderId", orderId),
			zap.String("carImei", req.CarCmd.Imei),
			zap.String("trace_id", traceID))
	}
	zap.L().Info("ReturnCar parts", zap.Strings("parts", req.CarCmd.Parts), zap.Any("result", req.Result))
}

func ReturnCar(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindReturnCarCmd(c)
	if !ok {
		return
	}
	logReturnCarRequest(c, req)
	respondReturnCar(c, tenantId, req, true)
}

func RidingCar(c *gin.Context) {
	initGlobals()
	var req dto.RideCarCmd
	if !web.BindJSON(c, &req, map[string]string{
		"commandContext": "must not be null",
		"serviceAreaId":  "must not be null",
		"carLocation":    "must not be null",
		"carCmd":         "must not be null",
	}) {
		return
	}
	if req.CarCmd == nil {
		web.WriteParamError(c, "carCmd must not be null")
		return
	}

	cmdCtx, ok := middleware.ApplyCommand(c, &req.Command)
	if !ok {
		return
	}

	res, err := globalRidingCarService.RidingCar(c.Request.Context(), cmdCtx.TenantId, req)
	if err != nil {
		if biz, ok := err.(*service.BizError); ok {
			web.WriteBizError(c, biz.Code, biz.Msg)
			return
		}
		zap.L().Error("RidingCar failed", zap.Error(err), zap.String("trace_id", cmdCtx.TraceId))
		web.WriteException(c, "Internal Server Error")
		return
	}

	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func GetFenceRelation(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindReturnCarCmd(c)
	if !ok {
		return
	}
	respondReturnCar(c, tenantId, req, false)
}
