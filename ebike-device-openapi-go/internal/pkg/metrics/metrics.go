package metrics

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var (
	httpRequestsTotal atomic.Int64
	decodeTotal       atomic.Int64
	cmdTotal          atomic.Int64
	kafkaPushTotal    atomic.Int64
	routeProxyTotal   atomic.Int64
)

func IncHTTPRequests() { httpRequestsTotal.Add(1) }
func IncDecode()       { decodeTotal.Add(1) }
func IncCmd()          { cmdTotal.Add(1) }
func IncKafkaPush()    { kafkaPushTotal.Add(1) }
func IncRouteProxy()   { routeProxyTotal.Add(1) }

func RegisterRoutes(r *gin.Engine) {
	r.GET("/actuator/metrics", handleMetrics)
}

func handleMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"http_requests_total": httpRequestsTotal.Load(),
		"decode_total":        decodeTotal.Load(),
		"cmd_total":           cmdTotal.Load(),
		"kafka_push_total":    kafkaPushTotal.Load(),
		"route_proxy_total":   routeProxyTotal.Load(),
	})
}
