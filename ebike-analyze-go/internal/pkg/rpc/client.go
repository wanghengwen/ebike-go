package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"ebike-analyze-go/internal/middleware"

	"github.com/go-resty/resty/v2"
)

var RestyClient *resty.Client

// DevicePaasClient uses a longer timeout for ECU/device queries.
var DevicePaasClient *resty.Client

// InitRestClient initializes the Resty HTTP Client.
// NOTE: no retry is configured on purpose — Spring Cloud OpenFeign defaults to
// Retryer.NEVER_RETRY, and endpoints may be non-idempotent.
func sharedHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          128,
		MaxIdleConnsPerHost:   32,
		MaxConnsPerHost:       64,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func InitRestClient() {
	RestyClient = resty.New()
	RestyClient.SetTimeout(10 * time.Second)
	RestyClient.SetTransport(sharedHTTPTransport())

	DevicePaasClient = resty.New()
	DevicePaasClient.SetTimeout(25 * time.Second)
	deviceTransport := sharedHTTPTransport()
	deviceTransport.ResponseHeaderTimeout = 25 * time.Second
	DevicePaasClient.SetTransport(deviceTransport)
}

// PostToDevicePaasService posts to ebike-device-paas with the longer ECU timeout.
func PostToDevicePaasService(ctx context.Context, serviceName string, path string, body interface{}, result interface{}) error {
	if DevicePaasClient == nil {
		InitRestClient()
	}
	return postWithClient(ctx, DevicePaasClient, serviceName, path, body, result)
}

func postWithClient(ctx context.Context, client *resty.Client, serviceName string, path string, body interface{}, result interface{}) error {
	addr, err := SelectOneHealthyInstance(serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service %s: %v", serviceName, err)
	}

	url := fmt.Sprintf("http://%s%s", addr, path)

	req := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body)

	if acceptLang := middleware.GetAcceptLanguageFromCtx(ctx); acceptLang != "" {
		req.SetHeader("Accept-Language", acceptLang)
	}
	if traceId := middleware.GetTraceIdFromCtx(ctx); traceId != "" {
		req.SetHeader("X-Trace-Id", traceId)
	}

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("RPC failed for %s: %v", url, err)
	}
	if resp.IsError() {
		return fmt.Errorf("RPC returned HTTP %d for %s", resp.StatusCode(), url)
	}
	if result != nil {
		if err := json.Unmarshal(resp.Body(), result); err != nil {
			return fmt.Errorf("failed to decode response from %s: %v", url, err)
		}
	}
	return nil
}

// IsServiceDiscoveryError reports Nacos discovery failures (no healthy instances).
// Callers may treat these as soft misses during shadow / partial deployments.
func IsServiceDiscoveryError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "instance list is empty") ||
		strings.Contains(msg, "failed to discover service")
}

// PostToService automatically discovers a service instance from Nacos and forwards a POST request.
// It mirrors Java's Feign client behavior, including Accept-Language header passthrough.
func PostToService(ctx context.Context, serviceName string, path string, body interface{}, result interface{}) error {
	if RestyClient == nil {
		InitRestClient()
	}
	return postWithClient(ctx, RestyClient, serviceName, path, body, result)
}

// GetFromService automatically discovers a service instance from Nacos and forwards a GET request
func GetFromService(ctx context.Context, serviceName string, path string, queryParams map[string]string, result interface{}) error {
	addr, err := SelectOneHealthyInstance(serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service %s: %v", serviceName, err)
	}

	url := fmt.Sprintf("http://%s%s", addr, path)

	req := RestyClient.R().
		SetContext(ctx).
		SetQueryParams(queryParams)

	// Forward Accept-Language header (matches Java's FeignHeaderInterceptor)
	if acceptLang := middleware.GetAcceptLanguageFromCtx(ctx); acceptLang != "" {
		req.SetHeader("Accept-Language", acceptLang)
	}

	// Propagate Trace Context
	if traceId := middleware.GetTraceIdFromCtx(ctx); traceId != "" {
		req.SetHeader("X-Trace-Id", traceId)
	}

	resp, err := req.Get(url)

	if err != nil {
		return fmt.Errorf("RPC failed for %s: %v", url, err)
	}

	if resp.IsError() {
		return fmt.Errorf("RPC returned HTTP %d for %s", resp.StatusCode(), url)
	}

	if result != nil {
		if err := json.Unmarshal(resp.Body(), result); err != nil {
			return fmt.Errorf("failed to decode response from %s: %v", url, err)
		}
	}

	return nil
}
