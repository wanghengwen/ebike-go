// Command paas-test is a CLI for exercising the Xiaoan-compatible open PaaS
// gateway (ebike-open-paas-go) as a third-party client would.
//
// Credentials are agentId + xc-access-token (not tenantId). Callback listen
// mode binds a local HTTP port; register the publicly reachable HTTPS URL that
// reverse-proxies to it.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return nil
	}

	cfg, rest, err := parseGlobal(args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		printUsage()
		return fmt.Errorf("missing command")
	}

	cmd, rest := rest[0], rest[1:]
	switch cmd {
	case "smoke":
		return runSmoke(cfg, rest)
	case "query":
		return runQuery(cfg, rest)
	case "cmd":
		return runCmd(cfg, rest)
	case "callback":
		return runCallback(cfg, rest)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q (try paas-test --help)", cmd)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `%s — Xiaoan PaaS open-platform test client

Usage:
  paas-test [global flags] <command> [command flags]

Global flags / env:
  --base-url     PAAS_BASE_URL      default https://paas.luopingtech.com
  --agent-id     PAAS_AGENT_ID      required for API calls
  --token        PAAS_ACCESS_TOKEN  xc-access-token (prefer env over argv)
  --token-file   path to a file containing the token
  --imei         PAAS_IMEI          default IMEI for device-scoped calls

Commands:
  smoke                         auth + read-only contract checks
  query <allDevices|deviceInfo|gps|address|battery|bms|currentBms|realtime>
  cmd   <lock|acc|defend|backWheel|batteryCompartment|reboot|mc|deviceVoice|bluetooth> --yes
  callback register  --event N --url https://...
  callback list      --event N
  callback unregister --event N --url https://...
  callback listen    --listen :8089 --public-url https://... [--events 1,2,3,4]

Examples:
  set PAAS_AGENT_ID=87
  set PAAS_ACCESS_TOKEN=...
  paas-test smoke
  paas-test query deviceInfo --imei 865067022403441
  paas-test callback listen --listen :8089 --public-url https://hook.example.com/paas --events 1,2,3,4
`, os.Args[0])
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return stripSurroundingQuotes(v)
	}
	return fallback
}

// stripSurroundingQuotes removes the quotes Windows cmd leaves behind when
// someone runs: set PAAS_ACCESS_TOKEN="secret"  (the " characters become part
// of the value).
func stripSurroundingQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
