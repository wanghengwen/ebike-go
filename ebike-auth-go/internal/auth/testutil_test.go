package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/rpc"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

type integrationEnv struct {
	router *gin.Engine
	redis  *miniredis.Miniredis
	mgmt   *httptest.Server
	user   *httptest.Server
}

func resetTenantMaps() {
	tenantMu.Lock()
	tenantAuthMap = make(map[string]*rpc.TenantAuthCo)
	tenantThirdAuthMap = make(map[string]*rpc.TenantThirdAuthCo)
	tenantMu.Unlock()
}

func seedTenantAuth(tenantID, grants, scope string) {
	tenantMu.Lock()
	tenantAuthMap[tenantID] = &rpc.TenantAuthCo{
		TenantId:             tenantID,
		AccessTokenValidity:  3600,
		RefreshTokenValidity: 86400,
		AuthorizedGrantTypes: grants,
		Scope:                scope,
	}
	tenantMu.Unlock()
}

func setupIntegrationEnv(t *testing.T, mode string) *integrationEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	resetTenantMaps()

	config.AppConfig.AuthMode = mode
	config.AppConfig.JwtSecret = testJwtSecret
	config.AppConfig.EnablePhoneVerify = false
	config.AppConfig.EnableEmailVerify = false
	config.AppConfig.EnableTenantSecretVerify = false
	config.AppConfig.GrayTenantIds = nil
	config.AppConfig.Aks = map[string]string{
		"service-key": "service-secret",
	}

	mr := miniredis.RunT(t)
	redis.Rdb = goredis.NewClient(&goredis.Options{Addr: mr.Addr()})

	rpc.InitRPCClient()
	mgmt := newManagementMockServer(t, mode)
	user := newUserMockServer(t, mode)
	os.Setenv("EBIKE_MANAGEMENT_URL", mgmt.URL)
	os.Setenv("EBIKE_USER_URL", user.URL)
	t.Cleanup(func() {
		os.Unsetenv("EBIKE_MANAGEMENT_URL")
		os.Unsetenv("EBIKE_USER_URL")
		mgmt.Close()
		user.Close()
		mr.Close()
	})

	seedTenantAuth(testTenantID, allGrantsForMode(mode), "all")

	r := gin.New()
	RegisterRoutes(r)
	return &integrationEnv{router: r, redis: mr, mgmt: mgmt, user: user}
}

func allGrantsForMode(mode string) string {
	if mode == "business" {
		return "password,phone_code,email_code,phone_secret,phone_password,refresh_token,ak"
	}
	return "password,phone_code,phone_secret,email_code,refresh_token,apple,google,facebook,wechat_miniapp,wechat_third_miniapp,alipay_miniapp,union_pay_miniapp,social_bound_phone,yudaoxing_app"
}

func newManagementMockServer(t *testing.T, mode string) *httptest.Server {
	t.Helper()
	active := true
	passwordHash := "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8" // sha256("password")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user/getUserByPin":
			pin := extractPinFromBody(r)
			user := rpc.UserCO{
				Pin: pin, Name: "Admin", Password: passwordHash,
				Authorities: []string{"ADMIN"}, SaasCode: []string{"saas-1"}, Status: &active,
			}
			if pin == "noperm001" {
				user.SaasCode = nil
			}
			writeJSON(w, rpc.ManagementResult[rpc.UserCO]{Code: "SUCCESS", Data: user})
		case "/user/getUserByPhone":
			writeJSON(w, rpc.ManagementResult[rpc.UserCO]{
				Code: "SUCCESS",
				Data: rpc.UserCO{
					Pin: "phone001", Name: "PhoneUser", Phone: "+86-13300000001", Password: passwordHash,
					Authorities: []string{"NORMAL"}, SaasCode: []string{"saas-1"}, Status: &active,
				},
			})
		case "/user/getUserByEmail":
			writeJSON(w, rpc.ManagementResult[rpc.UserCO]{
				Code: "SUCCESS",
				Data: rpc.UserCO{
					Pin: "email001", Name: "EmailUser", Email: "user@example.com", Password: passwordHash,
					Authorities: []string{"NORMAL"}, SaasCode: []string{"saas-1"}, Status: &active,
				},
			})
		case "/user/updateLastLogin":
			writeJSON(w, rpc.ManagementResult[string]{Code: "SUCCESS", Data: "ok"})
		case "/code/verify", "/email/code/verify":
			writeJSON(w, rpc.ManagementResult[bool]{Code: "SUCCESS", Data: true})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func extractPinFromBody(r *http.Request) string {
	var req struct {
		Pin string `json:"pin"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	return req.Pin
}

func newUserMockServer(t *testing.T, mode string) *httptest.Server {
	t.Helper()
	passwordHash := "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/userAuthInfo/selectUserByPin":
			writeJSON(w, rpc.ManagementResult[rpc.UserAuthInfoCo]{
				Code: "SUCCESS",
				Data: rpc.UserAuthInfoCo{
					Pin: "client001", Nickname: "ClientUser", Password: passwordHash, Status: 1,
				},
			})
		case "/userAuthInfo/selectUserByPhone":
			writeJSON(w, rpc.ManagementResult[rpc.UserAuthInfoCo]{Code: "SUCCESS", Data: rpc.UserAuthInfoCo{}})
		case "/userAuthInfo/phoneRegister":
			writeJSON(w, rpc.ManagementResult[rpc.UserAuthInfoCo]{
				Code: "SUCCESS",
				Data: rpc.UserAuthInfoCo{
					Pin: "newphone001", Nickname: "NewPhone", Phone: "+86-13300001111", Status: 1,
				},
			})
		case "/userAuthInfo/selectUserBySocial", "/userAuthInfo/socialUserRegister", "/userAuthInfo/addThirdUserInfo":
			writeJSON(w, rpc.ManagementResult[rpc.UserAuthInfoCo]{Code: "SUCCESS", Data: rpc.UserAuthInfoCo{}})
		case "/user/updateLastLogin", "/gray/syncUserData":
			writeJSON(w, rpc.ManagementResult[string]{Code: "SUCCESS", Data: "ok"})
		default:
			if mode == "client" {
				t.Logf("unexpected user path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

type apiResult struct {
	Success bool            `json:"success"`
	Code    string          `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

func postOAuthToken(t *testing.T, router *gin.Engine, tenantID, password string, body map[string]interface{}, headers map[string]string) (int, apiResult) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(payload))
	req.SetBasicAuth(tenantID, password)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result apiResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return w.Code, result
}

func postLogout(t *testing.T, router *gin.Engine, authorities string, body map[string]interface{}) apiResult {
	t.Helper()
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/oauth/logout", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if authorities != "" {
		req.Header.Set("authorities", authorities)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var result apiResult
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

func encodeAuthorities(t *testing.T, payload map[string]interface{}) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal authorities payload: %v", err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}
