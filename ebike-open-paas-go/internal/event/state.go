package event

import (
	"sort"
	"strconv"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/pkg/redis"
	"ebike-open-paas-go/internal/pkg/rediskey"
)

// Hash fields in the per-device notify state.
const (
	fieldOutOfFence = "outOfFence"
	fieldSocStep    = "socStep"
)

// socStepAboveAll is the stored marker for "SOC is above every configured
// threshold". It is a percentage rather than a notify code so that comparing two
// states is unambiguous — a deeper discharge is always the smaller number,
// whereas notify codes happen to run the other way (21 = 50%, 22 = 30%).
const socStepAboveAll = 101

// edge is the outcome of one state comparison: whether an event fires, and what
// to persist. Persisting without firing is the normal case for a device's first
// report and for a battery charging back up past a threshold.
type edge struct {
	notify    int
	hasNotify bool
	field     string
	value     string
	write     bool
}

// DerivedNotifies returns the notify codes implied by how this record differs
// from the device's previously seen state, and persists the new state.
//
// These codes exist because saas_0 reports them nowhere: the device tells us its
// current fence and charge state on every GPS/BMS report but never announces the
// transition, and ebike-device-consume derives them from state it keeps for its
// own shadow. Detecting the edge ourselves keeps the dependency at Redis instead
// of pulling in ebike-fence and ebike-management.
//
// The tradeoff is inherent to edge detection on an at-least-once stream: a
// dropped or reordered report means a missed or duplicated edge. Codes sourced
// straight from an alarm packet are strictly more reliable.
func DerivedNotifies(r *Record) []int {
	cfg := config.GlobalConfig().Open.Notify
	wantFence := cfg.FenceEnabled && r.DeviceDataType == typeGPS
	soc := recordSoc(r)
	wantSoc := cfg.SocStepsEnabled && len(cfg.SocSteps) > 0 && soc != nil
	if !wantFence && !wantSoc {
		// Nothing to read and nothing to write. The whole state hash stops being
		// refreshed and expires on its own TTL.
		return nil
	}

	key := rediskey.NotifyState(r.TenantID(), r.Imei)
	prev, err := redis.HGetAll(key)
	if err != nil {
		// A failed read is not an empty state. Treating it as one would make
		// every device look like a first report, and the report after Redis
		// recovers would then compare against a stale value and announce an
		// edge that never happened.
		return nil
	}
	dropDisabledState(key, prev, cfg)

	var edges []edge
	if wantFence {
		edges = append(edges, fenceEdge(r, prev))
	}
	if wantSoc {
		edges = append(edges, socEdge(*soc, cfg.SocSteps, prev))
	}

	var out []int
	updates := make([]interface{}, 0, 2*len(edges))
	for _, e := range edges {
		if e.hasNotify {
			out = append(out, e.notify)
		}
		if e.write {
			updates = append(updates, e.field, e.value)
		}
	}
	if len(updates) > 0 {
		ttl := time.Duration(cfg.StateTTLSeconds) * time.Second
		if ttl <= 0 {
			ttl = 24 * time.Hour
		}
		redis.HSetTTL(key, ttl, updates...)
	}
	return out
}

// dropDisabledState removes state left behind by a feature that has since been
// switched off, so a later re-enable starts from "unknown" and records the
// device's current state instead of comparing against a value from before the
// gap and announcing a crossing nobody was watching for.
//
// It only fires when the stale field is actually present, which makes it a
// one-off per device rather than a write on every record. The case it exists for
// is one feature off and the other on: the surviving feature keeps refreshing the
// hash TTL, so the disabled feature's field would otherwise never expire.
func dropDisabledState(key string, prev map[string]string, cfg config.NotifyConfig) {
	var fields []string
	if !cfg.FenceEnabled {
		if _, stale := prev[fieldOutOfFence]; stale {
			fields = append(fields, fieldOutOfFence)
		}
	}
	if !cfg.SocStepsEnabled {
		if _, stale := prev[fieldSocStep]; stale {
			fields = append(fields, fieldSocStep)
		}
	}
	if len(fields) == 0 {
		return
	}
	for _, f := range fields {
		delete(prev, f)
	}
	redis.HDel(key, fields...)
}

// fenceEdge detects a service-area crossing from GPS switch bit 15.
//
// The device only maintains that bit while fencing is enabled (bit 14), so a
// device with fencing off is skipped entirely rather than read as "inside" —
// otherwise enabling a fence later would fire a spurious enter event.
//
// A device's first report only records state: without a previous value there is
// no edge, and treating "unknown" as "inside" would announce an exit to every
// third party whenever this service restarts with a cold state key.
func fenceEdge(r *Record, prev map[string]string) edge {
	if enabled, has := intOf(r.IsFenceEnable); !has || enabled != 1 {
		return edge{}
	}
	current, has := intOf(r.IsOutofServAera)
	if !has {
		return edge{}
	}
	value := strconv.Itoa(current)
	before, seen := prev[fieldOutOfFence]
	switch {
	case !seen:
		return edge{field: fieldOutOfFence, value: value, write: true}
	case before == value:
		return edge{}
	case current == 1:
		return edge{notify: NotifyFenceExit, hasNotify: true, field: fieldOutOfFence, value: value, write: true}
	default:
		return edge{notify: NotifyFenceEnter, hasNotify: true, field: fieldOutOfFence, value: value, write: true}
	}
}

// socEdge detects a downward crossing of a configured SOC threshold.
//
// Only downward crossings fire, and charging back up rearms the step silently:
// Xiaoan's 21/22 mean "battery dropped to 50%/30%", so a recharge must not
// announce them in reverse. A jump that skips a threshold reports only the
// deepest one reached, since that is the state the battery is actually in.
func socEdge(soc int, steps []config.SocStep, prev map[string]string) edge {
	current := currentSocStep(soc, steps)
	value := strconv.Itoa(current)

	raw, seen := prev[fieldSocStep]
	if !seen {
		// Record where the battery already sits, so a device discovered below
		// 30% does not immediately emit both steps.
		return edge{field: fieldSocStep, value: value, write: true}
	}
	before, err := strconv.Atoi(raw)
	if err != nil {
		return edge{field: fieldSocStep, value: value, write: true}
	}
	if current >= before {
		// Same threshold, or recharged above one: rearm without notifying.
		if current == before {
			return edge{}
		}
		return edge{field: fieldSocStep, value: value, write: true}
	}
	notify, ok := notifyForSocStep(current, steps)
	if !ok {
		return edge{field: fieldSocStep, value: value, write: true}
	}
	return edge{notify: notify, hasNotify: true, field: fieldSocStep, value: value, write: true}
}

// currentSocStep returns the percentage of the lowest threshold this SOC has
// reached, or socStepAboveAll when it is above all of them.
func currentSocStep(soc int, steps []config.SocStep) int {
	step := socStepAboveAll
	for _, s := range steps {
		if soc <= s.Percent && s.Percent < step {
			step = s.Percent
		}
	}
	return step
}

// notifyForSocStep resolves a threshold percentage back to its notify code.
// Duplicate percentages resolve to the lowest notify code so the result does not
// depend on config ordering.
func notifyForSocStep(percent int, steps []config.SocStep) (int, bool) {
	matches := make([]int, 0, 1)
	for _, s := range steps {
		if s.Percent == percent {
			matches = append(matches, s.Notify)
		}
	}
	if len(matches) == 0 {
		return 0, false
	}
	sort.Ints(matches)
	return matches[0], true
}

// recordSoc picks whichever state-of-charge the record carries: a standalone BMS
// report uses soc, a GPS report carries bmsSoc in a TLV.
func recordSoc(r *Record) *int {
	if r.DeviceDataType == typeBMS {
		if v, ok := intOf(r.Soc); ok {
			return &v
		}
	}
	if v, ok := intOf(r.BmsSoc); ok {
		return &v
	}
	return nil
}
