package repository

import (
	"log"
	"sync"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/pkg/redis"
	"ebike-open-paas-go/internal/pkg/rediskey"
)

const rateTimeLayout = "200601021504" // yyyyMMddHHmm

// AgentAllowed applies a sliding-window quota to the agent.
//
// The window is evaluated and incremented in a single Redis script, so replicas
// and concurrent requests cannot all observe the same pre-increment count and
// burst past the quota together.
//
// When Redis is unreachable the quota falls back to an in-process counter with
// the same limits. That is weaker than the shared one — each replica gets its
// own budget — but the alternatives are both worse: failing open leaves the
// public API unmetered for the length of the outage, and failing closed takes it
// offline over a dependency that no request otherwise needs.
func AgentAllowed(agentID string, rl config.RateLimitConfig) (bool, string) {
	if agentID == "" || rl.MaxRequests <= 0 {
		return true, ""
	}
	now := time.Now()
	windowSecs := rl.WindowSeconds
	if windowSecs <= 0 {
		windowSecs = 300
	}
	windowMinutes := (windowSecs + 59) / 60
	if windowMinutes < 1 {
		windowMinutes = 1
	}
	current := now.Format(rateTimeLayout)
	oldest := now.Add(-time.Duration(windowMinutes-1) * time.Minute).Format(rateTimeLayout)
	ttl := time.Duration(windowMinutes*2) * time.Minute

	allowed, _, err := redis.SlidingWindowAllow(
		rediskey.RateLimit(agentID), current, oldest, rl.MaxRequests, ttl)
	if err != nil {
		allowed = localLimiter.allow(agentID, current, oldest, rl.MaxRequests, windowMinutes)
		if !allowed {
			return false, "rate limit exceeded"
		}
		return true, ""
	}
	if !allowed {
		return false, "rate limit exceeded"
	}
	return true, ""
}

// localLimiter is the degraded-mode counter used while Redis is unreachable.
var localLimiter = &localWindows{agents: map[string]map[string]int{}}

type localWindows struct {
	mu     sync.Mutex
	agents map[string]map[string]int
	warned time.Time
}

func (l *localWindows) allow(agentID, current, oldest string, limit, windowMinutes int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if time.Since(l.warned) > time.Minute {
		log.Printf("[ratelimit] redis unavailable, enforcing per-replica quota locally (limit=%d/%dm)",
			limit, windowMinutes)
		l.warned = time.Now()
	}

	buckets := l.agents[agentID]
	if buckets == nil {
		buckets = map[string]int{}
		l.agents[agentID] = buckets
	}
	total := 0
	for bucket, count := range buckets {
		if bucket < oldest {
			delete(buckets, bucket)
			continue
		}
		if bucket <= current {
			total += count
		}
	}
	if total >= limit {
		return false
	}
	buckets[current]++
	return true
}
