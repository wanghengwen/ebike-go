package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type checkResult struct {
	name string
	ok   bool
	msg  string
}

func runSmoke(cfg Config, args []string) error {
	fs := flag.NewFlagSet("smoke", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	verbose := fs.Bool("v", false, "print raw responses")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := cfg.requireAuth(); err != nil {
		return err
	}

	c := newClient(cfg)
	c.verbose = *verbose
	var results []checkResult

	// Missing token must fail closed with UNAUTHORIZED_ERROR.
	bad := newClient(Config{BaseURL: cfg.BaseURL, AgentID: cfg.AgentID, Token: "definitely-wrong-token"})
	env, status, err := bad.get("/ebike/v1/allDevices", url.Values{"pageSize": {"1"}, "pageNumber": {"1"}})
	results = append(results, check("auth rejects bad token",
		err == nil && status == 200 && env != nil && !env.Success && env.errorType() == "UNAUTHORIZED_ERROR",
		fmt.Sprintf("status=%d success=%v err=%s (%v)", status, env != nil && env.Success, safeErrType(env), err)))

	// Happy-path list should return the Xiaoan envelope (never bare HTML).
	env, status, err = c.get("/ebike/v1/allDevices", url.Values{"pageSize": {"1"}, "pageNumber": {"1"}})
	results = append(results, check("allDevices returns envelope", err == nil && status == 200 && env != nil,
		fmt.Sprintf("status=%d err=%v", status, err)))
	if env != nil && env.Success {
		results = append(results, check("allDevices success shape", json.Valid(env.Data), "data is valid JSON"))
	}

	// Unsupported endpoints must fail explicitly, not 404.
	for _, path := range []string{"/ebike/v1/sms", "/ebike/v1/lbs2gps", "/ebike/v1/batteryPowerSwitch"} {
		body := map[string]interface{}{"imei": "123456789012345"}
		env, status, err = c.post(path, body)
		results = append(results, check(path+" rejects unsupported",
			err == nil && status == 200 && env != nil && !env.Success,
			fmt.Sprintf("status=%d success=%v type=%s (%v)", status, env != nil && env.Success, safeErrType(env), err)))
	}

	// event=5 must be refused at registration.
	env, status, err = c.post("/ebike/v1/callback", map[string]interface{}{
		"event": 5,
		"url":   "https://example.com/hook",
	})
	results = append(results, check("callback rejects event 5 (UART)",
		err == nil && status == 200 && env != nil && !env.Success,
		fmt.Sprintf("status=%d success=%v type=%s promot=%s", status, env != nil && env.Success, safeErrType(env), safePromot(env))))

	// http:// callback must be refused (HTTPS only).
	env, status, err = c.post("/ebike/v1/callback", map[string]interface{}{
		"event": 1,
		"url":   "http://example.com/hook",
	})
	results = append(results, check("callback rejects http url",
		err == nil && status == 200 && env != nil && !env.Success,
		fmt.Sprintf("status=%d success=%v type=%s promot=%s", status, env != nil && env.Success, safeErrType(env), safePromot(env))))

	if cfg.IMEI != "" {
		imei, ierr := cfg.requireIMEI("")
		if ierr == nil {
			env, status, err = c.get("/ebike/v1/deviceInfo", url.Values{"imei": {imei}})
			results = append(results, check("deviceInfo envelope", err == nil && status == 200 && env != nil,
				fmt.Sprintf("status=%d success=%v (%v)", status, env != nil && env.Success, err)))
		}
	}

	failed := 0
	for _, r := range results {
		mark := "PASS"
		if !r.ok {
			mark = "FAIL"
			failed++
		}
		fmt.Printf("%s  %-40s  %s\n", mark, r.name, r.msg)
	}
	fmt.Printf("\n%d/%d checks passed\n", len(results)-failed, len(results))
	if failed > 0 {
		return fmt.Errorf("%d smoke check(s) failed", failed)
	}
	return nil
}

func check(name string, ok bool, msg string) checkResult {
	return checkResult{name: name, ok: ok, msg: msg}
}

func safeErrType(env *Envelope) string {
	if env == nil {
		return ""
	}
	return env.errorType()
}

func safePromot(env *Envelope) string {
	if env == nil {
		return ""
	}
	return env.promot()
}

func parseEvents(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("events is empty")
	}
	var out []int
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err != nil || n < 1 || n > 5 {
			return nil, fmt.Errorf("invalid event %q (want 1..5)", p)
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no events parsed from %q", raw)
	}
	return out, nil
}
