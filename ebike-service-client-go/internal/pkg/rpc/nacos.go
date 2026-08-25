package rpc

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"ebike-service-client-go/internal/pkg/buildinfo"
	"ebike-service-client-go/internal/pkg/config"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

func nacosStartupVerbose() bool {
	return strings.EqualFold(os.Getenv("CLIENT_GO_NACOS_VERBOSE"), "true")
}

// InitNacos initializes the Nacos naming client to register the service and discover downstream services
func InitNacos() {
	nacosCfg := config.GlobalConfig.Nacos

	serverOpts := []constant.ServerOption{}
	if cp := strings.TrimSpace(nacosCfg.ContextPath); cp != "" {
		serverOpts = append(serverOpts, constant.WithContextPath(cp))
	}
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(nacosCfg.ServerAddr, nacosCfg.Port, serverOpts...),
	}

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

	merged, failed := 0, 0
	for _, dataId := range nacosExtensionDataIds(config.GlobalConfig) {
		for _, group := range nacosExtensionGroups(config.GlobalConfig) {
			switch loadRemoteConfig(dataId, group) {
			case loadMerged:
				merged++
			case loadFailed:
				failed++
			}
		}
	}

	config.ApplyEnvOverrides()

	mapURL := strings.TrimSpace(config.GlobalConfig.Xyy.MapServiceConfig.Url)
	mapSecretOK := config.GlobalConfig.Xyy.MapServiceConfig.Secret != ""
	dryRun := os.Getenv("DRY_RUN") == "true"

	if mapURL != "" {
		log.Printf("Nacos ready build=%s namespace=%s configs=%d failed=%d map-service=%s secret=%t dry_run=%t",
			buildinfo.Version, nacosCfg.Namespace, merged, failed, mapURL, mapSecretOK, dryRun)
	} else {
		log.Printf("[WARN] Nacos ready build=%s namespace=%s configs=%d failed=%d map-service=empty secret=%t dry_run=%t",
			buildinfo.Version, nacosCfg.Namespace, merged, failed, mapSecretOK, dryRun)
	}

	if dryRun {
		return
	}

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
	log.Printf("Registered %s to Nacos at %s", config.GlobalConfig.Server.Name, localIp)
}

type loadResult int

const (
	loadEmpty loadResult = iota
	loadMerged
	loadFailed
)

func nacosExtensionDataIds(nacosCfg config.AppConfig) []string {
	seen := map[string]struct{}{}
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
	}
	add(nacosCfg.Nacos.DataId)
	for _, id := range nacosCfg.Nacos.ExtensionDataIds {
		add(id)
	}
	add("ebike-service-client.yml")
	add("ebike-service-client.yaml")
	add("ebike-service-client-go.yaml")
	add("ebike-service-client-go.yml")
	add("map-service.yaml")
	add("map-service.yml")

	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func nacosExtensionGroups(nacosCfg config.AppConfig) []string {
	seen := map[string]struct{}{}
	add := func(g string) {
		g = strings.TrimSpace(g)
		if g == "" {
			return
		}
		if _, ok := seen[g]; ok {
			return
		}
		seen[g] = struct{}{}
	}
	add(nacosCfg.Nacos.Group)
	for _, g := range nacosCfg.Nacos.ExtensionGroups {
		add(g)
	}
	if g := strings.TrimSpace(nacosCfg.Nacos.Group); g != "" {
		add(g + "_ops")
	}

	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	return out
}

func loadRemoteConfig(dataId, group string) loadResult {
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		log.Printf("[WARN] Nacos get failed dataId=%s group=%s: %v", dataId, group, err)
		return loadFailed
	}
	if strings.TrimSpace(content) == "" {
		return loadEmpty
	}

	beforeURL := config.GlobalConfig.Xyy.MapServiceConfig.Url
	config.MergeRemoteYAML(content)
	if nacosStartupVerbose() {
		log.Printf("Nacos merged dataId=%s group=%s bytes=%d", dataId, group, len(content))
		if after := config.GlobalConfig.Xyy.MapServiceConfig.Url; after != beforeURL {
			log.Printf("Nacos mapServiceConfig.url from dataId=%s group=%s: %s", dataId, group, after)
		}
	}
	return loadMerged
}

// GetLocalIP retrieves the actual local pod/host IP, prioritizing K8s POD_IP env var.
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
