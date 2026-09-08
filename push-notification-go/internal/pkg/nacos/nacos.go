package nacos

import (
	"log"
	"net"
	"push-notification-go/internal/pkg/config"
	"push-notification-go/internal/pkg/mysql"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const (
	mysqlDataID = "mysql.yaml"
	mysqlGroup  = "xyy_ops"

	appDataID = "push-notification.yml"
)

var (
	configClient config_client.IConfigClient
	namingClient naming_client.INamingClient
)

// InitConfigClient creates the Nacos config client, fetches remote configs,
// merges them into GlobalConfig, applies env overrides, and registers listeners
// for live-reload.
func InitConfigClient() {
	// Read env vars first to get correct Nacos server address
	config.ApplyEnvOverrides()

	nc := config.GlobalConfig.Nacos

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nc.ServerAddr, nc.Port),
	}
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(nc.Namespace),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)

	var err error
	configClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		log.Fatalf("[nacos] failed to create config client: %v", err)
	}
	log.Printf("[nacos] config client created, addr=%s:%d ns=%s", nc.ServerAddr, nc.Port, nc.Namespace)

	// --- Fetch mysql.yaml ---
	mysqlContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: mysqlDataID,
		Group:  mysqlGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", mysqlDataID, err)
	} else {
		config.MergeNacosMySQLConfig(mysqlContent)
	}

	// --- Fetch push-notification.yml ---
	appGroup := nc.Group
	appContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  appGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", appDataID, err)
	} else {
		config.MergeNacosAppConfig(appContent)
	}

	// Apply env overrides as the final layer.
	config.ApplyEnvOverrides()

	// --- Register listeners for live-reload ---
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: mysqlDataID,
		Group:  mysqlGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] %s changed, re-merging", mysqlDataID)
			config.MergeNacosMySQLConfig(data)
			config.ApplyEnvOverrides()
			mysql.Reinit()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", mysqlDataID, err)
	}

	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  appGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] %s changed, re-merging", appDataID)
			config.MergeNacosAppConfig(data)
			config.ApplyEnvOverrides()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", appDataID, err)
	}
}

// InitNamingClient creates the Nacos naming client and registers this service
// instance. Registration is skipped when DryRun is true.
func InitNamingClient() {
	nc := config.GlobalConfig.Nacos

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nc.ServerAddr, nc.Port),
	}
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(nc.Namespace),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)

	var err error
	namingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		log.Fatalf("[nacos] failed to create naming client: %v", err)
	}

	if config.GlobalConfig.DryRun {
		log.Printf("[nacos] dry-run mode, skipping service registration")
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
	} else {
		log.Printf("[nacos] registered instance %s:%d service=%s group=%s",
			ip, port, config.GlobalConfig.Server.Name, nc.Group)
	}
}

// Deregister removes this service instance from Nacos. It is a no-op when
// DryRun is true or the naming client was never created.
func Deregister() {
	if namingClient == nil || config.GlobalConfig.DryRun {
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
	} else {
		log.Printf("[nacos] deregistered instance %s:%d", ip, port)
	}
}

// GetLocalIP returns the preferred outbound IP of this machine.
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("[nacos] failed to detect local IP: %v, falling back to 127.0.0.1", err)
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
