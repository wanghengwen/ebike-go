package fastid

import (
	"os"
	"path/filepath"
	"strings"
)

// Config mirrors com.xyy.fastid.core.config.FastIdConfig.
type Config struct {
	URL                      string `mapstructure:"url"`
	Namespace                string `mapstructure:"namespace"`
	GroupID                  string `mapstructure:"group_id"`
	AppName                  string `mapstructure:"app_name"`
	Port                     int    `mapstructure:"port"`
	Secret                   string `mapstructure:"secret"`
	InstanceNoBits           int64  `mapstructure:"instance_no_bits"`
	SequenceBits             int64  `mapstructure:"sequence_bits"`
	DriftTime                int    `mapstructure:"drift_time"`
	InstanceNoLocalDirectory string `mapstructure:"instance_no_local_directory"`
	MachineUUID              string `mapstructure:"machine_uuid"`
	UseHTTPS                 bool   `mapstructure:"use_https"`
	UseHTTPSExplicit         bool   `mapstructure:"use_https_explicit"`
}

func (c *Config) applyDefaults() {
	if c.InstanceNoBits == 0 {
		c.InstanceNoBits = 12
	}
	if c.SequenceBits == 0 {
		c.SequenceBits = 19
	}
	if c.DriftTime == 0 {
		c.DriftTime = 10
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.InstanceNoLocalDirectory == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = "."
		}
		c.InstanceNoLocalDirectory = filepath.Join(home, "fastid")
	}
	c.normalizeURLScheme()
}

func (c *Config) normalizeURLScheme() {
	url := strings.TrimSpace(c.URL)
	switch {
	case strings.HasPrefix(url, "http://"):
		c.URL = strings.TrimPrefix(url, "http://")
		c.UseHTTPS = false
	case strings.HasPrefix(url, "https://"):
		c.URL = strings.TrimPrefix(url, "https://")
		c.UseHTTPS = true
	default:
		c.URL = url
		if !c.UseHTTPSExplicit {
			// Bare host defaults to HTTPS for external Java-compatible deployments.
			c.UseHTTPS = true
		}
	}
}
