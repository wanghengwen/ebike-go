package management

import (
	"testing"

	"ebike-service-client-go/internal/pkg/config"
)

func TestResolveMapServiceBaseURLFromConfig(t *testing.T) {
	config.GlobalConfig = config.AppConfig{}
	config.GlobalConfig.Xyy.MapServiceConfig.Url = "http://map-service.test:8080/"

	baseURL, source := resolveMapServiceBaseURL()
	if baseURL != "http://map-service.test:8080" {
		t.Fatalf("baseURL = %q", baseURL)
	}
	if source != "config" {
		t.Fatalf("source = %q", source)
	}
}

func TestResolveMapServiceBaseURLEmptyWithoutNacos(t *testing.T) {
	config.GlobalConfig = config.AppConfig{}
	baseURL, source := resolveMapServiceBaseURL()
	if baseURL != "" || source != "" {
		t.Fatalf("expected empty, got baseURL=%q source=%q", baseURL, source)
	}
}
