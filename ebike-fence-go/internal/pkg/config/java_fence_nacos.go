package config

import (
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

// Java Spring Cloud Nacos uses prefix + file-extension => ebike-fence.yml (not .yaml).
var JavaFenceNacosDataIDs = []string{"ebike-fence.yml", "ebike-fence.yaml"}

// JavaFenceNacosDataID is kept for callers that need a primary id.
const JavaFenceNacosDataID = "ebike-fence.yml"

type javaFenceYAML struct {
	Xyy XyyConfig `yaml:"xyy"`
}

// mergeXyyIncoming overlays non-empty fields from ebike-fence-go.yaml onto GlobalConfig.
func mergeXyyIncoming(dst *XyyConfig, src XyyConfig) {
	if src.RemoteLockDistance != nil {
		dst.RemoteLockDistance = src.RemoteLockDistance
	}
	if len(src.HelmetAuditWhite) > 0 {
		dst.HelmetAuditWhite = src.HelmetAuditWhite
	}
	if len(src.DirectionIgnoreTenantIds) > 0 {
		dst.DirectionIgnoreTenantIds = src.DirectionIgnoreTenantIds
	}
	if len(src.CancelAuthTenantIds) > 0 {
		dst.CancelAuthTenantIds = src.CancelAuthTenantIds
	}
}

// fillXyyFromJavaFence fills missing xyy fields from Java ebike-fence.yml.
func fillXyyFromJavaFence(dst *XyyConfig, src XyyConfig) {
	if dst.RemoteLockDistance == nil && src.RemoteLockDistance != nil {
		dst.RemoteLockDistance = src.RemoteLockDistance
	}
	if len(dst.HelmetAuditWhite) == 0 && len(src.HelmetAuditWhite) > 0 {
		dst.HelmetAuditWhite = src.HelmetAuditWhite
	}
	if len(dst.DirectionIgnoreTenantIds) == 0 && len(src.DirectionIgnoreTenantIds) > 0 {
		dst.DirectionIgnoreTenantIds = src.DirectionIgnoreTenantIds
	}
	if len(dst.CancelAuthTenantIds) == 0 && len(src.CancelAuthTenantIds) > 0 {
		dst.CancelAuthTenantIds = src.CancelAuthTenantIds
	}
}

// ApplyJavaFenceYAML merges xyy.* from Java ebike-fence.yml without overwriting values already set.
func ApplyJavaFenceYAML(content string) error {
	var incoming javaFenceYAML
	if err := yaml.Unmarshal([]byte(content), &incoming); err != nil {
		return err
	}
	fillXyyFromJavaFence(&GlobalConfig.Xyy, incoming.Xyy)
	return nil
}

func fetchJavaFenceContent(client config_client.IConfigClient, group string) (content, dataID string, err error) {
	for _, id := range JavaFenceNacosDataIDs {
		content, err = client.GetConfig(vo.ConfigParam{
			DataId: id,
			Group:  group,
		})
		if err != nil {
			log.Printf("[WARN] Failed to load %s from Nacos (group=%s): %v", id, group, err)
			continue
		}
		if content != "" {
			return content, id, nil
		}
		log.Printf("[WARN] %s is empty in Nacos group %s", id, group)
	}
	return "", "", nil
}

// LogXyyConfig prints merged xyy settings after local + Nacos bootstrap.
func LogXyyConfig(stage string) {
	log.Printf("[INFO] xyy config (%s): namespace=%s group=%s cancelAuthTenantIds=%v",
		stage, GlobalConfig.Nacos.Namespace, GlobalConfig.Nacos.Group, GlobalConfig.Xyy.CancelAuthTenantIds)
}

// MergeJavaFenceFromNacos loads Java ebike-fence.yml and back-fills xyy fields missing locally.
func MergeJavaFenceFromNacos(client config_client.IConfigClient, group string) {
	if client == nil {
		return
	}
	before := len(GlobalConfig.Xyy.CancelAuthTenantIds)
	content, dataID, err := fetchJavaFenceContent(client, group)
	if err != nil {
		return
	}
	if content == "" {
		log.Printf("[WARN] Java fence config not found in Nacos group %s (tried %v)", group, JavaFenceNacosDataIDs)
		return
	}
	if err := ApplyJavaFenceYAML(content); err != nil {
		log.Printf("[WARN] Failed to parse %s: %v", dataID, err)
		return
	}
	log.Printf("Loaded Java fence config from Nacos %s (group=%s)", dataID, group)
	if before == 0 && len(GlobalConfig.Xyy.CancelAuthTenantIds) > 0 {
		log.Printf("Loaded cancelAuthTenantIds from %s: %v", dataID, GlobalConfig.Xyy.CancelAuthTenantIds)
	}
}

// ListenJavaFenceFromNacos hot-reloads xyy.* when Java ebike-fence.yml changes.
func ListenJavaFenceFromNacos(client config_client.IConfigClient, group string) {
	if client == nil {
		return
	}
	for _, dataID := range JavaFenceNacosDataIDs {
		id := dataID
		if err := client.ListenConfig(vo.ConfigParam{
			DataId: id,
			Group:  group,
			OnChange: func(namespace, grp, changedID, data string) {
				if data == "" {
					return
				}
				if err := ApplyJavaFenceYAML(data); err != nil {
					log.Printf("[WARN] Failed to apply %s change: %v", changedID, err)
					return
				}
				log.Printf("Reloaded Java fence config from Nacos %s (group=%s)", changedID, grp)
				LogXyyConfig("hot-reload")
			},
		}); err != nil {
			log.Printf("[WARN] Failed to listen Nacos %s: %v", id, err)
		}
	}
}
