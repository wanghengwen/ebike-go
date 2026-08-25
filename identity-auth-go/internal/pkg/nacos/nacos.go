package nacos

import (
	"log"
	"net"
	"identity-auth-go/internal/pkg/config"
	"identity-auth-go/internal/pkg/mysql"
	"identity-auth-go/internal/pkg/oss"
	goredis "identity-auth-go/internal/pkg/redis"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const (
	mysqlDataID = "mysql.yaml"
	redisDataID = "redis.yaml"
	ossDataID   = "aliyun_oss.yaml"
	opsGroup    = "xyy_ops"

	appDataID = "identity-auth.yml"
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
		Group:  opsGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", mysqlDataID, err)
	} else {
		config.MergeNacosMySQLConfig(mysqlContent)
	}

	// --- Fetch redis.yaml ---
	redisContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: redisDataID,
		Group:  opsGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", redisDataID, err)
	} else {
		config.MergeNacosRedisConfig(redisContent)
	}

	// --- Fetch aliyun_oss.yaml ---
	ossContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: ossDataID,
		Group:  opsGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", ossDataID, err)
	} else {
		config.MergeNacosIdentityAuthConfig(ossContent)
	}

	// --- Fetch identity-auth.yml ---
	appGroup := nc.Group
	appContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  appGroup,
	})
	if err != nil {
		log.Printf("[nacos] failed to get %s: %v", appDataID, err)
	} else {
		config.MergeNacosIdentityAuthConfig(appContent)
	}

	// Apply env overrides as the final layer.
	config.ApplyEnvOverrides()

	// --- Register listeners for live-reload ---
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: mysqlDataID,
		Group:  opsGroup,
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
		DataId: redisDataID,
		Group:  opsGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] %s changed, re-merging", redisDataID)
			config.MergeNacosRedisConfig(data)
			config.ApplyEnvOverrides()
			goredis.Reinit()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", redisDataID, err)
	}

	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: ossDataID,
		Group:  opsGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] %s changed, re-merging", ossDataID)
			config.MergeNacosIdentityAuthConfig(data)
			config.ApplyEnvOverrides()
			oss.Init()
		},
	}); err != nil {
		log.Printf("[nacos] failed to listen %s: %v", ossDataID, err)
	}

	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  appGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] %s changed, re-merging", appDataID)
			config.MergeNacosIdentityAuthConfig(data)
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
		GroupName:    nc.Group,
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
		GroupName:    nc.Group,
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
