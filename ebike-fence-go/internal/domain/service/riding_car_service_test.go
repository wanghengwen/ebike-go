package service

import (
	"context"
	"testing"

	"ebike-fence-go/internal/api/dto"
	appconfig "ebike-fence-go/internal/pkg/config"
)

func TestRidingCarRejectsRemoteUnlockDistance(t *testing.T) {
	dist := 100
	appconfig.GlobalConfig.Xyy.RemoteLockDistance = &dist

	svc := NewRidingCarService(nil, nil, nil)
	izRemote := false
	req := dto.RideCarCmd{
		Command: dto.Command{
			CommandContext: &dto.CommandContext{TenantId: "t1", TraceId: "trace-1"},
		},
		ServiceAreaId:  int64Ptr(1),
		CarLocation:    &dto.LocationCmd{Lng: 116.40, Lat: 39.90},
		UserLocation:   &dto.LocationCmd{Lng: 116.50, Lat: 39.90},
		CarCmd:         &dto.CarCmd{CarId: "c1"},
		IzRemoteUnlock: &izRemote,
	}

	_, err := svc.RidingCar(context.Background(), "t1", req)
	if err == nil {
		t.Fatal("expected distance biz error")
	}
	biz, ok := err.(*BizError)
	if !ok || biz.Code != codeCanNotRemoteUnlock {
		t.Fatalf("unexpected error: %v", err)
	}
}

func int64Ptr(v int64) *int64 { return &v }
