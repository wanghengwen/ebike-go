package configsvc

import (
	"testing"

	"ebike-fence-go/internal/api/dto"
)

func TestApplyUseCarQueryDefaults_RedisPath(t *testing.T) {
	co := &dto.ConfigUseCarCO{}
	applyUseCarQueryDefaults(co, true)

	if co.HideCarConfig == nil || *co.HideCarConfig != defaultHideCarConfigJSON {
		t.Fatalf("hideCarConfig = %v, want %q", co.HideCarConfig, defaultHideCarConfigJSON)
	}
	if co.OutServiceAreaAutoLock == nil || *co.OutServiceAreaAutoLock != 10 {
		t.Fatalf("outServiceAreaAutoLock = %v, want 10", co.OutServiceAreaAutoLock)
	}
	if co.IzAuth == nil || !*co.IzAuth {
		t.Fatalf("izAuth = %v, want true", co.IzAuth)
	}
}

func TestApplyUseCarQueryDefaults_DBPath(t *testing.T) {
	co := &dto.ConfigUseCarCO{}
	applyUseCarQueryDefaults(co, false)

	if co.HideCarConfig == nil || *co.HideCarConfig != defaultHideCarConfigJSON {
		t.Fatalf("hideCarConfig = %v, want %q", co.HideCarConfig, defaultHideCarConfigJSON)
	}
	if co.OutServiceAreaAutoLock != nil {
		t.Fatalf("outServiceAreaAutoLock = %v, want nil on DB path", co.OutServiceAreaAutoLock)
	}
	if co.IzAuth != nil {
		t.Fatalf("izAuth = %v, want nil on DB path", co.IzAuth)
	}
}
