package controller

import (
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/pkg/config"
	"ebike-fence-go/internal/pkg/rpc"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// DeregisterService mirrors Java NacosEndpoint.deregisterService (localhost only).
func DeregisterService(c *gin.Context) {
	if c.Request.Host != "localhost" && c.Request.Host != "localhost:"+c.Request.URL.Port() {
		// Java checks request.getServerName() == "localhost"
		host := c.Request.URL.Hostname()
		if host != "localhost" && host != "127.0.0.1" {
			result := dto.NewErrorResult(dto.CodeException, "只允许localhost请求")
			web.RespondResult(c, &result, nil)
			return
		}
	}

	if rpc.NamingClient != nil {
		_, _ = rpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          rpc.GetLocalIP(),
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
	}
	time.Sleep(40 * time.Second)

	result := dto.NewSuccessResult("ok")
	web.RespondResult(c, &result, nil)
}
