package redis

import (
	"testing"
)

func TestGetLoginDeviceKey(t *testing.T) {
	prefix := "tenantId123"
	pin := "userPIN"
	expected := "login_device_tenantId123_userPIN"
	got := GetLoginDeviceKey(prefix, pin)
	if got != expected {
		t.Errorf("GetLoginDeviceKey(%q, %q) = %q; want %q", prefix, pin, got, expected)
	}
}

func TestGetLoginDeviceOldKey(t *testing.T) {
	pin := "userPIN"
	expected := "login_device_userPIN"
	got := GetLoginDeviceOldKey(pin)
	if got != expected {
		t.Errorf("GetLoginDeviceOldKey(%q) = %q; want %q", pin, got, expected)
	}
}

func TestGetUserLoginTagKey(t *testing.T) {
	tenantId := "tenantId123"
	pin := "userPIN"
	expected := "user_login_tag_tenantId123_userPIN"
	got := GetUserLoginTagKey(tenantId, pin)
	if got != expected {
		t.Errorf("GetUserLoginTagKey(%q, %q) = %q; want %q", tenantId, pin, got, expected)
	}
}
