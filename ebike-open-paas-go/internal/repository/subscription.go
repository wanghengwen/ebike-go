package repository

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

// snapshot is an immutable view of every callback subscription, indexed the way
// the event path queries it: tenantId -> event -> urls.
//
// The Kafka consumer sees every device event for every tenant on saas_0, so the
// per-event lookup must not touch Redis. Redis stays the source of truth and
// this snapshot is rebuilt on a timer plus immediately after a CRUD write.
type snapshot struct {
	byTenant map[string]map[int][]string
}

var (
	subs atomic.Pointer[snapshot]
	// refreshMu serialises rebuilds. Registration refreshes eagerly on the
	// request goroutine while the timer refreshes on its own, and two
	// overlapping rebuilds can finish out of order — the older one then
	// publishes last and resurrects a subscription that was just removed.
	refreshMu sync.Mutex
)

// HasTenant reports whether any agent of this tenant has any subscription. The
// consumer calls it first so events for unsubscribed tenants — the vast
// majority of saas_0 traffic — are dropped before any payload is built.
func HasTenant(tenantID string) bool {
	s := subs.Load()
	if s == nil {
		return false
	}
	return len(s.byTenant[tenantID]) > 0
}

// SubscribersFor returns the callback URLs registered for a tenant + event.
// The returned slice must not be mutated: it is shared with every other caller
// holding the same snapshot.
func SubscribersFor(tenantID string, event int) []string {
	s := subs.Load()
	if s == nil {
		return nil
	}
	return s.byTenant[tenantID][event]
}

// RefreshSubscriptions rebuilds the snapshot from Redis. A Redis failure leaves
// the previous snapshot in place rather than blanking it, so a blip does not
// silently stop callbacks.
func RefreshSubscriptions() {
	refreshMu.Lock()
	defer refreshMu.Unlock()

	byTenant := map[string]map[int][]string{}
	for _, agent := range config.EnabledAgents() {
		for event := EventPing; event <= EventUART; event++ {
			urls, err := ListCallbacks(agent.AgentID, event)
			if err != nil {
				log.Printf("[subs] refresh failed agent=%s event=%d: %v (keeping previous snapshot)",
					agent.AgentID, event, err)
				return
			}
			if len(urls) == 0 {
				continue
			}
			byEvent := byTenant[agent.TenantID]
			if byEvent == nil {
				byEvent = map[int][]string{}
				byTenant[agent.TenantID] = byEvent
			}
			// Two agents on one tenant both get delivered, and a URL registered
			// under both must not be posted twice.
			byEvent[event] = appendUnique(byEvent[event], urls)
		}
	}
	subs.Store(&snapshot{byTenant: byTenant})
}

// StartSubscriptionRefresh seeds the snapshot and keeps it fresh. Registration
// through the API refreshes eagerly; this timer covers writes made by another
// replica of this service.
func StartSubscriptionRefresh() {
	RefreshSubscriptions()
	interval := time.Duration(config.GlobalConfig().Open.Callback.SubscriptionTTLSeconds) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}
	go func() {
		for range time.Tick(interval) {
			RefreshSubscriptions()
		}
	}()
	log.Printf("[subs] subscription cache refreshing every %s", interval)
}

func appendUnique(dst []string, src []string) []string {
	for _, u := range src {
		found := false
		for _, existing := range dst {
			if existing == u {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, u)
		}
	}
	return dst
}
