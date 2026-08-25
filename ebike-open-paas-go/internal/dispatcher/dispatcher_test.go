package dispatcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRetryableStatus is the load-safety case: a 4xx means the endpoint has
// already answered about this body, so resending it only multiplies the load on a
// customer who told us no. Only a timeout, a throttle or a server fault can
// change on a resend.
func TestRetryableStatus(t *testing.T) {
	retryable := []int{
		http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, code := range retryable {
		if !retryableStatus(code) {
			t.Errorf("retryableStatus(%d) = false, want true", code)
		}
	}

	terminal := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusGone,
		http.StatusRequestEntityTooLarge,
		http.StatusUnprocessableEntity,
		http.StatusMovedPermanently,
	}
	for _, code := range terminal {
		if retryableStatus(code) {
			t.Errorf("retryableStatus(%d) = true, want false", code)
		}
	}
}

// TestLaneForIsStableAndSpreads: a URL must always land in the same lane, or a
// stalled customer would spread across every lane and take the whole pool down
// with it — the thing sharding exists to prevent.
func TestLaneForIsStableAndSpreads(t *testing.T) {
	shards = make([]*shard, 16)
	for i := range shards {
		shards[i] = &shard{queue: make(chan job, 1)}
	}

	const url = "https://open.example.com/callback"
	first := laneFor(url)
	for i := 0; i < 100; i++ {
		if got := laneFor(url); got != first {
			t.Fatalf("laneFor(%q) returned %d then %d, want a stable lane", url, first, got)
		}
	}

	used := map[int]bool{}
	for i := 0; i < 200; i++ {
		used[laneFor("https://tenant"+strings.Repeat("x", i%7)+".example.com/hook/"+string(rune('a'+i%26)))] = true
	}
	if len(used) < 8 {
		t.Errorf("200 distinct URLs used only %d of %d lanes; the hash is not spreading", len(used), len(shards))
	}
}

// TestEnqueueDropsWhenLaneStaysFull: dropping is the intended outcome of a lane
// that will not drain, because blocking the enqueue would stall the Kafka
// consumer behind whichever customer happens to be slowest.
func TestEnqueueDropsWhenLaneStaysFull(t *testing.T) {
	shards = []*shard{{queue: make(chan job, 1)}}
	enqueueWait = 0
	before, _, _ := Stats()

	enqueue(job{url: "https://open.example.com/hook"}, Event{Imei: "x", Event: 1})
	enqueue(job{url: "https://open.example.com/hook"}, Event{Imei: "x", Event: 1})

	after, _, _ := Stats()
	if after-before != 1 {
		t.Errorf("dropped advanced by %d, want 1 (the second event, with the lane full)", after-before)
	}
	if len(shards[0].queue) != 1 {
		t.Errorf("queue holds %d jobs, want the first one still buffered", len(shards[0].queue))
	}
}

// TestDeliverSendsIdempotencyHeaders: retries repost a byte-identical body (two
// GPS reports a second apart can be identical), so the receiver needs a stable id
// plus the attempt number to tell a retry from a genuinely new event.
func TestDeliverSendsIdempotencyHeaders(t *testing.T) {
	type got struct {
		id      string
		attempt string
	}
	var seen []got
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, got{r.Header.Get("X-Event-Id"), r.Header.Get("X-Event-Attempt")})
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client = srv.Client()
	deliver(job{url: srv.URL, body: []byte(`{"success":true}`), eventID: "inst-42"}, 3, 0)

	if len(seen) != 3 {
		t.Fatalf("server saw %d attempts, want 3 for a retryable 500", len(seen))
	}
	for i, s := range seen {
		if s.id != "inst-42" {
			t.Errorf("attempt %d had X-Event-Id %q, want the id stable across retries", i+1, s.id)
		}
	}
	if seen[0].attempt != "1" || seen[2].attempt != "3" {
		t.Errorf("attempt headers = %q…%q, want 1…3", seen[0].attempt, seen[2].attempt)
	}
}

// TestDeliverStopsOnClientError pairs with retryableStatus at the delivery level.
func TestDeliverStopsOnClientError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client = srv.Client()
	deliver(job{url: srv.URL, body: []byte(`{}`), eventID: "inst-1"}, 5, 0)

	if attempts != 1 {
		t.Errorf("server saw %d attempts for a 401, want 1", attempts)
	}
}

func TestDeliverCountsSuccessOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client = srv.Client()
	_, before, beforeFailed := Stats()
	deliver(job{url: srv.URL, body: []byte(`{}`), eventID: "inst-2"}, 3, 0)
	_, after, afterFailed := Stats()

	if after-before != 1 {
		t.Errorf("delivered advanced by %d, want 1", after-before)
	}
	if afterFailed != beforeFailed {
		t.Errorf("failed advanced by %d on a 200", afterFailed-beforeFailed)
	}
}

// TestGuardRedirectRequiresHTTPS: a 302 to http://… would otherwise drop the
// TLS requirement that registration enforced. Private destinations over HTTPS
// are allowed.
func TestGuardRedirectRequiresHTTPS(t *testing.T) {
	for _, raw := range []string{
		"http://open.example.com/hook",
		"http://10.0.0.5:8848/nacos/v1/cs/configs",
	} {
		req, err := http.NewRequest(http.MethodPost, raw, nil)
		if err != nil {
			t.Fatalf("NewRequest(%q): %v", raw, err)
		}
		if err := guardRedirect(req, nil); err == nil {
			t.Errorf("guardRedirect allowed an http hop to %q", raw)
		}
	}

	for _, raw := range []string{
		"https://open.example.com/hook2",
		"https://10.0.0.5:8443/hook",
	} {
		req, err := http.NewRequest(http.MethodPost, raw, nil)
		if err != nil {
			t.Fatalf("NewRequest(%q): %v", raw, err)
		}
		if err := guardRedirect(req, nil); err != nil {
			t.Errorf("guardRedirect rejected https hop %q: %v", raw, err)
		}
	}
}

func TestGuardRedirectStopsRedirectLoops(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://open.example.com/hook", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	via := make([]*http.Request, 3)
	if err := guardRedirect(req, via); err == nil {
		t.Error("guardRedirect followed a fourth redirect, want it capped")
	}
}

// TestNextEventIDIsUnique backs the idempotency header: a repeated id would make
// a receiver discard a genuinely new event as a duplicate.
func TestNextEventIDIsUnique(t *testing.T) {
	instanceID = newInstanceID()
	if instanceID == "" {
		t.Fatal("newInstanceID returned an empty id")
	}
	ids := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := nextEventID()
		if ids[id] {
			t.Fatalf("nextEventID repeated %q", id)
		}
		ids[id] = true
	}
}
