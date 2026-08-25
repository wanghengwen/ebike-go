package config

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// remoteConfigShape captures keys loaded from Nacos ebike-service-client*.yml.
// Java may bind mapServiceConfig under xyy or xiaoantech; both are merged here.
type remoteConfigShape struct {
	MapServiceConfig mapServiceFields `yaml:"mapServiceConfig"`
	Xyy              struct {
		MapServiceConfig mapServiceFields `yaml:"mapServiceConfig"`
		GatewayFilter    struct {
			Exclude string `yaml:"exclude"`
		} `yaml:"gatewayfilter"`
	} `yaml:"xyy"`
	Xiaoantech struct {
		MapServiceConfig mapServiceFields `yaml:"mapServiceConfig"`
	} `yaml:"xiaoantech"`
	Gaode struct {
		Navigate struct {
			Url string `yaml:"url"`
		} `yaml:"navigate"`
	} `yaml:"gaode"`
}

type mapServiceFields struct {
	Url    string `yaml:"url"`
	Secret string `yaml:"secret"`
}

// MergeRemoteYAML overlays Nacos YAML onto GlobalConfig (local file is the base).
func MergeRemoteYAML(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	var remote remoteConfigShape
	if err := yaml.Unmarshal([]byte(content), &remote); err != nil {
		log.Printf("[WARN] failed to parse Nacos config YAML: %v", err)
		return
	}

	if remote.Xyy.GatewayFilter.Exclude != "" {
		GlobalConfig.Xyy.GatewayFilter.Exclude = remote.Xyy.GatewayFilter.Exclude
	}

	url := firstNonEmpty(
		remote.MapServiceConfig.Url,
		remote.Xyy.MapServiceConfig.Url,
		remote.Xiaoantech.MapServiceConfig.Url,
	)
	secret := firstNonEmpty(
		remote.MapServiceConfig.Secret,
		remote.Xyy.MapServiceConfig.Secret,
		remote.Xiaoantech.MapServiceConfig.Secret,
	)
	if url != "" {
		if strings.Contains(url, "${") {
			log.Printf("[WARN] mapServiceConfig.url looks like unresolved placeholder: %q — set XYY_MAP_SERVICE_URL or fix Nacos value", url)
		} else {
			GlobalConfig.Xyy.MapServiceConfig.Url = url
		}
	}
	if secret != "" {
		GlobalConfig.Xyy.MapServiceConfig.Secret = secret
	}
	if remote.Gaode.Navigate.Url != "" {
		GlobalConfig.Gaode.Navigate.Url = remote.Gaode.Navigate.Url
	}
}

// ApplyEnvOverrides lets operators inject map-service settings without editing Nacos.
func ApplyEnvOverrides() {
	if v := firstNonEmpty(os.Getenv("XYY_MAP_SERVICE_URL"), os.Getenv("MAP_SERVICE_URL")); v != "" {
		GlobalConfig.Xyy.MapServiceConfig.Url = v
	}
	if v := firstNonEmpty(os.Getenv("XYY_MAP_SERVICE_SECRET"), os.Getenv("MAP_SERVICE_SECRET")); v != "" {
		GlobalConfig.Xyy.MapServiceConfig.Secret = v
	}
	if v := os.Getenv("GAODE_NAVIGATE_URL"); v != "" {
		GlobalConfig.Gaode.Navigate.Url = v
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
