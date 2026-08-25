package message_test

import (
	"encoding/json"
	"testing"
	"time"

	"ebike-device-worker-go/internal/domain/message"
)

func TestReceiveMillisTimeUnmarshal(t *testing.T) {
	var dto message.DeviceMessageDTO
	raw := `{"imei":"860123456789012","receiveDataTime":1700000123456,"data":{"cmd":2}}`
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		t.Fatal(err)
	}
	if dto.ReceiveDataTime == nil {
		t.Fatal("expected receiveDataTime")
	}
	got := dto.ReceiveDataTime.TruncateToSecond()
	want := time.Unix(1700000123, 0).UTC()
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestReceiveMillisTimeStringMillis(t *testing.T) {
	var r message.ReceiveMillisTime
	if err := json.Unmarshal([]byte(`"1700000123456"`), &r); err != nil {
		t.Fatal(err)
	}
	if r.TruncateToSecond().Unix() != 1700000123 {
		t.Fatalf("unexpected %d", r.TruncateToSecond().Unix())
	}
}
