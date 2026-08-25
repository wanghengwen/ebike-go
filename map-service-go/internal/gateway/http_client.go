package gateway

import (
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// NewMapHTTPClient mirrors Java RestTemplateConfig: connect 2s, read 5s.
func NewMapHTTPClient() *resty.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		IdleConnTimeout:     90 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	return resty.New().
		SetTransport(transport).
		SetTimeout(5 * time.Second)
}
