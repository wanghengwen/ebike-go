package main

import (
	"flag"
	"fmt"
	"os"
)

func runCmd(cfg Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: paas-test cmd <lock|acc|defend|backWheel|batteryCompartment|reboot|mc|deviceVoice|bluetooth> --yes")
	}
	sub, args := args[0], args[1:]
	if err := cfg.requireAuth(); err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "lock":
		return cmdBool(cfg, c, args, "/ebike/v1/lock", "locked")
	case "acc":
		return cmdBool(cfg, c, args, "/ebike/v1/acc", "acc")
	case "defend":
		return cmdBool(cfg, c, args, "/ebike/v1/defend", "defend")
	case "backWheel":
		return cmdBool(cfg, c, args, "/ebike/v1/backWheel", "locked")
	case "batteryCompartment":
		return cmdBool(cfg, c, args, "/ebike/v1/batteryCompartment", "locked")
	case "reboot":
		return cmdSimple(cfg, c, args, "/ebike/v1/reboot", nil)
	case "mc":
		return cmdMC(cfg, c, args)
	case "deviceVoice":
		return cmdVoice(cfg, c, args)
	case "bluetooth":
		return cmdBluetooth(cfg, c, args)
	default:
		return fmt.Errorf("unknown cmd %q", sub)
	}
}

func cmdBool(cfg Config, c *Client, args []string, path, field string) error {
	fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	val := fs.Int(field, -1, field+" 0|1")
	yes := fs.Bool("yes", false, "required to send a real device command")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		return fmt.Errorf("refusing to send %s without --yes (this moves a real device)", path)
	}
	if *val != 0 && *val != 1 {
		return fmt.Errorf("--%s must be 0 or 1", field)
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	env, status, err := c.post(path, map[string]interface{}{"imei": imei, field: *val})
	if err != nil {
		return err
	}
	printEnvelope(path, env, status)
	if !env.Success {
		return fmt.Errorf("%s failed: %s (%s)", path, env.errorType(), env.promot())
	}
	return nil
}

func cmdSimple(cfg Config, c *Client, args []string, path string, extra map[string]interface{}) error {
	fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	yes := fs.Bool("yes", false, "required to send a real device command")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		return fmt.Errorf("refusing to send %s without --yes", path)
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	body := map[string]interface{}{"imei": imei}
	for k, v := range extra {
		body[k] = v
	}
	env, status, err := c.post(path, body)
	if err != nil {
		return err
	}
	printEnvelope(path, env, status)
	if !env.Success {
		return fmt.Errorf("%s failed: %s (%s)", path, env.errorType(), env.promot())
	}
	return nil
}

func cmdMC(cfg Config, c *Client, args []string) error {
	fs := flag.NewFlagSet("cmd mc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	speed := fs.Int("speed", -1, "limit speed percent 0..100")
	yes := fs.Bool("yes", false, "required to send a real device command")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		return fmt.Errorf("refusing to send /ebike/v1/mc without --yes")
	}
	if *speed < 0 || *speed > 100 {
		return fmt.Errorf("--speed must be 0..100")
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	env, status, err := c.post("/ebike/v1/mc", map[string]interface{}{"imei": imei, "speed": *speed})
	if err != nil {
		return err
	}
	printEnvelope("/ebike/v1/mc", env, status)
	if !env.Success {
		return fmt.Errorf("mc failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

func cmdVoice(cfg Config, c *Client, args []string) error {
	fs := flag.NewFlagSet("cmd deviceVoice", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	idx := fs.Int("idx", -1, "Xiaoan ringtone slot")
	yes := fs.Bool("yes", false, "required to send a real device command")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		return fmt.Errorf("refusing to send /ebike/v1/deviceVoice without --yes")
	}
	if *idx < 0 {
		return fmt.Errorf("--idx is required")
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	env, status, err := c.post("/ebike/v1/deviceVoice", map[string]interface{}{"imei": imei, "idx": *idx})
	if err != nil {
		return err
	}
	printEnvelope("/ebike/v1/deviceVoice", env, status)
	if !env.Success {
		return fmt.Errorf("deviceVoice failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

func cmdBluetooth(cfg Config, c *Client, args []string) error {
	fs := flag.NewFlagSet("cmd bluetooth", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	yes := fs.Bool("yes", false, "required to send a real device command")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*yes {
		return fmt.Errorf("refusing to send /ebike/v1/bluetooth without --yes")
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	env, status, err := c.post("/ebike/v1/bluetooth", map[string]interface{}{"imei": imei})
	if err != nil {
		return err
	}
	printEnvelope("/ebike/v1/bluetooth", env, status)
	if !env.Success {
		return fmt.Errorf("bluetooth failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}
