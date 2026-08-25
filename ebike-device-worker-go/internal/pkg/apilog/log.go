package apilog

import (
	"encoding/json"
	"log"
	"strings"
)

const maxLen = 2048

// Params logs request body for traffic analysis (path + JSON payload).
func Params(path string, body []byte) {
	s := strings.TrimSpace(string(body))
	if s == "" {
		log.Printf("[API] path=%s params={}", path)
		return
	}
	if len(s) > maxLen {
		s = s[:maxLen] + "...(truncated)"
	}
	log.Printf("[API] path=%s params=%s", path, s)
}

// Object marshals a bound request DTO and logs it.
func Object(path string, obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		log.Printf("[API] path=%s params=<marshal error: %v>", path, err)
		return
	}
	Params(path, b)
}
