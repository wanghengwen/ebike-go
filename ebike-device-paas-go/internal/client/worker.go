package client

import (
	"encoding/json"
	"fmt"

	"ebike-device-paas-go/internal/pkg/config"
)

// PostWorker forwards body to the ebike-device-worker gateway at path and unwraps
// the Result envelope, decoding the inner data into out. On a non-success
// envelope it returns a *GatewayError (mirrors DeviceResultHelper.getResultData).
//
// Trajectory / metric queries are live RPCs to ebike-device-worker, so (like the
// other forwards) they are NOT shadow-compared.
func PostWorker(path string, body interface{}, out interface{}, tenantID string) error {
	base := config.GlobalConfig.Xyy.WorkerURL
	if base == "" {
		return fmt.Errorf("worker url not configured (xyy.workerUrl)")
	}
	env, err := doJSON("POST", base+path, body, tenantID, nil)
	if err != nil {
		return err
	}
	if !env.Success {
		return &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode worker data: %w", err)
	}
	return nil
}
