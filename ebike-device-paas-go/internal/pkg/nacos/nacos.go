// Package nacos wires the Nacos config + naming clients, modeled on
// identity-auth-go. It fetches remote ops/app configs, merges them into
// config.GlobalConfig with live-reload listeners, and registers this service
// instance for discovery (skipped in DryRun / when registration is disabled).
package nacos

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"ebike-device-paas-go/internal/pkg/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const (
	redisDataID = "redis.yaml"
	kafkaDataID = "kafka.yaml"
)

var (
	configClient config_client.IConfigClient
	namingClient naming_client.INamingClient
)

func clientParams() vo.NacosClientParam {
	nc := config.GlobalConfig.Nacos
	port := nc.Port
	if port == 0 {
		port = 8848
	}
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nc.ServerAddr, port),
	}
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(nc.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(filepath.Join(os.TempDir(), "nacos", "log")),
		constant.WithCacheDir(filepath.Join(os.TempDir(), "nacos", "cache")),
		constant.WithLogLevel("warn"),
	)
	return vo.NacosClientParam{ClientConfig: cc, ServerConfigs: sc}
}

// InitConfigClient creates the config client, fetches remote configs, merges
// them, applies env overrides and registers live-reload listeners. Individual
// fetch failures are logged (non-fatal) so the service can still start.
func InitConfigClient() {
	config.ApplyEnvOverrides() // resolve Nacos addr from env first

	nc := config.GlobalConfig.Nacos
	var err error
	configClient, err = clients.NewConfigClient(clientParams())
	if err != nil {
		log.Printf("[nacos] failed to create config client: %v (continuing with local config)", err)
		config.ApplyEnvOverrides()
		return
	}
	log.Printf("[nacos] config client created, addr=%s:%d ns=%s group=%s",
		nc.ServerAddr, nc.Port, nc.Namespace, nc.Group)

	// redis.yaml @ {group}_ops
	if content, err := configClient.GetConfig(vo.ConfigParam{DataId: redisDataID, Group: config.OpsGroup()}); err != nil {
		log.Printf("[nacos] get %s @ %s failed: %v", redisDataID, config.OpsGroup(), err)
	} else {
		config.MergeNacosRedisConfig(content)
	}

	// kafka.yaml @ {group}_ops (C34 producer bootstrap servers)
	if content, err := configClient.GetConfig(vo.ConfigParam{DataId: kafkaDataID, Group: config.OpsGroup()}); err != nil {
		log.Printf("[nacos] get %s @ %s failed: %v", kafkaDataID, config.OpsGroup(), err)
	} else {
		config.MergeNacosKafkaConfig(content)
	}

	// ebike-device-paas.yml @ {group}
	appDataID := config.AppDataID()
	if content, err := configClient.GetConfig(vo.ConfigParam{DataId: appDataID, Group: nc.Group}); err != nil {
		log.Printf("[nacos] get %s @ %s failed: %v", appDataID, nc.Group, err)
	} else {
		config.MergeNacosAppConfig(content)
	}

	// Env overrides are the final layer.
	config.ApplyEnvOverrides()

	// Live-reload listeners.
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: redisDataID,
		Group:  config.OpsGroup(),
		OnChange: func(_, _, _, data string) {
			log.Printf("[nacos] %s changed, re-merging", redisDataID)
			config.MergeNacosRedisConfig(data)
			config.ApplyEnvOverrides()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", redisDataID, err)
	}
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: kafkaDataID,
		Group:  config.OpsGroup(),
		OnChange: func(_, _, _, data string) {
			log.Printf("[nacos] %s changed, re-merging", kafkaDataID)
			config.MergeNacosKafkaConfig(data)
			config.ApplyEnvOverrides()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", kafkaDataID, err)
	}
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  nc.Group,
		OnChange: func(_, _, _, data string) {
			log.Printf("[nacos] %s changed, re-merging", appDataID)
			config.MergeNacosAppConfig(data)
			config.ApplyEnvOverrides()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", appDataID, err)
	}
}

// InitNamingClient creates the naming client and registers this instance.
// Registration is skipped in DryRun mode or when RegisterEnabled is false.
func InitNamingClient() {
	nc := config.GlobalConfig.Nacos
	var err error
	namingClient, err = clients.NewNamingClient(clientParams())
	if err != nil {
		log.Printf("[nacos] failed to create naming client: %v", err)
		return
	}

	if !nc.RegisterEnabled {
		log.Printf("[nacos] registerEnabled=false, skipping service registration")
		return
	}

	ip := GetLocalIP()
	port := uint64(config.GlobalConfig.Server.Port)
	ok, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: config.GlobalConfig.Server.Name,
		GroupName:   nc.Group,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})
	if err != nil || !ok {
		log.Printf("[nacos] failed to register instance: %v", err)
		return
	}
	log.Printf("[nacos] registered instance %s:%d service=%s group=%s",
		ip, port, config.GlobalConfig.Server.Name, nc.Group)
}

// Deregister removes this instance from Nacos. No-op in DryRun or when the
// naming client was never created.
func Deregister() {
	if namingClient == nil || !config.GlobalConfig.Nacos.RegisterEnabled {
		return
	}
	nc := config.GlobalConfig.Nacos
	ip := GetLocalIP()
	port := uint64(config.GlobalConfig.Server.Port)
	ok, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: config.GlobalConfig.Server.Name,
		GroupName:   nc.Group,
		Ephemeral:   true,
	})
	if err != nil || !ok {
		log.Printf("[nacos] failed to deregister instance: %v", err)
		return
	}
	log.Printf("[nacos] deregistered instance %s:%d", ip, port)
}

// GetLocalIP returns the preferred outbound IP of this machine.
func GetLocalIP() string {
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("[nacos] failed to detect local IP: %v, falling back to 127.0.0.1", err)
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
