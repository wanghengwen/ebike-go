package middleware

import "testing"

func TestWriteEndpointsNotMirrorable(t *testing.T) {
	writePaths := []string{
		"/client/ebike-marketing/user_reward/user_watch_ad_judgement",
		"/client/code/verify",
		"/client/pay/score/permission",
		"/client/ebike-pay/pay/getopenidbyjscode",
		"/client/messagecenter/getbyid",
	}
	for _, path := range writePaths {
		if isMirrorableRequest("POST", path) {
			t.Errorf("POST %s must not be mirrorable (has side effects)", path)
		}
	}
}

func TestReadOnlyEndpointsStillMirrorable(t *testing.T) {
	readPaths := []string{
		"/client/order/list",
		"/client/pay/score/getpermissionrecord",
		"/client/pay/score/getconfig",
		"/client/ebike-pay/pay/withdraw/page",
	}
	for _, path := range readPaths {
		if !isMirrorableRequest("POST", path) {
			t.Errorf("POST %s should remain mirrorable", path)
		}
	}
}

func TestUnregisteredGoBFTEndpointsNotMirrorable(t *testing.T) {
	unregistered := []string{
		"/client/fence/servicearea/returncar",
		"/client/fence/servicearea/ridingcar",
	}
	for _, path := range unregistered {
		if isMirrorableRequest("POST", path) {
			t.Errorf("POST %s must not be mirrorable (Go BFF route not implemented)", path)
		}
	}
}

func TestDestructiveMethodsNeverMirrorable(t *testing.T) {
	for _, method := range []string{"PUT", "PATCH", "DELETE"} {
		if isMirrorableRequest(method, "/client/order/list") {
			t.Errorf("%s must never be mirrorable", method)
		}
	}
}
