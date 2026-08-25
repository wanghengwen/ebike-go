package timefmt

import (
	"encoding/json"
	"testing"
)

func TestJavaLocalTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`"2024-07-15T16:19:08"`, "2024-07-15T16:19:08"},
		{`"2024-07-19 10:04:52"`, "2024-07-19 10:04:52"},
		{`"2025-12-30T17:19:04"`, "2025-12-30T17:19:04"},
	}
	for _, tc := range tests {
		var got JavaLocalTime
		if err := json.Unmarshal([]byte(tc.raw), &got); err != nil {
			t.Fatalf("unmarshal %s: %v", tc.raw, err)
		}
		if got.IsZero() {
			t.Fatalf("expected parsed time for %s", tc.raw)
		}
	}
}

func TestJavaLocalTime_UnmarshalConfigBaseDO(t *testing.T) {
	raw := `[{"updatedAt":"2024-07-15T16:19:08","createdAt":"2024-07-15T16:19:08","tenantId":"1004"}]`
	type row struct {
		UpdatedAt JavaLocalTime `json:"updatedAt"`
		CreatedAt JavaLocalTime `json:"createdAt"`
		TenantID  string        `json:"tenantId"`
	}
	var rows []row
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		t.Fatal(err)
	}
	if rows[0].UpdatedAt.IsZero() {
		t.Fatal("expected updatedAt parsed")
	}
}
