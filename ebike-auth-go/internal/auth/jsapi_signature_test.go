package auth

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

func TestGenerateJsApiSignature(t *testing.T) {
	ticket := "jsapi_ticket_value"
	nonce := "nonce_value"
	timestamp := int64(1718356800)
	pageURL := "https://example.com/h5/scan"

	got := generateJsApiSignature(ticket, nonce, timestamp, pageURL)
	raw := "jsapi_ticket=jsapi_ticket_value&noncestr=nonce_value&timestamp=1718356800&url=https://example.com/h5/scan"
	wantHash := sha1.Sum([]byte(raw))
	want := hex.EncodeToString(wantHash[:])
	if got != want {
		t.Fatalf("signature mismatch: got %s want %s", got, want)
	}
}

func TestGenerateJsApiNonce(t *testing.T) {
	nonce := generateJsApiNonce()
	if len(nonce) == 0 {
		t.Fatal("expected non-empty nonce")
	}
	nonce2 := generateJsApiNonce()
	if nonce == nonce2 {
		t.Fatal("expected random nonce values to differ")
	}
}

func TestBuildJsApiSignatureDto(t *testing.T) {
	dto := buildJsApiSignatureDto("ticket", "https://example.com")
	if dto.Signature == "" || dto.Nonce == "" || dto.URL != "https://example.com" || dto.Timestamp <= 0 {
		t.Fatalf("unexpected dto: %+v", dto)
	}
}

func TestClientJsApiSignatureHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	weixinMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":7200}`))
		case "/cgi-bin/ticket/getticket":
			_, _ = w.Write([]byte(`{"ticket":"mock-ticket","expires_in":7200}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer weixinMock.Close()

	config.AppConfig.WeixinPublicPlatform = config.WeixinPublicPlatformConfig{
		AppId:      "wx-test",
		AppSecret:  "test-secret",
		APIBaseURL: weixinMock.URL,
	}

	mr := miniredis.RunT(t)
	redis.Rdb = goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(mr.Close)

	r := gin.New()
	RegisterRoutes(r)

	t.Run("empty config returns success with null data", func(t *testing.T) {
		saved := config.AppConfig.WeixinPublicPlatform
		config.AppConfig.WeixinPublicPlatform = config.WeixinPublicPlatformConfig{}
		defer func() { config.AppConfig.WeixinPublicPlatform = saved }()

		body := `{"url":"https://client.example.com/h5/scan"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/jsapi/signature", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var result apiResult
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		// 对齐 Java：获取失败返回 success=true、data=null
		if !result.Success {
			t.Fatalf("expected success=true on failure (aligned with Java), got %+v", result)
		}
		if len(result.Data) != 0 && string(result.Data) != "null" {
			t.Fatalf("expected null data, got %s", string(result.Data))
		}
	})

	t.Run("success", func(t *testing.T) {
		body := `{"url":"https://client.example.com/h5/scan"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/jsapi/signature", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}

		var result apiResult
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !result.Success {
			t.Fatalf("expected success, got %+v", result)
		}

		var dto JsApiSignatureDto
		if err := json.Unmarshal(result.Data, &dto); err != nil {
			t.Fatalf("decode data: %v", err)
		}
		if dto.Signature == "" || dto.Nonce == "" || dto.URL != "https://client.example.com/h5/scan" || dto.Timestamp <= 0 {
			t.Fatalf("unexpected dto: %+v", dto)
		}
	})
}
