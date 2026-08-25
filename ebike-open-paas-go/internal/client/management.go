package client

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

type tenantIotPlatformCo struct {
	IotPlatformAppId     string `json:"iotPlatformAppId"`
	IotPlatformAppKey    string `json:"iotPlatformAppKey"`
	IotPlatformAppSecret string `json:"iotPlatformAppSecret"`
	TenantID             string `json:"tenantId"`
}

var (
	iotMu        sync.RWMutex
	iotPlatforms map[string]tenantIotPlatformCo
	iotFetchedA  time.Time
)

const iotAppIDTTL = 5 * time.Minute

// PreloadIotPlatforms eagerly loads tenant IoT credentials on startup.
func PreloadIotPlatforms() {
	refreshIotAppIDs()
	iotMu.RLock()
	n := len(iotPlatforms)
	iotMu.RUnlock()
	log.Printf("[client] iot platform cache: %d tenants (managementUrl=%q)", n, config.GlobalConfig().Xyy.ManagementURL)
}

func iotPlatform(tenantID string) (tenantIotPlatformCo, bool) {
	if tenantID == "" {
		return tenantIotPlatformCo{}, false
	}
	iotMu.RLock()
	cached, ok := iotPlatforms[tenantID]
	fresh := time.Since(iotFetchedA) < iotAppIDTTL
	cacheSize := len(iotPlatforms)
	iotMu.RUnlock()
	if ok && fresh {
		return cached, true
	}
	if !fresh || cacheSize == 0 || !ok {
		refreshIotAppIDs()
	}
	iotMu.RLock()
	defer iotMu.RUnlock()
	co, ok := iotPlatforms[tenantID]
	return co, ok
}

func refreshIotAppIDs() {
	base := config.GlobalConfig().Xyy.ManagementURL
	if base == "" {
		log.Printf("[client] iot platform refresh skipped: managementUrl not configured")
		iotMu.Lock()
		iotFetchedA = time.Now()
		iotMu.Unlock()
		return
	}
	reqBody := map[string]interface{}{
		"commandContext": map[string]interface{}{
			"tenantId": "@paas",
			"pin":      "@paas",
			"traceId":  newTraceID(),
		},
	}
	env, err := postJSON(base+"/tenant/tenantIotPlatform/queryAll", reqBody)
	if err != nil {
		log.Printf("[client] iot platform refresh failed: %v", err)
		iotMu.Lock()
		iotFetchedA = time.Now()
		iotMu.Unlock()
		return
	}
	if !env.Success {
		log.Printf("[client] iot platform refresh failed: code=%s msg=%s", env.Code, env.Msg)
		iotMu.Lock()
		iotFetchedA = time.Now()
		iotMu.Unlock()
		return
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		iotMu.Lock()
		iotFetchedA = time.Now()
		iotMu.Unlock()
		return
	}
	var list []tenantIotPlatformCo
	if err := json.Unmarshal(env.Data, &list); err != nil {
		log.Printf("[client] iot platform refresh decode failed: %v", err)
		iotMu.Lock()
		iotFetchedA = time.Now()
		iotMu.Unlock()
		return
	}
	m := make(map[string]tenantIotPlatformCo, len(list))
	for _, co := range list {
		if co.TenantID != "" {
			m[co.TenantID] = co
		}
	}
	iotMu.Lock()
	iotPlatforms = m
	iotFetchedA = time.Now()
	iotMu.Unlock()
	log.Printf("[client] iot platform refresh ok: %d tenants loaded", len(m))
}

// NewTraceID returns a random UUID-like trace id for commandContext.
func NewTraceID() string { return newTraceID() }

func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("openpaas-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
