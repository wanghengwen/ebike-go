package middleware

import (
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

func TestStripFenceParkingCarCount(t *testing.T) {
	goMap := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"parkings": []interface{}{
				map[string]interface{}{"id": "1", "carCount": 7},
				map[string]interface{}{"id": "2", "carCount": 35},
			},
		},
	}
	javaMap := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"parkings": []interface{}{
				map[string]interface{}{"id": "1", "carCount": 8},
				map[string]interface{}{"id": "2", "carCount": 36},
			},
		},
	}
	stripFenceParkingCarCount(goMap)
	stripFenceParkingCarCount(javaMap)
	if !jsondiff.ValuesEqual(goMap, javaMap) {
		t.Fatalf("expected equal after stripping carCount:\nGo=%v\nJava=%v", goMap, javaMap)
	}
}
