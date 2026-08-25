package management

// Ports Java GaodeMapController (@RequestMapping("/client/management/gaode")):
//   - POST /getAddress  -> ebike-management /gaode/map/getAddressByLocation
//   - POST /getLocation -> ebike-management /gaode/map/getLocationByAddress
//   - POST /navigate    -> external ${gaode.navigate.url} (plain JSON POST)
//   - POST /v2/navigate -> external map-service /map/mapNavigate
//   - POST /v3/navigate -> external map-service /map/mapNavigate
//
// GaodeMapServiceImpl / GaodeMapGatewayImpl contain real logic (entity/CO field
// mapping, swallowing downstream failures into an empty entity, polyline
// parsing) which is ported faithfully below.

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// gaodeMapDTO matches Java GaodeMapDto extends ClientDTO. No own-field
// validation annotations; @Validated only enforces ClientDTO's traceId/tenantId.
type gaodeMapDTO struct {
	dto.ClientDTO
	Id         *javacompat.LongStr  `json:"id"`
	Address    *string              `json:"address"`
	Country    *string              `json:"country"`
	Province   *string              `json:"province"`
	CityCode   *string              `json:"cityCode"`
	City       *string              `json:"city"`
	AdCode     *string              `json:"adCode"`
	District   *string              `json:"district"`
	TownCode   *string              `json:"townCode"`
	Township   *string              `json:"township"`
	SeaArea    *string              `json:"seaArea"`
	Remarks    *string              `json:"remarks"`
	UpdateTime *javacompat.DateTime `json:"updateTime"`
}

// gaodeMapCmd matches the downstream GaodeMapCmd (sent to ebike-management).
// All fields are emitted (Java serializes nulls); id is Long -> string via the
// global ToStringSerializer, updateTime -> "yyyy-MM-dd HH:mm:ss".
// commandContext is injected by rpc.ForwardCommand.
type gaodeMapCmd struct {
	Id         *javacompat.LongStr  `json:"id"`
	Longitude  *string              `json:"longitude"`
	Dimension  *string              `json:"dimension"`
	Address    *string              `json:"address"`
	Country    *string              `json:"country"`
	Province   *string              `json:"province"`
	CityCode   *string              `json:"cityCode"`
	City       *string              `json:"city"`
	AdCode     *string              `json:"adCode"`
	District   *string              `json:"district"`
	TownCode   *string              `json:"townCode"`
	Township   *string              `json:"township"`
	SeaArea    *string              `json:"seaArea"`
	Remarks    *string              `json:"remarks"`
	UpdateTime *javacompat.DateTime `json:"updateTime"`
}

// gaodeMapEntity is the lenient decode target for the downstream Result data
// (mirrors GaodeMapEntity after ConvertorHelper.copyProperties).
type gaodeMapEntity struct {
	Id         *javacompat.LongStr  `json:"id"`
	Longitude  *string              `json:"longitude"`
	Dimension  *string              `json:"dimension"`
	Address    *string              `json:"address"`
	Country    *string              `json:"country"`
	Province   *string              `json:"province"`
	CityCode   *string              `json:"cityCode"`
	City       *string              `json:"city"`
	AdCode     *string              `json:"adCode"`
	District   *string              `json:"district"`
	TownCode   *string              `json:"townCode"`
	Township   *string              `json:"township"`
	SeaArea    *string              `json:"seaArea"`
	Remarks    *string              `json:"remarks"`
	UpdateTime *javacompat.DateTime `json:"updateTime"`
}

// gaodeMapCO matches clientobject.GaodeMapCo (field order = Java declaration
// order; no omitempty because Java does not enable NON_NULL).
type gaodeMapCO struct {
	Id         *javacompat.LongStr  `json:"id"`
	Longitude  *float64             `json:"longitude"`
	Latitude   *float64             `json:"latitude"`
	Address    *string              `json:"address"`
	Country    *string              `json:"country"`
	Province   *string              `json:"province"`
	CityCode   *string              `json:"cityCode"`
	City       *string              `json:"city"`
	AdCode     *string              `json:"adCode"`
	District   *string              `json:"district"`
	TownCode   *string              `json:"townCode"`
	Township   *string              `json:"township"`
	SeaArea    *string              `json:"seaArea"`
	Remarks    *string              `json:"remarks"`
	UpdateTime *javacompat.DateTime `json:"updateTime"`
}

// gaodeNavigateDTO matches Java GaodeNavigateDTO extends ClientDTO
// (origin/destination @NotBlank, type @NotNull).
type gaodeNavigateDTO struct {
	dto.ClientDTO
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Type        *int   `json:"type" binding:"required"`
}

var gaodeNavigateMsgs = mergeMsgs(map[string]string{
	"origin":      "must not be blank",
	"destination": "must not be blank",
	"type":        "must not be null",
})

// gaodeNavigateCO matches clientobject.GaodeNavigateCO {polyline, duration, distance}.
type gaodeNavigateCO struct {
	Polyline []locationVO `json:"polyline"`
	Duration *string      `json:"duration"`
	Distance *string      `json:"distance"`
}

// stepCO matches clientobject.Step {instruction, orientation, roadName, stepDistance, polyline}.
type stepCO struct {
	Instruction  *string `json:"instruction"`
	Orientation  *string `json:"orientation"`
	RoadName     *string `json:"roadName"`
	StepDistance *int    `json:"stepDistance"`
	Polyline     *string `json:"polyline"`
}

// navigateCO matches clientobject.NavigateCO {duration, polyline, distance, paths}.
type navigateCO struct {
	Duration *string      `json:"duration"`
	Polyline []locationVO `json:"polyline"`
	Distance *string      `json:"distance"`
	Paths    []stepCO     `json:"paths"`
}

func buildGaodeCmd(req *gaodeMapDTO) *gaodeMapCmd {
	return &gaodeMapCmd{
		Id:         req.Id,
		Address:    req.Address,
		Country:    req.Country,
		Province:   req.Province,
		CityCode:   req.CityCode,
		City:       req.City,
		AdCode:     req.AdCode,
		District:   req.District,
		TownCode:   req.TownCode,
		Township:   req.Township,
		SeaArea:    req.SeaArea,
		Remarks:    req.Remarks,
		UpdateTime: req.UpdateTime,
	}
}

// callGaodeMap ports GaodeMapGatewayImpl.getAddressByLocation/getLocation:
// any transport error / failed Result / undecodable data is swallowed
// (Java try/catch logs the error) and an empty entity is returned.
func callGaodeMap(c *gin.Context, path string, cmd *gaodeMapCmd, cmdCtx *dto.CommandContext) *gaodeMapEntity {
	empty := &gaodeMapEntity{}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, path, cmd, cmdCtx)
	if err != nil || result == nil || !result.Success {
		return empty
	}
	if len(result.Data) == 0 || string(result.Data) == "null" {
		return empty
	}
	var entity gaodeMapEntity
	if json.Unmarshal(result.Data, &entity) != nil {
		return empty
	}
	return &entity
}

// gaodeEntityToCO ports GaodeMapServiceImpl's CO assembly: same-name fields are
// copied; longitude/latitude come from Double.parseDouble of the entity's
// longitude/dimension strings (an unparsable value throws NumberFormatException
// -> code "00001", exactly like Java).
func gaodeEntityToCO(c *gin.Context, entity *gaodeMapEntity) (*gaodeMapCO, bool) {
	co := &gaodeMapCO{
		Id:         entity.Id,
		Address:    entity.Address,
		Country:    entity.Country,
		Province:   entity.Province,
		CityCode:   entity.CityCode,
		City:       entity.City,
		AdCode:     entity.AdCode,
		District:   entity.District,
		TownCode:   entity.TownCode,
		Township:   entity.Township,
		SeaArea:    entity.SeaArea,
		Remarks:    entity.Remarks,
		UpdateTime: entity.UpdateTime,
	}
	if entity.Longitude != nil {
		v, ok := parseDoubleOrFail(c, string(*entity.Longitude))
		if !ok {
			return nil, false
		}
		co.Longitude = &v
	}
	if entity.Dimension != nil {
		v, ok := parseDoubleOrFail(c, string(*entity.Dimension))
		if !ok {
			return nil, false
		}
		co.Latitude = &v
	}
	return co, true
}

// gaodeGetAddress ports GaodeMapController.getAddress (reverse geocoding).
func gaodeGetAddress(c *gin.Context) {
	var req gaodeMapDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	// Java: entity.setLongitude(dto.getLongitude().toString()) -> NPE when null
	if req.Longitude == nil {
		web.WriteException(c, `NullPointerException:Cannot invoke "java.lang.Double.toString()" because the return value of "com.xyy.dto.ClientDTO.getLongitude()" is null`)
		return
	}
	if req.Latitude == nil {
		web.WriteException(c, `NullPointerException:Cannot invoke "java.lang.Double.toString()" because the return value of "com.xyy.dto.ClientDTO.getLatitude()" is null`)
		return
	}

	cmd := buildGaodeCmd(&req)
	lng := javaDoubleToString(*req.Longitude)
	lat := javaDoubleToString(*req.Latitude)
	cmd.Longitude = &lng
	cmd.Dimension = &lat

	entity := callGaodeMap(c, "/gaode/map/getAddressByLocation", cmd, cmdCtx)
	co, ok := gaodeEntityToCO(c, entity)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// gaodeGetLocation ports GaodeMapController.getLocation (geocoding by address).
func gaodeGetLocation(c *gin.Context) {
	var req gaodeMapDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	// Java copies dto -> entity WITHOUT setting longitude/dimension here
	// (the Double longitude is not assignable to the String field).
	cmd := buildGaodeCmd(&req)

	entity := callGaodeMap(c, "/gaode/map/getLocationByAddress", cmd, cmdCtx)
	co, ok := gaodeEntityToCO(c, entity)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// gaodeNavigate ports GaodeMapController.navigate (v1): GaodeMapGatewayImpl
// POSTs {origin,destination,type} to ${gaode.navigate.url}; on any HTTP/parse
// failure it throws "调用中台高德导航失败" (code 00001). A response with
// code != 0 yields data=null -> an all-null CO inside a success Result.
func gaodeNavigate(c *gin.Context) {
	var req gaodeNavigateDTO
	if !web.BindJSON(c, &req, gaodeNavigateMsgs) {
		return
	}
	if !web.NotBlank(c, "origin", req.Origin) || !web.NotBlank(c, "destination", req.Destination) {
		return
	}

	body := map[string]interface{}{
		"origin":      req.Origin,
		"destination": req.Destination,
		"type":        req.Type,
	}
	resp, err := rpc.RestyClient.R().
		SetContext(c.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(getGaodeNavigateURL())
	if err != nil || resp.IsError() {
		web.WriteException(c, "调用中台高德导航失败")
		return
	}

	var outer struct {
		Code *int            `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(resp.Body(), &outer) != nil {
		// Java: RestTemplate message conversion failure -> caught -> BizException
		web.WriteException(c, "调用中台高德导航失败")
		return
	}
	// fastjson getIntValue("code") returns 0 when missing
	code := 0
	if outer.Code != nil {
		code = *outer.Code
	}

	co := &gaodeNavigateCO{}
	if code == 0 && len(outer.Data) > 0 && string(outer.Data) != "null" {
		var data struct {
			Count       *int    `json:"count"`
			Duration    *string `json:"duration"`
			Distance    *string `json:"distance"`
			PolylineApi *string `json:"polylineApi"`
		}
		if json.Unmarshal(outer.Data, &data) != nil {
			web.WriteException(c, "调用中台高德导航失败")
			return
		}
		// Java: Integer count unboxing -> NPE when absent
		if data.Count == nil {
			web.WriteException(c, `NullPointerException:Cannot invoke "java.lang.Integer.intValue()" because "count" is null`)
			return
		}
		if *data.Count == 0 {
			// throw new BizException(MsgCodeEnum.EXCEPTION, "步行路程过长,建议采用其他出行方式")
			writeBizError(c, "00001", "步行路程过长,建议采用其他出行方式")
			return
		}
		co.Duration = data.Duration
		co.Distance = data.Distance
		if data.PolylineApi != nil && *data.PolylineApi != "" {
			points, ok := parsePolylineLngFirst(c, *data.PolylineApi)
			if !ok {
				return
			}
			co.Polyline = points
		}
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// parsePolylineLngFirst replicates the v1 navigate parsing order:
// lng (split[0]) is parsed before accessing split[1].
func parsePolylineLngFirst(c *gin.Context, polylineApi string) ([]locationVO, bool) {
	points := []locationVO{}
	for _, s := range javaSplit(polylineApi, ";") {
		parts := javaSplit(s, ",")
		if len(parts) < 1 {
			web.WriteException(c, "ArrayIndexOutOfBoundsException:Index 0 out of bounds for length 0")
			return nil, false
		}
		lng, ok := parseDoubleOrFail(c, parts[0])
		if !ok {
			return nil, false
		}
		if len(parts) < 2 {
			web.WriteException(c, fmt.Sprintf("ArrayIndexOutOfBoundsException:Index 1 out of bounds for length %d", len(parts)))
			return nil, false
		}
		lat, ok := parseDoubleOrFail(c, parts[1])
		if !ok {
			return nil, false
		}
		points = append(points, locationVO{Lng: lng, Lat: lat})
	}
	return points, true
}

// gaodeNavigateV2 ports GaodeMapController.navigateV2: map-service path
// planning, flattening all step polylines into one list.
func gaodeNavigateV2(c *gin.Context) {
	var req gaodeNavigateDTO
	if !web.BindJSON(c, &req, gaodeNavigateMsgs) {
		return
	}
	if !web.NotBlank(c, "origin", req.Origin) || !web.NotBlank(c, "destination", req.Destination) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	path, ok := selectNavigationPath(c, req.Origin, req.Destination, "AMap", req.Type, cmdCtx.TenantId)
	if !ok {
		return
	}
	if path.Steps == nil {
		// Java: for (StepDO step : path.getSteps()) with a null list -> NPE
		web.WriteException(c, `NullPointerException:Cannot invoke "java.util.List.iterator()" because "steps" is null`)
		return
	}

	co := &gaodeNavigateCO{Distance: path.Distance, Duration: path.Duration}
	polyline := []locationVO{}
	for _, step := range path.Steps {
		stepPolyline := ""
		if step != nil && step.Polyline != nil {
			stepPolyline = *step.Polyline
		}
		points, ok := parsePolylineLatFirst(c, stepPolyline)
		if !ok {
			return
		}
		polyline = append(polyline, points...)
	}
	co.Polyline = polyline
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// gaodeNavigateV3 ports GaodeMapController.navigateV3: like v2 but also returns
// the per-step path details (instruction/orientation/roadName/stepDistance/polyline).
func gaodeNavigateV3(c *gin.Context) {
	var req gaodeNavigateDTO
	if !web.BindJSON(c, &req, gaodeNavigateMsgs) {
		return
	}
	if !web.NotBlank(c, "origin", req.Origin) || !web.NotBlank(c, "destination", req.Destination) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	path, ok := selectNavigationPath(c, req.Origin, req.Destination, "AMap", req.Type, cmdCtx.TenantId)
	if !ok {
		return
	}
	if path.Steps == nil {
		web.WriteException(c, `NullPointerException:Cannot invoke "java.util.List.iterator()" because the return value of "com.xyy.ebike.service.client.infrastructure.gateway.map.dto.PathDO.getSteps()" is null`)
		return
	}

	co := &navigateCO{Duration: path.Duration, Distance: path.Distance}
	polyline := []locationVO{}
	paths := []stepCO{}
	for _, step := range path.Steps {
		s := stepCO{}
		stepPolyline := ""
		if step != nil {
			s = stepCO{
				Instruction:  step.Instruction,
				Orientation:  step.Orientation,
				RoadName:     step.RoadName,
				StepDistance: step.StepDistance,
				Polyline:     step.Polyline,
			}
			if step.Polyline != nil {
				stepPolyline = *step.Polyline
			}
		}
		paths = append(paths, s)

		points, ok := parsePolylineLatFirst(c, stepPolyline)
		if !ok {
			return
		}
		polyline = append(polyline, points...)
	}
	co.Polyline = polyline
	co.Paths = paths
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}
