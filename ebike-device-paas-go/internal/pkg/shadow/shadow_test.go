package shadow

import (
	"context"
	"testing"

	"ebike-device-paas-go/internal/pkg/config"
)

func TestCompareWithJavaNoop(t *testing.T) {
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })

	config.GlobalConfig = &config.Config{
		DryRun: true,
		Xyy: config.XyyConfig{
			JavaServiceURL: "http://java.example:8080",
		},
	}

	// Compatibility shim: must not panic or dial Java.
	CompareWithJava("POST", "/device/paas/deviceInfo", []byte(`{}`), []byte(`{"success":true}`))
}

func TestWithShadowTest(t *testing.T) {
	ctx := context.Background()
	if IsShadowTest(ctx) {
		t.Fatal("expected false")
	}
	ctx = WithShadowTest(ctx)
	if !IsShadowTest(ctx) {
		t.Fatal("expected true")
	}
}

func TestReportMatch(t *testing.T) {
	saved := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = saved })
	config.GlobalConfig = &config.Config{}

	a := []byte(`{"success":true,"code":"0","data":{"x":1}}`)
	b := []byte(`{"success":true,"code":"0","data":{"x":1}}`)
	if !Report("/t", nil, a, b) {
		t.Fatal("expected MATCH")
	}
}
