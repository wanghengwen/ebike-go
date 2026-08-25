package rpc

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type PaasClient struct {
	baseURL string
	enabled bool
	client  *http.Client
}

func NewPaasClient(baseURL string, enabled bool) *PaasClient {
	return &PaasClient{
		baseURL: baseURL,
		enabled: enabled,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PaasClient) Restart(imei, tenantID string) {
	if p == nil || !p.enabled || p.baseURL == "" {
		return
	}
	traceID := uuid.NewString()
	body := map[string]interface{}{
		"async": true,
		"imei":  imei,
		"commandContext": map[string]interface{}{
			"traceId":  traceID,
			"tenantId": tenantID,
			"pin":      "@system",
		},
	}
	raw, _ := json.Marshal(body)
	url := p.baseURL + "/device/paas/restart"
	resp, err := p.client.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		log.Printf("[paas] restart failed imei=%s: %v", imei, err)
		return
	}
	_ = resp.Body.Close()
	log.Printf("[paas] restart requested imei=%s tenant=%s", imei, tenantID)
}
