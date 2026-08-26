package api

import (
	"net"
	"net/http"

	"ebike-device-paas-go/internal/pkg/nacos"

	"github.com/gin-gonic/gin"
)

// DeregisterService mirrors Java NacosEndpoint.deregisterService (localhost only, for K8s preStop).
func DeregisterService(c *gin.Context) {
	if !isLocalhostRequest(c) {
		c.JSON(http.StatusOK, Fail(CodeParamError, "只允许localhost请求"))
		return
	}
	nacos.Deregister()
	c.JSON(http.StatusOK, Ok("ok"))
}

func isLocalhostRequest(c *gin.Context) bool {
	host := c.Request.URL.Hostname()
	if host == "" {
		var err error
		host, _, err = net.SplitHostPort(c.Request.Host)
		if err != nil {
			host = c.Request.Host
		}
	}
	return host == "localhost" || host == "127.0.0.1"
}
