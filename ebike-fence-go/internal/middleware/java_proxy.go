package middleware

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Shared transport for Java reverse proxy: bounded connections + header timeout so
// slow/hung upstream cannot accumulate one goroutine per request indefinitely.
var javaProxyTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     false,
	MaxIdleConns:          256,
	MaxIdleConnsPerHost:   64,
	MaxConnsPerHost:       128,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 12 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

type javaProxyCache struct {
	mu     sync.RWMutex
	target string
	proxy  *httputil.ReverseProxy
}

var cachedJavaProxy javaProxyCache

func getJavaReverseProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.Scheme + "://" + target.Host

	cachedJavaProxy.mu.RLock()
	if cachedJavaProxy.proxy != nil && cachedJavaProxy.target == key {
		p := cachedJavaProxy.proxy
		cachedJavaProxy.mu.RUnlock()
		return p
	}
	cachedJavaProxy.mu.RUnlock()

	cachedJavaProxy.mu.Lock()
	defer cachedJavaProxy.mu.Unlock()
	if cachedJavaProxy.proxy != nil && cachedJavaProxy.target == key {
		return cachedJavaProxy.proxy
	}

	p := httputil.NewSingleHostReverseProxy(target)
	p.Transport = javaProxyTransport
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[PROXY] upstream error path=%s err=%v", r.URL.Path, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	cachedJavaProxy.proxy = p
	cachedJavaProxy.target = key
	return p
}

func cloneJavaProxy(target *url.URL, modify func(*http.Response) error) httputil.ReverseProxy {
	base := *getJavaReverseProxy(target)
	base.ModifyResponse = modify
	return base
}
