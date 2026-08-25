package configsvc

import (
	"context"
	"errors"
	"testing"

	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/pkg/config"
)

func TestRidingCarConfigGetNullContextWhenCancelAuthConfigured(t *testing.T) {
	config.GlobalConfig.Xyy.CancelAuthTenantIds = []string{"1007"}
	t.Cleanup(func() { config.GlobalConfig.Xyy.CancelAuthTenantIds = nil })

	svc := NewRidingCarConfigService()
	_, err := svc.Get(context.Background(), "1004", 123)
	if !errors.Is(err, service.ErrRidingCarConfigNullContext) {
		t.Fatalf("expected ErrRidingCarConfigNullContext, got %v", err)
	}
}
