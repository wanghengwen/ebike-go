package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/message"

	"github.com/IBM/sarama"
)

const (
	defaultBatchSize      = 500
	progressLogInterval   = 10 * time.Minute
)

type MessageHandler interface {
	HandleMessage(messages []message.DeviceMessageDTO)
}

type claimBatch struct {
	messages   []message.DeviceMessageDTO
	okMsgs     []*sarama.ConsumerMessage
	poisonMsgs []*sarama.ConsumerMessage
	closed     bool
}

type Manager struct {
	mu        sync.Mutex
	cfg       *config.Config
	groups    []sarama.ConsumerGroup
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	deviceLog func([]byte, *message.DeviceMessageDTO)
	dataH     MessageHandler
	eventH    MessageHandler
	alarmH    MessageHandler
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg, deviceLog: logDeviceMessage}
}

func logDeviceMessage(raw []byte, dto *message.DeviceMessageDTO) {
	if os.Getenv("KAFKA_DEVICE_LOG_VERBOSE") != "true" {
		return
	}
	if dto == nil {
		log.Printf("[xyy-device] message len=%d", len(raw))
		return
	}
	log.Printf("[xyy-device] %s", raw)
}

func (m *Manager) Start(dataH, eventH, alarmH MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dataH = dataH
	m.eventH = eventH
	m.alarmH = alarmH
	m.startLocked()
}

// Restart stops active consumers and starts them again with the latest config.
func (m *Manager) Restart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
	m.startLocked()
}

func (m *Manager) startLocked() {
	if m.cfg == nil || !m.cfg.Kafka.Enabled || len(m.cfg.Kafka.Brokers) == 0 {
		log.Println("[kafka] disabled or no brokers configured")
		return
	}
	log.Printf("[kafka] starting consumers brokers=%v data_topic=%s event_topic=%s alarm_topic=%s",
		m.cfg.Kafka.Brokers, m.cfg.Kafka.Topics.DataTopic, m.cfg.Kafka.Topics.EventTopic, m.cfg.Kafka.Topics.AlarmTopic)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	m.startConsumer(ctx, m.cfg.Kafka.Topics.DataTopic, m.cfg.Kafka.Consumers.GroupData, m.cfg.Kafka.Consumers.DataConcurrency, m.dataH)
	m.startConsumer(ctx, m.cfg.Kafka.Topics.EventTopic, m.cfg.Kafka.Consumers.GroupEvent, m.cfg.Kafka.Consumers.EventConcurrency, m.eventH)
	m.startConsumer(ctx, m.cfg.Kafka.Topics.AlarmTopic, m.cfg.Kafka.Consumers.GroupAlarm, m.cfg.Kafka.Consumers.AlarmConcurrency, m.alarmH)
}

func (m *Manager) startConsumer(ctx context.Context, topic, group string, concurrency int, handler MessageHandler) {
	if topic == "" || group == "" || handler == nil {
		return
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	saramaCfg, err := buildSaramaConfig(m.cfg)
	if err != nil {
		log.Printf("[kafka] invalid sarama config topic=%s group=%s: %v", topic, group, err)
		return
	}
	for i := 0; i < concurrency; i++ {
		cg, err := sarama.NewConsumerGroup(m.cfg.Kafka.Brokers, group, saramaCfg)
		if err != nil {
			log.Printf("[kafka] create consumer group failed topic=%s group=%s: %v", topic, group, err)
			continue
		}
		m.groups = append(m.groups, cg)
		m.wg.Add(1)
		go m.consumeGroupLoop(ctx, cg, topic, group, handler)
	}
	log.Printf("[kafka] consumer started topic=%s group=%s concurrency=%d brokers=%v start_offset=%s",
		topic, group, concurrency, m.cfg.Kafka.Brokers, resolveStartOffsetLabel(m.cfg.Kafka.Consumers.StartOffset))
}

// buildSaramaConfig aligns with Spring Boot 2.3.5 / Java kafka-clients defaults used by ebike-device-worker.
func buildSaramaConfig(cfg *config.Config) (*sarama.Config, error) {
	s := sarama.NewConfig()
	s.ClientID = "ebike-device-worker-go"
	s.Version = sarama.V2_5_0_0
	s.Consumer.Group.Session.Timeout = 10 * time.Second
	s.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	s.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	s.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	s.Consumer.Offsets.Initial = resolveInitialOffset(cfg.Kafka.Consumers.StartOffset)
	s.Consumer.Return.Errors = true
	s.Producer.RequiredAcks = sarama.WaitForLocal
	s.Producer.Return.Successes = true
	s.Producer.Return.Errors = true
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

func resolveStartOffsetLabel(raw string) string {
	if resolveInitialOffset(raw) == sarama.OffsetOldest {
		return "earliest"
	}
	return "latest"
}

func (m *Manager) consumeGroupLoop(ctx context.Context, cg sarama.ConsumerGroup, topic, group string, handler MessageHandler) {
	defer m.wg.Done()
	batchSize := m.cfg.Kafka.Consumers.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	wait := time.Duration(m.cfg.Kafka.Consumers.BatchWaitMs) * time.Millisecond
	if wait <= 0 {
		wait = 100 * time.Millisecond
	}

	consumer := &batchConsumerHandler{
		manager:   m,
		topic:     topic,
		group:     group,
		handler:   handler,
		batchSize: batchSize,
		wait:      wait,
	}

	var lastRebalanceLog time.Time
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		err := cg.Consume(ctx, []string{topic}, consumer)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return
			}
			if isRebalanceGroupErr(err) {
				if time.Since(lastRebalanceLog) >= 30*time.Second {
					log.Printf("[kafka] group rebalancing topic=%s group=%s (waiting to rejoin)", topic, group)
					lastRebalanceLog = time.Now()
				}
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("[kafka] group consume error topic=%s group=%s brokers=%v: %v", topic, group, m.cfg.Kafka.Brokers, err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

type batchConsumerHandler struct {
	manager   *Manager
	topic     string
	group     string
	handler   MessageHandler
	batchSize int
	wait      time.Duration
}

func (h *batchConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[kafka] joined topic=%s group=%s generation=%d member=%s assignments=%s",
		h.topic, h.group, session.GenerationID(), session.MemberID(), formatSessionClaims(session.Claims()))
	return nil
}

func (h *batchConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *batchConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var progress consumeProgress
	consumingLogged := false

	for {
		if err := session.Context().Err(); err != nil {
			return nil
		}

		fetched := h.fetchClaimBatch(session, claim)
		if fetched.closed && len(fetched.messages) == 0 && len(fetched.poisonMsgs) == 0 {
			return nil
		}
		if len(fetched.poisonMsgs) > 0 {
			markMessages(session, fetched.poisonMsgs)
			session.Commit()
		}
		if len(fetched.messages) == 0 {
			continue
		}

		if !consumingLogged {
			consumingLogged = true
			log.Printf("[kafka] consuming topic=%s group=%s partition=%d", h.topic, h.group, claim.Partition())
		}

		handledOK := true
		func() {
			defer func() {
				if r := recover(); r != nil {
					handledOK = false
					log.Printf("[kafka] handler panic: %v", r)
				}
			}()
			h.handler.HandleMessage(fetched.messages)
		}()
		if !handledOK {
			log.Printf("[kafka] skip commit for %d messages after handler failure", len(fetched.okMsgs))
			continue
		}
		markMessages(session, fetched.okMsgs)
		session.Commit()
		progress.add(len(fetched.okMsgs))
		progress.maybeLog(h.topic, h.group)
	}
}

func (h *batchConsumerHandler) fetchClaimBatch(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) claimBatch {
	var out claimBatch
	deadline := time.NewTimer(h.wait)
	defer deadline.Stop()

	for len(out.messages) < h.batchSize {
		select {
		case <-session.Context().Done():
			return out
		case <-deadline.C:
			return out
		case msg, ok := <-claim.Messages():
			if !ok {
				out.closed = true
				return out
			}
			var dto message.DeviceMessageDTO
			if err := json.Unmarshal(msg.Value, &dto); err != nil {
				log.Printf("[kafka] unmarshal failed, skip poison message offset=%d: %v", msg.Offset, err)
				out.poisonMsgs = append(out.poisonMsgs, msg)
				continue
			}
			h.manager.deviceLog(msg.Value, &dto)
			out.messages = append(out.messages, dto)
			out.okMsgs = append(out.okMsgs, msg)
		}
	}
	return out
}

func markMessages(session sarama.ConsumerGroupSession, msgs []*sarama.ConsumerMessage) {
	for _, msg := range msgs {
		session.MarkMessage(msg, "")
	}
}

func formatSessionClaims(claims map[string][]int32) string {
	if len(claims) == 0 {
		return "none (standby)"
	}
	parts := make([]string, 0)
	for topic, partitions := range claims {
		for _, p := range partitions {
			parts = append(parts, fmt.Sprintf("%s:p%d", topic, p))
		}
	}
	return strings.Join(parts, ",")
}

type consumeProgress struct {
	batches int
	msgs    int
	lastLog time.Time
}

func (p *consumeProgress) add(msgs int) {
	p.batches++
	p.msgs += msgs
}

func (p *consumeProgress) maybeLog(topic, group string) {
	if p.lastLog.IsZero() {
		p.lastLog = time.Now()
	}
	if time.Since(p.lastLog) < progressLogInterval {
		return
	}
	log.Printf("[kafka] progress topic=%s group=%s batches=%d msgs=%d (last 10m)",
		topic, group, p.batches, p.msgs)
	p.batches = 0
	p.msgs = 0
	p.lastLog = time.Now()
}

func isRebalanceGroupErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sarama.ErrRebalanceInProgress) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "rebalance in progress") || strings.Contains(lower, "group load in progress")
}

// BuildSaramaConfigForTest exposes sarama config for unit tests.
func BuildSaramaConfigForTest(cfg *config.Config) (*sarama.Config, error) {
	return buildSaramaConfig(cfg)
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
}

func (m *Manager) stopLocked() {
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

type Producer struct {
	mu       sync.Mutex
	producer sarama.SyncProducer
}

func NewProducer(cfg *config.Config) *Producer {
	p := &Producer{}
	_ = p.Reinit(cfg)
	return p
}

// Reinit rebuilds the kafka producer from the latest broker list.
func (p *Producer) Reinit(cfg *config.Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.producer != nil {
		_ = p.producer.Close()
		p.producer = nil
	}
	if cfg == nil || len(cfg.Kafka.Brokers) == 0 {
		return nil
	}
	saramaCfg, err := buildSaramaConfig(cfg)
	if err != nil {
		return err
	}
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, saramaCfg)
	if err != nil {
		return err
	}
	p.producer = producer
	return nil
}

func (p *Producer) Publish(topic, key, data string) {
	p.mu.Lock()
	producer := p.producer
	p.mu.Unlock()
	if producer == nil || topic == "" {
		return
	}
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(data),
	})
	if err != nil {
		log.Printf("[kafka] publish failed topic=%s key=%s: %v", topic, key, err)
	}
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.producer != nil {
		err := p.producer.Close()
		p.producer = nil
		return err
	}
	return nil
}
