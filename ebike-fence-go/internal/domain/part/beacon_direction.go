package part

// BeaconReturnResult mirrors Java BluetoothBeaconEntity.returnAnalysisResult.
func BeaconReturnResult(event *int, tBeaconAddr string) bool {
	return event != nil && *event == 3 && tBeaconAddr != ""
}

// DirectionReturnResult mirrors Java DirectionEntity.returnAnalysisResult.
func DirectionReturnResult(headingAngle, fenceDirection, formulateDirection *float64, directionIgnoreTenant bool) bool {
	if headingAngle == nil {
		if directionIgnoreTenant {
			return true
		}
		return false
	}
	if *headingAngle <= 0 {
		if directionIgnoreTenant {
			return true
		}
		return false
	}
	if fenceDirection == nil || formulateDirection == nil {
		return false
	}
	return directionAngleCalculation(*fenceDirection, *headingAngle, *formulateDirection)
}

// directionAngleCalculation mirrors Java DirectionEntity.angleCalculation.
func directionAngleCalculation(angle1, angle2, formulateDirection float64) bool {
	up := angle1 + formulateDirection
	down := angle1 - formulateDirection
	if angle1 < 360-formulateDirection && angle1 > formulateDirection {
		return angle2 >= down && angle2 <= up
	}
	if angle1 >= 360-formulateDirection {
		return (angle2 >= 0 && angle2 <= up-360) || (angle2 >= down)
	}
	if angle1 <= formulateDirection {
		return (angle2 >= 0 && angle2 <= up) || angle2 >= down+360
	}
	return false
}

// NormalizeHeadingAngle converts device heading (0.1° units) to degrees when needed.
func NormalizeHeadingAngle(v float64) float64 {
	if v > 360 {
		return v / 10
	}
	return v
}
