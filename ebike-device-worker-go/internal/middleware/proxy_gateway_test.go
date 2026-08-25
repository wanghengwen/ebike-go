package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ebike-device-worker-go/internal/config"

	"github.com/gin-gonic/gin"
)

type trackingWriter struct {
	header http.Header
	code   int
	body   []byte
}

func (t *trackingWriter) Header() http.Header {
	if t.header == nil {
		t.header = make(http.Header)
	}
	return t.header
}

func (t *trackingWriter) Write(b []byte) (int, error) {
	if t.code == 0 {
		t.code = http.StatusOK
	}
	t.body = append(t.body, b...)
	return len(b), nil
}

func (t *trackingWriter) WriteHeader(statusCode int) {
	t.code = statusCode
}

func (t *trackingWriter) CloseNotify() <-chan bool { return nil }

func (t *trackingWriter) Pusher() http.Pusher { return nil }

func (t *trackingWriter) Flush() {}

func (t *trackingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}

func (t *trackingWriter) Size() int { return len(t.body) }

func (t *trackingWriter) Status() int { return t.code }

func (t *trackingWriter) Written() bool { return t.code != 0 || len(t.body) > 0 }

func (t *trackingWriter) WriteString(s string) (int, error) {
	return t.Write([]byte(s))
}

func (t *trackingWriter) WriteHeaderNow() {}

func TestResponseRecorderDoesNotMutateUnderlyingWriter(t *testing.T) {
	base := &trackingWriter{}
	recorder := &responseRecorder{
		ResponseWriter: base,
		body:           bytes.NewBuffer(nil),
	}

	recorder.Header().Set("X-Go-Shadow", "1")
	recorder.WriteHeader(http.StatusCreated)
	_, _ = recorder.Write([]byte(`{"ok":true}`))

	if base.header != nil && base.header.Get("X-Go-Shadow") != "" {
		t.Fatalf("underlying writer header was mutated: %v", base.header)
	}
	if base.code != 0 {
		t.Fatalf("underlying writer status was committed: %d", base.code)
	}
	if len(base.body) != 0 {
		t.Fatalf("underlying writer body was written: %q", base.body)
	}
	if recorder.Header().Get("X-Go-Shadow") != "1" {
		t.Fatalf("recorder should capture header")
	}
	if recorder.statusCode != http.StatusCreated {
		t.Fatalf("recorder status=%d want %d", recorder.statusCode, http.StatusCreated)
	}
}

type closeNotifierRecorder struct {
	*httptest.ResponseRecorder
}

func (c *closeNotifierRecorder) CloseNotify() <-chan bool {
	return nil
}

func TestProxyGatewayDryRunMode(t *testing.T) {
	// 1. Setup mock Java backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"from":"java"}`))
	}))
	defer backend.Close()

	// 2. Setup Config
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Proxy.TargetUrl = backend.URL

	// 3. Enable DRY_RUN
	t.Setenv("DRY_RUN", "true")

	// 4. Setup Gin Router
	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/test-endpoint", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"from": "go"})
	})

	// 5. Perform request
	req := httptest.NewRequest("POST", "/test-endpoint", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cnw := &closeNotifierRecorder{ResponseRecorder: w}

	r.ServeHTTP(cnw, req)

	// 6. Verify client receives Java response, not Go response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body != `{"from":"java"}` {
		t.Fatalf("expected body %q, got %q (did Go response leak?)", `{"from":"java"}`, body)
	}
}

func TestProxyGatewayForwardsRequestBody(t *testing.T) {
	wantBody := []byte(`{"imei":"860123456789012","startTime":1700000000000,"endTime":1700003600000}`)
	var gotBody []byte
	var gotContentLength string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentLength = r.Header.Get("Content-Length")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer backend.Close()

	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Proxy.TargetUrl = backend.URL
	t.Setenv("DRY_RUN", "false")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ProxyGateway())
	r.POST("/cmd/getEbikeCmd", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"from": "go"})
	})

	req := httptest.NewRequest(http.MethodPost, "/cmd/getEbikeCmd", bytes.NewReader(wantBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(wantBody))
	w := httptest.NewRecorder()
	cnw := &closeNotifierRecorder{ResponseRecorder: w}
	r.ServeHTTP(cnw, req)

	if string(gotBody) != string(wantBody) {
		t.Fatalf("upstream body=%q want=%q", gotBody, wantBody)
	}
	if gotContentLength != strconv.Itoa(len(wantBody)) {
		t.Fatalf("upstream Content-Length=%q want=%d", gotContentLength, len(wantBody))
	}
	if w.Body.String() != `{"success":true}` {
		t.Fatalf("client body=%q", w.Body.String())
	}
}
