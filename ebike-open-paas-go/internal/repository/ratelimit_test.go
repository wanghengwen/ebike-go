package repository

import (
	"sync"
	"testing"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

func newLocalWindows() *localWindows {
	return &localWindows{agents: map[string]map[string]int{}}
}

// TestLocalLimiterEnforcesQuota covers the degraded path: with Redis unreachable
// the public API still has to be metered, because failing open would leave it
// unmetered for the whole outage.
func TestLocalLimiterEnforcesQuota(t *testing.T) {
	l := newLocalWindows()
	now := time.Now()
	current := now.Format(rateTimeLayout)
	oldest := now.Add(-4 * time.Minute).Format(rateTimeLayout)

	for i := 1; i <= 3; i++ {
		if !l.allow("a1", current, oldest, 3, 5) {
			t.Fatalf("request %d of 3 was refused inside the quota", i)
		}
	}
	if l.allow("a1", current, oldest, 3, 5) {
		t.Error("the 4th request was allowed past a quota of 3")
	}
}

// TestLocalLimiterIsPerAgent keeps one noisy customer from spending another's
// budget.
func TestLocalLimiterIsPerAgent(t *testing.T) {
	l := newLocalWindows()
	now := time.Now()
	current := now.Format(rateTimeLayout)
	oldest := now.Add(-4 * time.Minute).Format(rateTimeLayout)

	for i := 0; i < 2; i++ {
		l.allow("a1", current, oldest, 2, 5)
	}
	if l.allow("a1", current, oldest, 2, 5) {
		t.Fatal("a1 exceeded its own quota")
	}
	if !l.allow("a2", current, oldest, 2, 5) {
		t.Error("a2 was refused because a1 had spent its quota")
	}
}

// TestLocalLimiterExpiresOldBuckets: the window has to slide, or an agent that
// hit the limit once would stay blocked and the map would grow without bound.
func TestLocalLimiterExpiresOldBuckets(t *testing.T) {
	l := newLocalWindows()
	base := time.Now()

	stale := base.Add(-10 * time.Minute).Format(rateTimeLayout)
	staleOldest := base.Add(-14 * time.Minute).Format(rateTimeLayout)
	for i := 0; i < 3; i++ {
		l.allow("a1", stale, staleOldest, 3, 5)
	}

	current := base.Format(rateTimeLayout)
	oldest := base.Add(-4 * time.Minute).Format(rateTimeLayout)
	if !l.allow("a1", current, oldest, 3, 5) {
		t.Fatal("a request was refused on counts that fell out of the window")
	}
	if got := len(l.agents["a1"]); got != 1 {
		t.Errorf("the agent holds %d buckets, want the stale ones pruned to 1", got)
	}
}

// TestLocalLimiterIsConcurrencySafe: every request goes through this on the
// degraded path, from all of Gin's handler goroutines at once.
func TestLocalLimiterIsConcurrencySafe(t *testing.T) {
	l := newLocalWindows()
	now := time.Now()
	current := now.Format(rateTimeLayout)
	oldest := now.Add(-4 * time.Minute).Format(rateTimeLayout)

	const limit = 50
	var mu sync.Mutex
	granted := 0
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.allow("a1", current, oldest, limit, 5) {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if granted != limit {
		t.Errorf("%d of 200 concurrent requests were granted, want exactly the quota of %d", granted, limit)
	}
}

// TestAgentAllowedSkipsUnlimitedConfigs: an unset quota must not fall through to
// a Redis call on every request.
func TestAgentAllowedSkipsUnlimitedConfigs(t *testing.T) {
	if ok, _ := AgentAllowed("", config.RateLimitConfig{MaxRequests: 10}); !ok {
		t.Error("an empty agentId was rate limited")
	}
	if ok, _ := AgentAllowed("a1", config.RateLimitConfig{MaxRequests: 0}); !ok {
		t.Error("maxRequests=0 was treated as a quota of zero rather than unlimited")
	}
}

// TestAgentAllowedFallsBackWhenRedisIsDown exercises the real entry point with no
// Redis configured, which is what the fallback exists for.
func TestAgentAllowedFallsBackWhenRedisIsDown(t *testing.T) {
	prev := localLimiter
	localLimiter = newLocalWindows()
	t.Cleanup(func() { localLimiter = prev })

	rl := config.RateLimitConfig{MaxRequests: 2, WindowSeconds: 300}
	for i := 1; i <= 2; i++ {
		if ok, msg := AgentAllowed("agent-x", rl); !ok {
			t.Fatalf("request %d refused with %q, want it inside the quota", i, msg)
		}
	}
	ok, msg := AgentAllowed("agent-x", rl)
	if ok {
		t.Error("the 3rd request was allowed past a quota of 2 with Redis unavailable")
	}
	if msg == "" {
		t.Error("a refusal came back with no reason")
	}
}
