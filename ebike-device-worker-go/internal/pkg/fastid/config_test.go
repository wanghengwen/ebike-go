package fastid

import "testing"

func TestApplyDefaultsHTTPURL(t *testing.T) {
	cfg := Config{URL: "http://xyy-fastid:8080"}
	cfg.applyDefaults()
	if cfg.UseHTTPS {
		t.Fatal("expected UseHTTPS=false for http:// URL")
	}
	if cfg.URL != "xyy-fastid:8080" {
		t.Fatalf("url=%s", cfg.URL)
	}
}

func TestApplyDefaultsHTTPSURL(t *testing.T) {
	cfg := Config{URL: "https://fastid.luopingtech.com"}
	cfg.applyDefaults()
	if !cfg.UseHTTPS {
		t.Fatal("expected UseHTTPS=true for https:// URL")
	}
	if cfg.URL != "fastid.luopingtech.com" {
		t.Fatalf("url=%s", cfg.URL)
	}
}

func TestApplyDefaultsBareHostDefaultsHTTPS(t *testing.T) {
	cfg := Config{URL: "fastid.luopingtech.com"}
	cfg.applyDefaults()
	if !cfg.UseHTTPS {
		t.Fatal("expected bare host to default to HTTPS")
	}
}

func TestApplyDefaultsBareHostExplicitHTTP(t *testing.T) {
	cfg := Config{
		URL:              "xyy-fastid.prod.svc.cluster.local:8080",
		UseHTTPS:         false,
		UseHTTPSExplicit: true,
	}
	cfg.applyDefaults()
	if cfg.UseHTTPS {
		t.Fatal("expected explicit UseHTTPS=false to be preserved")
	}
}
