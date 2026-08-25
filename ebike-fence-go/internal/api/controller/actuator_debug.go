package controller

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// registerActuatorDebug wires in-process diagnostics under /actuator/* (skipped by ProxyGateway).
func registerActuatorDebug(r *gin.RouterGroup) {
	r.GET("/actuator/memstats", Memstats)
	r.POST("/actuator/memstats", Memstats)

	pprofGroup := r.Group("/actuator/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
}

type memstatsResponse struct {
	Goroutines      int     `json:"goroutines"`
	HeapAllocBytes  uint64  `json:"heapAllocBytes"`
	HeapInuseBytes  uint64  `json:"heapInuseBytes"`
	HeapIdleBytes   uint64  `json:"heapIdleBytes"`
	HeapSysBytes    uint64  `json:"heapSysBytes"`
	StackInuseBytes uint64  `json:"stackInuseBytes"`
	HeapAllocMB     float64 `json:"heapAllocMB"`
	HeapInuseMB     float64 `json:"heapInuseMB"`
	HeapSysMB       float64 `json:"heapSysMB"`
	NumGC           uint32  `json:"numGC"`
	LastGCPauseMs   float64 `json:"lastGCPauseMs"`
	GCCPUFraction   float64 `json:"gcCPUFraction"`
	GomemlimitBytes int64   `json:"gomemlimitBytes"`
	GomemlimitMB    float64 `json:"gomemlimitMB"`
}

// Memstats exposes runtime.MemStats for quick heap/GC inspection (kubectl exec / port-forward).
func Memstats(c *gin.Context) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	limit := debug.SetMemoryLimit(-1)
	resp := memstatsResponse{
		Goroutines:      runtime.NumGoroutine(),
		HeapAllocBytes:  ms.HeapAlloc,
		HeapInuseBytes:  ms.HeapInuse,
		HeapIdleBytes:   ms.HeapIdle,
		HeapSysBytes:    ms.HeapSys,
		StackInuseBytes: ms.StackInuse,
		HeapAllocMB:     bytesToMB(ms.HeapAlloc),
		HeapInuseMB:     bytesToMB(ms.HeapInuse),
		HeapSysMB:       bytesToMB(ms.HeapSys),
		NumGC:           ms.NumGC,
		LastGCPauseMs:   float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1e6,
		GCCPUFraction:   ms.GCCPUFraction,
		GomemlimitBytes: limit,
	}
	if limit > 0 {
		resp.GomemlimitMB = float64(limit) / (1024 * 1024)
	}
	c.JSON(http.StatusOK, resp)
}

func bytesToMB(b uint64) float64 {
	return float64(b) / (1024 * 1024)
}
