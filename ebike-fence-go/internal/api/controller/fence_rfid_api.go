package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func fenceRfidHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/rfid/bind/queryRFIDsByFenceId":       handleRfidQueryByFenceID,
		"/rfid/bind/queryFenceIdByRfidCache":   handleRfidQueryFenceByRfid,
		"/rfid/bind/queryFenceByRfids":         handleRfidQueryFenceByRfids,
		"/rfid/bind/saveOrUpdate":              handleRfidSaveOrUpdate,
		"/rfid/bind/change":                    handleRfidChange,
		"/rfid/bind/batchSaveOrUpdate":         handleRfidBatchSaveOrUpdate,
	}
}

func handleRfidQueryByFenceID(c *gin.Context) {
	var req dto.FenceRfidCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := rfidAdmin.QueryRFIDsByFenceID(c.Request.Context(), tenantID, req.FenceId)
	respondRaw(c, res, err)
}

func handleRfidQueryFenceByRfid(c *gin.Context) {
	var req dto.RfidCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := rfidAdmin.QueryFenceIDByRfidCache(c.Request.Context(), tenantID, req.RfidCode)
	respondRaw(c, res, err)
}

func handleRfidQueryFenceByRfids(c *gin.Context) {
	var req dto.FenceInfoByRfidsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := rfidAdmin.QueryFenceByRfids(c.Request.Context(), tenantID, req.RfidCodes)
	respondRaw(c, res, err)
}

func handleRfidSaveOrUpdate(c *gin.Context) {
	var req dto.FenceRfidSaveCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := rfidAdmin.SaveOrUpdate(c.Request.Context(), tenantID, cmdPin(c), req)
	respondRaw(c, res, err)
}

func handleRfidChange(c *gin.Context) {
	var req dto.FenceRfidChangeCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := rfidAdmin.Change(c.Request.Context(), tenantID, cmdPin(c), req)
	respondRaw(c, res, err)
}

func handleRfidBatchSaveOrUpdate(c *gin.Context) {
	// Java FenceRfidController.batchSaveOrUpdate is a no-op stub that returns null.
	//
	// NOTE: the original Go behavior delegated to the real saveOrUpdate handler.
	// Preserved here for reference:
	//
	//	handleRfidSaveOrUpdate(c)
	respondRaw(c, nil, nil)
}
