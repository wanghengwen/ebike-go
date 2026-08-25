package metrics

import (
	"strconv"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/proxy"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "ebike_gateway"

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests processed by the gateway.",
	}, []string{"mode", "method", "route_id", "status"})

	gatewayDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "Gateway end-to-end request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"mode", "method", "route_id"})

	upstreamDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "upstream_request_duration_seconds",
		Help:      "Upstream reverse-proxy duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"mode", "method", "route_id"})

	routesLoaded = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "routes_loaded",
		Help:      "Number of routes currently loaded from Nacos.",
	})

	proxyCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "proxy_cache_size",
		Help:      "Number of cached reverse proxy instances.",
	})
)

// Middleware records request metrics after handler completion.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.Request.URL.Path
		if path == "/actuator/health" || path == "/health" || path == "/actuator/prometheus" {
			return
		}

		routeID := "unmatched"
		if routeObj, exists := c.Get("matchedRoute"); exists {
			if matchedRoute, ok := routeObj.(*proxy.Route); ok && matchedRoute.Id != "" {
				routeID = matchedRoute.Id
			}
		}

		mode := config.AppConfig.GatewayMode
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		requestsTotal.WithLabelValues(mode, method, routeID, status).Inc()
		gatewayDuration.WithLabelValues(mode, method, routeID).Observe(time.Since(start).Seconds())

		if val, exists := c.Get("upstreamDuration"); exists {
			if d, ok := val.(time.Duration); ok {
				upstreamDuration.WithLabelValues(mode, method, routeID).Observe(d.Seconds())
			}
		}
	}
}

// SetRoutesLoaded updates the routes_loaded gauge.
func SetRoutesLoaded(count int) {
	routesLoaded.Set(float64(count))
}

// SetProxyCacheSize updates the proxy_cache_size gauge.
func SetProxyCacheSize(count int) {
	proxyCacheSize.Set(float64(count))
}
