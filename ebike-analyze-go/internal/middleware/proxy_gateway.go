package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"ebike-analyze-go/internal/pkg/config"
	"ebike-analyze-go/internal/pkg/env"
	"ebike-analyze-go/internal/pkg/shadow"

	"github.com/gin-gonic/gin"
)

type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return len(b), nil
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return len(s), nil
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	// Do nothing to prevent sending headers to client from Go handler
}

func (r *responseRecorder) Flush() {
	// Do nothing
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if strings.HasPrefix(item, v) || strings.HasPrefix(v, item) || item == v {
			return true
		}
	}
	return false
}

// ProxyGateway implements the 3-tier routing strategy (Live, Shadow, Record, Default).
func ProxyGateway() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxyCfg := config.GlobalConfig.Xyy.Proxy
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

		// K8s/management probes: native Go handlers only — no shadow compare or DIFF/MATCH logs.
		if strings.HasPrefix(path, "/actuator/") {
			c.Next()
			return
		}

		// 1. LiveList: Full native Go execution, no proxy
		if contains(proxyCfg.LiveList, path) {
			c.Next()
			return
		}

		// Read request body for proxy and logging
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// 2. RecordList: Proxy and record Request/Response
		if contains(proxyCfg.RecordList, path) {
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

		// 3. Default (Shadow): Dual execution & compare when DRY_RUN=true
		if env.DryRun() {
			// Clone context for Shadow Test
			newCtx := shadow.WithShadowTest(c.Request.Context())
			c.Request = c.Request.WithContext(newCtx)

			// Capture Go response
			origWriter := c.Writer
			recorder := &responseRecorder{
				ResponseWriter: origWriter,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = recorder

			// Execute Go handlers
			c.Next()

			// Get Go result
			goResBody := recorder.body.Bytes()

			// Restore original writer
			c.Writer = origWriter

			// Proxy to Java
			proxy := cloneJavaProxy(target, func(r *http.Response) error {
				javaResBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(javaResBytes))

				// Compare
				goResStr := string(goResBody)
				javaResStr := string(javaResBytes)
				if goResStr != javaResStr && !isJSONEq(goResBody, javaResBytes) {
					log.Printf("[SHADOW DIFF] Path: %s\nReq: %s\nJava: %s\nGo: %s\n", path, string(reqBodyBytes), javaResStr, goResStr)
				} else {
					log.Printf("[SHADOW MATCH] Path: %s", path)
				}
				return nil
			})

			// Restore request body for proxy
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort() // Prevent further Gin handlers if any
		} else {
			// If not DRY_RUN, simply proxy to Java (No dual execution)
			proxy := *getJavaReverseProxy(target)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

func isJSONEq(a, b []byte) bool {
	var aObj, bObj map[string]interface{}
	if err := json.Unmarshal(a, &aObj); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bObj); err != nil {
		return false
	}
	stripNil(aObj)
	stripNil(bObj)
	return jsonValuesEqual(aObj, bObj)
}

func jsonValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return mapJSONEqual(av, bv)
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonValuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		return floatJSONEqual(av, bv)
	case json.Number:
		bf, err := av.Float64()
		if err != nil {
			return false
		}
		switch bv := b.(type) {
		case float64:
			return floatJSONEqual(bf, bv)
		case json.Number:
			bf2, err := bv.Float64()
			return err == nil && floatJSONEqual(bf, bf2)
		default:
			return false
		}
	default:
		return reflect.DeepEqual(a, b)
	}
}

func mapJSONEqual(a, b map[string]interface{}) bool {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	for k := range seen {
		av, aOk := a[k]
		bv, bOk := b[k]
		if !aOk {
			av = nil
		}
		if !bOk {
			bv = nil
		}
		if av == nil && bv == nil {
			continue
		}
		if !jsonValuesEqual(av, bv) {
			return false
		}
	}
	return true
}

// floatJSONEqual tolerates tiny floating-point drift and small geo distance deltas vs Java JTS.
func floatJSONEqual(a, b float64) bool {
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	if diff <= 1e-5 {
		return true
	}
	return false
}

func stripNil(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		} else if vMap, ok := v.(map[string]interface{}); ok {
			stripNil(vMap)
		} else if vSlice, ok := v.([]interface{}); ok {
			for _, item := range vSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					stripNil(itemMap)
				}
			}
		}
	}
}
