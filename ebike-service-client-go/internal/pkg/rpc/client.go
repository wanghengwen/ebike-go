package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ebike-service-client-go/internal/middleware"
	"github.com/go-resty/resty/v2"
)

var RestyClient *resty.Client

// RestyUploadClient is used for multipart file uploads where downstream may
// need more time than the default 10s RPC timeout (Java Feign read timeout).
var RestyUploadClient *resty.Client

// InitRestClient initializes the Resty HTTP Client.
// NOTE: no retry is configured on purpose — Spring Cloud OpenFeign defaults to
// Retryer.NEVER_RETRY, and endpoints like repair/record/add are non-idempotent
// (a retry could create duplicate repair tickets).
func InitRestClient() {
	RestyClient = resty.New()
	RestyClient.SetTimeout(10 * time.Second)

	RestyUploadClient = resty.New()
	RestyUploadClient.SetTimeout(60 * time.Second)
}

// PostToService automatically discovers a service instance from Nacos and forwards a POST request.
// It mirrors Java's Feign client behavior, including Accept-Language header passthrough
// (matching Java's FeignHeaderInterceptor).
// The response body is decoded with json.Unmarshal into result; when result contains
// json.RawMessage fields (see dto.Result.Data), the downstream payload is preserved
// verbatim (Long-as-string values, null fields, large numbers).
func PostToService(ctx context.Context, serviceName string, path string, body interface{}, result interface{}) error {
	addr, err := SelectOneHealthyInstance(serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service %s: %v", serviceName, err)
	}

	url := fmt.Sprintf("http://%s%s", addr, path)

	req := RestyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body)

	// Forward Accept-Language header (matches Java's FeignHeaderInterceptor)
	if acceptLang := middleware.GetAcceptLanguageFromCtx(ctx); acceptLang != "" {
		req.SetHeader("Accept-Language", acceptLang)
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
