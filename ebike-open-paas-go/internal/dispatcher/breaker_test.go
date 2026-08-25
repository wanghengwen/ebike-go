package dispatcher

import (
	"fmt"
	"testing"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

func withCircuit(t *testing.T, failThreshold, cooldownSec int) {
	t.Helper()
	resetCircuitsForTest()
	config.MergeNacosAppConfig(fmt.Sprintf(`
open:
  callback:
    circuitEnabled: true
    circuitFailThreshold: %d
    circuitCooldownSeconds: %d
`, failThreshold, cooldownSec))
	t.Cleanup(func() {
		resetCircuitsForTest()
		config.MergeNacosAppConfig(`
open:
  callback:
    circuitEnabled: true
    circuitFailThreshold: 20
    circuitCooldownSeconds: 600
`)
	})
}

// TestCircuitOpensAfterConsecutiveFailures: a dead endpoint must stop being
// hammered without deleting the Redis subscription.
func TestCircuitOpensAfterConsecutiveFailures(t *testing.T) {
	withCircuit(t, 3, 60)
	const url = "https://slow.example.com/hook"

	for i := 0; i < 3; i++ {
		if !circuitAllow(url) {
			t.Fatalf("allow failed on attempt %d, want deliveries to proceed while closed", i+1)
		}
		circuitFailure(url)
	}
	state, fails, _ := circuitStateForTest(url)
	if state != circuitOpen {
		t.Fatalf("state=%d, want open after 3 failures", state)
	}
	if fails != 3 {
		t.Fatalf("fails=%d, want 3", fails)
	}
	if circuitAllow(url) {
		t.Fatal("open circuit allowed a delivery during cooldown")
	}
	suppressed, tripped := CircuitStats()
	if suppressed == 0 {
		t.Error("suppressed counter did not advance")
	}
	if tripped == 0 {
		t.Error("tripped counter did not advance")
	}
}

// TestCircuitHalfOpenProbeAndRecovery: after the cooldown a single probe is
// allowed; success must close the circuit for subsequent events.
func TestCircuitHalfOpenProbeAndRecovery(t *testing.T) {
	withCircuit(t, 2, 1)
	const url = "https://recover.example.com/hook"

	circuitFailure(url)
	circuitFailure(url)
	if circuitAllow(url) {
		t.Fatal("expected open during cooldown")
	}

	time.Sleep(1100 * time.Millisecond)
	if !circuitAllow(url) {
		t.Fatal("expected half-open probe after cooldown")
	}
	// Second event while probe in flight must stay suppressed.
	if circuitAllow(url) {
		t.Fatal("second event was allowed while probe in flight")
	}

	circuitSuccess(url)
	state, fails, probe := circuitStateForTest(url)
	if state != circuitClosed || fails != 0 || probe {
		t.Fatalf("after success state=%d fails=%d probe=%v, want closed", state, fails, probe)
	}
	if !circuitAllow(url) {
		t.Fatal("closed circuit refused delivery")
	}
}

// TestCircuitHalfOpenFailureReopens: a failed probe must not leave the circuit
// closed and resume hammering.
func TestCircuitHalfOpenFailureReopens(t *testing.T) {
	withCircuit(t, 2, 1)
	const url = "https://still-dead.example.com/hook"

	circuitFailure(url)
	circuitFailure(url)
	time.Sleep(1100 * time.Millisecond)
	if !circuitAllow(url) {
		t.Fatal("expected probe")
	}
	circuitFailure(url)
	state, _, _ := circuitStateForTest(url)
	if state != circuitOpen {
		t.Fatalf("state=%d after failed probe, want open", state)
	}
	if circuitAllow(url) {
		t.Fatal("re-opened circuit allowed delivery immediately")
	}
}

// TestResetCircuitOnReregister: POST /callback for the same URL is the
// customer's signal that the endpoint is ready again.
func TestResetCircuitOnReregister(t *testing.T) {
	withCircuit(t, 2, 600)
	const url = "https://fixed.example.com/hook"

	circuitFailure(url)
	circuitFailure(url)
	if circuitAllow(url) {
		t.Fatal("expected open")
	}
	ResetCircuit(url)
	state, fails, _ := circuitStateForTest(url)
	if state != circuitClosed || fails != 0 {
		t.Fatalf("after reset state=%d fails=%d, want closed", state, fails)
	}
	if !circuitAllow(url) {
		t.Fatal("reset circuit still refused delivery")
	}
}

func TestCircuitSuccessClearsConsecutiveFails(t *testing.T) {
	withCircuit(t, 3, 60)
	const url = "https://flaky.example.com/hook"

	circuitFailure(url)
	circuitFailure(url)
	circuitSuccess(url)
	_, fails, _ := circuitStateForTest(url)
	if fails != 0 {
		t.Fatalf("fails=%d after success, want 0", fails)
	}
	// Two more failures must not open (threshold 3, counter was reset).
	circuitFailure(url)
	circuitFailure(url)
	state, _, _ := circuitStateForTest(url)
	if state != circuitClosed {
		t.Fatalf("state=%d, want still closed after only 2 fails post-reset", state)
	}
}
