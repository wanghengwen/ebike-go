package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

const fastIDNacosDataID = "fast_id.yaml"

type fastIDNacosFile struct {
	Spring struct {
		Xyy struct {
			Fastid FastIDSettings `yaml:"fastid"`
		} `yaml:"xyy"`
	} `yaml:"spring"`
}

// LoadFastIDFromNacos pulls spring.xyy.fastid (including secret) from Nacos fast_id.yaml.
func LoadFastIDFromNacos() error {
	return LoadExtensionConfigsFromNacos()
}

func applyFastIDFromNacos(client config_client.IConfigClient) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	group := OpsGroup()
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: fastIDNacosDataID,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("get %s from nacos group %s: %w", fastIDNacosDataID, group, err)
	}
	if content == "" {
		return fmt.Errorf("%s is empty in nacos group %s", fastIDNacosDataID, group)
	}
	fastid, err := parseFastIDNacosYAML(content)
	if err != nil {
		return err
	}
	mergeFastIDSettings(&GlobalConfig.Spring.Xyy.Fastid, fastid)
	log.Printf("[nacos] loaded fastid from %s (group=%s, secret=%t)", fastIDNacosDataID, group, fastid.Secret != "")
	return nil
}

func parseFastIDNacosYAML(content string) (FastIDSettings, error) {
	var raw fastIDNacosFile
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return FastIDSettings{}, fmt.Errorf("parse %s: %w", fastIDNacosDataID, err)
	}
	return raw.Spring.Xyy.Fastid, nil
}

func mergeFastIDSettings(dst *FastIDSettings, from FastIDSettings) {
	if from.Enabled {
		dst.Enabled = true
	}
	if from.ServerAddr != "" {
		dst.ServerAddr = from.ServerAddr
	}
	if from.ServerURL != "" {
		dst.ServerURL = from.ServerURL
	}
	if from.URL != "" {
		dst.URL = from.URL
	}
	if from.Namespace != "" {
		dst.Namespace = from.Namespace
	}
	if from.GroupID != "" {
		dst.GroupID = from.GroupID
	}
	if from.Secret != "" {
		dst.Secret = from.Secret
	}
	if from.AppName != "" && !isUnresolvedPlaceholder(from.AppName) {
		dst.AppName = from.AppName
	}
	if from.Port != 0 {
		dst.Port = from.Port
	}
	if from.InstanceNoLocalDirectory != "" {
		dst.InstanceNoLocalDirectory = from.InstanceNoLocalDirectory
	}
	if from.MachineUUID != "" {
		dst.MachineUUID = from.MachineUUID
	}
	if from.DriftTime != 0 {
		dst.DriftTime = from.DriftTime
	}
	if from.UseHTTPSExplicit {
		dst.UseHTTPS = from.UseHTTPS
		dst.UseHTTPSExplicit = true
	}
	applyFastIDSchemeFromAddr(dst)
}

func applyFastIDSchemeFromAddr(dst *FastIDSettings) {
	// Prefer server-url (Go-only), then server-addr, then legacy url.
	switch {
	case strings.HasPrefix(dst.ServerURL, "http://"):
		dst.ServerURL = strings.TrimPrefix(dst.ServerURL, "http://")
		dst.UseHTTPS = false
		dst.UseHTTPSExplicit = true
	case strings.HasPrefix(dst.ServerURL, "https://"):
		dst.ServerURL = strings.TrimPrefix(dst.ServerURL, "https://")
		dst.UseHTTPS = true
		dst.UseHTTPSExplicit = true
	case strings.HasPrefix(dst.ServerAddr, "http://"):
		dst.ServerAddr = strings.TrimPrefix(dst.ServerAddr, "http://")
		dst.UseHTTPS = false
		dst.UseHTTPSExplicit = true
	case strings.HasPrefix(dst.URL, "http://"):
		dst.URL = strings.TrimPrefix(dst.URL, "http://")
		dst.UseHTTPS = false
		dst.UseHTTPSExplicit = true
	case strings.HasPrefix(dst.ServerAddr, "https://"):
		dst.ServerAddr = strings.TrimPrefix(dst.ServerAddr, "https://")
		dst.UseHTTPS = true
		dst.UseHTTPSExplicit = true
	case strings.HasPrefix(dst.URL, "https://"):
		dst.URL = strings.TrimPrefix(dst.URL, "https://")
		dst.UseHTTPS = true
		dst.UseHTTPSExplicit = true
	}
}
