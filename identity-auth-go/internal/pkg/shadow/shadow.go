package shadow

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"identity-auth-go/internal/pkg/config"
	"identity-auth-go/internal/pkg/jsondiff"
	"identity-auth-go/internal/pkg/mask"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second, // P1 fix: prevent goroutine leak on slow/dead Java service
}

// CompareWithJava sends an async HTTP request to the old Java service when dry-run
// mode is enabled, and compares its response JSON with the provided goResponseJSON.
func CompareWithJava(method string, path string, reqBody []byte, goResponseJSON []byte) {
	if !config.GlobalConfig.DryRun {
		return
	}
	if config.GlobalConfig.Xyy.JavaServiceUrl == "" {
		return // skip shadow testing if Java URL is not configured
	}

	go func() {
		targetUrl := config.GlobalConfig.Xyy.JavaServiceUrl + path

		var bodyReader io.Reader
		if len(reqBody) > 0 && method == "POST" {
			bodyReader = bytes.NewBuffer(reqBody)
		}

		req, err := http.NewRequest(method, targetUrl, bodyReader)
		if err != nil {
			log.Printf("[SHADOW_DIFF_ERROR] Path: %s | Failed to create request: %v", path, err)
			return
		}
		if method == "POST" {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("[SHADOW_DIFF_ERROR] Path: %s | Failed to send request: %v", path, err)
			return
		}
		defer resp.Body.Close()

		// P2 fix: check Java response status before comparing
		if resp.StatusCode != http.StatusOK {
			log.Printf("[SHADOW_DIFF_ERROR] Path: %s | Java returned HTTP %d", path, resp.StatusCode)
			return
		}

		javaResponseJSON, err := io.ReadAll(resp.Body) // P2 fix: use io.ReadAll instead of deprecated ioutil.ReadAll
		if err != nil {
			log.Printf("[SHADOW_DIFF_ERROR] Path: %s | Failed to read Java response: %v", path, err)
			return
		}

		if jsondiff.Equal(goResponseJSON, javaResponseJSON) {
			log.Printf("[SHADOW MATCH] Path: %s", path)
		} else {
			log.Printf("[SHADOW DIFF] Path: %s\nReq: %s\nJava: %s\nGo: %s",
				path, mask.JSONForLog(reqBody), mask.JSONForLog(javaResponseJSON), mask.JSONForLog(goResponseJSON))
		}
	}()
}
