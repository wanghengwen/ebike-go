package part

import "testing"

func TestHelmetReturnResultFourCoreLocked(t *testing.T) {
	lock := 1
	if !HelmetReturnResult("unlock-token", false, &lock, nil, nil, nil, "tenant") {
		t.Fatal("helmet should pass when lock=1 and helmet was opened")
	}
}

func TestHelmetReturnResultNoOpenHelmet(t *testing.T) {
	if !HelmetReturnResult("", false, nil, nil, nil, nil, "tenant") {
		t.Fatal("helmet should pass when helmet was never opened")
	}
}

func TestRFIDReturnResult(t *testing.T) {
	event := 1
	if !RFIDReturnResult(&event) {
		t.Fatal("rfid event=1 should pass")
	}
	event = 0
	if RFIDReturnResult(&event) {
		t.Fatal("rfid event=0 should fail")
	}
}

func TestCameraReturnResultEventMode(t *testing.T) {
	event := 1
	if !CameraReturnResult(&event, nil, nil) {
		t.Fatal("camera event=1 should pass when angle unset")
	}
}

func TestCameraReturnResultAngleMode(t *testing.T) {
	event := 1
	angleDet := 95
	cameraAngle := 10
	if !CameraReturnResult(&event, &angleDet, &cameraAngle) {
		t.Fatal("camera angle within tolerance should pass")
	}
}
