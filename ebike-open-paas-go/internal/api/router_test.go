package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestIsLocalhostRequestIgnoresSpoofableHeaders is the reason this check moved off
// the Host header: the header is chosen by the caller, so `curl -H "Host:
// localhost"` through the ingress would otherwise let anyone pull this pod out of
// Nacos and take it out of the load balancer.
func TestIsLocalhostRequestIgnoresSpoofableHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		remoteAddr string
		host       string
		forwarded  string
		want       bool
	}{
		{"prestop hook", "127.0.0.1:54321", "127.0.0.1:8080", "", true},
		{"prestop hook by name", "127.0.0.1:54321", "localhost:8080", "", true},
		{"ipv6 loopback", "[::1]:54321", "localhost:8080", "", true},
		{"loopback range", "127.0.0.53:54321", "localhost:8080", "", true},
		{"spoofed host header", "203.0.113.9:44444", "localhost:8080", "", false},
		{"spoofed host header no port", "203.0.113.9:44444", "127.0.0.1", "", false},
		{"spoofed forwarded for", "203.0.113.9:44444", "paas.luopingtech.com", "127.0.0.1", false},
		{"pod ip", "10.42.0.7:33333", "paas.luopingtech.com", "", false},
		{"remote addr without port", "203.0.113.9", "localhost", "", false},
	}

	for _, tc := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "http://"+tc.host+"/actuator/deregisterService", nil)
		c.Request.RemoteAddr = tc.remoteAddr
		c.Request.Host = tc.host
		if tc.forwarded != "" {
			c.Request.Header.Set("X-Forwarded-For", tc.forwarded)
		}

		if got := isLocalhostRequest(c); got != tc.want {
			t.Errorf("%s: isLocalhostRequest(remote=%s, host=%s, xff=%s) = %v, want %v",
				tc.name, tc.remoteAddr, tc.host, tc.forwarded, got, tc.want)
		}
	}
}

// TestDeregisterServiceRefusesRemoteCallers checks the handler and not just the
// predicate, since the response shape is what the preStop hook reads.
func TestDeregisterServiceRefusesRemoteCallers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "http://localhost:8080/actuator/deregisterService", nil)
	c.Request.RemoteAddr = "203.0.113.9:44444"

	DeregisterService(c)

	if body := rec.Body.String(); !strings.Contains(body, `"success":false`) {
		t.Errorf("body = %s, want success:false for a remote caller", body)
	}
}
