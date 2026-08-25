package contract

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func recordJSON(t *testing.T, fn func(c *gin.Context)) (int, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)
	fn(c)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response %q is not JSON: %v", rec.Body.String(), err)
	}
	return rec.Code, body
}

// TestFailUsesXiaoanEnvelope pins the field names and the HTTP status. Xiaoan's
// clients read `error.errormessage` off a 200, so answering with a non-200 or a
// differently-spelled key reads as a transport failure rather than the business
// error it is.
func TestFailUsesXiaoanEnvelope(t *testing.T) {
	status, body := recordJSON(t, func(c *gin.Context) {
		Fail(c, ErrImeiIllegal, "")
	})

	if status != 200 {
		t.Errorf("status = %d, want 200 — business errors ride on a 200", status)
	}
	if body["success"] != false {
		t.Errorf("success = %#v, want false", body["success"])
	}
	if _, present := body["data"]; present {
		t.Errorf("data present on a failure: %#v", body["data"])
	}
	err, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("error = %#v, want an object", body["error"])
	}
	if err["errormessage"] != ErrImeiIllegal {
		t.Errorf("errormessage = %#v, want %q", err["errormessage"], ErrImeiIllegal)
	}
	if err["promot"] != promotByErrorType[ErrImeiIllegal] {
		t.Errorf("promot = %#v, want the default %q", err["promot"], promotByErrorType[ErrImeiIllegal])
	}
}

func TestFailKeepsExplicitPromot(t *testing.T) {
	_, body := recordJSON(t, func(c *gin.Context) {
		Fail(c, ErrBizError, "rate limit exceeded")
	})
	err := body["error"].(map[string]interface{})
	if err["promot"] != "rate limit exceeded" {
		t.Errorf("promot = %#v, want the caller's message", err["promot"])
	}
}

// TestEveryErrorTypeHasAPromot: an error type with no entry produces an empty
// promot, which shows up in a customer's logs as a blank reason.
func TestEveryErrorTypeHasAPromot(t *testing.T) {
	for _, errorType := range []string{
		ErrInputParamsMiss, ErrSendCmdError, ErrDeviceInfoIsNull,
		ErrUnauthorized, ErrImeiIllegal, ErrBizError, ErrRangeTooLarge,
	} {
		if promotByErrorType[errorType] == "" {
			t.Errorf("%s has no default promot", errorType)
		}
	}
}

func TestOKOmitsErrorObject(t *testing.T) {
	_, body := recordJSON(t, func(c *gin.Context) {
		OK(c, map[string]interface{}{"imei": "865067022403441"})
	})
	if body["success"] != true {
		t.Errorf("success = %#v, want true", body["success"])
	}
	if _, present := body["error"]; present {
		t.Errorf("error present on a success: %#v", body["error"])
	}
}

// TestOKDeviceCodeKeepsZero: command endpoints answer with data.code, and 0 is
// the success value — dropping it as empty would leave the caller with no code
// at all.
func TestOKDeviceCodeKeepsZero(t *testing.T) {
	_, body := recordJSON(t, func(c *gin.Context) {
		OKDeviceCode(c, CodeOK)
	})
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data = %#v, want an object", body["data"])
	}
	if data["code"] != float64(0) {
		t.Errorf("data.code = %#v, want 0", data["code"])
	}
}

// TestMapEcuCodeDefaultsToServerInternal: an unrecognised upstream code must not
// be forwarded verbatim, since the two numbering schemes overlap with different
// meanings and a plausible-looking wrong code is worse than a generic one.
func TestMapEcuCodeDefaultsToServerInternal(t *testing.T) {
	for _, raw := range []string{"99999", "abc", "-1", "10099"} {
		if got := MapEcuCode(raw); got != CodeServerInternal {
			t.Errorf("MapEcuCode(%q) = %d, want %d", raw, got, CodeServerInternal)
		}
	}
}

// TestMapEcuCodeTreatsBlankAsSuccess covers the upstreams that omit ecuCode
// entirely on success.
func TestMapEcuCodeTreatsBlankAsSuccess(t *testing.T) {
	for _, raw := range []string{"", "0", "00"} {
		if got := MapEcuCode(raw); got != CodeOK {
			t.Errorf("MapEcuCode(%q) = %d, want %d", raw, got, CodeOK)
		}
	}
}

// TestMapEcuCodeMapsOfflineAndOwnership are the two a third party branches on:
// 109 tells them to retry later, 1001 that the device is not theirs.
func TestMapEcuCodeMapsOfflineAndOwnership(t *testing.T) {
	if got := MapEcuCode("10012"); got != CodeDeviceOffline {
		t.Errorf("MapEcuCode(\"10012\") = %d, want %d", got, CodeDeviceOffline)
	}
	if got := MapEcuCode("10006"); got != CodeNotBelongAgent {
		t.Errorf("MapEcuCode(\"10006\") = %d, want %d", got, CodeNotBelongAgent)
	}
}
