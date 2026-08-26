package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/pkg/shadow"

	"github.com/gin-gonic/gin"
)

// responseRecorder captures the Go handler body without flushing it to the client
// (fence-go style). During shadow dual-run the client receives the Java response.
type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	return r.body.WriteString(s)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	// Swallow: do not send Go status/headers to the client during shadow.
}

func (r *responseRecorder) Flush() {}

// shadowUnorderedByPath lists per-endpoint UnorderedKeys for jsondiff (top-level
// "data" lists whose Java order follows HashMap / unstable sort ties).
var shadowUnorderedByPath = map[string][]string{
	"/device/paas/device/car_count":              {"data"},
	"/device/paas/device/getCarNumByServiceId":   {"data"},
	"/device/paas/device/carStatistics":          {"data"},
	"/device/paas/device/getRackCarNumAll":       {"data"},
	"/device/paas/device/queryDeviceByBattery":   {"data"},
	"/device/paas/device/queryDeviceByTotalMiles": {"data"},
	"/device/paas/device/queryDeviceByNoOrderTime": {"data"},
	"/device/paas/device/queryDeviceByStaticTime":  {"data"},
	"/device/paas/device/carParkingStatistics":   {"data"},
}

// ProxyGateway implements fence-go 3-tier routing:
//
//  1. liveList  → Go native only (no proxy / no shadow)
//  2. recordList → reverse-proxy Java + [RECORD] (Go handlers skipped)
//  3. neither + DryRun → run Go (captured) + proxy Java to client + [SHADOW *]
//     neither + !DryRun → proxy Java only
func ProxyGateway() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/actuator/") {
			c.Next()
			return
		}

		targetURL := config.ProxyTargetURL()
		if targetURL == "" {
			c.Next()
			return
		}

		target, err := url.Parse(targetURL)
		if err != nil {
			log.Printf("[PROXY] invalid target URL %q: %v", targetURL, err)
			c.Next()
			return
		}

		// 1. LiveList: full native Go
		if config.IsLivePath(path) {
			c.Next()
			return
		}

		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// 2. RecordList: Java only + [RECORD]
		if config.PathInProxyList(config.GlobalConfig.Xyy.Proxy.RecordList, path) {
			proxy := cloneJavaProxy(target, func(r *http.Response) error {
				javaResBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaResBytes))
				log.Printf("[RECORD] Path: %s\nReq: %s\nRes: %s\n", path, string(reqBodyBytes), string(javaResBytes))
				return nil
			})
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// 3. Default: shadow dual-run when DryRun, else Java-only proxy
		if config.GlobalConfig.DryRun {
			c.Request = c.Request.WithContext(shadow.WithShadowTest(c.Request.Context()))

			origWriter := c.Writer
			recorder := &responseRecorder{
				ResponseWriter: origWriter,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = recorder
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			c.Next()
			goResBody := recorder.body.Bytes()
			c.Writer = origWriter

			extra := shadowUnorderedByPath[path]
			proxy := cloneJavaProxy(target, func(r *http.Response) error {
				javaResBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaResBytes))
				shadow.Report(path, reqBodyBytes, javaResBytes, goResBody, extra...)
				return nil
			})
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		proxy := *getJavaReverseProxy(target)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
