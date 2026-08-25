package fence

import (
	"context"
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Request DTOs (FenceController)
// ---------------------------------------------------------------------------

// fenceIdDTO mirrors FenceIdDTO (@NotNull id) + ClientDTO (@NotEmpty traceId/tenantId).
type fenceIdDTO struct {
	dto.ClientDTO
	Id *int64 `json:"id" binding:"required"`
}

// fenceTypeDTO mirrors FenceTypeDTO extends FenceIdDTO.
type fenceTypeDTO struct {
	dto.ClientDTO
	Id    *int64 `json:"id" binding:"required"`
	Types []int  `json:"types"`
}

// locationDTO mirrors LocationDTO (@NotNull lat/lng).
type locationDTO struct {
	dto.ClientDTO
	Lat *float64 `json:"lat" binding:"required"`
	Lng *float64 `json:"lng" binding:"required"`
}

// nearLocationInner carries the nested LocationDTO of NearLocationDTO; the
// nested object itself is @NotNull but its fields are NOT validated (no @Valid).
type nearLocationInner struct {
	Lat *float64 `json:"lat"`
	Lng *float64 `json:"lng"`
}

// nearLocationDTO mirrors NearLocationDTO (@NotNull serviceId/locationDTO).
type nearLocationDTO struct {
	dto.ClientDTO
	ServiceId   *int64             `json:"serviceId" binding:"required"`
	LocationDTO *nearLocationInner `json:"locationDTO" binding:"required"`
	Radius      *float64           `json:"radius"`
}

// parkingDTO mirrors ParkingDTO (and its parent FenceDTO). The endpoints use
// group validation (@Validated(CreateGroup)/(UpdateGroup)), under which the
// Default-group constraints of ClientDTO are NOT enforced, so the loose
// client DTO is embedded and required checks are done per-handler.
// pointList is bound raw so the client's number tokens are forwarded verbatim
// (Java re-serializes List<List<Double>> with fastjson into a JSON string).
type parkingDTO struct {
	clientLoose
	Id                     *int64           `json:"id"`
	Name                   *string          `json:"name"`
	ShapeType              *string          `json:"shapeType"`
	CenterLat              *float64         `json:"centerLat"`
	CenterLng              *float64         `json:"centerLng"`
	PointList              *json.RawMessage `json:"pointList"`
	Pics                   *string          `json:"pics"`
	MaxParkingNumber       *int             `json:"maxParkingNumber"`
	ServiceId              *int64           `json:"serviceId"`
	Tbeacon                *bool            `json:"tbeacon"`
	Directional            *bool            `json:"directional"`
	Direction              *float64         `json:"direction"`
	Rfid                   *bool            `json:"rfid"`
	IzEnable               *bool            `json:"izEnable"`
	CoefficientOfDifficult *float64         `json:"coefficientOfDifficult"`
	BufferDistance         *float64         `json:"bufferDistance"`
	Camera                 *bool            `json:"camera"`
}

// noParkingDTO mirrors NoParkingDTO extends FenceDTO.
type noParkingDTO struct {
	clientLoose
	Id        *int64           `json:"id"`
	Name      *string          `json:"name"`
	ShapeType *string          `json:"shapeType"`
	CenterLat *float64         `json:"centerLat"`
	CenterLng *float64         `json:"centerLng"`
	PointList *json.RawMessage `json:"pointList"`
	Pics      *string          `json:"pics"`
	// NoParkingDTO.serviceId is @NotNull WITHOUT a group, so it is NOT
	// validated under @Validated(CreateGroup)/(UpdateGroup).
	ServiceId *int64 `json:"serviceId"`
}

// checkPointListType mirrors Jackson failing to bind a non-array pointList
// (HttpMessageNotReadableException -> 00002).
func checkPointListType(c *gin.Context, raw *json.RawMessage) bool {
	if raw == nil || isJSONNull(*raw) {
		return true
	}
	var v [][]float64
	if err := json.Unmarshal(*raw, &v); err != nil {
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
		return false
	}
	return true
}

// pointListString mirrors FenceDTO.getPointList(): the bound List<List<Double>>
// is re-serialized into a JSON string for the downstream Cmd; a null list
// becomes the literal string "null" (fastjson JSONObject.toJSONString(null)).
func pointListString(raw *json.RawMessage) string {
	if raw == nil || isJSONNull(*raw) {
		return "null"
	}
	return string(*raw)
}

// Validation messages (hibernate-validator English defaults).
var fenceMsgs = map[string]string{
	"traceId":     "must not be empty",
	"tenantId":    "must not be empty",
	"id":          "must not be null",
	"lat":         "must not be null",
	"lng":         "must not be null",
	"serviceId":   "must not be null",
	"locationDTO": "must not be null",
}

// ---------------------------------------------------------------------------
// Downstream payloads
// ---------------------------------------------------------------------------

// idCmd mirrors CommandContextHolder.genParam(IdCmd.class) + setId(...): only
// the id field is populated.
type idCmd struct {
	Id *int64 `json:"id"`
}

// idRawCmd is idCmd with a passthrough id token (used when the id comes from a
// downstream Long field).
type idRawCmd struct {
	Id *jLong `json:"id"`
}

type locationCmd struct {
	Lat *float64 `json:"lat"`
	Lng *float64 `json:"lng"`
}

// nearLocationCmd mirrors com.xyy.ebike.fence.api.dto.packing.NearLocationCmd.
type nearLocationCmd struct {
	ServiceId   *int64       `json:"serviceId"`
	LocationCmd *locationCmd `json:"locationCmd"`
	IzClient    *bool        `json:"izClient,omitempty"`
	Radius      *float64     `json:"radius"`
}

// parkingCmdPayload mirrors Convertor.getParkCmdByDTO: BeanUtils same-name copy
// of ParkingDTO plus pointList in JSON-string form.
type parkingCmdPayload struct {
	Id                     *int64   `json:"id"`
	Name                   *string  `json:"name"`
	ShapeType              *string  `json:"shapeType"`
	CenterLat              *float64 `json:"centerLat"`
	CenterLng              *float64 `json:"centerLng"`
	PointList              string   `json:"pointList"`
	Pics                   *string  `json:"pics"`
	MaxParkingNumber       *int     `json:"maxParkingNumber"`
	ServiceId              *int64   `json:"serviceId"`
	Tbeacon                *bool    `json:"tbeacon"`
	Directional            *bool    `json:"directional"`
	Direction              *float64 `json:"direction"`
	Rfid                   *bool    `json:"rfid"`
	IzEnable               *bool    `json:"izEnable"`
	CoefficientOfDifficult *float64 `json:"coefficientOfDifficult"`
	BufferDistance         *float64 `json:"bufferDistance"`
	Camera                 *bool    `json:"camera"`
}

type noParkingCmdPayload struct {
	Id        *int64   `json:"id"`
	Name      *string  `json:"name"`
	ShapeType *string  `json:"shapeType"`
	CenterLat *float64 `json:"centerLat"`
	CenterLng *float64 `json:"centerLng"`
	PointList string   `json:"pointList"`
	Pics      *string  `json:"pics"`
	ServiceId *int64   `json:"serviceId"`
}

// ---------------------------------------------------------------------------
// Fetch helpers
// ---------------------------------------------------------------------------

// fetchFenceList performs a downstream call that Java wraps in
// ResultHelper.getResultData (failure -> BizException with downstream code/msg).
// Returns (list, finished); when finished is true a response was written.
func fetchFenceList(c *gin.Context, ctx context.Context, path string, payload interface{}, cmdCtx *dto.CommandContext) ([]*fenceCOIn, bool) {
	result, err := rpc.ForwardCommand(ctx, rpc.ServiceFence, path, payload, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return nil, true
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return nil, true
	}
	if isJSONNull(result.Data) {
		return nil, false
	}
	var cos []*fenceCOIn
	if err := json.Unmarshal(result.Data, &cos); err != nil {
		web.WriteException(c, err.Error())
		return nil, true
	}
	return cos, false
}

// fetchServiceAreaOne ports ServiceAreaGatewayImpl.getOne: any failure
// (transport error or unsuccessful Result) yields an empty ServiceAreaCO with
// type=1; a successful null payload yields nil.
func fetchServiceAreaOne(ctx context.Context, id *int64, cmdCtx *dto.CommandContext) *fenceCOIn {
	emptyCO := &fenceCOIn{Type: json.RawMessage("1")}
	result, err := rpc.ForwardCommand(ctx, rpc.ServiceFence, "/serviceArea/getById", idCmd{Id: id}, cmdCtx)
	if err != nil || !result.Success {
		return emptyCO
	}
	if isJSONNull(result.Data) {
		return nil
	}
	var co fenceCOIn
	if err := json.Unmarshal(result.Data, &co); err != nil {
		return emptyCO
	}
	return &co
}

// ---------------------------------------------------------------------------
// ServiceArea handlers
// ---------------------------------------------------------------------------

// getAllServiceAreas ports FenceController.getAll (POST /client/fence/serviceArea/getAll).
// No @Validated on the Java handler -> no validation.
// Downstream: ebike-fence POST /serviceArea/getList with {commandContext} only.
func getAllServiceAreas(c *gin.Context) {
	var req clientLoose
	if !web.BindJSON(c, &req, nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.client())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/serviceArea/getList", struct{}{}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	if isJSONNull(result.Data) {
		// ConvertorHelper.convert(null, ...) -> null
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	var cos []*fenceCOIn
	if err := json.Unmarshal(result.Data, &cos); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	out := make([]*serviceAreaClientO, 0, len(cos))
	for _, co := range cos {
		o, convErr := toServiceAreaClientO(co)
		if convErr != nil {
			web.WriteException(c, convErr.Error())
			return
		}
		out = append(out, o)
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// getServiceAreaByLocation ports FenceController.getServiceAreaByLocation
// (POST /client/fence/serviceArea/getByLocation).
// Downstream: ebike-fence POST /serviceArea/getNearServiceByLocation; the
// commandContext tenantId is overridden with the request-body tenantId
// (ServiceAreaGatewayImpl.getByLocation), since this path is login-exempt.
func getServiceAreaByLocation(c *gin.Context) {
	var req locationDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmdCtx.TenantId = req.TenantId
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/serviceArea/getNearServiceByLocation",
		locationCmd{Lat: req.Lat, Lng: req.Lng}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	if isJSONNull(result.Data) {
		// Java: Convertor.copyServiceToClientO(null) -> NPE -> 00001
		writeNPE(c)
		return
	}
	var co fenceCOIn
	if err := json.Unmarshal(result.Data, &co); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	o, convErr := toServiceAreaClientO(&co)
	if convErr != nil {
		web.WriteException(c, convErr.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(o))
}

// getFenceByServiceId ports FenceController.getFenceByServiceId
// (POST /client/fence/serviceArea/getFenceByServiceId).
// Aggregates serviceArea(1)/parking(2)/noParking(4)/banRiding(9) fences,
// mirroring ServiceAreaServiceImpl.getFenceByServiceId.
func getFenceByServiceId(c *gin.Context) {
	var req fenceTypeDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	ctx := c.Request.Context()

	var fences fencesCO

	addServiceArea := func() bool {
		co := fetchServiceAreaOne(ctx, req.Id, cmdCtx)
		if co == nil {
			fences.ServiceAreas = nil
			return true
		}
		o, convErr := toServiceAreaClientO(co)
		if convErr != nil {
			web.WriteException(c, convErr.Error())
			return false
		}
		fences.ServiceAreas = []*serviceAreaClientO{o}
		return true
	}
	addParkings := func() bool {
		cos, finished := fetchFenceList(c, ctx, "/parking/getListByServiceId", idCmd{Id: req.Id}, cmdCtx)
		if finished {
			return false
		}
		if len(cos) == 0 {
			// ParkingServiceImpl returns null for an empty list
			fences.Parkings = nil
			return true
		}
		out := make([]*parkingClientO, 0, len(cos))
		for _, co := range cos {
			o, convErr := toParkingClientO(co)
			if convErr != nil {
				web.WriteException(c, convErr.Error())
				return false
			}
			out = append(out, o)
		}
		fences.Parkings = out
		return true
	}
	addNoParkings := func() bool {
		cos, finished := fetchFenceList(c, ctx, "/noParking/getListByServiceId", idCmd{Id: req.Id}, cmdCtx)
		if finished {
			return false
		}
		if len(cos) == 0 {
			fences.NoParkings = nil
			return true
		}
		out := make([]*noParkingClientO, 0, len(cos))
		for _, co := range cos {
			o, convErr := toNoParkingClientO(co)
			if convErr != nil {
				web.WriteException(c, convErr.Error())
				return false
			}
			out = append(out, o)
		}
		fences.NoParkings = out
		return true
	}
	addBanRidings := func() bool {
		cos, finished := fetchFenceList(c, ctx, "/banRiding/getListByServiceId", idCmd{Id: req.Id}, cmdCtx)
		if finished {
			return false
		}
		// BanRidingServiceImpl returns an empty list (not null) when empty
		out := make([]*banRidingClientO, 0, len(cos))
		for _, co := range cos {
			o, convErr := toBanRidingClientO(co)
			if convErr != nil {
				web.WriteException(c, convErr.Error())
				return false
			}
			out = append(out, o)
		}
		fences.BanRidings = out
		return true
	}

	if len(req.Types) == 0 {
		if !addServiceArea() || !addParkings() || !addNoParkings() || !addBanRidings() {
			return
		}
	} else {
		for _, t := range req.Types {
			switch t {
			case 1:
				if !addServiceArea() {
					return
				}
			case 2:
				if !addParkings() {
					return
				}
			case 4:
				if !addNoParkings() {
					return
				}
			case 9:
				if !addBanRidings() {
					return
				}
			}
		}
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(fences))
}

// getServiceAreaById ports FenceController.getServiceAreaById
// (POST /client/fence/serviceArea/getServiceAreaById).
// ServiceAreaGatewayImpl.getOne swallows every failure and returns an empty
// ServiceAreaCO (type=1); a successful null payload triggers the Java NPE in
// Convertor (00001).
func getServiceAreaById(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	co := fetchServiceAreaOne(c.Request.Context(), req.Id, cmdCtx)
	if co == nil {
		writeNPE(c)
		return
	}
	o, convErr := toServiceAreaClientO(co)
	if convErr != nil {
		web.WriteException(c, convErr.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(o))
}

// getNearFence ports FenceController.getNearFence
// (POST /client/fence/serviceArea/getNearFence).
// Downstream: /parking/getNearParking (izClient=true) then
// /noParking/getNearNoParking; both with commandContext tenantId taken from
// the request body (login-exempt path).
func getNearFence(c *gin.Context) {
	var req nearLocationDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmdCtx.TenantId = req.TenantId
	ctx := c.Request.Context()
	loc := &locationCmd{}
	if req.LocationDTO != nil {
		loc.Lat = req.LocationDTO.Lat
		loc.Lng = req.LocationDTO.Lng
	}

	izClient := true
	var fences fencesCO

	// parkings first (ServiceAreaServiceImpl.getNearFence order)
	cos, finished := fetchFenceList(c, ctx, "/parking/getNearParking",
		nearLocationCmd{ServiceId: req.ServiceId, LocationCmd: loc, IzClient: &izClient, Radius: req.Radius}, cmdCtx)
	if finished {
		return
	}
	if cos != nil {
		out := make([]*parkingClientO, 0, len(cos))
		for _, co := range cos {
			o, convErr := toParkingClientO(co)
			if convErr != nil {
				web.WriteException(c, convErr.Error())
				return
			}
			out = append(out, o)
		}
		fences.Parkings = out
	}

	cos, finished = fetchFenceList(c, ctx, "/noParking/getNearNoParking",
		nearLocationCmd{ServiceId: req.ServiceId, LocationCmd: loc, Radius: req.Radius}, cmdCtx)
	if finished {
		return
	}
	if cos != nil {
		out := make([]*noParkingClientO, 0, len(cos))
		for _, co := range cos {
			o, convErr := toNoParkingClientO(co)
			if convErr != nil {
				web.WriteException(c, convErr.Error())
				return
			}
			out = append(out, o)
		}
		fences.NoParkings = out
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(fences))
}

// ---------------------------------------------------------------------------
// Parking handlers
// ---------------------------------------------------------------------------

// getParkingById ports FenceController.getParkingById (POST /client/fence/parking/getById).
// Downstream: /parking/getById.
func getParkingById(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/parking/getById", idCmd{Id: req.Id}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	if isJSONNull(result.Data) {
		writeNPE(c) // Convertor.copyParkToClientO(null) -> NPE
		return
	}
	var co fenceCOIn
	if err := json.Unmarshal(result.Data, &co); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	o, convErr := toParkingClientO(&co)
	if convErr != nil {
		web.WriteException(c, convErr.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(o))
}

// getParkingByServiceId ports FenceController.getParkingByServiceId
// (POST /client/fence/parking/getByServiceId). Downstream: /parking/getListByServiceId.
// An empty list becomes null (ParkingServiceImpl.getByServiceId).
func getParkingByServiceId(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cos, finished := fetchFenceList(c, c.Request.Context(), "/parking/getListByServiceId", idCmd{Id: req.Id}, cmdCtx)
	if finished {
		return
	}
	if len(cos) == 0 {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	out := make([]*parkingClientO, 0, len(cos))
	for _, co := range cos {
		o, convErr := toParkingClientO(co)
		if convErr != nil {
			web.WriteException(c, convErr.Error())
			return
		}
		out = append(out, o)
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// createParking ports FenceController.createParking
// (POST /client/fence/parking/createParking, @Validated(CreateGroup)).
// CreateGroup constraints: name/shapeType/centerLat/centerLng/pointList/serviceId.
// Downstream: /parking/createParking, response passthrough (Result<Long>).
func createParking(c *gin.Context) {
	var req parkingDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if !checkPointListType(c, req.PointList) {
		return
	}
	if !requireNotNull(c, "name", req.Name != nil) ||
		!requireNotNull(c, "shapeType", req.ShapeType != nil) ||
		!requireNotNull(c, "centerLat", req.CenterLat != nil) ||
		!requireNotNull(c, "centerLng", req.CenterLng != nil) ||
		!requireNotNull(c, "pointList", req.PointList != nil && !isJSONNull(*req.PointList)) ||
		!requireNotNull(c, "serviceId", req.ServiceId != nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.client())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/parking/createParking", parkingPayload(&req), cmdCtx)
	web.RespondResult(c, result, err)
}

// updateParking ports FenceController.updateParking
// (POST /client/fence/parking/updateParking, @Validated(UpdateGroup): only id).
// Downstream: /parking/updateParking; Java returns ResultHelper.success() (data null).
func updateParking(c *gin.Context) {
	var req parkingDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if !checkPointListType(c, req.PointList) {
		return
	}
	if !requireNotNull(c, "id", req.Id != nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.client())
	respondVoidSuccess(c, rpc.ServiceFence, "/parking/updateParking", parkingPayload(&req), cmdCtx)
}

func parkingPayload(req *parkingDTO) parkingCmdPayload {
	return parkingCmdPayload{
		Id: req.Id, Name: req.Name, ShapeType: req.ShapeType,
		CenterLat: req.CenterLat, CenterLng: req.CenterLng,
		PointList: pointListString(req.PointList), Pics: req.Pics,
		MaxParkingNumber: req.MaxParkingNumber, ServiceId: req.ServiceId,
		Tbeacon: req.Tbeacon, Directional: req.Directional, Direction: req.Direction,
		Rfid: req.Rfid, IzEnable: req.IzEnable,
		CoefficientOfDifficult: req.CoefficientOfDifficult, BufferDistance: req.BufferDistance,
		Camera: req.Camera,
	}
}

// respondVoidSuccess handles the Java pattern
// "gateway call via getResultData (discarding data); controller returns
// ResultHelper.success()": success -> data null, failure -> downstream code/msg.
func respondVoidSuccess(c *gin.Context, service, path string, payload interface{}, cmdCtx *dto.CommandContext) {
	result, err := rpc.ForwardCommand(c.Request.Context(), service, path, payload, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

// deleteParking ports FenceController.deleteParking (POST /client/fence/parking/deleteParking).
func deleteParking(c *gin.Context) {
	parkingIdAction(c, "/parking/deleteParking")
}

// enableParking ports FenceController.enableParking (POST /client/fence/parking/enable).
func enableParking(c *gin.Context) {
	parkingIdAction(c, "/parking/enable")
}

// disableParking ports FenceController.disableParking (POST /client/fence/parking/disable).
func disableParking(c *gin.Context) {
	parkingIdAction(c, "/parking/disable")
}

func parkingIdAction(c *gin.Context, path string) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	respondVoidSuccess(c, rpc.ServiceFence, path, idCmd{Id: req.Id}, cmdCtx)
}

// getNearParkingNum ports FenceController.getNearParkingNum
// (POST /client/fence/parking/nearParkingNum).
// Downstream: /parking/getNearParkingNum (izClient=true, body tenantId).
func getNearParkingNum(c *gin.Context) {
	var req nearLocationDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmdCtx.TenantId = req.TenantId
	loc := &locationCmd{}
	if req.LocationDTO != nil {
		loc.Lat = req.LocationDTO.Lat
		loc.Lng = req.LocationDTO.Lng
	}
	izClient := true
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/parking/getNearParkingNum",
		nearLocationCmd{ServiceId: req.ServiceId, LocationCmd: loc, IzClient: &izClient, Radius: req.Radius}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	if isJSONNull(result.Data) {
		// ConvertorHelper.copyProperties(null, ...) -> null
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	var co nearParkingClientO
	if err := json.Unmarshal(result.Data, &co); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// ---------------------------------------------------------------------------
// NoParking handlers
// ---------------------------------------------------------------------------

// getNoParkingById ports FenceController.getNoParkingById (POST /client/fence/noParking/getById).
func getNoParkingById(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/noParking/getById", idCmd{Id: req.Id}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		c.JSON(http.StatusOK, failResult(result))
		return
	}
	if isJSONNull(result.Data) {
		writeNPE(c)
		return
	}
	var co fenceCOIn
	if err := json.Unmarshal(result.Data, &co); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	o, convErr := toNoParkingClientO(&co)
	if convErr != nil {
		web.WriteException(c, convErr.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(o))
}

// getNoParkingByServiceId ports FenceController.getNoParkingByServiceId
// (POST /client/fence/noParking/getByServiceId). Empty list -> null.
func getNoParkingByServiceId(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cos, finished := fetchFenceList(c, c.Request.Context(), "/noParking/getListByServiceId", idCmd{Id: req.Id}, cmdCtx)
	if finished {
		return
	}
	if len(cos) == 0 {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	out := make([]*noParkingClientO, 0, len(cos))
	for _, co := range cos {
		o, convErr := toNoParkingClientO(co)
		if convErr != nil {
			web.WriteException(c, convErr.Error())
			return
		}
		out = append(out, o)
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// createNoParking ports FenceController.createNoParking
// (POST /client/fence/noParking/createNoParking, @Validated(CreateGroup)).
// Note: NoParkingDTO.serviceId is a Default-group constraint -> NOT validated here.
func createNoParking(c *gin.Context) {
	var req noParkingDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if !checkPointListType(c, req.PointList) {
		return
	}
	if !requireNotNull(c, "name", req.Name != nil) ||
		!requireNotNull(c, "shapeType", req.ShapeType != nil) ||
		!requireNotNull(c, "centerLat", req.CenterLat != nil) ||
		!requireNotNull(c, "centerLng", req.CenterLng != nil) ||
		!requireNotNull(c, "pointList", req.PointList != nil && !isJSONNull(*req.PointList)) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.client())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/noParking/createNoParking", noParkingPayload(&req), cmdCtx)
	web.RespondResult(c, result, err)
}

// updateNoParking ports FenceController.updateNoParking
// (POST /client/fence/noParking/updateNoParking, @Validated(UpdateGroup): only id).
func updateNoParking(c *gin.Context) {
	var req noParkingDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if !checkPointListType(c, req.PointList) {
		return
	}
	if !requireNotNull(c, "id", req.Id != nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, req.client())
	respondVoidSuccess(c, rpc.ServiceFence, "/noParking/updateNoParking", noParkingPayload(&req), cmdCtx)
}

func noParkingPayload(req *noParkingDTO) noParkingCmdPayload {
	return noParkingCmdPayload{
		Id: req.Id, Name: req.Name, ShapeType: req.ShapeType,
		CenterLat: req.CenterLat, CenterLng: req.CenterLng,
		PointList: pointListString(req.PointList), Pics: req.Pics,
		ServiceId: req.ServiceId,
	}
}

// deleteNoParking ports FenceController.deleteNoParking
// (POST /client/fence/noParking/deleteNoParking).
func deleteNoParking(c *gin.Context) {
	var req fenceIdDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	respondVoidSuccess(c, rpc.ServiceFence, "/noParking/deleteNoParking", idCmd{Id: req.Id}, cmdCtx)
}
