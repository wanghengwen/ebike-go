package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Configuration
var (
	javaBaseURL = "https://fastid.luopingtech.com"
	goBaseURL   = "http://127.0.0.1:8080"
	secret      = "8v1i656gulgw0vdhxtt6pkvx3r25mbkfm4j6kd63lxgymhsymiwi007y1hjud5oh"
)

var (
	passed int
	failed int
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║   FastId Java vs Go Compatibility Test              ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Printf("Java: %s\n", javaBaseURL)
	fmt.Printf("Go:   %s\n", goBaseURL)
	fmt.Println()

	testAppName := fmt.Sprintf("go-compat-test-%d", time.Now().Unix())
	fmt.Printf("Using test app name: %s\n\n", testAppName)

	// ========== Test Group 1: machineId - Response Format ==========
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Test Group 1: POST /fastid/machineId")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Test 1.1: First machine registration
	fmt.Println("\n[1.1] First machine registration")
	jResp := postMachineID(javaBaseURL, "", "", testAppName, "machine-A", secret)
	gResp := postMachineID(goBaseURL, "", "", testAppName, "machine-A", secret)
	compareResponses("1.1", jResp, gResp)

	// Test 1.2: Second machine (different UUID, same app)
	fmt.Println("\n[1.2] Second machine registration")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "machine-B", secret)
	gResp = postMachineID(goBaseURL, "", "", testAppName, "machine-B", secret)
	compareResponses("1.2", jResp, gResp)

	// Test 1.3: Duplicate registration (same UUID)
	fmt.Println("\n[1.3] Duplicate registration (same machine-A)")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "machine-A", secret)
	gResp = postMachineID(goBaseURL, "", "", testAppName, "machine-A", secret)
	compareResponses("1.3", jResp, gResp)

	// Test 1.4: Third machine
	fmt.Println("\n[1.4] Third machine registration")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "machine-C", secret)
	gResp = postMachineID(goBaseURL, "", "", testAppName, "machine-C", secret)
	compareResponses("1.4", jResp, gResp)

	// ========== Test Group 2: Error Cases ==========
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Test Group 2: Error Cases")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Test 2.1: Wrong secret
	fmt.Println("\n[2.1] Wrong secret")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "machine-D", "wrong-secret")
	gResp = postMachineID(goBaseURL, "", "", testAppName, "machine-D", "wrong-secret")
	compareResponses("2.1", jResp, gResp)

	// Test 2.2: Empty secret
	fmt.Println("\n[2.2] Empty secret")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "machine-D", "")
	gResp = postMachineID(goBaseURL, "", "", testAppName, "machine-D", "")
	compareResponses("2.2", jResp, gResp)

	// Test 2.3: Empty appName
	fmt.Println("\n[2.3] Empty appName")
	jResp = postMachineID(javaBaseURL, "", "", "", "machine-D", secret)
	gResp = postMachineID(goBaseURL, "", "", "", "machine-D", secret)
	compareResponses("2.3", jResp, gResp)

	// Test 2.4: Empty machineUuid
	fmt.Println("\n[2.4] Empty machineUuid")
	jResp = postMachineID(javaBaseURL, "", "", testAppName, "", secret)
	gResp = postMachineID(goBaseURL, "", "", testAppName, "", secret)
	compareResponses("2.4", jResp, gResp)

	// ========== Test Group 3: Namespace & Group ==========
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Test Group 3: Custom Namespace & Group")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	nsTestApp := testAppName + "-ns"
	fmt.Println("\n[3.1] Custom namespace and group")
	jResp = postMachineID(javaBaseURL, "test-ns", "test-group", nsTestApp, "machine-ns-1", secret)
	gResp = postMachineID(goBaseURL, "test-ns", "test-group", nsTestApp, "machine-ns-1", secret)
	compareResponses("3.1", jResp, gResp)

	// ========== Test Group 4: appList JSON Format ==========
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Test Group 4: GET /fastid/appList")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\n[4.1] appList JSON structure")
	jAppList := getEndpoint(javaBaseURL, "/fastid/appList")
	gAppList := getEndpoint(goBaseURL, "/fastid/appList")
	compareJSONFieldNames("4.1", jAppList.body, gAppList.body, "appList")

	// ========== Test Group 5: machineList JSON Format ==========
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Test Group 5: GET /fastid/machineList")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// We need an appId. Get the first one from appList.
	appID := extractFirstAppID(gAppList.body)
	if appID != "" {
		fmt.Printf("\n[5.1] machineList for appId=%s\n", appID)
		jML := getEndpoint(javaBaseURL, "/fastid/machineList?appId="+appID)
		gML := getEndpoint(goBaseURL, "/fastid/machineList?appId="+appID)
		compareJSONFieldNames("5.1", jML.body, gML.body, "machineList")
	} else {
		fmt.Println("\n[5.1] SKIP: Could not extract appId from appList")
	}

	fmt.Println("\n[5.2] machineList without appId")
	jML := getEndpoint(javaBaseURL, "/fastid/machineList")
	gML := getEndpoint(goBaseURL, "/fastid/machineList")
	compareResponses("5.2", jML, gML)

	// ========== Summary ==========
	fmt.Println("\n══════════════════════════════════════════════════════")
	fmt.Printf("  Results: %d PASSED, %d FAILED\n", passed, failed)
	fmt.Println("══════════════════════════════════════════════════════")

	if failed > 0 {
		os.Exit(1)
	}
}

type httpResponse struct {
	statusCode  int
	contentType string
	body        string
	err         error
}

func postMachineID(baseURL, namespace, groupID, appName, machineUUID, secretVal string) httpResponse {
	form := url.Values{}
	form.Set("namespace", namespace)
	form.Set("groupId", groupID)
	form.Set("appName", appName)
	form.Set("machineUuid", machineUUID)
	form.Set("secret", secretVal)

	resp, err := http.Post(baseURL+"/fastid/machineId", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return httpResponse{err: err}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")
	return httpResponse{
		statusCode:  resp.StatusCode,
		contentType: ct,
		body:        string(body),
	}
}

func getEndpoint(baseURL, path string) httpResponse {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return httpResponse{err: err}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")
	return httpResponse{
		statusCode:  resp.StatusCode,
		contentType: ct,
		body:        string(body),
	}
}

func compareResponses(testID string, java, golang httpResponse) {
	if java.err != nil {
		fmt.Printf("  ⚠ JAVA ERROR: %v\n", java.err)
		failed++
		return
	}
	if golang.err != nil {
		fmt.Printf("  ⚠ GO ERROR:   %v\n", golang.err)
		failed++
		return
	}

	fmt.Printf("  Java:  [%d] %q\n", java.statusCode, java.body)
	fmt.Printf("  Go:    [%d] %q\n", golang.statusCode, golang.body)

	match := true

	// Compare status code
	if java.statusCode != golang.statusCode {
		fmt.Printf("  ✗ Status code mismatch: Java=%d Go=%d\n", java.statusCode, golang.statusCode)
		match = false
	}

	// Compare body content
	if java.body != golang.body {
		fmt.Printf("  ✗ Body mismatch!\n")
		match = false
	}

	// Compare Content-Type (normalize: ignore params like charset differences)
	jCT := normalizeContentType(java.contentType)
	gCT := normalizeContentType(golang.contentType)
	if jCT != gCT {
		fmt.Printf("  ⚠ Content-Type diff: Java=%q Go=%q (non-blocking)\n", java.contentType, golang.contentType)
	}

	if match {
		fmt.Printf("  ✓ PASS [%s]\n", testID)
		passed++
	} else {
		fmt.Printf("  ✗ FAIL [%s]\n", testID)
		failed++
	}
}

func compareJSONFieldNames(testID string, javaBody, goBody, label string) {
	jFields := extractJSONFields(javaBody)
	gFields := extractJSONFields(goBody)

	fmt.Printf("  Java %s fields: %v\n", label, jFields)
	fmt.Printf("  Go   %s fields: %v\n", label, gFields)

	if len(jFields) == 0 && len(gFields) == 0 {
		fmt.Printf("  ⚠ Both empty, cannot compare field names\n")
		passed++
		return
	}

	match := true
	for _, f := range jFields {
		found := false
		for _, g := range gFields {
			if f == g {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  ✗ Java field %q missing in Go response\n", f)
			match = false
		}
	}

	if match {
		fmt.Printf("  ✓ PASS [%s] - JSON field names match\n", testID)
		passed++
	} else {
		fmt.Printf("  ✗ FAIL [%s] - JSON field names mismatch\n", testID)
		failed++
	}
}

func extractJSONFields(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" || body == "null" {
		return nil
	}

	// Try as array
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(body), &arr); err == nil && len(arr) > 0 {
		fields := make([]string, 0)
		for k := range arr[0] {
			fields = append(fields, k)
		}
		return fields
	}

	// Try as single object
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(body), &obj); err == nil {
		fields := make([]string, 0)
		for k := range obj {
			fields = append(fields, k)
		}
		return fields
	}

	return nil
}

func extractFirstAppID(appListBody string) string {
	var apps []map[string]interface{}
	if err := json.Unmarshal([]byte(appListBody), &apps); err != nil || len(apps) == 0 {
		return ""
	}
	if id, ok := apps[0]["id"]; ok {
		return fmt.Sprintf("%.0f", id)
	}
	return ""
}

func normalizeContentType(ct string) string {
	parts := strings.Split(ct, ";")
	return strings.TrimSpace(parts[0])
}
