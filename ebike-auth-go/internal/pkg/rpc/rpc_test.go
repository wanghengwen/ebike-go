package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ebike-auth-go/internal/pkg/config"
)

func TestRPCMethods(t *testing.T) {
	InitRPCClient()
	config.AppConfig.EnablePhoneVerify = true

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/tenant/tenantAuth/queryAll" {
			res := ManagementResult[[]TenantAuthCo]{
				Code: "SUCCESS",
				Msg:  "success",
				Data: []TenantAuthCo{
					{
						TenantId:             "tenantA",
						TenantSecret:         "secretA",
						AccessTokenValidity:  3600,
						RefreshTokenValidity: 7200,
						AuthorizedGrantTypes: "password",
						Scope:                "all",
					},
				},
			}
			data, _ := json.Marshal(res)
			w.Write(data)
			return
		}

		if r.URL.Path == "/tenant/tenantThirdAuth/queryAll" {
			res := ManagementResult[[]TenantThirdAuthCo]{
				Code: "SUCCESS",
				Msg:  "success",
				Data: []TenantThirdAuthCo{
					{
						TenantId:  "tenantA",
						ThirdType: 4,
						AppId:     "wxAppId",
						AppSecret: "wxSecret",
					},
				},
			}
			data, _ := json.Marshal(res)
			w.Write(data)
			return
		}

		if r.URL.Path == "/user/getUserByPin" {
			status := true
			res := ManagementResult[UserCO]{
				Code: "SUCCESS",
				Msg:  "success",
				Data: UserCO{
					Pin:         "userPinA",
					Name:        "NicknameA",
					Phone:       "13388888888",
					Email:       "test@example.com",
					Password:    "hashedPassword",
					Authorities: []string{"ADMIN"},
					SaasCode:    []string{"code1"},
					Status:      &status,
				},
			}
			data, _ := json.Marshal(res)
			w.Write(data)
			return
		}

		if r.URL.Path == "/userAuthInfo/phoneRegister" {
			res := ManagementResult[UserAuthInfoCo]{
				Code: "SUCCESS",
				Msg:  "success",
				Data: UserAuthInfoCo{
					Pin:         "newPin",
					Nickname:    "NewUser",
					Phone:       "13399999999",
					Authorities: []string{"USER"},
					Status:      1,
				},
			}
			data, _ := json.Marshal(res)
			w.Write(data)
			return
		}

		if r.URL.Path == "/code/verify" {
			res := ManagementResult[bool]{
				Code: "SUCCESS",
				Msg:  "success",
				Data: true,
			}
			data, _ := json.Marshal(res)
			w.Write(data)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	os.Setenv("EBIKE_MANAGEMENT_URL", ts.URL)
	os.Setenv("EBIKE_USER_URL", ts.URL)
	defer func() {
		os.Unsetenv("EBIKE_MANAGEMENT_URL")
		os.Unsetenv("EBIKE_USER_URL")
	}()

	ctx := context.Background()
	cmdCtx := NewCommandContext("tenantA", "trace-test", "ios", "device-1", "")

	t.Run("QueryAllTenantAuth", func(t *testing.T) {
		auths, err := QueryAllTenantAuth()
		if err != nil {
			t.Fatalf("QueryAllTenantAuth failed: %v", err)
		}
		if len(auths) != 1 || auths[0].TenantId != "tenantA" {
			t.Errorf("unexpected QueryAllTenantAuth result: %+v", auths)
		}
	})

	t.Run("QueryAllTenantThirdAuth", func(t *testing.T) {
		thirds, err := QueryAllTenantThirdAuth()
		if err != nil {
			t.Fatalf("QueryAllTenantThirdAuth failed: %v", err)
		}
		if len(thirds) != 1 || thirds[0].AppId != "wxAppId" {
			t.Errorf("unexpected QueryAllTenantThirdAuth result: %+v", thirds)
		}
	})

	t.Run("BusinessGetUserByPin", func(t *testing.T) {
		user, err := BusinessGetUserByPin(ctx, cmdCtx, "userPinA")
		if err != nil {
			t.Fatalf("BusinessGetUserByPin failed: %v", err)
		}
		if user.Pin != "userPinA" || user.Nickname != "NicknameA" || user.Status != 1 {
			t.Errorf("unexpected BusinessGetUserByPin result: %+v", user)
		}
	})

	t.Run("ClientUserPhoneRegister", func(t *testing.T) {
		user, err := ClientUserPhoneRegister(ctx, cmdCtx, "13399999999")
		if err != nil {
			t.Fatalf("ClientUserPhoneRegister failed: %v", err)
		}
		if user.Pin != "newPin" || user.Phone != "13399999999" {
			t.Errorf("unexpected ClientUserPhoneRegister result: %+v", user)
		}
	})

	t.Run("VerifyPhoneMessageCode", func(t *testing.T) {
		ok, err := VerifyPhoneMessageCode(ctx, cmdCtx, "1", "13388888888", "123456")
		if err != nil {
			t.Fatalf("VerifyPhoneMessageCode failed: %v", err)
		}
		if !ok {
			t.Errorf("expected verification to succeed")
		}
	})
}
