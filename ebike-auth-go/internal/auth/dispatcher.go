package auth

import (
	"ebike-auth-go/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/actuator/health", healthHandler)

	if config.AppConfig.AuthMode == "business" {
		r.POST("/oauth/token", BusinessTokenHandler)
		r.POST("/oauth/logout", BusinessLogoutHandler)
	} else {
		r.POST("/oauth/token", ClientTokenHandler)
		r.POST("/oauth/logout", ClientLogoutHandler)
		r.POST("/oauth/phoneLogout", ClientPhoneLogoutHandler)
		r.POST("/oauth/jsapi/signature", ClientJsApiSignatureHandler)
	}
}
