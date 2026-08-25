package service

import (
	"encoding/json"
	"strings"
	"testing"

	"ebike-fence-go/internal/api/dto"
)

func TestPartAnalysisByAppSuccessMatchesJavaShape(t *testing.T) {
	co := partAnalysisByAppSuccess()
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, `"name":"helmet"`) {
		t.Fatalf("success must not include part name, got %s", s)
	}
	if !strings.Contains(s, `"name":null`) {
		t.Fatalf("expected name:null, got %s", s)
	}
	if !strings.Contains(s, `"result":true`) || !strings.Contains(s, `"canUse":true`) {
		t.Fatalf("expected result/canUse true, got %s", s)
	}
}

func TestPartAnalysisByAppFailureKeepsPartName(t *testing.T) {
	fail := false
	co := partAnalysisByAppFailure("helmet", &fail)
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"name":"helmet"`) {
		t.Fatalf("expected part name on failure, got %s", s)
	}
	if !strings.Contains(s, `"canUse":true`) {
		t.Fatalf("expected default canUse true, got %s", s)
	}
}

func TestPartHelmetTempParkingResultCanUseDefaultTrue(t *testing.T) {
	co := partHelmetTempParkingResult(true, true)
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"canUse":true`) {
		t.Fatalf("expected canUse true like Java field default, got %s", s)
	}
	if !strings.Contains(s, `"name":"helmet"`) || !strings.Contains(s, `"result":true`) || !strings.Contains(s, `"izExist":true`) {
		t.Fatalf("unexpected shape: %s", s)
	}
}

func TestPartNamePtrJSONNull(t *testing.T) {
	b, err := json.Marshal(dto.PartAnalysisResultCO{Name: dto.PartNamePtr("")})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":null,"izExist":null,"canUse":null,"useType":null,"result":null,"state":null,"ext":null}` {
		t.Fatalf("unexpected json: %s", string(b))
	}
}
