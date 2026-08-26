package api

import (
	"encoding/json"
	"io"
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerEcuCommandRoutes attaches the EcuCommand endpoints (EcuCommandApi).
//
// Each forwards a live device command to the openapi gateway (with Redis write
// side effects), so they are NOT shadow-compared.
func registerEcuCommandRoutes(r *gin.Engine) {
	r.POST("/device/paas/lock", lockCommand)
	r.POST("/device/paas/helmetLock", helmetCommand)
	r.POST("/device/paas/rearWheelLock", rearWheelLock)
	r.POST("/device/paas/defend", defendCommand)
	r.POST("/device/paas/setInnerParam", setInnerParam)
	r.POST("/device/paas/updateInnerFence", updateInnerFence)
	r.POST("/device/paas/deviceUpgrade", deviceUpgrade)
	r.POST("/device/paas/voiceUpgrade", voiceUpgrade)
	r.POST("/device/paas/voice", voiceCommand)
	r.POST("/device/paas/batteryCompartment", batteryCompartmentCommand)
	r.POST("/device/paas/bluetooth", bluetoothCommand)
	r.POST("/device/paas/scanBleHelmet", scanBleHelmet)
	r.POST("/device/paas/restart", restartCommand)
	r.POST("/device/paas/transmission", transmission)
	r.POST("/device/paas/tempUnLock", tempUnLock)
	r.POST("/device/paas/dashboard", dashboard)
	r.POST("/device/paas/replyStopMove", replyStopMove)
	r.POST("/device/paas/triggerTempState", triggerTempState)
}

// bindCommand reads the request body into a generic map (forwarded to the
// gateway) and extracts the commandContext for tenant/trace resolution.
func bindCommand(c *gin.Context) (map[string]interface{}, *dto.CommandContext, bool) {
	body, _ := io.ReadAll(c.Request.Body)
	raw := map[string]interface{}{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &raw); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return nil, nil, false
		}
	}
	var ctx struct {
		CommandContext *dto.CommandContext `json:"commandContext"`
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &ctx)
	}
	// Ensure the context carries a tenantId (header fallback) so the forwarded
	// gateway call can sign its AES auth headers (FeignClientConfig).
	backfillTenant(c, &ctx.CommandContext)
	return raw, ctx.CommandContext, true
}

// writeSuccessResult mirrors ResultHelper.success(commandResult): always wrap as
// success (used by transmission / tempUnLock / replyStopMove / triggerTempState),
// degrading to a gateway-error envelope when the forward fails.
func writeSuccessResult(c *gin.Context, cr dto.CommandResult, err error) {
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(cr))
}

func lockCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.LockCommand(resolveTenant(c, cc), raw, cc)
	writeDeviceResult(c, cr, err)
}

func defendCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.DefendCommand(resolveTenant(c, cc), raw, cc)
	writeDeviceResult(c, cr, err)
}

func helmetCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.HelmetCommand(raw, cc)
	writeDeviceResult(c, cr, err)
}

func rearWheelLock(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.RearWheelLock(raw, cc)
	writeDeviceResult(c, cr, err)
}

func setInnerParam(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.SetInnerParam(raw, cc)
	writeDeviceResult(c, cr, err)
}

func updateInnerFence(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.UpdateInnerFence(raw, cc)
	writeDeviceResult(c, cr, err)
}

func deviceUpgrade(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.DeviceUpgrade(raw, cc)
	writeDeviceResult(c, cr, err)
}

func voiceUpgrade(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.VoiceUpgrade(raw, cc)
	writeDeviceResult(c, cr, err)
}

func voiceCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.VoiceCommand(resolveTenant(c, cc), raw, cc)
	writeDeviceResult(c, cr, err)
}

func batteryCompartmentCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.BatteryCompartmentCommand(raw, cc)
	writeDeviceResult(c, cr, err)
}

func bluetoothCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.Bluetooth(raw, cc)
	writeDeviceResult(c, cr, err)
}

func scanBleHelmet(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.ScanBleHelmet(raw, cc)
	writeDeviceResult(c, cr, err)
}

func restartCommand(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.RestartCommand(raw, cc)
	writeDeviceResult(c, cr, err)
}

func dashboard(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.Dashboard(raw, cc)
	writeDeviceResult(c, cr, err)
}

func transmission(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.Transmission(raw, cc)
	writeSuccessResult(c, cr, err)
}

func tempUnLock(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.TempUnLock(resolveTenant(c, cc), raw, cc)
	writeSuccessResult(c, cr, err)
}

func replyStopMove(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.ReplyStopMove(raw, cc)
	writeSuccessResult(c, cr, err)
}

func triggerTempState(c *gin.Context) {
	raw, cc, ok := bindCommand(c)
	if !ok {
		return
	}
	cr, err := service.TriggerTempState(resolveTenant(c, cc), raw, cc)
	writeSuccessResult(c, cr, err)
}
