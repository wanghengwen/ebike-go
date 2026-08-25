// Package dispatcher fans device events out to third-party callback URLs.
// Delivery is asynchronous with bounded queues and a fixed worker pool so a
// slow customer URL can never stall our saas_0 consumer, which shares no
// partitions with ebike-device-worker or ebike-device-consume but would still
// lag its own group if delivery ran inline.
package dispatcher

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/repository"
)

// Event is one Xiaoan callback event, built by internal/event from a saas_0
// record. Data is the already-shaped Xiaoan payload for the event type.
type Event struct {
	TenantID string          `json:"tenantId"`
	Imei     string          `json:"imei"`
	Event    int             `json:"event"` // 1..5
	Data     json.RawMessage `json:"data"`
	Tm       int64           `json:"tm,omitempty"` // seconds; UART only
}

// CallbackBody is the Xiaoan outbound webhook body.
type CallbackBody struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type job struct {
	url     string
	body    []byte
	eventID string
}

// shard is one delivery lane. URLs are pinned to a lane by hash so a customer
// whose endpoint has stopped responding fills only its own queue. With a single
// shared queue, one such customer's retries occupy every worker and the backlog
// then drops events belonging to everybody else.
type shard struct {
	queue chan job
	wg    sync.WaitGroup
}

var (
	shards       []*shard
	client       *http.Client
	started      int32
	dropped      uint64
	delivered    uint64
	failed       uint64
	seq          uint64
	instanceID   string
	enqueueWait  time.Duration
	startOnce    sync.Once
	shutdownOnce sync.Once
)

// Start initialises the delivery pool. Safe to call multiple times.
func Start() {
	startOnce.Do(func() {
		cfg := config.GlobalConfig().Open.Callback
		size := cfg.QueueSize
		if size <= 0 {
			size = 20000
		}
		lanes := cfg.Workers
		if lanes <= 0 {
			lanes = 16
		}
		timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		enqueueWait = time.Duration(cfg.EnqueueWaitMs) * time.Millisecond
		if enqueueWait < 0 {
			enqueueWait = 0
		}
		perLane := size / lanes
		if perLane < 64 {
			perLane = 64
		}

		instanceID = newInstanceID()
		client = &http.Client{Timeout: timeout, CheckRedirect: guardRedirect}
		shards = make([]*shard, lanes)
		for i := range shards {
			s := &shard{queue: make(chan job, perLane)}
			shards[i] = s
			s.wg.Add(1)
			go func(s *shard) {
				defer s.wg.Done()
				work(s)
			}(s)
		}
		atomic.StoreInt32(&started, 1)
		log.Printf("[dispatcher] started lanes=%d queuePerLane=%d timeout=%s enqueueWait=%s circuit=%v failThreshold=%d cooldown=%ds instance=%s",
			lanes, perLane, timeout, enqueueWait, cfg.CircuitEnabled, cfg.CircuitFailThreshold, cfg.CircuitCooldownSeconds, instanceID)
	})
}

// Stats returns the delivery counters: events discarded because a lane stayed
// full, deliveries that succeeded, and deliveries that exhausted their attempts.
func Stats() (droppedN, deliveredN, failedN uint64) {
	return atomic.LoadUint64(&dropped), atomic.LoadUint64(&delivered), atomic.LoadUint64(&failed)
}

// Dropped returns how many events were discarded because a lane was full.
func Dropped() uint64 { return atomic.LoadUint64(&dropped) }

// Dispatch looks up subscriptions for the event and enqueues deliveries.
func Dispatch(ev Event) {
	if atomic.LoadInt32(&started) == 0 {
		Start()
	}
	if ev.TenantID == "" || ev.Imei == "" || !repository.IsValidEvent(ev.Event) {
		return
	}

	urls := repository.SubscribersFor(ev.TenantID, ev.Event)
	if len(urls) == 0 {
		return
	}

	data := map[string]interface{}{
		"imei":  ev.Imei,
		"event": ev.Event,
		"data":  json.RawMessage(ev.Data),
	}
	if ev.Event == repository.EventUART && ev.Tm > 0 {
		data["tm"] = ev.Tm
	}
	raw, err := json.Marshal(CallbackBody{Success: true, Data: data})
	if err != nil {
		log.Printf("[dispatcher] marshal failed: %v", err)
		return
	}

	eventID := nextEventID()
	for _, u := range urls {
		if !circuitAllow(u) {
			continue
		}
		enqueue(job{url: u, body: raw, eventID: eventID}, ev)
	}
}

// enqueue hands the job to the URL's lane, waiting briefly if the lane is busy.
//
// The wait is bounded rather than indefinite: a short pause lets a lane recover
// from a burst without losing anything, while blocking outright would stall the
// Kafka consumer behind whichever customer happens to be slowest.
func enqueue(j job, ev Event) {
	s := shards[laneFor(j.url)]
	select {
	case s.queue <- j:
		return
	default:
	}
	if enqueueWait > 0 {
		timer := time.NewTimer(enqueueWait)
		defer timer.Stop()
		select {
		case s.queue <- j:
			return
		case <-timer.C:
		}
	}
	atomic.AddUint64(&dropped, 1)
	log.Printf("[dispatcher] lane full, drop url=%s imei=%s event=%d eventId=%s",
		j.url, ev.Imei, ev.Event, j.eventID)
}

func laneFor(url string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	return int(h.Sum32() % uint32(len(shards)))
}

func work(s *shard) {
	cfg := config.GlobalConfig().Open.Callback
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	backoff := time.Duration(cfg.RetryBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	for j := range s.queue {
		deliver(j, maxAttempts, backoff)
	}
}

func deliver(j job, maxAttempts int, backoff time.Duration) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPost, j.url, bytes.NewReader(j.body))
		if err != nil {
			lastErr = err
			break
		}
		req.Header.Set("Content-Type", "application/json")
		// Retries repost an identical body, so the receiver needs a stable id to
		// tell a retry from a second event with the same values (two GPS reports
		// one second apart can be byte-identical).
		req.Header.Set("X-Event-Id", j.eventID)
		req.Header.Set("X-Event-Attempt", strconv.Itoa(attempt))

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				atomic.AddUint64(&delivered, 1)
				circuitSuccess(j.url)
				return
			}
			lastErr = &statusError{code: resp.StatusCode}
			if !retryableStatus(resp.StatusCode) {
				// 401/404/410 and friends mean the endpoint will reject this
				// body however often we resend it. Retrying only multiplies the
				// load on a customer who has already answered.
				break
			}
		}
		if attempt < maxAttempts {
			time.Sleep(backoff * time.Duration(attempt))
		}
	}
	atomic.AddUint64(&failed, 1)
	circuitFailure(j.url)
	log.Printf("[dispatcher] deliver failed url=%s eventId=%s err=%v", j.url, j.eventID, lastErr)
}

// retryableStatus reports whether resending can plausibly change the outcome.
func retryableStatus(code int) bool {
	switch {
	case code == http.StatusRequestTimeout, code == http.StatusTooManyRequests:
		return true
	case code >= 500:
		return true
	default:
		return false
	}
}

// Shutdown stops accepting work and waits for the in-flight backlog, so events
// already taken off Kafka and committed are not lost on a rolling restart.
func Shutdown(timeout time.Duration) {
	if atomic.LoadInt32(&started) == 0 {
		return
	}
	shutdownOnce.Do(func() {
		for _, s := range shards {
			close(s.queue)
		}
		done := make(chan struct{})
		go func() {
			for _, s := range shards {
				s.wg.Wait()
			}
			close(done)
		}()
		select {
		case <-done:
			log.Println("[dispatcher] drained")
		case <-time.After(timeout):
			log.Printf("[dispatcher] drain timed out after %s, %d events dropped so far", timeout, Dropped())
		}
	})
}

type statusError struct{ code int }

func (e *statusError) Error() string { return "http status " + strconv.Itoa(e.code) }

// guardRedirect keeps every hop on HTTPS and caps redirect depth. A 302 to
// http://… would otherwise drop the TLS requirement that registration enforced.
func guardRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return fmt.Errorf("too many redirects")
	}
	return repository.ValidateCallbackURL(req.URL.String())
}

// nextEventID is unique per process run and per event: the random instance part
// keeps ids from colliding across replicas and restarts.
func nextEventID() string {
	return fmt.Sprintf("%s-%d", instanceID, atomic.AddUint64(&seq, 1))
}

func newInstanceID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b[:])
}
