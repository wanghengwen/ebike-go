package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

func runQuery(cfg Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: paas-test query <allDevices|deviceInfo|gps|address|battery|bms|currentBms|realtime>")
	}
	sub, args := args[0], args[1:]
	if err := cfg.requireAuth(); err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "allDevices":
		return queryAllDevices(c, args)
	case "deviceInfo":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.get("/ebike/v1/deviceInfo", url.Values{"imei": {imei}})
		})
	case "address":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.get("/ebike/v1/address", url.Values{"imei": {imei}})
		})
	case "battery":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.get("/ebike/v1/batteryInfo", url.Values{"imei": {imei}})
		})
	case "bms":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.get("/ebike/v1/bmsInfo", url.Values{"imei": {imei}})
		})
	case "currentBms":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.get("/ebike/v1/currentBmsInfo", url.Values{"imei": {imei}})
		})
	case "realtime":
		return queryWithIMEI(cfg, c, args, func(imei string) (*Envelope, int, error) {
			return c.post("/ebike/api/device", map[string]interface{}{"imei": imei})
		})
	case "gps":
		return queryGPS(cfg, c, args)
	default:
		return fmt.Errorf("unknown query %q", sub)
	}
}

func queryAllDevices(c *Client, args []string) error {
	fs := flag.NewFlagSet("query allDevices", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	pageSize := fs.Int("pageSize", 20, "page size")
	pageNumber := fs.Int("pageNumber", 1, "page number")
	if err := fs.Parse(args); err != nil {
		return err
	}
	env, status, err := c.get("/ebike/v1/allDevices", url.Values{
		"pageSize":   {strconv.Itoa(*pageSize)},
		"pageNumber": {strconv.Itoa(*pageNumber)},
	})
	if err != nil {
		return err
	}
	printEnvelope("allDevices", env, status)
	if !env.Success {
		return fmt.Errorf("allDevices failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

type imeiQuery func(imei string) (*Envelope, int, error)

func queryWithIMEI(cfg Config, c *Client, args []string, fn imeiQuery) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	if err := fs.Parse(args); err != nil {
		return err
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	env, status, err := fn(imei)
	if err != nil {
		return err
	}
	printEnvelope("query", env, status)
	if !env.Success {
		return fmt.Errorf("query failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}

func queryGPS(cfg Config, c *Client, args []string) error {
	fs := flag.NewFlagSet("query gps", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	imeiFlag := fs.String("imei", cfg.IMEI, "device IMEI")
	from := fs.Int64("from", 0, "start unix seconds (default: now-1h)")
	to := fs.Int64("to", 0, "end unix seconds (default: now)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	imei, err := cfg.requireIMEI(*imeiFlag)
	if err != nil {
		return err
	}
	end := *to
	start := *from
	now := time.Now().Unix()
	if end <= 0 {
		end = now
	}
	if start <= 0 {
		start = end - 3600
	}
	env, status, err := c.get("/ebike/v1/GPSPoints", url.Values{
		"imei":      {imei},
		"startTime": {strconv.FormatInt(start, 10)},
		"endTime":   {strconv.FormatInt(end, 10)},
	})
	if err != nil {
		return err
	}
	printEnvelope("GPSPoints", env, status)
	if !env.Success {
		return fmt.Errorf("GPSPoints failed: %s (%s)", env.errorType(), env.promot())
	}
	return nil
}
