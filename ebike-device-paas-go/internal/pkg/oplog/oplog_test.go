package oplog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestEmitWritesJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "operation.log")
	t.Setenv("OPERATION_LOG_PATH", path)

	// Reset the lazily-initialized writer so the env var takes effect.
	initOnce = sync.Once{}
	file = nil
	// Windows cannot remove a file that is still open, so close it before the
	// t.TempDir cleanup runs.
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		if file != nil {
			_ = file.Close()
			file = nil
		}
	})

	Emit(Entry{
		TenantId:  "100",
		Pin:       "u1",
		Platform:  "wechat",
		Imei:      "imei-1",
		CarId:     "car-1",
		EventType: "17522",
		EventName: "播放寻车音",
		Content:   `{"idx":6}`,
	})

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var got Entry
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal line %q: %v", b, err)
	}
	if got.EventType != "17522" || got.EventName != "播放寻车音" {
		t.Fatalf("event mismatch: %+v", got)
	}
	if got.TenantId != "100" || got.Platform != "wechat" || got.Imei != "imei-1" {
		t.Fatalf("context fields mismatch: %+v", got)
	}
	if got.Time == "" {
		t.Fatal("time should be auto-filled")
	}
}
