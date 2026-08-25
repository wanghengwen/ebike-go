package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/pkg/apilog"
	"ebike-device-worker-go/internal/pkg/env"
	"ebike-device-worker-go/internal/pkg/jsondiff"
	"ebike-device-worker-go/internal/pkg/shadow"

	"github.com/gin-gonic/gin"
)

type responseRecorder struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	header     http.Header
	statusCode int
}

func (r *responseRecorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return r.body.Write(b)
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseRecorder) WriteHeaderNow() {}

func (r *responseRecorder) Flush() {}

func contains(list []string, item string) bool {
	for _, v := range list {
		if strings.HasPrefix(item, v) || strings.HasPrefix(v, item) || item == v {
			return true
		}
	}
	return false
}

func setProxyRequestBody(req *http.Request, body []byte) {
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	if len(body) == 0 {
		req.Header.Del("Content-Length")
	} else {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}
	req.Header.Del("Transfer-Encoding")
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
}

func newSingleHostProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = target.Host
	}
	return proxy
}

// ProxyGateway implements LiveList / RecordList / Shadow (DRY_RUN) routing, mirroring ebike-fence-go.
func ProxyGateway() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxyCfg := config.GlobalConfig.Proxy
		if proxyCfg.TargetUrl == "" {
			c.Next()
			return
		}

		target, err := url.Parse(proxyCfg.TargetUrl)
		if err != nil {
			log.Printf("[PROXY] Error parsing target URL: %v", err)
			c.Next()
			return
		}

		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/actuator/") {
			c.Next()
			return
		}

		if contains(proxyCfg.LiveList, path) {
			// Native handler logs bound params in api.BindAndServe.
			c.Next()
			return
		}

		var reqBodyBytes []byte
		if c.Request.Body != nil {
			var err error
			reqBodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				log.Printf("[PROXY] read request body failed path=%s: %v", path, err)
			}
		}
		setProxyRequestBody(c.Request, reqBodyBytes)

		if contains(proxyCfg.RecordList, path) {
			proxy := newSingleHostProxy(target)
			proxy.ModifyResponse = func(r *http.Response) error {
				javaResBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaResBytes))
				log.Printf("[RECORD] Path: %s\nReq: %s\nRes: %s\n", path, string(reqBodyBytes), string(javaResBytes))
				return nil
			}
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		if env.IsDryRun() {
			newCtx := shadow.WithShadowTest(c.Request.Context())
			c.Request = c.Request.WithContext(newCtx)

			origWriter := c.Writer
			recorder := &responseRecorder{
				ResponseWriter: origWriter,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = recorder
			c.Next()
			goResBody := recorder.body.Bytes()
			c.Writer = origWriter

			proxy := newSingleHostProxy(target)
			proxy.ModifyResponse = func(r *http.Response) error {
				javaResBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaResBytes))
				goResStr := string(goResBody)
				javaResStr := string(javaResBytes)
				if goResStr != javaResStr && !jsondiff.Equal(goResBody, javaResBytes) {
					log.Printf("[SHADOW DIFF] Path: %s\nReq: %s\nJava: %s\nGo: %s\n", path, string(reqBodyBytes), javaResStr, goResStr)
				} else {
					log.Printf("[SHADOW MATCH] Path: %s Req: %s", path, string(reqBodyBytes))
				}
				return nil
			}
			setProxyRequestBody(c.Request, reqBodyBytes)
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		apilog.Params(path, reqBodyBytes)
		proxy := newSingleHostProxy(target)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
