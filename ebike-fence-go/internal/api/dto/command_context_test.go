package dto

import "testing"

func TestEnsureCommandContextNilUsesTenant(t *testing.T) {
	got := EnsureCommandContext(nil, "t1")
	if got.TenantId != "t1" || got.TraceId != "" {
		t.Fatalf("unexpected context: %+v", got)
	}
}

func TestEnsureCommandContextPreservesTraceId(t *testing.T) {
	in := &CommandContext{TenantId: "t1", TraceId: "trace-1", DeviceId: "dev-1"}
	got := EnsureCommandContext(in, "t1")
	if got != in {
		t.Fatalf("expected same pointer when tenant already set")
	}
	if got.TraceId != "trace-1" || got.DeviceId != "dev-1" {
		t.Fatalf("trace fields lost: %+v", got)
	}
}

func TestEnsureCommandContextFillsTenant(t *testing.T) {
	in := &CommandContext{TraceId: "trace-1"}
	got := EnsureCommandContext(in, "t1")
	if got == in {
		t.Fatalf("expected copied context when tenant missing")
	}
	if got.TenantId != "t1" || got.TraceId != "trace-1" {
		t.Fatalf("unexpected context: %+v", got)
	}
}
