package config

import (
	"fmt"
	"log"
)

// LoadExtensionConfigsFromNacos loads fast_id.yaml, redis.yaml and kafka.yaml from {group}_ops.
func LoadExtensionConfigsFromNacos() error {
	if GlobalConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	// Nacos connection settings (addr/port/namespace/group) must come from env
	// before the client is created; FinalizeConfig runs too late for this.
	applyNacosEnv(GlobalConfig)
	if GlobalConfig.Nacos.ServerAddr == "" {
		return fmt.Errorf("nacos.serverAddr is empty")
	}
	port := GlobalConfig.Nacos.Port
	if port == 0 {
		port = 8848
	}
	log.Printf("[nacos] connecting to %s:%d namespace=%s group=%s",
		GlobalConfig.Nacos.ServerAddr, port, GlobalConfig.Nacos.Namespace, GlobalConfig.Nacos.Group)

	client, err := InitNacosConfigClient()
	if err != nil {
		return err
	}

	if err := applyMainFromNacos(client); err != nil {
		log.Printf("[WARN] main config: %v", err)
	}

	if GlobalConfig.FastIDEnabled() {
		if err := applyFastIDFromNacos(client); err != nil {
			if GlobalConfig.Spring.Xyy.Fastid.Secret == "" {
				FinalizeConfig(GlobalConfig)
				if GlobalConfig.Spring.Xyy.Fastid.Secret == "" {
					return fmt.Errorf("fast_id.yaml: %w", err)
				}
			}
			log.Printf("[WARN] fast_id.yaml: %v, using secret from env/local config", err)
		}
	} else {
		log.Println("[nacos] fastid disabled, skip fast_id.yaml")
	}

	if err := applyRedisFromNacos(client); err != nil {
		log.Printf("[WARN] redis.yaml: %v", err)
	}

	if err := applyKafkaFromNacos(client); err != nil {
		log.Printf("[WARN] kafka.yaml: %v", err)
	}

	FinalizeConfig(GlobalConfig)
	LogEffectiveConfigSource()
	return nil
}
