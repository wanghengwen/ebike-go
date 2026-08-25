package fastid

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func fetchMachineID(cfg Config, machineUUID string) (int64, error) {
	scheme := "http"
	if cfg.UseHTTPS {
		scheme = "https"
	}
	endpoint := fmt.Sprintf("%s://%s/fastid/machineId", scheme, strings.TrimLeft(cfg.URL, "/"))

	form := url.Values{}
	form.Set("namespace", cfg.Namespace)
	form.Set("groupId", cfg.GroupID)
	form.Set("appName", cfg.AppName)
	form.Set("machineUuid", machineUUID)
	form.Set("secret", cfg.Secret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fastid register http %d: %s", resp.StatusCode, string(body))
	}
	parsed := parseParam(string(body))
	if parsed["success"] != "true" {
		return 0, fmt.Errorf("fastid register failed: %s", parsed["msg"])
	}
	return strconv.ParseInt(parsed["result"], 10, 64)
}

func parseParam(param string) map[string]string {
	result := map[string]string{}
	if param == "" {
		return result
	}
	for _, kv := range strings.Split(param, "&") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}

func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			return ip.String()
		}
	}
	return ""
}
