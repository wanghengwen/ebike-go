package apilog_test

import (
	"strings"
	"testing"

	"ebike-device-worker-go/internal/pkg/apilog"
)

func TestParamsTruncatesLongBody(t *testing.T) {
	long := strings.Repeat("a", 3000)
	apilog.Params("/ebike/gps/getBatchOrderTrajectory", []byte(`{"x":"`+long+`"}`))
}

func TestObjectMarshalsStruct(t *testing.T) {
	apilog.Object("/ebike/gps/getTrajectory", map[string]interface{}{
		"imei":      "860123456789012",
		"startTime": int64(1700000000000),
		"endTime":   int64(1700003600000),
	})
}
