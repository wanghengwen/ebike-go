package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/metrics"

	kafkalib "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Pusher handles sending messages to Kafka
type Pusher struct {
	DryRun bool
}

var (
	writer   *kafkalib.Writer
	writerMu sync.RWMutex
	writerWG sync.WaitGroup
)

// InitKafkaWriter initializes the global Kafka writer with production-grade settings.
// Configuration is read from Nacos (GlobalConfig) or from environment variables.
func InitKafkaWriter() {
	brokersEnv := config.GetConfig().Spring.Kafka.BootstrapServers
	if brokersEnv == "" {
		brokersEnv = os.Getenv("KAFKA_BROKERS")
	}
	if brokersEnv == "" {
		brokersEnv = "127.0.0.1:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	writerMu.Lock()
	writer = &kafkalib.Writer{
		Addr: kafkalib.TCP(brokers...),
		// Murmur2Balancer reproduces the Java kafka-clients DefaultPartitioner that
		// the Java ebike-device-openapi gets from KafkaTemplate, so the same IMEI
		// lands on the same partition in both runtimes. Must stay set: kafka-go
		// falls back to round-robin when Balancer is nil, which ignores the key
		// entirely and destroys the per-device ordering the callers below rely on.
		Balancer: &kafkalib.Murmur2Balancer{},
		// AllowAutoTopicCreation prevents error when topic doesn't exist yet
		AllowAutoTopicCreation: true,
		// RequiredAcks=1: leader acknowledges write. Balances throughput vs. durability.
		// Java spring-kafka default is also 1 (acks=1).
		RequiredAcks: kafkalib.RequireOne,
		// MaxAttempts: retry up to 3 times on transient network errors
		MaxAttempts: 3,
		// BatchTimeout: flush every 50ms at most (tradeoff between latency and throughput)
		BatchTimeout: 50 * time.Millisecond,
		// WriteTimeout: individual write deadline
		WriteTimeout: 10 * time.Second,
		// ReadTimeout: read deadline for broker responses
		ReadTimeout: 10 * time.Second,
		// Compression: Snappy for space efficiency (matches common Spring Kafka default)
		Compression: kafkalib.Snappy,
	}
	writerMu.Unlock()
	logger.Log.Info("Kafka writer initialized", zap.String("brokers", brokersEnv))
}

// CloseKafkaWriter flushes pending messages and closes the writer.
// Must be called during graceful shutdown to prevent message loss.
func CloseKafkaWriter() {
	// Wait for any in-flight async writes to finish
	writerWG.Wait()

	writerMu.Lock()
	defer writerMu.Unlock()
	if writer != nil {
		if err := writer.Close(); err != nil {
			logger.Log.Error("Kafka writer close error", zap.Error(err))
		} else {
			logger.Log.Info("Kafka writer closed gracefully")
		}
		writer = nil
	}
}

// PushMessage sends a DeviceReportMessage to the appropriate Kafka topic.
// The partition key is the device IMEI, ensuring ordered processing per device
// in ebike-device-worker, which consumes these data/event/alarm topics.
//
// This method runs the Kafka write in an async goroutine to avoid blocking the
// HTTP handler. The goroutine is protected by a defer recover() to prevent
// any panic from propagating and crashing the process.
func (p *Pusher) PushMessage(imei string, msg *dto.DeviceReportMessage, shadowMode bool) {
	writerMu.RLock()
	w := writer
	writerMu.RUnlock()
	if w == nil {
		logger.Log.Warn("kafka pusher not initialized, skipping message")
		return
	}
	if shadowMode {
		logger.Log.Info("shadow mode intercept kafka push", zap.Any("msg", msg))
		return
	}
	if msg == nil || msg.Data == "" || msg.MsgType == "" {
		logger.Log.Warn("PushMessage invalid message", zap.String("imei", imei))
		return
	}
	stampReceiveDataTime(&msg.ReceiveDataTime)

	// msgType determines the target topic: data / alarm / event
	cfg := config.GetConfig()
	topic := cfg.Spring.Kafka.Producer.DataTopic
	if msg.MsgType == "alarm" {
		topic = cfg.Spring.Kafka.Producer.AlarmTopic
	} else if msg.MsgType == "event" {
		topic = cfg.Spring.Kafka.Producer.EventTopic
	}

	dataBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error("PushMessage marshal failed", zap.Any("msg", msg), zap.Error(err))
		return
	}

	if p.DryRun || !cfg.Spring.Kafka.Switch {
		logger.Log.Info("PushMessage dry-run/switch-off",
			zap.String("imei", imei),
			zap.String("topic", topic),
			zap.String("msgType", msg.MsgType),
			zap.Int("payloadBytes", len(dataBytes)),
		)
		return
	}

	writerMu.RLock()
	w = writer
	writerMu.RUnlock()
	if w == nil {
		logger.Log.Warn("PushMessage kafka writer closed before send", zap.String("imei", imei))
		return
	}

	metrics.IncKafkaPush()

	// Async write: never block the HTTP handler goroutine.
	// The partition key is imei, guaranteeing per-device message ordering.
	msgBytes := make([]byte, len(dataBytes))
	copy(msgBytes, dataBytes)
	topicCopy := topic
	imeiCopy := imei

	writerWG.Add(1)
	go func() {
		// Panic guard: a panic in this goroutine must not crash the entire process.
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("PushMessage panic recovered", zap.String("imei", imeiCopy), zap.Any("panic", r))
			}
			writerWG.Done()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := w.WriteMessages(ctx,
			kafkalib.Message{
				Topic: topicCopy,
				// Key = IMEI ensures all messages for one device go to the same partition,
				// preserving per-device order in downstream consumers.
				Key:   []byte(imeiCopy),
				Value: msgBytes,
			},
		)
		if err != nil {
			fields := []zap.Field{
				zap.String("imei", imeiCopy),
				zap.String("topic", topicCopy),
				zap.Int("payloadBytes", len(msgBytes)),
				zap.Error(err),
			}
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logger.Log.Error("kafka send message timed out", fields...)
			} else {
				logger.Log.Error("kafka send message failed", fields...)
			}
			return
		}
		// logger.Log.Info("PushMessage kafka sent",
		// 	zap.String("imei", imeiCopy),
		// 	zap.String("topic", topicCopy),
		// 	zap.Int("payloadBytes", len(msgBytes)),
		// )
	}()
}

// PushReplyMessage sends a DeviceReplyMessage directly to Kafka.
//
// Java context: DeviceReplyMessage extends DeviceReportMessage. When Java calls
// messagePusher.pushMessage(imei, deviceReplyMessage), Jackson serializes ALL fields
// (imei, msgType, bussinessType, data, msgId, identifier, payload) as a flat JSON object.
// Go must replicate this flat serialization — no wrapping or nesting.
func (p *Pusher) PushReplyMessage(imei string, msg *dto.DeviceReplyMessage, shadowMode bool) {
	if msg == nil || msg.MsgType == "" {
		logger.Log.Warn("PushReplyMessage invalid message", zap.String("imei", imei))
		return
	}
	if shadowMode {
		logger.Log.Info("shadow mode intercept kafka reply push", zap.Any("msg", msg))
		return
	}
	stampReceiveDataTime(&msg.ReceiveDataTime)

	topic := config.GetConfig().Spring.Kafka.Producer.EventTopic

	dataBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error("PushReplyMessage marshal failed", zap.Error(err))
		return
	}

	if p.DryRun || !config.GetConfig().Spring.Kafka.Switch {
		return
	}

	writerMu.RLock()
	w := writer
	writerMu.RUnlock()
	if w == nil {
		logger.Log.Warn("PushReplyMessage kafka writer not initialized", zap.String("imei", imei))
		return
	}

	metrics.IncKafkaPush()

	msgBytes := make([]byte, len(dataBytes))
	copy(msgBytes, dataBytes)
	topicCopy := topic
	imeiCopy := imei

	writerWG.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("PushReplyMessage panic recovered", zap.String("imei", imeiCopy), zap.Any("panic", r))
			}
			writerWG.Done()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := w.WriteMessages(ctx,
			kafkalib.Message{
				Topic: topicCopy,
				Key:   []byte(imeiCopy),
				Value: msgBytes,
			},
		)
		if err != nil {
			fields := []zap.Field{
				zap.String("imei", imeiCopy),
				zap.String("topic", topicCopy),
				zap.Int("payloadBytes", len(msgBytes)),
				zap.Error(err),
			}
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logger.Log.Error("kafka reply send timed out", fields...)
			} else {
				logger.Log.Error("kafka reply send failed", fields...)
			}
		}
	}()
}

func stampReceiveDataTime(ms *int64) {
	if ms == nil || *ms != 0 {
		return
	}
	*ms = time.Now().UnixMilli()
}
