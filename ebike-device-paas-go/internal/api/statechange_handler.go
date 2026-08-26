package api

import (
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerStateChangeRoutes attaches the StateChangeApi endpoints. These are
// device-state writes (Redis SETRANGE / C34 Kafka) -> NOT shadow-compared.
func registerStateChangeRoutes(r *gin.Engine) {
	r.POST("/device/ridingState/change", ridingStateChange)
	r.POST("/device/alarmState/change", alarmStateChange)
	r.POST("/device/location/change", locationChange)
	r.POST("/device/scanLocation/change", scanLocationChange)
}

// backfillTenant ensures the command context carries a tenantId, falling back to
// the gateway-injected header (mirrors resolveTenant for the write endpoints).
func backfillTenant(c *gin.Context, cc **dto.CommandContext) {
	if *cc == nil {
		*cc = &dto.CommandContext{}
	}
	if (*cc).TenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			(*cc).TenantID, _ = v.(string)
		}
	}
}

func ridingStateChange(c *gin.Context) {
	var cmd dto.RidingStateChangeCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	service.RidingStateChange(&cmd)
	c.JSON(http.StatusOK, Ok(nil))
}

func alarmStateChange(c *gin.Context) {
	var cmd dto.AlarmStateChangeCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	service.AlarmStateChange(&cmd)
	c.JSON(http.StatusOK, Ok(nil))
}

func locationChange(c *gin.Context) {
	var cmd dto.LocationChangeCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	service.LocationChange(&cmd)
	c.JSON(http.StatusOK, Ok(1))
}

func scanLocationChange(c *gin.Context) {
	var cmd dto.ScanLocationChangeCmd
	if !bindBody(c, &cmd) {
		return
	}
	backfillTenant(c, &cmd.CommandContext)
	service.ScanLocationChange(&cmd)
	c.JSON(http.StatusOK, Ok(1))
}
