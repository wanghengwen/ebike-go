package kafka_test

import (
	"testing"
	"time"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/infrastructure/kafka"

	"github.com/IBM/sarama"
)

func TestBuildSaramaConfig(t *testing.T) {
	cfg := &config.Config{}
	cfg.Kafka.Brokers = []string{"192.168.2.20:9092"}
	cfg.Kafka.Consumers.StartOffset = "latest"

	sc, err := kafka.BuildSaramaConfigForTest(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if sc.ClientID != "ebike-device-worker-go" {
		t.Fatalf("ClientID=%q", sc.ClientID)
	}
	if sc.Consumer.Group.Session.Timeout != 10*time.Second {
		t.Fatalf("SessionTimeout=%v", sc.Consumer.Group.Session.Timeout)
	}
	if sc.Consumer.Group.Heartbeat.Interval != 3*time.Second {
		t.Fatalf("HeartbeatInterval=%v", sc.Consumer.Group.Heartbeat.Interval)
	}
	if sc.Consumer.Group.Rebalance.Timeout != 60*time.Second {
		t.Fatalf("RebalanceTimeout=%v", sc.Consumer.Group.Rebalance.Timeout)
	}
	if sc.Consumer.Offsets.Initial != sarama.OffsetNewest {
		t.Fatalf("Initial offset=%d", sc.Consumer.Offsets.Initial)
	}
	if len(sc.Consumer.Group.Rebalance.GroupStrategies) != 1 {
		t.Fatalf("GroupStrategies=%d want 1", len(sc.Consumer.Group.Rebalance.GroupStrategies))
	}
	if sc.Producer.RequiredAcks != sarama.WaitForLocal {
		t.Fatalf("RequiredAcks=%v", sc.Producer.RequiredAcks)
	}
}

func TestBuildSaramaConfigEarliestOffset(t *testing.T) {
	cfg := &config.Config{}
	cfg.Kafka.Consumers.StartOffset = "earliest"
	sc, err := kafka.BuildSaramaConfigForTest(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Consumer.Offsets.Initial != sarama.OffsetOldest {
		t.Fatalf("Initial offset=%d", sc.Consumer.Offsets.Initial)
	}
}

func TestManagerRestartWhenDisabled(t *testing.T) {
	cfg := &config.Config{}
	cfg.Kafka.Enabled = false
	m := kafka.NewManager(cfg)
	m.Start(nil, nil, nil)
	m.Restart()
	m.Stop()
}

func TestProducerReinit(t *testing.T) {
	cfg := &config.Config{}
	cfg.Kafka.Brokers = []string{"127.0.0.1:9092"}
	p := kafka.NewProducer(cfg)
	if err := p.Reinit(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Kafka.Brokers = nil
	if err := p.Reinit(cfg); err != nil {
		t.Fatal(err)
	}
	_ = p.Close()
}
