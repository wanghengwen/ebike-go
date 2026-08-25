package part

import "math"

// HelmetReturnResult mirrors Java HelmetEntity.returnAnalysisResult.
func HelmetReturnResult(openHelmet string, helmetAudit bool, helmetLock, helmet6React, helmet6Lock *int, helmetAuditWhite []string, tenantId string) bool {
	if openHelmet == "" || helmetAudit {
		return true
	}
	if helmet6React == nil && helmet6Lock == nil {
		return helmetLock != nil && *helmetLock == 1
	}
	for _, t := range helmetAuditWhite {
		if t == tenantId {
			return (helmet6React != nil && *helmet6React == 1) || (helmet6Lock != nil && *helmet6Lock == 1)
		}
	}
	return (helmet6React != nil && *helmet6React == 1) && (helmet6Lock != nil && *helmet6Lock == 1)
}

// RFIDReturnResult mirrors Java RFIDBeaconEntity.returnAnalysisResult (event == 1).
func RFIDReturnResult(event *int) bool {
	return event != nil && *event == 1
}

// CameraReturnResult mirrors Java CameraEntity.returnAnalysisResult.
func CameraReturnResult(event, angleDet, cameraAngle *int) bool {
	angle := -1
	if cameraAngle != nil {
		angle = *cameraAngle
	}
	det := -1
	if angleDet != nil {
		det = *angleDet
	}
	evt := 0
	if event != nil {
		evt = *event
	}
	if angle == -1 {
		return evt == 1
	}
	return math.Abs(float64(det-90)) <= float64(angle)
}
