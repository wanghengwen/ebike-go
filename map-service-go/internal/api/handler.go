package api

import (
	"map-service-go/internal/model"
	"map-service-go/internal/service"

	"github.com/gin-gonic/gin"
)

type MapController struct {
	mapSvc *service.MapService
}

func NewMapController() *MapController {
	return &MapController{
		mapSvc: service.NewMapService(),
	}
}

// Result mimics com.xyy.dto.Result
func Success(data interface{}) gin.H {
	return gin.H{
		"success": true,
		"code":    "0",
		"msg":     "成功",
		"data":    data,
	}
}

func Error(code string, msg string) gin.H {
	return gin.H{
		"success": false,
		"code":    code,
		"msg":     msg,
		"data":    nil,
	}
}

func (ctrl *MapController) Location(c *gin.Context) {
	c.JSON(200, Success("武汉"))
}

func (ctrl *MapController) Regeo(c *gin.Context) {
	var cmd model.MapCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		handleBindError(c, err)
		return
	}

	addr, err := ctrl.mapSvc.GetAddressByLocation(c.Request.Context(), &cmd)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(200, Success(addr))
}

func (ctrl *MapController) BatchRegeo(c *gin.Context) {
	var cmd model.MapCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		handleBindError(c, err)
		return
	}

	if cmd.Locations == nil {
		c.JSON(200, Error("-3", "locations 不能为空"))
		return
	}

	addrs, err := ctrl.mapSvc.GetBatchAddress(c.Request.Context(), &cmd)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(200, Success(addrs))
}

func (ctrl *MapController) Geo(c *gin.Context) {
	var cmd model.MapCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		handleBindError(c, err)
		return
	}

	if cmd.Address == "" {
		c.JSON(200, Error("-3", "address 不能为空"))
		return
	}

	loc, err := ctrl.mapSvc.GetLocationByAddress(c.Request.Context(), &cmd)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(200, Success(loc))
}

func (ctrl *MapController) MapNavigate(c *gin.Context) {
	var cmd model.MapCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		handleBindError(c, err)
		return
	}

	if cmd.Origin == "" {
		c.JSON(200, Error("-3", "origin 不能为空"))
		return
	}
	if cmd.Destination == "" {
		c.JSON(200, Error("-3", "destination 不能为空"))
		return
	}

	nav, err := ctrl.mapSvc.Navigate(c.Request.Context(), &cmd)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(200, Success(nav))
}
