package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var nacosConfigClient config_client.IConfigClient

// NewNacosClientOptions builds shared Nacos client options for config/naming clients.
func NewNacosClientOptions() (*constant.ClientConfig, []constant.ServerConfig, error) {
	if GlobalConfig == nil {
		return nil, nil, fmt.Errorf("config not loaded")
	}
	nacosCfg := GlobalConfig.Nacos
	if nacosCfg.ServerAddr == "" {
		return nil, nil, fmt.Errorf("nacos.serverAddr is empty")
	}
	port := nacosCfg.Port
	if port == 0 {
		port = 8848
	}

	logDir := filepath.Join(os.TempDir(), "nacos", "log")
	cacheDir := filepath.Join(os.TempDir(), "nacos", "cache")
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(nacosCfg.ServerAddr, port),
	}
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(nacosCfg.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(logDir),
		constant.WithCacheDir(cacheDir),
		constant.WithLogLevel("error"),
	)
	return clientConfig, serverConfigs, nil
}

// OpsGroup returns the Nacos group for shared ops configs (fast_id.yaml, redis.yaml, etc.).
func OpsGroup() string {
	if GlobalConfig == nil || GlobalConfig.Nacos.Group == "" {
		return "xyy_ops"
	}
	return GlobalConfig.Nacos.Group + "_ops"
}

// InitNacosConfigClient creates the Nacos config client from local bootstrap settings.
func InitNacosConfigClient() (config_client.IConfigClient, error) {
	clientConfig, serverConfigs, err := NewNacosClientOptions()
	if err != nil {
		return nil, err
	}

	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos config client: %w", err)
	}
	nacosConfigClient = client
	return client, nil
}

// NacosConfigClient returns the initialized Nacos config client, if any.
func NacosConfigClient() config_client.IConfigClient {
	return nacosConfigClient
}
