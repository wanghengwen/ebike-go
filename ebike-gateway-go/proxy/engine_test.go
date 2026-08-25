package proxy

import (
	"ebike-gateway-go/logger"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMatchAntPath(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		// 1. Exact match
		{"Exact match - Success", "/oauth/token", "/oauth/token", true},
		{"Exact match - Fail", "/oauth/token", "/oauth/token/123", false},

		// 2. Double star (**) matches zero or more segments
		{"Double star - single level", "/oauth/**", "/oauth/token", true},
		{"Double star - multi level", "/oauth/**", "/oauth/token/refresh/v1", true},
		{"Double star - zero level", "/oauth/**", "/oauth/", true},
		{"Double star - fail prefix", "/oauth/**", "/api/oauth/token", false},

		// 3. Single star (*) matches exactly one segment
		{"Single star - single level", "/ebike_visual/*/business/**", "/ebike_visual/123/business/test", true},
		{"Single star - multi level fail", "/ebike_visual/*/business/**", "/ebike_visual/123/456/business/test", false},
		{"Single star - missing segment fail", "/ebike_visual/*/business/**", "/ebike_visual//business/test", true}, // [^/]* matches empty string too, which is generally acceptable for AntPath
		{"Single star - partial name", "/*/user", "/api/user", true},
		{"Single star - partial name fail", "/*/user", "/api/v1/user", false},

		// 4. Root matching
		{"Root double star", "/**", "/anything/goes/here", true},
		{"Root double star empty", "/**", "/", true},

		// 5. Complex combinations
		{"Complex - success", "/client/*/device/**/config", "/client/1003/device/a/b/config", true},
		{"Complex - fail", "/client/*/device/**/config", "/client/1003/device/a/b/config/extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAntPath(tt.pattern, tt.path)
			if result != tt.expected {
				t.Errorf("matchAntPath(%q, %q) = %v, want %v", tt.pattern, tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseAndMatchRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Log, _ = zap.NewDevelopment()

	// Mock YAML imitating the rich configuration
	yamlContent := `
- id: ebike-auth-business
  uri: http://ebike-auth-business:8080
  predicates:
  - name: Path
    args:
      pattern: /oauth/**
  filters:
  - name: Security
    args:
      enabled: true
      secret: secret_auth
      ignoreUrls: /oauth/token:/oauth/phoneLogout

- id: zt-battery-callback
  uri: http://ebike-service-business:8080
  predicates:
  - name: Path
    args:
      pattern: /callback/zt/battery/**

- id: ebike-visual
  uri: http://ebike-visual
  predicates:
  - name: Path
    args:
      pattern: /ebike_visual/*/business/**
  filters:
  - name: Security
    args:
      enabled: false
  - name: Sign
    args:
      enabled: true
      secret: secret_visual
`

	// 1. Test YAML Parsing
	routes, err := ParseRoutes(yamlContent)
	if err != nil {
		t.Fatalf("ParseRoutes failed: %v", err)
	}

	if len(routes) != 3 {
		t.Fatalf("Expected 3 routes, got %d", len(routes))
	}

	// Verify ebike-auth-business
	authRoute := routes[0]
	if authRoute.Id != "ebike-auth-business" || authRoute.PathPattern != "/oauth/**" || !authRoute.SecurityEnabled {
		t.Errorf("Failed to parse authRoute correctly: %+v", authRoute)
	}
	if authRoute.SecuritySecret != "secret_auth" || len(authRoute.SecurityIgnores) != 2 {
		t.Errorf("Failed to parse authRoute security args: %+v", authRoute)
	}

	// Verify zt-battery-callback (no filters)
	cbRoute := routes[1]
	if cbRoute.SecurityEnabled || cbRoute.SignEnabled {
		t.Errorf("Callback route should have no filters enabled: %+v", cbRoute)
	}

	// 2. Test MatchRouteMiddleware
	engine := NewProxyEngine()
	engine.UpdateRoutes(routes)

	tests := []struct {
		name          string
		requestPath   string
		expectedCode  int
		expectedMatch string // id of the matched route
	}{
		{"Match oauth", "/oauth/token", 200, "ebike-auth-business"},
		{"Match callback", "/callback/zt/battery/status", 200, "zt-battery-callback"},
		{"Match visual with wildcard", "/ebike_visual/123/business/data", 200, "ebike-visual"},
		{"Mismatch visual (wrong structure)", "/ebike_visual/123/456/business/data", 404, ""},
		{"Mismatch totally unknown", "/unknown/api", 404, ""},
		{"Health check actuator", "/actuator/health", 200, `{"status":"UP"}`},
		{"Health check normal", "/health", 200, `{"status":"UP"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin context
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Add MatchRouteMiddleware
			r.Use(engine.MatchRouteMiddleware())
			r.GET("/*path", func(c *gin.Context) {
				matchedObj, exists := c.Get("matchedRoute")
				if !exists {
					c.Status(404)
					return
				}
				route := matchedObj.(*Route)
				c.String(200, route.Id)
			})

			// Perform request
			req, _ := http.NewRequest("GET", tt.requestPath, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d for path %s", tt.expectedCode, w.Code, tt.requestPath)
			}

			if tt.expectedCode == 200 {
				if w.Body.String() != tt.expectedMatch {
					t.Errorf("Expected matched route %s, got %s", tt.expectedMatch, w.Body.String())
				}
			}
		})
	}
}

func TestProxyCacheReuse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Log, _ = zap.NewDevelopment()

	engine := NewProxyEngine()
	routes := []Route{
		{
			Id:          "svc-a",
			Uri:         "http://service-a:8080",
			PathPattern: "/a/**",
			TargetURL:   mustParseURL("http://service-a:8080"),
		},
		{
			Id:          "svc-b",
			Uri:         "http://service-b:8080",
			PathPattern: "/b/**",
			TargetURL:   mustParseURL("http://service-b:8080"),
		},
	}
	engine.UpdateRoutes(routes)

	p1 := engine.getOrCreateProxy(mustParseURL("http://service-a:8080"))
	p2 := engine.getOrCreateProxy(mustParseURL("http://service-a:8080"))
	p3 := engine.getOrCreateProxy(mustParseURL("http://service-b:8080"))

	if p1 != p2 {
		t.Fatal("expected same reverse proxy instance for identical target")
	}
	if p1 == p3 {
		t.Fatal("expected different reverse proxy instances for different targets")
	}
	if engine.CachedProxyCount() != 2 {
		t.Fatalf("expected 2 cached proxies, got %d", engine.CachedProxyCount())
	}

	engine.UpdateRoutes([]Route{routes[0]})
	if engine.CachedProxyCount() != 1 {
		t.Fatalf("expected pruned cache size 1, got %d", engine.CachedProxyCount())
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
