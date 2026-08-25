package middleware

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"ebike-service-client-go/internal/pkg/config"
	"ebike-service-client-go/internal/pkg/env"

	"github.com/gin-gonic/gin"
)

type proxyResponseRecorder struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	header     http.Header
	statusCode int
}

func (r *proxyResponseRecorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *proxyResponseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return r.body.Write(b)
}

func (r *proxyResponseRecorder) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func (r *proxyResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *proxyResponseRecorder) WriteHeaderNow() {}

func (r *proxyResponseRecorder) Flush() {}

// pathInProxyList reports whether path matches any entry in list.
// Exact match always wins; entries ending with "/" match by prefix.
func pathInProxyList(list []string, path string) bool {
	for _, entry := range list {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == path {
			return true
		}
		if strings.HasSuffix(entry, "/") && strings.HasPrefix(path, entry) {
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

func readRequestBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	if cached, ok := c.Get("proxyReqBody"); ok {
		return cached.([]byte), nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Set("proxyReqBody", body)
	setProxyRequestBody(c.Request, body)
	return body, nil
}

// ProxyGateway routes traffic like ebike-device-worker-go:
//   - live_list: native Go handler (client gets Go response)
//   - record_list: proxy to Java, log [RECORD] (client gets Java; Go handler not run)
//   - DRY_RUN + other paths: Go shadow run then proxy Java, log [SHADOW MATCH/DIFF]
//   - default: proxy to Java
//
// Skips when proxy.target_url is empty, or for /actuator/* (native health/deregister).
func ProxyGateway() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GlobalConfig.Proxy
		if cfg.TargetURL == "" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/actuator/") {
			c.Next()
			return
		}

		if pathInProxyList(cfg.LiveList, path) {
			c.Next()
			return
		}

		target, err := url.Parse(cfg.TargetURL)
		if err != nil {
			log.Printf("[PROXY] invalid target_url: %v", err)
			c.Next()
			return
		}

		reqBody, err := readRequestBody(c)
		if err != nil {
			log.Printf("[PROXY] read request body failed path=%s: %v", path, err)
			c.Next()
			return
		}

		if pathInProxyList(cfg.RecordList, path) {
			proxy := newSingleHostProxy(target)
			proxy.ModifyResponse = func(r *http.Response) error {
				javaRes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaRes))
				log.Printf("[RECORD] Path: %s\nReq: %s\nRes: %s\n", path, string(reqBody), string(javaRes))
				return nil
			}
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		if env.IsDryRun() {
			origWriter := c.Writer
			recorder := &proxyResponseRecorder{
				ResponseWriter: origWriter,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = recorder
			c.Next()
			goResBody := append([]byte(nil), recorder.body.Bytes()...)
			c.Writer = origWriter

			proxy := newSingleHostProxy(target)
			proxy.ModifyResponse = func(r *http.Response) error {
				javaRes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaRes))
				LogShadowDiff(path, string(reqBody), goResBody, javaRes)
				return nil
			}
			setProxyRequestBody(c.Request, reqBody)
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		proxy := newSingleHostProxy(target)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

// logShadowDiffFromHeader compares Go body with Java body from gateway mirror header.
func logShadowDiffFromHeader(path, javaResultBase64, goRes string) {
	if javaResultBase64 == "" {
		return
	}
	javaResultBytes, err := base64.StdEncoding.DecodeString(javaResultBase64)
	if err != nil {
		log.Printf("[SHADOW_DIFF_ERROR] Path: %s | Base64 Decode Error: %v", path, err)
		return
	}
	LogShadowDiff(path, "", []byte(goRes), javaResultBytes)
}
