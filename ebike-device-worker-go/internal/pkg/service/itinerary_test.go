package service_test

import (
	"testing"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/service"
	"ebike-device-worker-go/internal/pkg/web"
)

func TestValidateTrajectoryTimeRangeErrors(t *testing.T) {
	now := time.Now().UnixMilli()

	_, err := service.GetTrajectory(dto.TrajectoryCmd{
		Imei: "x", StartTime: now, EndTime: now - 1000, Type: new(int),
	})
	if biz, ok := err.(*web.BizError); !ok || biz.Code != dto.CodeTrajectoryQueryDateOutOfRange {
		t.Fatalf("expected start>end error, got %v", err)
	}

	old := now - int64(200*24*3600*1000) // ~200 days ago
	_, err = service.GetTrajectory(dto.TrajectoryCmd{
		Imei: "x", StartTime: old, EndTime: now, Type: new(int),
	})
	if biz, ok := err.(*web.BizError); !ok || biz.Code != dto.CodeTrajectoryQueryDateOutOfRange {
		t.Fatalf("expected 6-month error, got %v", err)
	}
}

func TestValidateMetricTimeRangeErrors(t *testing.T) {
	now := time.Now().UnixMilli()
	week := int64(8 * 24 * 3600 * 1000)

	_, err := service.GetMetric(dto.MetricCmd{
		Imei: "x", StartTime: now - week, EndTime: now, Type: new(int),
	})
	if biz, ok := err.(*web.BizError); !ok || biz.Code != dto.CodeTrajectoryQueryDateOutOfRange {
		t.Fatalf("expected week limit error, got %v", err)
	}
}
