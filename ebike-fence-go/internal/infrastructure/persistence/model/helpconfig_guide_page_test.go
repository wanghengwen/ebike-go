package model

import (
	"encoding/json"
	"testing"
)

func TestTConfigGuidePageRedisMarshalRoundTrip(t *testing.T) {
	pages := `[{"chainType":3,"picUrl":"https://example.com/a.png"}]`
	row := TConfigGuidePage{
		ServiceID:  364881328207300695,
		GuidePages: pages,
	}
	b, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	got, ok := m["guidePages"]
	if !ok {
		t.Fatalf("guidePages missing from marshaled JSON: %s", string(b))
	}
	var s string
	if err := json.Unmarshal(got, &s); err != nil {
		t.Fatalf("guidePages not a string: %v", err)
	}
	if s != pages {
		t.Fatalf("expected guidePages %q, got %q", pages, s)
	}

	var cached TConfigGuidePage
	if err := json.Unmarshal(b, &cached); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if cached.GuidePages != pages {
		t.Fatalf("expected guidePages %q, got %q", pages, cached.GuidePages)
	}
}

func TestTConfigGuidePageUnmarshalIgnoresGuidePagesArray(t *testing.T) {
	raw := `{"serviceId":364881328207300695,"guidePages":[{"chainType":3}],"pageNumEsc":1}`
	var row TConfigGuidePage
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if row.ServiceID != 364881328207300695 {
		t.Fatalf("unexpected serviceID %d", row.ServiceID)
	}
	if row.GuidePages != "" {
		t.Fatalf("expected guidePages ignored, got %q", row.GuidePages)
	}
	if row.PageNumEsc != 1 {
		t.Fatalf("expected pageNumEsc=1, got %d", row.PageNumEsc)
	}
}
