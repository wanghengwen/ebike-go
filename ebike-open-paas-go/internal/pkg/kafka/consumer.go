// Package kafka consumes the saas_0 topic that ebike-device-worker publishes
// flattened device events to, and feeds them to the callback dispatcher.
//
// It uses IBM/sarama consumer groups, the same client ebike-device-worker-go and
// ebike-device-consume-go use, so the Java Spring-Kafka semantics those services
// were ported against (manual commit, range assignor, earliest/latest) carry over
// unchanged. We join with our own group id, so our offsets and rebalances are
// independent of theirs and both read the full topic.
package kafka

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ebike-open-paas-go/internal/pkg/config"

	"github.com/IBM/sarama"
)

// RecordHandler consumes one raw saas_0 record value.
type RecordHandler func(value []byte)

// Manager owns the consumer-group goroutines.
type Manager struct {
	mu      sync.Mutex
	handler RecordHandler
	groups  []sarama.ConsumerGroup
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
}

// current is the manager the readiness probe reports on. There is one consumer
// per process, so the probe does not need it threaded through the router.
var current atomic.Pointer[Manager]

// NewManager builds a consumer manager for the given record handler.
func NewManager(handler RecordHandler) *Manager {
	m := &Manager{handler: handler}
	current.Store(m)
	return m
}

// Enabled reports whether event callbacks are configured to be consumed at all.
// When they are not, readiness must not fail: the query and command APIs work
// without the consumer.
func Enabled() bool {
	return config.GlobalConfig().Kafka.Enabled
}

// Running reports whether consumer groups are live. A consumer that died leaves
// subscribers silently receiving nothing, which readiness should surface.
func Running() bool {
	m := current.Load()
	return m != nil && m.running.Load()
}

// Start validates the config and launches consumers. A configuration problem is
// reported and leaves the consumer off rather than aborting the process: the
// Xiaoan query and command APIs stay usable without event callbacks.
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startLocked()
}

// Restart rebuilds consumers from the current config, for a kafka.yaml change.
func (m *Manager) Restart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
	m.startLocked()
}

func (m *Manager) startLocked() {
	cfg := config.GlobalConfig().Kafka
	if !cfg.Enabled {
		log.Println("[kafka] disabled (kafka.enabled=false); event callbacks will not be delivered")
		return
	}
	if err := config.ValidateKafka(); err != nil {
		log.Printf("[kafka] not starting: %v", err)
		return
	}
	config.LogEffectiveKafka()

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	saramaCfg, err := buildSaramaConfig()
	if err != nil {
		log.Printf("[kafka] invalid sarama config: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	for i := 0; i < concurrency; i++ {
		cg, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
		if err != nil {
			log.Printf("[kafka] create consumer group failed group=%s: %v", cfg.GroupID, err)
			continue
		}
		m.groups = append(m.groups, cg)
		m.wg.Add(2)
		go m.consumeLoop(ctx, cg, cfg.Topics, cfg.GroupID)
		go m.drainErrors(cg, cfg.GroupID)
	}
	if len(m.groups) == 0 {
		log.Printf("[kafka] no consumer group could be created group=%s", cfg.GroupID)
		cancel()
		m.cancel = nil
		return
	}
	m.running.Store(true)
	log.Printf("[kafka] consumer started topics=%v group=%s concurrency=%d", cfg.Topics, cfg.GroupID, len(m.groups))
}

// buildSaramaConfig mirrors ebike-device-consume-go so this consumer behaves the
// same way against the same brokers.
//
// AutoCommit is off because we commit explicitly per batch; leaving sarama's
// default on would let its 1s ticker race session.Commit() and acknowledge
// records the dispatcher has not been handed yet.
func buildSaramaConfig() (*sarama.Config, error) {
	s := sarama.NewConfig()
	s.ClientID = "ebike-open-paas-go"
	s.Version = sarama.V2_5_0_0
	s.Consumer.Group.Session.Timeout = 10 * time.Second
	s.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	s.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	s.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	s.Consumer.Offsets.Initial = resolveInitialOffset(config.GlobalConfig().Kafka.StartOffset)
	s.Consumer.Offsets.AutoCommit.Enable = false
	s.Consumer.Return.Errors = true
	return s, s.Validate()
}

func resolveInitialOffset(raw string) int64 {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "earliest", "first", "beginning":
		return sarama.OffsetOldest
	default:
		return sarama.OffsetNewest
	}
}

func (m *Manager) consumeLoop(ctx context.Context, cg sarama.ConsumerGroup, topics []string, group string) {
	defer m.wg.Done()
	cfg := config.GlobalConfig().Kafka
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}
	wait := time.Duration(cfg.BatchWaitMs) * time.Millisecond
	if wait <= 0 {
		wait = 100 * time.Millisecond
	}
	handler := &groupHandler{
		manager:   m,
		group:     group,
		batchSize: batchSize,
		wait:      wait,
	}

	var lastRebalanceLog time.Time
	for {
		if ctx.Err() != nil {
			return
		}
		err := cg.Consume(ctx, topics, handler)
		if err == nil {
			continue
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}
		if isRebalanceErr(err) {
			if time.Since(lastRebalanceLog) >= 30*time.Second {
				log.Printf("[kafka] group rebalancing topics=%v group=%s", topics, group)
				lastRebalanceLog = time.Now()
			}
			time.Sleep(2 * time.Second)
			continue
		}
		log.Printf("[kafka] consume error topics=%v group=%s: %v", topics, group, err)
		time.Sleep(5 * time.Second)
	}
}

// drainErrors consumes the group's error channel. Config.Consumer.Return.Errors
// is on, which means sarama publishes every broker-level failure here and blocks
// on a full channel — leaving it unread stalls the consumer instead of just
// losing the diagnostics.
func (m *Manager) drainErrors(cg sarama.ConsumerGroup, group string) {
	defer m.wg.Done()
	for err := range cg.Errors() {
		log.Printf("[kafka] group=%s error: %v", group, err)
	}
}

type groupHandler struct {
	manager   *Manager
	group     string
	batchSize int
	wait      time.Duration
}

func (h *groupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[kafka] joined group=%s generation=%d member=%s claims=%v",
		h.group, session.GenerationID(), session.MemberID(), session.Claims())
	return nil
}

func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim drains a partition in batches and commits after each one.
//
// Handling is best-effort by design: the dispatcher enqueues onto a bounded
// queue and drops on overflow, so a record can never fail in a way that retrying
// the batch would fix. Committing unconditionally keeps a poison record from
// stalling the partition forever.
func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		if session.Context().Err() != nil {
			return nil
		}
		batch, closed := h.fetchBatch(session, claim)
		for _, msg := range batch {
			h.handle(msg)
		}
		if len(batch) > 0 {
			for _, msg := range batch {
				session.MarkMessage(msg, "")
			}
			session.Commit()
		}
		if closed {
			return nil
		}
	}
}

func (h *groupHandler) handle(msg *sarama.ConsumerMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[kafka] handler panic partition=%d offset=%d: %v", msg.Partition, msg.Offset, r)
		}
	}()
	h.manager.handler(msg.Value)
}

func (h *groupHandler) fetchBatch(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) ([]*sarama.ConsumerMessage, bool) {
	out := make([]*sarama.ConsumerMessage, 0, h.batchSize)
	deadline := time.NewTimer(h.wait)
	defer deadline.Stop()
	for len(out) < h.batchSize {
		select {
		case <-session.Context().Done():
			return out, true
		case <-deadline.C:
			return out, false
		case msg, ok := <-claim.Messages():
			if !ok {
				return out, true
			}
			out = append(out, msg)
		}
	}
	return out, false
}

func isRebalanceErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sarama.ErrRebalanceInProgress) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "rebalance in progress") || strings.Contains(lower, "group load in progress")
}

// Stop closes all consumer groups and waits for the goroutines to exit.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
}

func (m *Manager) stopLocked() {
	m.running.Store(false)
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	for _, g := range m.groups {
		_ = g.Close()
	}
	m.wg.Wait()
	m.groups = nil
}
