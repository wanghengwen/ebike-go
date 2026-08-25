package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ebike-service-client-go/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

type closeNotifierRecorder struct {
	*httptest.ResponseRecorder
}

func (c *closeNotifierRecorder) CloseNotify() <-chan bool {
	return nil
}

func serveRequest(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(&closeNotifierRecorder{ResponseRecorder: w}, req)
	return w
}

func setProxyTestConfig(target string, live, record []string) {
	config.GlobalConfig.Proxy.TargetURL = target
	config.GlobalConfig.Proxy.LiveList = live
	config.GlobalConfig.Proxy.RecordList = record
}

func TestProxyResponseRecorderCapturesBody(t *testing.T) {
	recorder := &proxyResponseRecorder{
		body: bytes.NewBuffer(nil),
	}

	recorder.WriteHeader(http.StatusCreated)
	_, _ = recorder.Write([]byte(`{"ok":true}`))

	if recorder.statusCode != http.StatusCreated {
		t.Fatalf("recorder status=%d want %d", recorder.statusCode, http.StatusCreated)
	}
	if recorder.body.String() != `{"ok":true}` {
		t.Fatalf("body=%q", recorder.body.String())
	}
}

func TestPathInProxyList(t *testing.T) {
	list := []string{"/client/order/detail", "/callback/pay/score/"}
	if !pathInProxyList(list, "/client/order/detail") {
		t.Fatal("exact match")
	}
	if !pathInProxyList(list, "/callback/pay/score/1000/paySuccessNotify") {
		t.Fatal("prefix match")
	}
	if pathInProxyList(list, "/client/order/list") {
		t.Fatal("should not match")
	}
}

func TestProxyGatewayLiveList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Java backend should not be called for live_list")
	}))
	defer backend.Close()

	setProxyTestConfig(backend.URL, []string{"/client/order/detail"}, nil)
	t.Setenv("DRY_RUN", "true")

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/client/order/detail", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/client/order/detail", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := serveRequest(r, req)

	if w.Body.String() != `{"from":"go"}` {
		t.Fatalf("body=%q want go response", w.Body.String())
	}
}

func TestProxyGatewayDryRunShadow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"from":"java"}`))
	}))
	defer backend.Close()

	setProxyTestConfig(backend.URL, nil, nil)
	t.Setenv("DRY_RUN", "true")

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/client/order/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/client/order/list", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := serveRequest(r, req)

	if w.Body.String() != `{"from":"java"}` {
		t.Fatalf("client should receive Java body=%q", w.Body.String())
	}
}

func TestProxyGatewayRecordList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	wantBody := []byte(`{"write":true}`)
	var gotBody []byte

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"from":"java"}`))
	}))
	defer backend.Close()

	setProxyTestConfig(backend.URL, nil, []string{"/client/order/closeOrder"})
	t.Setenv("DRY_RUN", "true")

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/client/order/closeOrder", func(c *gin.Context) {
		t.Fatal("Go handler must not run for record_list")
	})

	req := httptest.NewRequest(http.MethodPost, "/client/order/closeOrder", bytes.NewReader(wantBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(wantBody))
	w := serveRequest(r, req)

	if string(gotBody) != string(wantBody) {
		t.Fatalf("upstream body=%q want=%q", gotBody, wantBody)
	}
	if w.Body.String() != `{"from":"java"}` {
		t.Fatalf("body=%q", w.Body.String())
	}
}

func TestProxyGatewayForwardsRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	wantBody := []byte(`{"imei":"860123456789012"}`)
	var gotContentLength string
	var gotBody []byte

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentLength = r.Header.Get("Content-Length")
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer backend.Close()

	setProxyTestConfig(backend.URL, nil, nil)
	t.Setenv("DRY_RUN", "false")

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/client/order/list", func(c *gin.Context) {
		t.Fatal("should proxy without running handler")
	})

	req := httptest.NewRequest(http.MethodPost, "/client/order/list", bytes.NewReader(wantBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(wantBody))
	w := serveRequest(r, req)

	if string(gotBody) != string(wantBody) {
		t.Fatalf("upstream body=%q", gotBody)
	}
	if gotContentLength != strconv.Itoa(len(wantBody)) {
		t.Fatalf("Content-Length=%q", gotContentLength)
	}
	if w.Body.String() != `{"success":true}` {
		t.Fatalf("client body=%q", w.Body.String())
	}
}

func TestProxyGatewaySkipsWhenTargetEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setProxyTestConfig("", nil, nil)

	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/client/order/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/client/order/list", nil)
	w := serveRequest(r, req)

	if w.Body.String() != `{"from":"go"}` {
		t.Fatalf("body=%q", w.Body.String())
	}
}
