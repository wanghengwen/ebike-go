package controller

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
	domainsvc "ebike-fence-go/internal/domain/service"
)

func parkingHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/parking/getById":                       handleParkingGetByID,
		"/parking/getByIds":                      handleParkingGetByIDs,
		"/parking/getByCarId":                    handleParkingGetByCarID,
		"/parking/getList":                       handleParkingGetList,
		"/parking/getListByServiceId":            handleParkingGetListByServiceID,
		"/parking/getListByServiceIdSimple":      handleParkingGetListByServiceIDSimple,
		"/parking/getPageListByServiceId":        handleParkingGetPageListByServiceID,
		"/parking/createParking":                 handleParkingCreate,
		"/parking/createParkingBatch":            handleParkingCreateBatch,
		"/parking/updateParking":                 handleParkingUpdate,
		"/parking/updateParkingBatch":            handleParkingUpdateBatch,
		"/parking/deleteParking":                 handleParkingDelete,
		"/parking/enable":                        handleParkingEnable,
		"/parking/disable":                       handleParkingDisable,
		"/parking/getNearParking":                handleParkingGetNear,
		"/parking/getByLocations":                handleParkingGetByLocations,
		"/parking/deleteParkingBatch":            handleParkingDeleteBatch,
		"/parking/enableBatch":                   handleParkingEnableBatch,
		"/parking/disableBatch":                  handleParkingDisableBatch,
		"/parking/bindParking":                   handleParkingBind,
		"/parking/parkingCarCount":               handleParkingCarCount,
		"/parking/copy":                          handleParkingCopy,
		"/parking/unBindParking":                 handleParkingUnBind,
		"/parking/page":                          handleParkingPage,
		"/parking/getNearParkingNum":             handleParkingGetNearNum,
		"/parking/parkingCarCountByLocation":     handleParkingCarCountByLocation,
		"/parking/inParkInfoByLocation":          handleParkingInParkInfo,
		"/parking/parkingDetailByMaintainAreaId": handleParkingDetailByMaintainArea,
		"/parking/refTags":                       handleParkingRefTags,
		"/parking/stationCars":                   handleParkingStationCars,
		"/parking/filterParkingList":             handleParkingFilterList,
		"/parking/getPageOrderBy":                handleParkingGetPageOrderBy,
		"/parking/copyByService":                 handleParkingCopyByService,
		"/parking/getParkingMonitorPage":         handleParkingMonitorPage,
		"/parking/getParkingMonitorList":         handleParkingMonitorList,
		"/parking/getParkingMonitorDetail":       handleParkingMonitorDetail,
		"/parking/createRadPacketParking":        handleParkingCreateRadPacket,
	}
}

func handleParkingGetByID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.GetByIDDetail(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleParkingGetByIDs(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetByIDs(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, res, err)
}

func handleParkingGetByCarID(c *gin.Context) {
	initGlobals()
	var req dto.CarIdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := domainsvc.GetParkingByCarId(c.Request.Context(), tenantID, req.Id, globalParkingDetailGateway)
	respondAdmin(c, res, err)
}

func handleParkingGetList(c *gin.Context) {
	var req dto.ParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetListByCmd(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingGetListByServiceID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.GetListByServiceID(c.Request.Context(), tenantID, *req.Id, true)
	respondAdmin(c, res, err)
}

func handleParkingGetListByServiceIDSimple(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.GetListByServiceID(c.Request.Context(), tenantID, *req.Id, false)
	respondAdmin(c, res, err)
}

func handleParkingGetPageListByServiceID(c *gin.Context) {
	var req dto.IdPageCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.GetPageListByServiceID(c.Request.Context(), tenantID, *req.Id, req.Name, req.AreaSize, req.FieldList, req.PageNum, req.PageSize)
	respondAdmin(c, res, err)
}

func handleParkingCreate(c *gin.Context) {
	var req dto.ParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

func handleParkingCreateBatch(c *gin.Context) {
	var req dto.CreateParkingBatchCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.CreateBatch(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

func handleParkingUpdate(c *gin.Context) {
	var req dto.ParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := parkingAdmin.Update(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleParkingUpdateBatch(c *gin.Context) {
	var req dto.UpdateParkingBatchCMD
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := parkingAdmin.UpdateBatch(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleParkingDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := parkingAdmin.Delete(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, nil, err)
}

func handleParkingEnable(c *gin.Context) {
	parkingSetEnable(c, true)
}

func handleParkingDisable(c *gin.Context) {
	parkingSetEnable(c, false)
}

func parkingSetEnable(c *gin.Context, enable bool) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := parkingAdmin.SetEnable(c.Request.Context(), tenantID, cmdPin(c), *req.Id, enable)
	respondAdmin(c, nil, err)
}

func handleParkingGetNear(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetNearParking(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingGetByLocations(c *gin.Context) {
	var req dto.LocationsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetByLocations(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingDeleteBatch(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := parkingAdmin.DeleteBatch(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, nil, err)
}

func handleParkingEnableBatch(c *gin.Context) {
	parkingSetEnableBatch(c, true)
}

func handleParkingDisableBatch(c *gin.Context) {
	parkingSetEnableBatch(c, false)
}

func parkingSetEnableBatch(c *gin.Context, enable bool) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := parkingAdmin.SetEnableBatch(c.Request.Context(), tenantID, cmdPin(c), req.Ids, enable)
	respondAdmin(c, nil, err)
}

func handleParkingBind(c *gin.Context) {
	var req dto.ParkingBindCarCmd
	if !bindJSONCmd(c, &req, map[string]string{
		"commandContext": "must not be null",
		"imei":           "must not be blank",
		"lat":            "纬度不能为空",
		"lng":            "经度不能为空",
	}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.BindParking(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingCarCount(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.ParkingCarCount(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleParkingCopy(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	// Java parking/copy: id is the service-area id; returns Result.success() (null data).
	err := parkingAdmin.Copy(c.Request.Context(), tenantID, cmdPin(c), *req.Id)
	respondAdmin(c, nil, err)
}

func handleParkingUnBind(c *gin.Context) {
	var req dto.ParkingUnBindCarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.UnBindParking(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingPage(c *gin.Context) {
	var req dto.ParkingPageQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.Page(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingGetNearNum(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetNearParkingNum(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingCarCountByLocation(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.ParkingCarCountByLocation(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingInParkInfo(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.InParkInfoByLocation(c.Request.Context(), tenantID, req)
	if err != nil {
		respondAdmin(c, nil, err)
		return
	}
	if res == nil {
		respondAdmin(c, nil, nil)
		return
	}
	// Align empty FenceCO strings / zero distance with Java Jackson nulls.
	respondAdmin(c, inParkInfoJavaShape(res), nil)
}

func handleParkingDetailByMaintainArea(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.ParkingDetailByMaintainAreaId(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleParkingRefTags(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.RefTags(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleParkingStationCars(c *gin.Context) {
	var req dto.ParkStationCarsPageQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.StationCars(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingFilterList(c *gin.Context) {
	var req dto.ParkingStationFilterCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.FilterParkingList(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingGetPageOrderBy(c *gin.Context) {
	var req dto.ParkingPageOrderByQry
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetPageOrderBy(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingCopyByService(c *gin.Context) {
	var req dto.CopyByServiceCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	// Java ParkingController.copyByServiceId returns Result.success() (null data).
	_, err := parkingAdmin.CopyByService(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleParkingMonitorPage(c *gin.Context) {
	var req dto.ParkingMonitorPageQry
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetParkingMonitorPage(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingMonitorList(c *gin.Context) {
	var req dto.ParkingMonitorPageQry
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.GetParkingMonitorList(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleParkingMonitorDetail(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := parkingAdmin.GetParkingMonitorDetail(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleParkingCreateRadPacket(c *gin.Context) {
	var req dto.CreateRadPacketParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := parkingAdmin.CreateRadPacketParking(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

// inParkInfoJavaShape rewrites Go ParkingCO JSON so empty FenceCO strings and zero
// distance match Java Jackson nulls produced by ConvertorHelper on Redis entities.
//
// UseNumber is required: encoding/json into map[string]interface{} stores numbers as
// float64 and corrupts snowflake IDs above 2^53 (e.g. ...267002 → ...267000).
func inParkInfoJavaShape(co *dto.ParkingCO) map[string]interface{} {
	b, err := json.Marshal(co)
	if err != nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil {
		return nil
	}
	for _, k := range []string{
		"pics", "createdPin", "createdAt", "updatedPin", "updatedAt", "tenantId",
		"carCount", "refTags", "distance", "fullCar",
	} {
		m[k] = nil
	}
	return m
}
