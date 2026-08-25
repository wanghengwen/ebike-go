package controller

import (
	"net/http"
	"strings"
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/pkg/config"
	pkgrpc "ebike-analyze-go/internal/pkg/rpc"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// DeregisterService mirrors Java NacosEndpoint.deregisterService (localhost only).
func DeregisterService(c *gin.Context) {
	host := c.Request.URL.Hostname()
	if host == "" {
		host = c.Request.Host
	}
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	// Java: request.getServerName().equals("localhost") - exact match only, no 127.0.0.1 allowance.
	if host != "localhost" {
		msg := "只允许localhost请求"
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, msg))
		return
	}
	if pkgrpc.RegisteredToNacos && pkgrpc.NamingClient != nil {
		localIP := pkgrpc.GetLocalIP()
		_, _ = pkgrpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          localIP,
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
	}
	time.Sleep(20 * time.Second)
	c.JSON(http.StatusOK, dto.NewSuccessResult("ok"))
}
