package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"luopingtech-fastid-register/internal/model"
	"luopingtech-fastid-register/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

// ---- Mock Service ----

type mockService struct {
	getMachineIDFunc func(namespace, groupID, appName, machineUUID string) (int, error)
	listAppsFunc     func() ([]model.App, error)
	listMachinesFunc func(appID uint) ([]model.Machine, error)
}

func (m *mockService) GetMachineID(namespace, groupID, appName, machineUUID string) (int, error) {
	return m.getMachineIDFunc(namespace, groupID, appName, machineUUID)
}

func (m *mockService) ListApps() ([]model.App, error) {
	return m.listAppsFunc()
}

func (m *mockService) ListMachines(appID uint) ([]model.Machine, error) {
	return m.listMachinesFunc(appID)
}

// ---- Response Format Tests ----

func TestResponseFormat_Success(t *testing.T) {
	got := response(true, "", 1)
	want := "success=true&msg=&result=1"
	if got != want {
		t.Errorf("response(true, \"\", 1) = %q, want %q", got, want)
	}
}

func TestResponseFormat_SuccessZero(t *testing.T) {
	got := response(true, "", 0)
	want := "success=true&msg=&result=0"
	if got != want {
		t.Errorf("response(true, \"\", 0) = %q, want %q", got, want)
	}
}

func TestResponseFormat_Failure(t *testing.T) {
	got := response(false, "秘钥验证失败", "")
	want := "success=false&msg=秘钥验证失败&result="
	if got != want {
		t.Errorf("response(false, \"秘钥验证失败\", \"\") = %q, want %q", got, want)
	}
}

func TestResponseFormat_ErrorMessage(t *testing.T) {
	got := response(false, "appName 和 machineUuid 都不能为空", "")
	want := "success=false&msg=appName 和 machineUuid 都不能为空&result="
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ---- Handler Tests ----

func setupRouter(svc *mockService, secrets []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.AppConfig{
		FastId: config.FastIdConfig{
			Secrets: secrets,
		},
	}
	h := NewFastIdHandler(svc, cfg)
	h.RegisterRoutes(r)
	return r
}

func TestMachineID_SecretVerificationFailed_EmptySecret(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc, []string{"valid-secret"})

	form := url.Values{}
	form.Set("appName", "testApp")
	form.Set("machineUuid", "192.168.1.1:8080")
	// No secret provided

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	want := "success=false&msg=秘钥验证失败&result="
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestMachineID_SecretVerificationFailed_WrongSecret(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc, []string{"valid-secret"})

	form := url.Values{}
	form.Set("appName", "testApp")
	form.Set("machineUuid", "192.168.1.1:8080")
	form.Set("secret", "wrong-secret")

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	want := "success=false&msg=秘钥验证失败&result="
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestMachineID_SecretVerificationFailed_NoSecretsConfigured(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc, nil) // No secrets configured

	form := url.Values{}
	form.Set("appName", "testApp")
	form.Set("machineUuid", "192.168.1.1:8080")
	form.Set("secret", "any-secret")

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	want := "success=false&msg=秘钥验证失败&result="
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestMachineID_Success(t *testing.T) {
	svc := &mockService{
		getMachineIDFunc: func(namespace, groupID, appName, machineUUID string) (int, error) {
			// Verify parameters are passed through correctly
			if appName != "testApp" {
				t.Errorf("appName = %q, want %q", appName, "testApp")
			}
			if machineUUID != "192.168.1.1:8080" {
				t.Errorf("machineUUID = %q, want %q", machineUUID, "192.168.1.1:8080")
			}
			return 1, nil
		},
	}
	r := setupRouter(svc, []string{"valid-secret"})

	form := url.Values{}
	form.Set("namespace", "")
	form.Set("groupId", "")
	form.Set("appName", "testApp")
	form.Set("machineUuid", "192.168.1.1:8080")
	form.Set("secret", "valid-secret")

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	want := "success=true&msg=&result=1"
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestMachineID_ServiceError(t *testing.T) {
	svc := &mockService{
		getMachineIDFunc: func(namespace, groupID, appName, machineUUID string) (int, error) {
			return 0, &testError{msg: "appName 和 machineUuid 都不能为空"}
		},
	}
	r := setupRouter(svc, []string{"valid-secret"})

	form := url.Values{}
	form.Set("appName", "")
	form.Set("machineUuid", "")
	form.Set("secret", "valid-secret")

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	want := "success=false&msg=appName 和 machineUuid 都不能为空&result="
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestMachineID_WithNamespaceAndGroup(t *testing.T) {
	svc := &mockService{
		getMachineIDFunc: func(namespace, groupID, appName, machineUUID string) (int, error) {
			if namespace != "prod" {
				t.Errorf("namespace = %q, want %q", namespace, "prod")
			}
			if groupID != "group1" {
				t.Errorf("groupID = %q, want %q", groupID, "group1")
			}
			return 5, nil
		},
	}
	r := setupRouter(svc, []string{"valid-secret"})

	form := url.Values{}
	form.Set("namespace", "prod")
	form.Set("groupId", "group1")
	form.Set("appName", "testApp")
	form.Set("machineUuid", "pod-abc-123")
	form.Set("secret", "valid-secret")

	req := httptest.NewRequest(http.MethodPost, "/fastid/machineId", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	want := "success=true&msg=&result=5"
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestAppList_Success(t *testing.T) {
	now := model.Now()
	svc := &mockService{
		listAppsFunc: func() ([]model.App, error) {
			return []model.App{
				{ID: 1, Namespace: "PUBLIC", GroupID: "PUBLIC", Name: "app1", CreateTime: now},
			}, nil
		},
	}
	r := setupRouter(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/fastid/appList", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify it's a JSON array
	body := w.Body.String()
	if !strings.HasPrefix(body, "[") || !strings.HasSuffix(body, "]") {
		t.Errorf("expected JSON array, got: %s", body)
	}
	if !strings.Contains(body, `"name":"app1"`) {
		t.Errorf("response should contain app name, got: %s", body)
	}
	if !strings.Contains(body, `"groupId":"PUBLIC"`) {
		t.Errorf("response should use camelCase 'groupId', got: %s", body)
	}
}

func TestMachineList_Success(t *testing.T) {
	now := model.Now()
	svc := &mockService{
		listMachinesFunc: func(appID uint) ([]model.Machine, error) {
			if appID != 1 {
				t.Errorf("appID = %d, want 1", appID)
			}
			return []model.Machine{
				{ID: 1, AppID: 1, MachineUUID: "10.0.0.1:8080", MachineID: 1, CreateTime: now, BeatTime: now},
			}, nil
		},
	}
	r := setupRouter(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/fastid/machineList?appId=1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"machineUuid":"10.0.0.1:8080"`) {
		t.Errorf("response should contain machineUuid, got: %s", body)
	}
	if !strings.Contains(body, `"machineId":1`) {
		t.Errorf("response should contain machineId, got: %s", body)
	}
}

func TestMachineList_NoAppId(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/fastid/machineList", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	// Java returns null when appId is null, Spring serializes as empty body
	if w.Body.String() != "" {
		t.Errorf("body = %q, want %q", w.Body.String(), "")
	}
}

func TestHealth_Success(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	want := `{"status":"UP"}`
	if w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}

// ---- Verify Secret Tests ----

func TestVerifySecret_EmptySecret(t *testing.T) {
	h := &FastIdHandler{cfg: &config.AppConfig{FastId: config.FastIdConfig{Secrets: []string{"abc"}}}}
	if h.verifySecret("") {
		t.Error("expected false for empty secret")
	}
}

func TestVerifySecret_ValidSecret(t *testing.T) {
	h := &FastIdHandler{cfg: &config.AppConfig{FastId: config.FastIdConfig{Secrets: []string{"abc", "def"}}}}
	if !h.verifySecret("abc") {
		t.Error("expected true for valid secret")
	}
	if !h.verifySecret("def") {
		t.Error("expected true for valid secret")
	}
}

func TestVerifySecret_InvalidSecret(t *testing.T) {
	h := &FastIdHandler{cfg: &config.AppConfig{FastId: config.FastIdConfig{Secrets: []string{"abc"}}}}
	if h.verifySecret("xyz") {
		t.Error("expected false for invalid secret")
	}
}

func TestVerifySecret_NilSecrets(t *testing.T) {
	h := &FastIdHandler{cfg: &config.AppConfig{FastId: config.FastIdConfig{Secrets: nil}}}
	if h.verifySecret("abc") {
		t.Error("expected false when secrets is nil")
	}
}

// ---- Test helper ----

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
