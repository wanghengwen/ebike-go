// Package oplog reproduces the Java operation-audit log
// (com.xyy.common.log.OperationLog + LoggerUtils.operationLog): each entry is
// serialized to a single JSON line and written to a dedicated "operation"
// logger. In ebike-device-paas the logback "operation" appender routes that
// logger to logs/operation/<app>.log with a bare "%msg%n" pattern, so the file
// contains one JSON object per line and nothing else. We mirror that exactly.
//
// These records are audit/analytics only (shipped to ELK); they never affect an
// API response, so the JSON shape — not byte-for-byte field ordering — is what
// matters for parity.
package oplog

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ebike-device-paas-go/internal/pkg/config"
)

// Entry mirrors the fields of com.xyy.common.log.OperationLog. fastjson omits
// null values, so omitempty keeps the output close to the Java line.
type Entry struct {
	TenantId  string      `json:"tenantId,omitempty"`
	TraceId   string      `json:"traceId,omitempty"`
	Platform  string      `json:"platform,omitempty"`
	Pin       string      `json:"pin,omitempty"`
	Name      string      `json:"name,omitempty"`
	CarId     string      `json:"carId,omitempty"`
	Imei      string      `json:"imei,omitempty"`
	Time      string      `json:"time,omitempty"`
	EventType string      `json:"eventType,omitempty"`
	EventName string      `json:"eventName,omitempty"`
	Content   string      `json:"content,omitempty"`
	Result    interface{} `json:"result,omitempty"`
}

var (
	initOnce sync.Once
	mu       sync.Mutex
	file     *os.File
)

// logPath resolves the operation-log file path. It honours OPERATION_LOG_PATH
// (full file path) and LOG_PATH (base dir, default ./logs), mirroring the Java
// "${log.path}/operation/<app>.log" layout.
func logPath() string {
	if p := os.Getenv("OPERATION_LOG_PATH"); p != "" {
		return p
	}
	base := os.Getenv("LOG_PATH")
	if base == "" {
		base = "./logs"
	}
	app := config.GlobalConfig.Server.Name
	if app == "" {
		app = "ebike-device-paas"
	}
	return filepath.Join(base, "operation", app+".log")
}

func ensureFile() {
	initOnce.Do(func() {
		p := logPath()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			log.Printf("[oplog] mkdir %s failed: %v — operation logs go to stderr", filepath.Dir(p), err)
			return
		}
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Printf("[oplog] open %s failed: %v — operation logs go to stderr", p, err)
			return
		}
		file = f
	})
}

// Emit fills the time and writes the entry as one JSON line, mirroring
// LoggerUtils.operationLog(JsonUtils.toJson(operationLog)). It never blocks the
// caller on I/O errors; on failure it falls back to the standard logger.
func Emit(e Entry) {
	if e.Time == "" {
		e.Time = time.Now().Format("2006-01-02 15:04:05")
	}
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	ensureFile()
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		log.Printf("[operation] %s", b)
		return
	}
	b = append(b, '\n')
	if _, err := file.Write(b); err != nil {
		log.Printf("[oplog] write failed: %v; entry=%s", err, b)
	}
}
