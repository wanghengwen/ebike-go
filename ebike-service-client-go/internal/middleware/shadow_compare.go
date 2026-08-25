package middleware

import (
	"encoding/json"
	"log"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

// ShadowResponsesEqual compares Go and Java JSON bodies for shadow traffic.
// path is used for endpoint-specific normalization (paidInfo, carCount, etc.).
func ShadowResponsesEqual(path string, goBody, javaBody []byte) bool {
	var goMap, javaMap map[string]interface{}
	errGo := json.Unmarshal(goBody, &goMap)
	errJava := json.Unmarshal(javaBody, &javaMap)

	if errGo != nil || errJava != nil {
		return string(goBody) == string(javaBody)
	}

	normalizeShadowPair(path, goMap, javaMap)
	return jsondiff.ValuesEqual(goMap, javaMap)
}

// LogShadowDiff logs [SHADOW MATCH] or [SHADOW DIFF] (and legacy [SHADOW_SUCCESS]/[SHADOW_DIFF_ERROR]).
func LogShadowDiff(path string, reqBody string, goBody, javaBody []byte) {
	if ShadowResponsesEqual(path, goBody, javaBody) {
		if reqBody != "" {
			log.Printf("[SHADOW MATCH] Path: %s Req: %s", path, reqBody)
		} else {
			log.Printf("[SHADOW MATCH] Path: %s", path)
		}
		log.Printf("[SHADOW_SUCCESS] Path: %s", path)
		return
	}
	log.Printf("[SHADOW DIFF] Path: %s\nReq: %s\nJava: %s\nGo: %s\n", path, reqBody, string(javaBody), string(goBody))
	log.Printf("[SHADOW_DIFF_ERROR] Path: %s", path)
}

func normalizeShadowPair(path string, goMap, javaMap map[string]interface{}) {
	if path == "/client/order/detail" || path == "/client/order/detailLast" {
		dataGo, okGo := goMap["data"].(map[string]interface{})
		dataJava, okJava := javaMap["data"].(map[string]interface{})
		if okGo && okJava {
			pGo := dataGo["paidInfo"]
			pJava := dataJava["paidInfo"]
			if (pGo == nil && pJava == "") || (pGo == "" && pJava == nil) {
				dataGo["paidInfo"] = ""
				dataJava["paidInfo"] = ""
			}
		}
	}

	if path == "/client/fence/serviceArea/getNearFence" || path == "/client/fence/serviceArea/getFenceByServiceId" {
		stripFenceParkingCarCount(goMap)
		stripFenceParkingCarCount(javaMap)
	}
}

// stripFenceParkingCarCount removes volatile carCount from parking entries before SHADOW diff.
func stripFenceParkingCarCount(root map[string]interface{}) {
	data, ok := root["data"].(map[string]interface{})
	if !ok {
		return
	}
	parkings, ok := data["parkings"].([]interface{})
	if !ok {
		return
	}
	for _, item := range parkings {
		if row, ok := item.(map[string]interface{}); ok {
			delete(row, "carCount")
		}
	}
}
