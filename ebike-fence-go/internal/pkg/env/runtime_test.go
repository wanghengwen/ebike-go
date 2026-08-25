package env

import "testing"

func TestDryRun(t *testing.T) {
	t.Setenv("DRY_RUN", "true")
	if !DryRun() {
		t.Fatal("expected DryRun true")
	}
	t.Setenv("DRY_RUN", "false")
	if DryRun() {
		t.Fatal("expected DryRun false")
	}
}

func TestNacosRegisterEnabled(t *testing.T) {
	t.Setenv("NACOS_REGISTER_ENABLED", "")
	if !NacosRegisterEnabled() {
		t.Fatal("expected default register enabled")
	}
	t.Setenv("NACOS_REGISTER_ENABLED", "false")
	if NacosRegisterEnabled() {
		t.Fatal("expected register disabled")
	}
	t.Setenv("NACOS_REGISTER_ENABLED", "true")
	if !NacosRegisterEnabled() {
		t.Fatal("expected register enabled")
	}
}
