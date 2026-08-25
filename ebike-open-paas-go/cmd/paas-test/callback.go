package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

func runCallback(cfg Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: paas-test callback <register|list|unregister|listen>")
	}
	sub, args := args[0], args[1:]
	switch sub {
	case "register":
		return callbackRegister(cfg, args)
	case "list":
		return callbackList(cfg, args)
	case "unregister", "delete":
		return callbackUnregister(cfg, args)
	case "listen":
		return callbackListen(cfg, args)
	default:
		return fmt.Errorf("unknown callback command %q", sub)
	}
}

func callbackRegister(cfg Config, args []string) error {
	if err := cfg.requireAuth(); err != nil {
		return err
	}
	fs := flag.NewFlagSet("callback register", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	event := fs.Int("event", 0, "event type 1..4")
	cbURL := fs.String("url", envOr("PAAS_CALLBACK_URL", ""), "public https callback URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *event < 1 || *event > 4 {
		return fmt.Errorf("--event must be 1..4 (5/UART is not deliverable)")
	}
	if err := mustHTTPS(*cbURL); err != nil {
		return err
	}
	c := newClient(cfg)
	env, status, err := c.post("/ebike/v1/callback", map[string]interface{}{
		"event": *event,
		"url":   *cbURL,
	})
	if err != nil {
		return err
	}
	printEnvelope("callback register", env, status)
	if !env.Success {
		return fmt.Errorf("register failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

func callbackList(cfg Config, args []string) error {
	if err := cfg.requireAuth(); err != nil {
		return err
	}
	fs := flag.NewFlagSet("callback list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	event := fs.Int("event", 0, "event type 1..5")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *event < 1 || *event > 5 {
		return fmt.Errorf("--event must be 1..5")
	}
	c := newClient(cfg)
	env, status, err := c.get("/ebike/v1/callback", url.Values{"event": {strconv.Itoa(*event)}})
	if err != nil {
		return err
	}
	printEnvelope("callback list", env, status)
	if !env.Success {
		return fmt.Errorf("list failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

func callbackUnregister(cfg Config, args []string) error {
	if err := cfg.requireAuth(); err != nil {
		return err
	}
	fs := flag.NewFlagSet("callback unregister", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	event := fs.Int("event", 0, "event type 1..5")
	cbURL := fs.String("url", envOr("PAAS_CALLBACK_URL", ""), "callback URL to remove")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *event < 1 || *event > 5 {
		return fmt.Errorf("--event must be 1..5")
	}
	if *cbURL == "" {
		return fmt.Errorf("--url / PAAS_CALLBACK_URL is required")
	}
	c := newClient(cfg)
	env, status, err := c.delete("/ebike/v1/callback", url.Values{
		"event": {strconv.Itoa(*event)},
		"url":   {*cbURL},
	})
	if err != nil {
		return err
	}
	printEnvelope("callback unregister", env, status)
	if !env.Success {
		return fmt.Errorf("unregister failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}
