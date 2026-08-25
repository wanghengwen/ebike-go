package clientconfig

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvertAdConfigDTOParsesStringIzOn(t *testing.T) {
	raw := json.RawMessage(`{
		"bannerAds":"1",
		"motivateAds":"2",
		"plaqueAds":"3",
		"videoAds":"4",
		"serviceId":"274684702135689202",
		"izOn":"[{\"type\":\"banner\",\"enable\":true}]"
	}`)
	out := convertAdConfigDTO(raw)
	s := string(out)
	if strings.Contains(s, `"izOn":"[`) {
		t.Fatalf("izOn should be array not string, got %s", s)
	}
	if !strings.Contains(s, `"izOn":[`) {
		t.Fatalf("expected izOn array in output, got %s", s)
	}
	if !strings.Contains(s, `"serviceId":"274684702135689202"`) {
		t.Fatalf("expected string serviceId, got %s", s)
	}
}

func TestConvertAdConfigDTOEmptyIzOn(t *testing.T) {
	raw := json.RawMessage(`{"serviceId":"1","izOn":""}`)
	out := convertAdConfigDTO(raw)
	if !strings.Contains(string(out), `"izOn":[]`) {
		t.Fatalf("expected empty izOn array, got %s", out)
	}
}
