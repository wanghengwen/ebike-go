// Package misc ports Java TestController (/business) and NacosEndpoint (/actuator).
package misc

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/config"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// testCmd mirrors Java TestCmd extends ClientDTO for the echo endpoints.
// Java TestController has NO @Valid/@Validated on its parameters, so the
// @NotEmpty annotations are NOT enforced — therefore no binding:"required"
// tags here. Pointer types reproduce Jackson's null output for absent params
// (the Java ObjectMapper does not enable NON_NULL).
type testCmd struct {
	TraceId       *string  `json:"traceId" form:"traceId"`
	TenantId      *string  `json:"tenantId" form:"tenantId"`
	Platform      *string  `json:"platform" form:"platform"`
	DeviceId      *string  `json:"deviceId" form:"deviceId"`
	Version       *string  `json:"version" form:"version"`
	Ip            *string  `json:"ip" form:"ip"`
	Longitude     *float64 `json:"longitude" form:"longitude"`
	Latitude      *float64 `json:"latitude" form:"latitude"`
	Source        *string  `json:"source" form:"source"`
	StressTesting bool     `json:"stressTesting" form:"stressTesting"`
	Name          *string  `json:"name" form:"name"`
}

// RegisterRoutes wires /business/test/* and /actuator/deregisterService.
func RegisterRoutes(r *gin.RouterGroup) {
	business := r.Group("/business")
	{
		business.GET("/test/get", testGet)
		business.POST("/test/postForm", testPostForm)
		business.POST("/test/postJson", testPostJson)
	}
	actuator := r.Group("/actuator")
	{
		actuator.GET("/health", actuatorHealth)
		actuator.POST("/deregisterService", deregisterService)
	}
}

// actuatorHealth mirrors Spring Boot Actuator health for K8S liveness/readiness probes.
func actuatorHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func testGet(c *gin.Context) {
	var cmd testCmd
	if !web.BindQuery(c, &cmd, nil) {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(cmd))
}

func testPostForm(c *gin.Context) {
	var cmd testCmd
	if !web.BindForm(c, &cmd, nil) {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(cmd))
}

func testPostJson(c *gin.Context) {
	var cmd testCmd
	if !web.BindJSON(c, &cmd, nil) {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(cmd))
}

// deregisterService mirrors Java NacosEndpoint.deregisterService:
// only callable via localhost, stops the Nacos registration, then sleeps 40s
// so in-flight traffic drains before the pod is terminated.
func deregisterService(c *gin.Context) {
	host := c.Request.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if host != "localhost" {
		// Java: ResultHelper.exception("只允许localhost请求") -> code "00001"
		web.WriteException(c, "只允许localhost请求")
		return
	}

	log.Println("nacos实例准备下线")
	if os.Getenv("DRY_RUN") != "true" && rpc.NamingClient != nil {
		_, err := rpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          rpc.GetLocalIP(),
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
		if err != nil {
			log.Printf("nacos deregister error: %v", err)
		}
	}
	time.Sleep(40 * time.Second)
	log.Println("nacos实例下线完成")
	c.JSON(http.StatusOK, dto.NewSuccessResult("ok"))
}
