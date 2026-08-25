package kafka

import (
	"encoding/json"
	"testing"

	"ebike-device-openapi-go/internal/api/dto"

	kafkalib "github.com/segmentio/kafka-go"
)

func TestStampReceiveDataTimeOnMarshalPath(t *testing.T) {
	msg := &dto.DeviceReportMessage{
		Imei:          "860123456789012",
		MsgType:       "data",
		BussinessType: "ebike",
		Data:          `{"cmd":3}`,
	}
	stampReceiveDataTime(&msg.ReceiveDataTime)
	if msg.ReceiveDataTime <= 0 {
		t.Fatalf("receiveDataTime=%d", msg.ReceiveDataTime)
	}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["receiveDataTime"]; !ok {
		t.Fatalf("missing receiveDataTime in %s", b)
	}
}

// initWriterForTest builds the global writer and hands back its balancer. It does
// not connect to a broker: kafka-go dials lazily on the first write.
func initWriterForTest(t *testing.T) kafkalib.Balancer {
	t.Helper()
	InitKafkaWriter()
	writerMu.RLock()
	w := writer
	writerMu.RUnlock()
	if w == nil {
		t.Fatal("InitKafkaWriter left writer nil")
	}
	t.Cleanup(func() {
		writerMu.Lock()
		writer = nil
		writerMu.Unlock()
	})
	return w.Balancer
}

// TestWriterBalancerIsKeyAffine guards the partitioning contract that PushMessage
// documents. kafka-go silently falls back to round-robin when Balancer is nil, and
// round-robin discards the message key, so dropping the Balancer field would
// scatter one device's messages across every partition without any error
// surfacing.
func TestWriterBalancerIsKeyAffine(t *testing.T) {
	balancer := initWriterForTest(t)
	if _, ok := balancer.(*kafkalib.Murmur2Balancer); !ok {
		t.Fatalf("Balancer = %T, want *kafka.Murmur2Balancer (nil falls back to round-robin, which ignores the key)", balancer)
	}

	partitions := []int{0, 1, 2, 3, 4, 5, 6, 7}
	const imei = "865067022403441"

	msg := kafkalib.Message{Key: []byte(imei)}
	want := balancer.Balance(msg, partitions...)
	for i := 0; i < 100; i++ {
		if got := balancer.Balance(msg, partitions...); got != want {
			t.Fatalf("imei %s routed to partition %d on call %d, first call gave %d", imei, got, i, want)
		}
	}
}

// TestWriterBalancerSpreadsDevices checks the balancer still distributes distinct
// devices, so key affinity is not achieved by collapsing everything onto one
// partition.
func TestWriterBalancerSpreadsDevices(t *testing.T) {
	balancer := initWriterForTest(t)

	partitions := []int{0, 1, 2, 3, 4, 5, 6, 7}
	imeis := []string{
		"865067022403441", "865067022403442", "865067022403443", "865067022403444",
		"865067022403445", "865067022403446", "865067022403447", "865067022403448",
	}

	seen := make(map[int]struct{}, len(partitions))
	for _, imei := range imeis {
		seen[balancer.Balance(kafkalib.Message{Key: []byte(imei)}, partitions...)] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("%d IMEIs all mapped to %d partition(s); expected the hash to spread them", len(imeis), len(seen))
	}
}
