package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/pkg/shadow"

	"github.com/gin-gonic/gin"
)

// closeNotifyRecorder satisfies http.CloseNotifier required by httputil.ReverseProxy
// when gin wraps the test recorder.
type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
	closeNotify chan bool
}

func newCloseNotifyRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotify:      make(chan bool, 1),
	}
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool { return r.closeNotify }

func TestPathInProxyList(t *testing.T) {
	list := []string{
		"/device/paas/lock",
		"/device/trajectory/saveDb",
	}
	if !config.PathInProxyList(list, "/device/paas/lock") {
		t.Fatal("expected exact match")
	}
	if !config.PathInProxyList(list, "/device/trajectory/saveDb") {
		t.Fatal("expected trajectory saveDb match")
	}
	if config.PathInProxyList(list, "/device/paas/device/detail") {
		t.Fatal("read endpoint should not match recordList")
	}
}

func TestIsLivePath(t *testing.T) {
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Xyy.Proxy.LiveList = []string{
		"/device/paas/device/detail",
		"/device/paas/device/list",
	}

	if !config.IsLivePath("/device/paas/device/detail") {
		t.Fatal("expected device/detail in liveList")
	}
	if !config.IsLivePath("/device/paas/device/list") {
		t.Fatal("expected device/list in liveList")
	}
	if config.IsLivePath("/device/paas/lock") {
		t.Fatal("lock should not be live")
	}
}

func TestProxyGateway_LiveRunsGoOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })

	java := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("liveList must not call Java")
	}))
	defer java.Close()

	config.GlobalConfig = &config.Config{
		DryRun: true,
		Xyy: config.XyyConfig{
			JavaServiceURL: java.URL,
			Proxy: config.ProxyConfig{
				TargetURL: java.URL,
				LiveList:  []string{"/device/paas/device/detail"},
			},
		},
	}

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/device/paas/device/detail", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/device/paas/device/detail", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := newCloseNotifyRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["from"] != "go" {
		t.Fatalf("client should get Go response: %s", w.Body.String())
	}
}

func TestProxyGateway_RecordProxiesJava(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })

	javaHits := 0
	java := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		javaHits++
		_, _ = w.Write([]byte(`{"success":true,"from":"java"}`))
	}))
	defer java.Close()

	config.GlobalConfig = &config.Config{
		DryRun: true,
		Xyy: config.XyyConfig{
			JavaServiceURL: java.URL,
			Proxy: config.ProxyConfig{
				TargetURL:  java.URL,
				RecordList: []string{"/device/paas/lock"},
			},
		},
	}

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/device/paas/lock", func(c *gin.Context) {
		t.Fatal("recordList must not run Go handler")
	})

	req := httptest.NewRequest(http.MethodPost, "/device/paas/lock", bytes.NewBufferString(`{"imei":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := newCloseNotifyRecorder()
	r.ServeHTTP(w, req)

	if javaHits != 1 {
		t.Fatalf("javaHits=%d", javaHits)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"from":"java"`)) {
		t.Fatalf("body=%s", w.Body.String())
	}
}

func TestProxyGateway_ShadowDualRunClientGetsJava(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })

	javaHits := 0
	java := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		javaHits++
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"imei":"x"}` {
			t.Errorf("java req body=%s", body)
		}
		_, _ = w.Write([]byte(`{"success":true,"code":"0","from":"java"}`))
	}))
	defer java.Close()

	config.GlobalConfig = &config.Config{
		DryRun: true,
		Xyy: config.XyyConfig{
			JavaServiceURL: java.URL,
			Proxy: config.ProxyConfig{
				TargetURL: java.URL,
			},
		},
	}

	goRan := false
	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/device/paas/deviceInfo", func(c *gin.Context) {
		goRan = true
		if !shadow.IsShadowTest(c.Request.Context()) {
			t.Error("expected shadow context")
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "code": "0", "from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/device/paas/deviceInfo", bytes.NewBufferString(`{"imei":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := newCloseNotifyRecorder()
	r.ServeHTTP(w, req)

	if !goRan {
		t.Fatal("Go handler should run in shadow mode")
	}
	if javaHits != 1 {
		t.Fatalf("javaHits=%d", javaHits)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"from":"java"`)) {
		t.Fatalf("client must receive Java response, got %s", w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte(`"from":"go"`)) {
		t.Fatal("client must not receive Go response during shadow")
	}
}
