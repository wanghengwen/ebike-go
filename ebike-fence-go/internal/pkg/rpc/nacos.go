package rpc

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/pkg/config"
	"ebike-fence-go/internal/pkg/env"
	"ebike-fence-go/internal/pkg/mysql"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

// RegisteredToNacos is true after a successful RegisterInstance in InitNacos.
var RegisteredToNacos bool

var nacosConfigListenOnce sync.Once

func registerNacosConfigListeners(dataID, group string) {
	nacosConfigListenOnce.Do(func() {
		config.ListenJavaFenceFromNacos(ConfigClient, group)

		if err := config.ListenRedisFromNacos(ConfigClient, func() {
			if err := pkgredis.Reinit(); err != nil {
				log.Printf("[ERROR] Redis reinit after Nacos change failed: %v", err)
				return
			}
			log.Printf("Redis reconnected after Nacos redis.yaml change")
		}); err != nil {
			log.Printf("[WARN] Failed to listen Nacos redis.yaml: %v", err)
		}

		if err := config.ListenMySQLFromNacos(ConfigClient, func() {
			if err := mysql.Reinit(); err != nil {
				log.Printf("[ERROR] MySQL reinit after Nacos change failed: %v", err)
				return
			}
			persistence.Init()
			log.Printf("MySQL reconnected after Nacos mysql.yaml change")
		}); err != nil {
			log.Printf("[WARN] Failed to listen Nacos mysql.yaml: %v", err)
		}

		err := ConfigClient.ListenConfig(vo.ConfigParam{
			DataId: dataID,
			Group:  group,
			OnChange: func(namespace, grp, changedID, data string) {
				log.Printf("Nacos config changed: namespace=%s group=%s dataId=%s", namespace, grp, changedID)
				prevRedis := config.GlobalConfig.Redis
				prevMySQLDSN := config.GlobalConfig.MySQL.DSN
				if err := config.ApplyMainNacosYAML(data); err != nil {
					log.Printf("[ERROR] Failed to apply changed Nacos config: %v", err)
					return
				}
				config.MergeJavaFenceFromNacos(ConfigClient, group)
				log.Printf("Successfully updated main config from Nacos")
				if config.GlobalConfig.Redis.Addr != "" && config.GlobalConfig.Redis != prevRedis {
					if err := pkgredis.Reinit(); err != nil {
						log.Printf("[ERROR] Redis reinit after main config change failed: %v", err)
					}
				}
				if config.GlobalConfig.MySQL.DSN != "" && config.GlobalConfig.MySQL.DSN != prevMySQLDSN {
					if err := mysql.Reinit(); err != nil {
						log.Printf("[ERROR] MySQL reinit after main config change failed: %v", err)
					} else {
						persistence.Init()
					}
				}
			},
		})
		if err != nil {
			log.Printf("[WARN] Failed to listen Nacos config: %v", err)
		}
	})
}

// InitNacos initializes the Nacos naming client to register the service and discover downstream services
func InitNacos() {
	// Re-apply env so NACOS_SERVER_ADDR wins even if local yaml still lists an old addr.
	config.ApplyEnvOverrides()
	nacosCfg := config.GlobalConfig.Nacos
	if nacosCfg.Port == 0 {
		nacosCfg.Port = 8848
		config.GlobalConfig.Nacos.Port = 8848
	}
	log.Printf("[nacos] connecting to %s:%d namespace=%s group=%s",
		nacosCfg.ServerAddr, nacosCfg.Port, nacosCfg.Namespace, nacosCfg.Group)

	// Configure Nacos server
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(nacosCfg.ServerAddr, nacosCfg.Port),
	}

	// Configure Nacos client properties
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(nacosCfg.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("error"),
	)

	var err error
	NamingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		log.Fatalf("Failed to initialize Nacos Naming Client: %v", err)
	}

	ConfigClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize Nacos Config Client: %v", err)
	}

	// Pull initial config
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: nacosCfg.DataId,
		Group:  nacosCfg.Group,
	})
	if err != nil {
		log.Printf("[WARN] Failed to get initial config from Nacos: %v", err)
	} else if content != "" {
		if err := config.ApplyMainNacosYAML(content); err != nil {
			log.Printf("[ERROR] Failed to unmarshal Nacos config: %v", err)
		} else {
			log.Printf("Successfully loaded initial config from Nacos (%s)", nacosCfg.DataId)
		}
	} else {
		log.Printf("[WARN] Nacos config %s is empty (group=%s)", nacosCfg.DataId, nacosCfg.Group)
	}
	config.MergeJavaFenceFromNacos(ConfigClient, nacosCfg.Group)
	config.LogXyyConfig("startup")

	// Redis: extension config redis.yaml in {group}_ops (same as Java ebike-fence bootstrap.yml).
	if err := config.ApplyRedisFromNacos(ConfigClient); err != nil {
		log.Printf("[WARN] %v", err)
	}

	// MySQL: extension config mysql.yaml in {group}_ops (same as Java ebike-fence bootstrap.yml).
	if err := config.ApplyMySQLFromNacos(ConfigClient); err != nil {
		log.Printf("[WARN] %v", err)
	}

	// FastId: extension config fast_id.yaml in {group}_ops (Java FenceKeyGenerator).
	if err := config.ApplyFastIDFromNacos(ConfigClient); err != nil {
		log.Printf("[WARN] %v", err)
	}
	config.ApplyFastIDEnvOverrides()

	registerNacosConfigListeners(nacosCfg.DataId, nacosCfg.Group)

	if !env.NacosRegisterEnabled() {
		log.Printf("Nacos config loaded; registration skipped (NACOS_REGISTER_ENABLED=false, dry_run=%t).", env.DryRun())
		return
	}

	// Register this service instance to Nacos using the actual local IP
	localIp := GetLocalIP()
	success, err := NamingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          localIp,
		Port:        uint64(config.GlobalConfig.Server.Port),
		ServiceName: config.GlobalConfig.Server.Name,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		GroupName:   config.GlobalConfig.Nacos.Group,
	})

	if !success || err != nil {
		log.Fatalf("Failed to register instance to Nacos: %v", err)
	}

	RegisteredToNacos = true
	log.Printf("Successfully registered instance %s to Nacos at %s (dry_run=%t).",
		config.GlobalConfig.Server.Name, localIp, env.DryRun())
}

// GetLocalIP retrieves the actual local pod/host IP, prioritizing K8s POD_IP env var.
// Implementation referenced from ebike-device-openapi-go.
func GetLocalIP() string {
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && isSiteLocal(ip) {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

// isSiteLocal checks if an IPv4 address is a private/site-local address.
func isSiteLocal(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	return ip[0] == 10 ||
		(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
		(ip[0] == 192 && ip[1] == 168)
}

// SelectOneHealthyInstance picks one healthy instance of the target service
func SelectOneHealthyInstance(serviceName string) (string, error) {
	instance, err := NamingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   config.GlobalConfig.Nacos.Group,
	})
	if err != nil {
		return "", err
	}
	return instance.Ip + ":" + strconv.FormatUint(instance.Port, 10), nil
}
