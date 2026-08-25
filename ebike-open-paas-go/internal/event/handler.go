package event

import (
	"encoding/json"
	"log"
	"sync/atomic"

	"ebike-open-paas-go/internal/dispatcher"
	"ebike-open-paas-go/internal/repository"
)

// Counters for the periodic consumer log. Callback delivery is invisible from
// the outside until something breaks, so the ratio of seen to matched records is
// the main signal that the subscription filter is behaving.
var (
	seen      uint64
	unmatched uint64
	malformed uint64
	emitted   uint64
	panicked  uint64
)

// Stats returns the running totals: records read, records dropped because no
// agent subscribes to that tenant, records that failed to parse, callback events
// handed to the dispatcher, and records abandoned by a panic.
func Stats() (seenN, unmatchedN, malformedN, emittedN, panickedN uint64) {
	return atomic.LoadUint64(&seen), atomic.LoadUint64(&unmatched),
		atomic.LoadUint64(&malformed), atomic.LoadUint64(&emitted),
		atomic.LoadUint64(&panicked)
}

// Handle processes one saas_0 record.
//
// The tenant filter runs before parsing costs anything beyond the unmarshal,
// because saas_0 carries every tenant's traffic while only the few tenants with
// a registered callback are relevant to us.
//
// A panic is contained to the record that caused it. Letting it escape would
// abandon the rest of the batch the consumer is holding, and since that batch is
// committed either way, every other record in it would be lost.
func Handle(value []byte) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddUint64(&panicked, 1)
			log.Printf("[event] panic handling record: %v", r)
		}
	}()
	handle(value)
}

func handle(value []byte) {
	atomic.AddUint64(&seen, 1)

	r, err := Parse(value)
	if err != nil {
		atomic.AddUint64(&malformed, 1)
		return
	}
	tenantID := r.TenantID()
	if tenantID == "" || r.Imei == "" {
		atomic.AddUint64(&malformed, 1)
		return
	}
	if !repository.HasTenant(tenantID) {
		atomic.AddUint64(&unmatched, 1)
		return
	}

	switch r.DeviceDataType {
	case typePing:
		emit(tenantID, r.Imei, repository.EventPing, PingPayload(r))
	case typeGPS:
		// Xiaoan v1.1.7: a report positioned at (0,0) is a device that has no
		// fix yet, and must not be forwarded. The test is on both coordinates
		// being present and zero — a record that simply omits them is a decoder
		// shape we do not recognise, not a device sitting in the Gulf of Guinea.
		if !isUnfixedGPS(r) {
			emit(tenantID, r.Imei, repository.EventGPS, GPSPayload(r))
		}
		emitNotifies(tenantID, r.Imei, DerivedNotifies(r))
	case typeBMS:
		emit(tenantID, r.Imei, repository.EventBMS, BMSPayload(r))
		emitNotifies(tenantID, r.Imei, DerivedNotifies(r))
	case typeAlarm:
		alarmType, ok := intOf(r.Type)
		if !ok {
			return
		}
		notify, ok := NotifyForAlarm(alarmType)
		if !ok {
			// Unmapped alarms are dropped rather than forwarded verbatim; see
			// notify.go for why passing the raw code through is unsafe.
			return
		}
		emitNotifies(tenantID, r.Imei, []int{notify})
	case typeLogin:
		emitNotifies(tenantID, r.Imei, []int{NotifyLogin})
	case typeLogout:
		// Xiaoan separates 4 (lost connection) from 10 (device closed the
		// connection itself), but saas_0 carries a single logout record with no
		// way to tell them apart. 4 is the safe reading: it is what a third
		// party acts on either way.
		emitNotifies(tenantID, r.Imei, []int{NotifyLostConnection})
	}
}

// isUnfixedGPS reports whether the record is the (0,0) placeholder a device
// sends before it has a fix.
func isUnfixedGPS(r *Record) bool {
	if r.Wgs84Lat == nil || r.Wgs84Lng == nil {
		return false
	}
	return *r.Wgs84Lat == 0 && *r.Wgs84Lng == 0
}

func emit(tenantID, imei string, event int, payload map[string]interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[event] marshal payload failed imei=%s event=%d: %v", imei, event, err)
		return
	}
	atomic.AddUint64(&emitted, 1)
	dispatcher.Dispatch(dispatcher.Event{
		TenantID: tenantID,
		Imei:     imei,
		Event:    event,
		Data:     raw,
	})
}

// emitNotifies sends each notify as its own event=3 callback, whose payload is
// the bare code rather than an object.
func emitNotifies(tenantID, imei string, notifies []int) {
	for _, n := range notifies {
		raw, err := json.Marshal(n)
		if err != nil {
			continue
		}
		atomic.AddUint64(&emitted, 1)
		dispatcher.Dispatch(dispatcher.Event{
			TenantID: tenantID,
			Imei:     imei,
			Event:    repository.EventNotify,
			Data:     raw,
		})
	}
}
