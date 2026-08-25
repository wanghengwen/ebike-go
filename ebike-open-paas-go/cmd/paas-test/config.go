package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config is the third-party client identity used against the open PaaS.
type Config struct {
	BaseURL string
	AgentID string
	Token   string
	IMEI    string
}

func parseGlobal(args []string) (Config, []string, error) {
	cfg := Config{
		BaseURL: strings.TrimRight(envOr("PAAS_BASE_URL", "https://paas.luopingtech.com"), "/"),
		AgentID: envOr("PAAS_AGENT_ID", ""),
		Token:   envOr("PAAS_ACCESS_TOKEN", ""),
		IMEI:    envOr("PAAS_IMEI", ""),
	}

	fs := flag.NewFlagSet("paas-test", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	base := fs.String("base-url", cfg.BaseURL, "PaaS base URL")
	agent := fs.String("agent-id", cfg.AgentID, "Xiaoan agentId")
	token := fs.String("token", "", "xc-access-token (overrides env)")
	tokenFile := fs.String("token-file", "", "file containing xc-access-token")
	imei := fs.String("imei", cfg.IMEI, "default device IMEI")

	// Stop at the first non-flag so subcommands keep their own flags.
	var i int
	for i = 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			i++
			break
		}
		if !strings.HasPrefix(a, "-") {
			break
		}
	}
	if err := fs.Parse(args[:i]); err != nil {
		return cfg, nil, err
	}

	cfg.BaseURL = strings.TrimRight(*base, "/")
	cfg.AgentID = strings.TrimSpace(*agent)
	cfg.IMEI = strings.TrimSpace(*imei)

	switch {
	case strings.TrimSpace(*token) != "":
		cfg.Token = stripSurroundingQuotes(strings.TrimSpace(*token))
	case strings.TrimSpace(*tokenFile) != "":
		raw, err := os.ReadFile(*tokenFile)
		if err != nil {
			return cfg, nil, fmt.Errorf("read --token-file: %w", err)
		}
		cfg.Token = stripSurroundingQuotes(strings.TrimSpace(string(raw)))
	default:
		cfg.Token = stripSurroundingQuotes(strings.TrimSpace(cfg.Token))
	}

	cfg.AgentID = stripSurroundingQuotes(cfg.AgentID)
	cfg.IMEI = stripSurroundingQuotes(cfg.IMEI)

	return cfg, args[i:], nil
}

func (c Config) requireAuth() error {
	if c.BaseURL == "" {
		return fmt.Errorf("--base-url / PAAS_BASE_URL is required")
	}
	if c.AgentID == "" {
		return fmt.Errorf("--agent-id / PAAS_AGENT_ID is required")
	}
	if c.Token == "" {
		return fmt.Errorf("--token / PAAS_ACCESS_TOKEN (or --token-file) is required")
	}
	return nil
}

func (c Config) requireIMEI(override string) (string, error) {
	imei := strings.TrimSpace(override)
	if imei == "" {
		imei = c.IMEI
	}
	if imei == "" {
		return "", fmt.Errorf("--imei / PAAS_IMEI is required")
	}
	if len(imei) != 15 {
		return "", fmt.Errorf("imei must be 15 digits, got %d", len(imei))
	}
	for _, r := range imei {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("imei must be digits only")
		}
	}
	return imei, nil
}
