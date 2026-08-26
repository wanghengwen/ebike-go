// Package kafka wraps a segmentio/kafka-go writer used to publish the C34
// device-state messages that the Java service emits via KafkaProducer.sendC34.
// When no brokers are configured the producer degrades to a no-op so the
// write/report endpoints still respond (the message is simply dropped + logged).
package kafka

import (
	"context"
	"log"
	"time"

	"ebike-device-paas-go/internal/pkg/config"

	kgo "github.com/segmentio/kafka-go"
)

var writer *kgo.Writer

// Init builds the shared writer from config.GlobalConfig.Kafka.Brokers. Safe to
// call after config load; a nil/empty broker list leaves the producer disabled.
func Init() {
	brokers := config.GlobalConfig.Kafka.Brokers
	if len(brokers) == 0 {
		log.Println("[kafka] no brokers configured, C34 producer disabled")
		return
	}
	writer = &kgo.Writer{
		Addr: kgo.TCP(brokers...),
		// Java calls kafkaTemplate.send(topic, value) with no key, so messages use
		// the default (keyless) partitioner — round-robin across partitions rather
		// than keyed/Hash partitioning.
		Balancer:     &kgo.RoundRobin{},
		WriteTimeout: 5 * time.Second,
		RequiredAcks: kgo.RequireOne,
		Async:        true,
	}
	log.Printf("[kafka] C34 producer ready (brokers=%v)", brokers)
}

// Send publishes value to topic with no message key, mirroring Java
// kafkaTemplate.send(topic, value). It is a no-op (logged) when the producer is
// disabled or the topic is empty.
func Send(topic string, value []byte) {
	if writer == nil {
		log.Printf("[kafka] producer disabled, dropping message (topic=%s)", topic)
		return
	}
	if topic == "" {
		log.Printf("[kafka] empty topic, dropping message")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := writer.WriteMessages(ctx, kgo.Message{
		Topic: topic,
		Value: value,
	}); err != nil {
		log.Printf("[kafka] send failed (topic=%s): %v", topic, err)
	}
}

// Close flushes and releases the writer.
func Close() error {
	if writer == nil {
		return nil
	}
	return writer.Close()
}
