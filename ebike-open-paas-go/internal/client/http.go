// Package client holds HTTP clients for upstream services. Outbound calls use
// static ClusterIP base URLs (same as ebike-device-paas-go); AES tenant headers
// are attached only for paths matching feign.client.auth-url-regex.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// ResultEnvelope mirrors the platform's internal {success,code,msg,data} Result.
type ResultEnvelope struct {
	Success bool            `json:"success"`
	Code    string          `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

// GatewayError carries an upstream business code.
type GatewayError struct {
	Code string
	Msg  string
}

func (e *GatewayError) Error() string {
	return fmt.Sprintf("gateway code=%s msg=%s", e.Code, e.Msg)
}

// CommandResult mirrors paas/openapi CommandResult after buildCommandResult.
type CommandResult struct {
	JobID   *string     `json:"jobId"`
	Async   *bool       `json:"async"`
	EcuCode *string     `json:"ecuCode"`
	Payload *string     `json:"payload"`
	Result  interface{} `json:"result"`
}

// EcuCodeString returns the ecuCode or "" when absent.
func (r *CommandResult) EcuCodeString() string {
	if r == nil || r.EcuCode == nil {
		return ""
	}
	return *r.EcuCode
}

func doJSON(method, url string, body interface{}, tenantID string, extra map[string]string) (*ResultEnvelope, error) {
	var reader io.Reader
	hasBody := body != nil
	if hasBody {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	if err := applyGatewayAuthHeaders(req, tenantID); err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream HTTP %d: %s", resp.StatusCode, string(data))
	}
	var env ResultEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("decode Result envelope: %w", err)
	}
	return &env, nil
}

func postJSON(url string, body interface{}) (*ResultEnvelope, error) {
	return doJSON(http.MethodPost, url, body, "", nil)
}

// PostPaas posts to ebike-device-paas and returns the raw Result envelope.
// Paas paths are /device/* and do not receive AES headers.
func PostPaas(path string, body interface{}) (*ResultEnvelope, error) {
	base := config.GlobalConfig().Xyy.PaasURL
	if base == "" {
		return nil, fmt.Errorf("paas url not configured (xyy.paasUrl)")
	}
	return doJSON(http.MethodPost, base+path, body, "", nil)
}

// PostPaasCommand posts a device command to paas and unwraps CommandResult.
// Business failures that still carry data (17002) are returned as CommandResult
// with EcuCode set, not as an error — matching Xiaoan's success+code model.
func PostPaasCommand(path string, body interface{}) (*CommandResult, error) {
	env, err := PostPaas(path, body)
	if err != nil {
		return nil, err
	}
	var cr CommandResult
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &cr); err != nil {
			return nil, fmt.Errorf("decode CommandResult: %w", err)
		}
	}
	if !env.Success && cr.EcuCode == nil {
		return nil, &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	return &cr, nil
}

// PostOpenapi posts to ebike-device-openapi with AES headers for /ebike/* paths.
func PostOpenapi(path string, body interface{}, tenantID string) (*ResultEnvelope, error) {
	base := config.GlobalConfig().Xyy.OpenapiURL
	if base == "" {
		return nil, fmt.Errorf("openapi url not configured (xyy.openapiUrl)")
	}
	return doJSON(http.MethodPost, base+path, body, tenantID, nil)
}

// PostWorker posts to ebike-device-worker (AES for /ebike/*).
func PostWorker(path string, body interface{}, tenantID string, out interface{}) error {
	base := config.GlobalConfig().Xyy.WorkerURL
	if base == "" {
		return fmt.Errorf("worker url not configured (xyy.workerUrl)")
	}
	env, err := doJSON(http.MethodPost, base+path, body, tenantID, nil)
	if err != nil {
		return err
	}
	if !env.Success {
		return &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// GetConsole GETs anvelink-console and unmarshals Result.data into out.
// Uses xyy.consoleAuth (/auth/login) or a static consoleAuthToken. On JWT
// expiry or auth failure (5001/5002/5003) it re-logins once and retries.
func GetConsole(path string, query url.Values, out interface{}) error {
	err := getConsoleOnce(path, query, out, false)
	if err != nil && isConsoleAuthError(err) {
		return getConsoleOnce(path, query, out, true)
	}
	return err
}

func getConsoleOnce(path string, query url.Values, out interface{}, forceRefresh bool) error {
	base := config.GlobalConfig().Xyy.ConsoleURL
	if base == "" {
		return fmt.Errorf("console url not configured (xyy.consoleUrl)")
	}
	u := strings.TrimRight(base, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	extra := map[string]string{}
	if tok, err := consoleBearer(forceRefresh); err != nil {
		return err
	} else if tok != "" {
		extra["Authorization"] = tok
	}
	env, err := doJSON(http.MethodGet, u, nil, "", extra)
	if err != nil {
		return err
	}
	if !env.Success {
		return &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// PostMap posts to map-service with the secret header.
func PostMap(path string, body interface{}, out interface{}) error {
	base := config.GlobalConfig().Xyy.MapServiceConfig.URL
	if base == "" {
		return fmt.Errorf("map service url not configured")
	}
	extra := map[string]string{}
	if s := config.GlobalConfig().Xyy.MapServiceConfig.Secret; s != "" {
		extra["secret"] = s
	}
	env, err := doJSON(http.MethodPost, base+path, body, "", extra)
	if err != nil {
		return err
	}
	if !env.Success {
		return &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}
