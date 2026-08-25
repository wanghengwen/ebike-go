package api

import (
	"net"
	"net/http"
	"strings"

	"ebike-open-paas-go/internal/middleware"
	"ebike-open-paas-go/internal/pkg/kafka"
	"ebike-open-paas-go/internal/pkg/nacos"
	"ebike-open-paas-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the Gin engine with Xiaoan public routes and internal event routes.
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.AccessLog())

	// Liveness only reports that the process is scheduling goroutines; a
	// dependency outage must not restart the pod, since a restart cannot fix it.
	r.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/actuator/health/readiness", readiness)
	r.POST("/actuator/deregisterService", DeregisterService)

	v1 := r.Group("/ebike/v1", middleware.Auth())
	{
		v1.GET("/allDevices", allDevices)
		v1.GET("/deviceInfo", deviceInfo)
		v1.GET("/GPSPoints", gpsPoints)
		v1.GET("/address", address)
		v1.GET("/batteryInfo", batteryInfo)
		v1.GET("/bmsInfo", bmsInfo)
		v1.GET("/currentBmsInfo", currentBmsInfo)

		v1.POST("/lock", lock)
		v1.POST("/acc", acc)
		v1.POST("/defend", defend)
		v1.POST("/backWheel", backWheel)
		v1.POST("/batteryCompartment", batteryCompartment)
		v1.POST("/deviceVoice", deviceVoice)
		v1.POST("/bluetooth", bluetooth)
		v1.POST("/reboot", reboot)
		v1.POST("/mc", mc)
		v1.POST("/batteryPowerSwitch", batteryPowerSwitch)
		v1.POST("/lbs2gps", lbs2gps)
		v1.POST("/sms", sms)

		v1.POST("/callback", registerCallback)
		v1.GET("/callback", getCallback)
		v1.DELETE("/callback", deleteCallback)
	}

	// Xiaoan realtime device query uses a non-/v1 path.
	r.POST("/ebike/api/device", middleware.Auth(), realtimeDevice)

	// Callback delivery has no user-visible surface, so expose the consumer and
	// dispatcher counters for operators to confirm events are flowing.
	internal := r.Group("/internal", middleware.InternalAuth())
	{
		internal.GET("/callback/stats", callbackStats)
		internal.GET("/debug/device-belong", debugDeviceBelong)
	}

	return r
}

// DeregisterService mirrors Java NacosEndpoint.deregisterService (localhost only).
// It is called by the preStop hook so the pod leaves Nacos before it stops
// serving.
func DeregisterService(c *gin.Context) {
	if !isLocalhostRequest(c) {
		c.JSON(http.StatusOK, gin.H{"success": false, "code": "00002", "msg": "只允许localhost请求"})
		return
	}
	nacos.Deregister()
	c.JSON(http.StatusOK, gin.H{"success": true, "code": "0", "msg": "成功", "data": "ok"})
}

// isLocalhostRequest reports whether the connection came from the pod itself.
//
// The check is on the peer address, not the Host header: the header is chosen by
// the caller, so `curl -H "Host: localhost"` through the ingress would otherwise
// let anyone on the internet pull this pod out of Nacos. RemoteAddr is the real
// socket peer and cannot be spoofed the same way; X-Forwarded-For is ignored for
// the same reason.
func isLocalhostRequest(c *gin.Context) bool {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.Request.RemoteAddr
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

// readiness fails while a dependency the traffic path needs is down, so the
// endpoint stops receiving requests it could only answer with an error.
func readiness(c *gin.Context) {
	checks := gin.H{}
	ready := true

	// Redis holds the callback subscriptions, the device→tenant mapping and the
	// device shadow, so every route depends on it.
	if redis.Available() {
		if healthy := redis.Healthy(); healthy {
			checks["redis"] = "UP"
		} else {
			checks["redis"] = "DOWN"
			ready = false
		}
	} else {
		checks["redis"] = "DISABLED"
	}

	// A consumer that has stopped means callbacks are silently not delivered,
	// which is worse than the pod being pulled from the pool.
	switch {
	case !kafka.Enabled():
		checks["kafka"] = "DISABLED"
	case kafka.Running():
		checks["kafka"] = "UP"
	default:
		checks["kafka"] = "DOWN"
		ready = false
	}

	status := "UP"
	code := http.StatusOK
	if !ready {
		status = "DOWN"
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, gin.H{"status": status, "components": checks})
}
