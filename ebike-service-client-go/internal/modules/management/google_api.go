package management

// Ports Java GoogleMapController (@RequestMapping("/client/management/google")):
//   - POST /getAddress  -> ebike-management /google/map/getAddressByLocation
//   - POST /getLocation -> ebike-management /google/map/getLocationByAddress
//
// GoogleMapServiceImpl/GoogleMapGatewayImpl behave exactly like their Gaode
// counterparts: downstream failures are swallowed into an empty entity and the
// CO is assembled locally (longitude/latitude parsed from strings).

import (
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// googleMapDTO matches Java GoogleMapDto extends ClientDTO (no own validation;
// @Validated enforces only ClientDTO's traceId/tenantId).
type googleMapDTO struct {
	dto.ClientDTO
	Id                *javacompat.LongStr  `json:"id"`
	AddressComponents *string              `json:"addressComponents"`
	FormattedAddress  *string              `json:"formattedAddress"`
	Geometry          *string              `json:"geometry"`
	PlaceId           *string              `json:"placeId"`
	Types             *string              `json:"types"`
	Remarks           *string              `json:"remarks"`
	UpdateTime        *javacompat.DateTime `json:"updateTime"`
}

// googleMapCmd matches the downstream GoogleMapCmd; commandContext is injected
// by rpc.ForwardCommand. All fields emitted (Java serializes nulls).
type googleMapCmd struct {
	Id                *javacompat.LongStr  `json:"id"`
	Longitude         *string              `json:"longitude"`
	Dimension         *string              `json:"dimension"`
	AddressComponents *string              `json:"addressComponents"`
	FormattedAddress  *string              `json:"formattedAddress"`
	Geometry          *string              `json:"geometry"`
	PlaceId           *string              `json:"placeId"`
	Types             *string              `json:"types"`
	Remarks           *string              `json:"remarks"`
	UpdateTime        *javacompat.DateTime `json:"updateTime"`
}

// googleMapEntity is the lenient decode target for the downstream Result data.
type googleMapEntity struct {
	Id                *javacompat.LongStr  `json:"id"`
	Longitude         *string              `json:"longitude"`
	Dimension         *string              `json:"dimension"`
	AddressComponents *string              `json:"addressComponents"`
	FormattedAddress  *string              `json:"formattedAddress"`
	Geometry          *string              `json:"geometry"`
	PlaceId           *string              `json:"placeId"`
	Types             *string              `json:"types"`
	Remarks           *string              `json:"remarks"`
	UpdateTime        *javacompat.DateTime `json:"updateTime"`
}

// googleMapCO matches clientobject.GoogleMapCo (declaration order, nulls kept).
type googleMapCO struct {
	Id                *javacompat.LongStr  `json:"id"`
	Longitude         *float64             `json:"longitude"`
	Latitude          *float64             `json:"latitude"`
	AddressComponents *string              `json:"addressComponents"`
	FormattedAddress  *string              `json:"formattedAddress"`
	Geometry          *string              `json:"geometry"`
	PlaceId           *string              `json:"placeId"`
	Types             *string              `json:"types"`
	Remarks           *string              `json:"remarks"`
	UpdateTime        *javacompat.DateTime `json:"updateTime"`
}

func buildGoogleCmd(req *googleMapDTO) *googleMapCmd {
	return &googleMapCmd{
		Id:                req.Id,
		AddressComponents: req.AddressComponents,
		FormattedAddress:  req.FormattedAddress,
		Geometry:          req.Geometry,
		PlaceId:           req.PlaceId,
		Types:             req.Types,
		Remarks:           req.Remarks,
		UpdateTime:        req.UpdateTime,
	}
}

// callGoogleMap ports GoogleMapGatewayImpl: failures are swallowed (Java
// try/catch logs) and an empty entity is returned.
func callGoogleMap(c *gin.Context, path string, cmd *googleMapCmd, cmdCtx *dto.CommandContext) *googleMapEntity {
	empty := &googleMapEntity{}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, path, cmd, cmdCtx)
	if err != nil || result == nil || !result.Success {
		return empty
	}
	if len(result.Data) == 0 || string(result.Data) == "null" {
		return empty
	}
	var entity googleMapEntity
	if json.Unmarshal(result.Data, &entity) != nil {
		return empty
	}
	return &entity
}

func googleEntityToCO(c *gin.Context, entity *googleMapEntity) (*googleMapCO, bool) {
	co := &googleMapCO{
		Id:                entity.Id,
		AddressComponents: entity.AddressComponents,
		FormattedAddress:  entity.FormattedAddress,
		Geometry:          entity.Geometry,
		PlaceId:           entity.PlaceId,
		Types:             entity.Types,
		Remarks:           entity.Remarks,
		UpdateTime:        entity.UpdateTime,
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

// googleGetAddress ports GoogleMapController.getAddress.
func googleGetAddress(c *gin.Context) {
	var req googleMapDTO
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

	cmd := buildGoogleCmd(&req)
	lng := javaDoubleToString(*req.Longitude)
	lat := javaDoubleToString(*req.Latitude)
	cmd.Longitude = &lng
	cmd.Dimension = &lat

	entity := callGoogleMap(c, "/google/map/getAddressByLocation", cmd, cmdCtx)
	co, ok := googleEntityToCO(c, entity)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// googleGetLocation ports GoogleMapController.getLocation.
func googleGetLocation(c *gin.Context) {
	var req googleMapDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	cmd := buildGoogleCmd(&req)

	entity := callGoogleMap(c, "/google/map/getLocationByAddress", cmd, cmdCtx)
	co, ok := googleEntityToCO(c, entity)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}
