package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"ebike-open-paas-go/internal/pkg/config"
)

var (
	consoleMu      sync.Mutex
	consoleToken   string
	consoleExpMs   int64
	consoleFetched time.Time
)

const consoleAuthReuseSkew = time.Minute

// PreloadConsoleAuth logs in to anvelink-console when consoleAuth credentials
// are configured, so the first allDevices call does not pay login latency.
func PreloadConsoleAuth() {
	cfg := config.GlobalConfig().Xyy
	if strings.TrimSpace(cfg.ConsoleAuthToken) != "" {
		log.Printf("[client] console auth: using static consoleAuthToken (len=%d)", len(cfg.ConsoleAuthToken))
		return
	}
	if strings.TrimSpace(cfg.ConsoleAuth.Phone) == "" {
		log.Printf("[client] console auth preload skipped: consoleAuth.phone not configured")
		return
	}
	if _, err := consoleBearer(false); err != nil {
		log.Printf("[client] console auth preload failed: %v", err)
		return
	}
	log.Printf("[client] console auth preload ok (phone=%q)", cfg.ConsoleAuth.Phone)
}

func isConsoleAuthError(err error) bool {
	ge, ok := err.(*GatewayError)
	if !ok {
		return false
	}
	switch ge.Code {
	case "5001", "5002", "5003":
		return true
	default:
		return false
	}
}

func consoleBearer(forceRefresh bool) (string, error) {
	cfg := config.GlobalConfig().Xyy
	if tok := strings.TrimSpace(cfg.ConsoleAuthToken); tok != "" && !forceRefresh {
		return formatBearer(tok), nil
	}
	phone := strings.TrimSpace(cfg.ConsoleAuth.Phone)
	password := cfg.ConsoleAuth.Password
	if phone == "" || password == "" {
		return "", fmt.Errorf("console auth: set xyy.consoleAuth.phone/password or consoleAuthToken")
	}

	consoleMu.Lock()
	defer consoleMu.Unlock()

	if !forceRefresh && consoleToken != "" && consoleTokenStillValid() {
		return formatBearer(consoleToken), nil
	}

	if err := loginConsoleLocked(phone, password); err != nil {
		return "", err
	}
	return formatBearer(consoleToken), nil
}

func consoleTokenStillValid() bool {
	if consoleExpMs == -1 {
		return true
	}
	if consoleExpMs > 0 {
		return time.Now().UnixMilli()+consoleAuthReuseSkew.Milliseconds() < consoleExpMs
	}
	// Login did not return exp; reuse for up to 23h (console default is 24h).
	return !consoleFetched.IsZero() && time.Since(consoleFetched) < 23*time.Hour
}

func loginConsoleLocked(phone, password string) error {
	base := strings.TrimRight(config.GlobalConfig().Xyy.ConsoleURL, "/")
	if base == "" {
		return fmt.Errorf("console url not configured (xyy.consoleUrl)")
	}
	env, err := postJSON(base+"/auth/login", map[string]interface{}{
		"phone":    phone,
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("console login: %w", err)
	}
	if !env.Success {
		return &GatewayError{Code: env.Code, Msg: env.Msg}
	}
	var data struct {
		Token string `json:"token"`
		Exp   int64  `json:"exp"`
	}
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return fmt.Errorf("console login decode: %w", err)
		}
	}
	if strings.TrimSpace(data.Token) == "" {
		return fmt.Errorf("console login: empty token in response")
	}
	consoleToken = data.Token
	consoleExpMs = data.Exp
	consoleFetched = time.Now()
	log.Printf("[client] console login ok (phone=%q exp=%d)", phone, data.Exp)
	return nil
}

func formatBearer(tok string) string {
	tok = strings.TrimSpace(tok)
	if strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		return tok
	}
	return "Bearer " + tok
}
