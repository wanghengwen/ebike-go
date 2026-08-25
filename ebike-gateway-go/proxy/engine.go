package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ebike-gateway-go/logger"
	"ebike-gateway-go/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var gatewayTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          2000,
	MaxIdleConnsPerHost:   500,
	MaxConnsPerHost:       0,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

type ProxyEngine struct {
	routes     atomic.Value // holds []Route
	proxyCache sync.Map     // key: target key (scheme://host), value: *httputil.ReverseProxy
}

var DefaultEngine *ProxyEngine

func NewProxyEngine() *ProxyEngine {
	pe := &ProxyEngine{}
	pe.routes.Store([]Route{})
	DefaultEngine = pe
	return pe
}

func (pe *ProxyEngine) UpdateRoutes(routes []Route) {
	pe.routes.Store(routes)
	pe.pruneProxyCache(routes)
}

func targetKey(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func (pe *ProxyEngine) pruneProxyCache(routes []Route) {
	active := make(map[string]struct{}, len(routes))
	for _, r := range routes {
		if r.TargetURL != nil {
			active[targetKey(r.TargetURL)] = struct{}{}
			continue
		}
		targetUri := r.Uri
		if strings.HasPrefix(targetUri, "lb://") {
			targetUri = "http://" + strings.TrimPrefix(targetUri, "lb://")
		}
		if parsed, err := url.Parse(targetUri); err == nil {
			active[targetKey(parsed)] = struct{}{}
		}
	}

	pe.proxyCache.Range(func(key, value any) bool {
		if _, ok := active[key.(string)]; !ok {
			pe.proxyCache.Delete(key)
		}
		return true
	})
}

func (pe *ProxyEngine) getOrCreateProxy(remote *url.URL) *httputil.ReverseProxy {
	key := targetKey(remote)
	if cached, ok := pe.proxyCache.Load(key); ok {
		return cached.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = gatewayTransport
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Log.Error("Proxy connection failed", zap.String("path", r.URL.Path), zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"success":false,"code":"00001","msg":"Bad Gateway: downstream service unreachable","data":null}`))
	}

	actual, _ := pe.proxyCache.LoadOrStore(key, proxy)
	return actual.(*httputil.ReverseProxy)
}

// matchAntPath converts Spring AntPath pattern to regex and matches the path
func matchAntPath(pattern, path string) bool {
	if pattern == "" {
		return false
	}
	p := strings.ReplaceAll(pattern, "**", "___DOUBLE_STAR___")
	p = strings.ReplaceAll(p, "*", "[^/]*")
	p = strings.ReplaceAll(p, "___DOUBLE_STAR___", ".*")
	matched, _ := regexp.MatchString("^"+p+"$", path)
	return matched
}

// MatchRouteMiddleware finds the route and puts it in the context
func (pe *ProxyEngine) MatchRouteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if reqPath == "/actuator/health" || reqPath == "/health" {
			c.JSON(http.StatusOK, gin.H{"status": "UP"})
			c.Abort()
			return
		}

		routes := pe.routes.Load().([]Route)

		var matchedRoute *Route
		for _, r := range routes {
			matched := false
			if r.PathRegexp != nil {
				matched = r.PathRegexp.MatchString(reqPath)
			} else {
				matched = matchAntPath(r.PathPattern, reqPath)
			}
			if matched {
				rCopy := r
				matchedRoute = &rCopy
				break
			}
		}

		if matchedRoute == nil {
			response.Abort(c, http.StatusNotFound, response.CodeException, "No route matched for "+reqPath)
			return
		}

		c.Set("matchedRoute", matchedRoute)
		c.Next()
	}
}

func (pe *ProxyEngine) HandleRequest(c *gin.Context) {
	matchedObj, exists := c.Get("matchedRoute")
	if !exists {
		response.JSON(c, http.StatusNotFound, response.CodeException, "No route matched")
		return
	}

	matchedRoute := matchedObj.(*Route)

	remote := matchedRoute.TargetURL
	if remote == nil {
		targetUri := matchedRoute.Uri
		if strings.HasPrefix(targetUri, "lb://") {
			serviceName := strings.TrimPrefix(targetUri, "lb://")
			targetUri = "http://" + serviceName
		}
		var err error
		remote, err = url.Parse(targetUri)
		if err != nil {
			logger.Log.Error("Invalid route target URI", zap.String("uri", targetUri), zap.Error(err))
			response.JSON(c, http.StatusInternalServerError, response.CodeException, "Invalid target URI")
			return
		}
	}

	proxy := pe.getOrCreateProxy(remote)

	c.Request.URL.Host = remote.Host
	c.Request.URL.Scheme = remote.Scheme
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
	c.Request.Header.Set("X-Real-IP", c.ClientIP())
	c.Request.Host = remote.Host

	startUpstream := time.Now()
	proxy.ServeHTTP(c.Writer, c.Request)
	c.Set("upstreamDuration", time.Since(startUpstream))
}

// CachedProxyCount returns the number of cached reverse proxy instances (for tests/metrics).
func (pe *ProxyEngine) CachedProxyCount() int {
	count := 0
	pe.proxyCache.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
