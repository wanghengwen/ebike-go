package service

import (
	"strconv"
	"testing"

	"ebike-open-paas-go/internal/contract"
)

func TestNormalizeCacheTenant(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2", "2"},
		{" 2 ", "2"},
		{"\"2\"", "2"},
		{" \"1003\" ", "1003"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeCacheTenant(tc.in); got != tc.want {
			t.Errorf("normalizeCacheTenant(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDiagnoseDeviceBelongsMissingInput(t *testing.T) {
	d := DiagnoseDeviceBelongs("", "862149061045441")
	if d.Reason != BelongMissingTenant {
		t.Fatalf("reason = %s", d.Reason)
	}
	d = DiagnoseDeviceBelongs("2", "")
	if d.Reason != BelongMissingIMEI {
		t.Fatalf("reason = %s", d.Reason)
	}
}

func TestEnsureDeviceBelongsRejectsEmpty(t *testing.T) {
	if err := EnsureDeviceBelongs("87", "", "862149061045441"); err == nil || err.Error() != contract.ErrImeiIllegal {
		t.Errorf("empty tenant: %v", err)
	}
	if err := EnsureDeviceBelongs("87", "2", ""); err == nil || err.Error() != contract.ErrImeiIllegal {
		t.Errorf("empty imei: %v", err)
	}
}

func TestEnsureDeviceBelongsWithoutRegistry(t *testing.T) {
	err := EnsureDeviceBelongs("87", "2", "862149061045441")
	if err == nil || err.Error() != strconv.Itoa(contract.CodeNotBelongAgent) {
		t.Errorf("EnsureDeviceBelongs without Redis = %v, want %d", err, contract.CodeNotBelongAgent)
	}
	d := DiagnoseDeviceBelongs("2", "862149061045441")
	if d.Reason != BelongRedisUnavailable {
		t.Errorf("Diagnose without Redis reason = %s, want %s", d.Reason, BelongRedisUnavailable)
	}
}
