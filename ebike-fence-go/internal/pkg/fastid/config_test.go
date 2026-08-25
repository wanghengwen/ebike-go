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

func TestSnowflakeCompose(t *testing.T) {
	g, err := NewForTest(Config{
		InstanceNoBits: 12,
		SequenceBits:   19,
		DriftTime:      10,
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	id1 := g.Next()
	id2 := g.Next()
	if id1 == id2 {
		t.Fatal("ids should differ")
	}
}
