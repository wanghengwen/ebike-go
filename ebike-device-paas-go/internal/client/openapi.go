// Package client holds HTTP clients for the upstream services that the Java
// ebike-device-paas reaches via Feign (resolved through Nacos): the
// ebike-device-openapi command gateway and ebike-management.
//
// These forwarded calls are live device-command queries / cross-service RPCs,
// so they are NOT shadow-compared (mirroring would trigger a second real
// command on the device).
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"ebike-device-paas-go/internal/pkg/config"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// CommandResultDo mirrors the Java CommandResultDo<String>: the openapi gateway
// returns this inside a Result envelope. Result is itself a JSON string.
type CommandResultDo struct {
	JobID   *string `json:"jobId"`
	Async   *bool   `json:"async"`
	Payload *string `json:"payload"`
	Result  string  `json:"result"`
}

// resultEnvelope mirrors com.xyy.dto.Result.
type resultEnvelope struct {
	Success bool            `json:"success"`
	Code    string          `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

// GatewayError carries the upstream business code so callers can map it to the
// paas-facing code (mirrors DeviceResultHelper.getResultData switch).
type GatewayError struct {
	Code string
	Msg  string
}

func (e *GatewayError) Error() string { return fmt.Sprintf("gateway code=%s msg=%s", e.Code, e.Msg) }

// PostGateway sends body to the openapi gateway at path and unwraps the Result
// envelope, returning the inner CommandResultDo. On a non-success envelope it
// returns a *GatewayError. tenantID identifies the IoT platform credentials used
// for the AES auth headers (FeignClientConfig) on the /ebike/.* path.
func PostGateway(path string, body interface{}, tenantID string) (*CommandResultDo, error) {
	base := config.GlobalConfig.Xyy.OpenapiURL
	if base == "" {
		return nil, fmt.Errorf("openapi gateway url not configured (xyy.openapiUrl)")
	}
	env, err := doJSON(http.MethodPost, base+path, body, tenantID, nil)
	if err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	var do CommandResultDo
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &do); err != nil {
			return nil, fmt.Errorf("decode CommandResultDo: %w", err)
		}
	}
	return &do, nil
}

// GetJobState mirrors DeviceCommandApiFeign.jobStateGateway:
// GET /ebike/cmd/get_job_result?jobId=... returning Result<String>, where the
// inner String is the raw job-result JSON ({"code":...,"result":{...}}). On a
// non-success envelope it returns a *GatewayError (mirrors DeviceResultHelper).
func GetJobState(jobID string, tenantID string) (string, error) {
	base := config.GlobalConfig.Xyy.OpenapiURL
	if base == "" {
		return "", fmt.Errorf("openapi gateway url not configured (xyy.openapiUrl)")
	}
	u := base + "/ebike/cmd/get_job_result?jobId=" + url.QueryEscape(jobID)
	env, err := doJSON(http.MethodGet, u, nil, tenantID, nil)
	if err != nil {
		return "", err
	}
	if !env.Success {
		return "", &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return "", nil
	}
	// Result<String>: Data is a JSON-encoded string.
	var s string
	if err := json.Unmarshal(env.Data, &s); err != nil {
		// Fall back to raw bytes if the upstream did not quote the payload.
		return string(env.Data), nil
	}
	return s, nil
}

// carRestBatteryCO mirrors ebike-management CarRestBatteryCO (only restBattery used).
type carRestBatteryCO struct {
	RestBattery *int `json:"restBattery"`
}

// ComputeRestBattery mirrors EbikeManageRpcImpl.computeRestBattery via
// ebike-management VoltagePlanApi.getRestBattery.
func ComputeRestBattery(imei string, voltage *int, soc *int, commandContext interface{}) (*int, error) {
	base := config.GlobalConfig.Xyy.ManagementURL
	if base == "" {
		return nil, fmt.Errorf("management url not configured (xyy.managementUrl)")
	}
	var restBattery *float64
	if soc != nil {
		f := float64(*soc)
		restBattery = &f
	}
	reqBody := map[string]interface{}{
		"imei":           imei,
		"voltage":        voltage,
		"restBattery":    restBattery,
		"commandContext": commandContext,
	}
	// Path must match management VoltagePlanApi.getRestBattery: "voltage-plan/get-rest-battery".
	env, err := postJSON(base+"/voltage-plan/get-rest-battery", reqBody)
	if err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	var co carRestBatteryCO
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &co); err != nil {
			return nil, fmt.Errorf("decode CarRestBatteryCO: %w", err)
		}
	}
	return co.RestBattery, nil
}

// doJSON performs an HTTP request to an upstream service and decodes the Result
// envelope. When body is non-nil it is JSON-encoded with a Content-Type header.
// tenantID drives the AES gateway auth headers for auth-url-regex paths (see
// applyGatewayAuthHeaders); extra carries per-call headers (e.g. map-service
// "secret"). A non-2xx response is returned as an error.
func doJSON(method, url string, body interface{}, tenantID string, extra map[string]string) (*resultEnvelope, error) {
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
	var env resultEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("decode Result envelope: %w", err)
	}
	return &env, nil
}

// postJSON is the no-auth POST helper for management / iot-platform calls (their
// paths never match auth-url-regex, so no gateway auth headers are added).
func postJSON(url string, body interface{}) (*resultEnvelope, error) {
	return doJSON(http.MethodPost, url, body, "", nil)
}
