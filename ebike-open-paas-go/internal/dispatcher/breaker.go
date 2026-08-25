package dispatcher

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

// Circuit states for one callback URL.
//
// closed  — deliver normally
// open    — skip delivery for a cooldown (subscription stays in Redis)
// half-open — allow a single probe after the cooldown; success closes, failure reopens
const (
	circuitClosed = iota
	circuitOpen
	circuitHalfOpen
)

type urlCircuit struct {
	mu               sync.Mutex
	state            int
	consecutiveFails int
	openedAt         time.Time
	probeInFlight    bool
}

var (
	circuits   sync.Map // string → *urlCircuit
	suppressed uint64   // skipped because the circuit was open
	tripped    uint64   // times a circuit transitioned to open
)

func circuitCfg() (enabled bool, failThreshold int, cooldown time.Duration) {
	cfg := config.GlobalConfig().Open.Callback
	enabled = cfg.CircuitEnabled
	failThreshold = cfg.CircuitFailThreshold
	if failThreshold <= 0 {
		failThreshold = 20
	}
	secs := cfg.CircuitCooldownSeconds
	if secs <= 0 {
		secs = 600
	}
	return enabled, failThreshold, time.Duration(secs) * time.Second
}

func circuitFor(url string) *urlCircuit {
	if v, ok := circuits.Load(url); ok {
		return v.(*urlCircuit)
	}
	c := &urlCircuit{}
	actual, _ := circuits.LoadOrStore(url, c)
	return actual.(*urlCircuit)
}

// circuitAllow reports whether a delivery to url may proceed.
//
// An open circuit suppresses the event without removing the Redis subscription:
// the customer still appears registered; we simply stop hammering a dead endpoint
// until the cooldown elapses or they re-register the same URL.
func circuitAllow(url string) bool {
	enabled, _, cooldown := circuitCfg()
	if !enabled {
		return true
	}
	c := circuitFor(url)
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case circuitClosed:
		return true
	case circuitOpen:
		if time.Since(c.openedAt) < cooldown {
			atomic.AddUint64(&suppressed, 1)
			return false
		}
		c.state = circuitHalfOpen
		c.probeInFlight = true
		log.Printf("[dispatcher] circuit half-open url=%s (probe)", url)
		return true
	case circuitHalfOpen:
		if c.probeInFlight {
			atomic.AddUint64(&suppressed, 1)
			return false
		}
		c.probeInFlight = true
		return true
	default:
		return true
	}
}

func circuitSuccess(url string) {
	enabled, _, _ := circuitCfg()
	if !enabled {
		return
	}
	c := circuitFor(url)
	c.mu.Lock()
	defer c.mu.Unlock()
	wasOpen := c.state != circuitClosed
	c.state = circuitClosed
	c.consecutiveFails = 0
	c.probeInFlight = false
	if wasOpen {
		log.Printf("[dispatcher] circuit closed url=%s", url)
	}
}

func circuitFailure(url string) {
	enabled, threshold, _ := circuitCfg()
	if !enabled {
		return
	}
	c := circuitFor(url)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.probeInFlight = false
	c.consecutiveFails++
	if c.state == circuitHalfOpen || c.consecutiveFails >= threshold {
		if c.state != circuitOpen {
			atomic.AddUint64(&tripped, 1)
			log.Printf("[dispatcher] circuit open url=%s consecutiveFails=%d", url, c.consecutiveFails)
		}
		c.state = circuitOpen
		c.openedAt = time.Now()
	}
}

// ResetCircuit clears the temporary mute for url.
//
// Re-registration is the customer's signal that the endpoint is ready again; we
// must not delete the Redis subscription on failure, but we do honour an
// explicit register as a manual reset.
func ResetCircuit(url string) {
	if url == "" {
		return
	}
	if v, ok := circuits.Load(url); ok {
		c := v.(*urlCircuit)
		c.mu.Lock()
		c.state = circuitClosed
		c.consecutiveFails = 0
		c.probeInFlight = false
		c.mu.Unlock()
		log.Printf("[dispatcher] circuit reset url=%s (re-register)", url)
	}
}

// CircuitStats returns events skipped while a URL was muted, and how many times
// a circuit tripped open.
func CircuitStats() (suppressedN, trippedN uint64) {
	return atomic.LoadUint64(&suppressed), atomic.LoadUint64(&tripped)
}

// circuitStateForTest exposes state for unit tests.
func circuitStateForTest(url string) (state, fails int, probe bool) {
	c := circuitFor(url)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state, c.consecutiveFails, c.probeInFlight
}

func resetCircuitsForTest() {
	circuits = sync.Map{}
	atomic.StoreUint64(&suppressed, 0)
	atomic.StoreUint64(&tripped, 0)
}
