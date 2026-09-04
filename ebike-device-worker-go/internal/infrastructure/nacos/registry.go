package nacos

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"ebike-device-worker-go/internal/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	namingClient   naming_client.INamingClient
	registeredIP   string
	deregisterOnce sync.Once
)

// InitNamingClient creates the Nacos naming client (service discovery registration).
func InitNamingClient() error {
	if config.GlobalConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	if config.GlobalConfig.Nacos.ServerAddr == "" {
		return fmt.Errorf("nacos.serverAddr is empty")
	}
	clientConfig, serverConfigs, err := config.NewNacosClientOptions()
	if err != nil {
		return err
	}
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return fmt.Errorf("create nacos naming client: %w", err)
	}
	namingClient = client
	return nil
}

// NamingClient returns the initialized naming client, if any.
func NamingClient() naming_client.INamingClient {
	return namingClient
}

// RegisterInstance registers this pod to Nacos when enabled.
func RegisterInstance() error {
	if config.GlobalConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	if !config.GlobalConfig.Nacos.RegisterEnabled {
		log.Println("[nacos] registration disabled (nacos.registerEnabled=false)")
		return nil
	}
	if namingClient == nil {
		return fmt.Errorf("nacos naming client is nil")
	}

	serviceName := ServiceName()
	port, err := ServerPort()
	if err != nil {
		return err
	}
	ip := LocalIP()
	registeredIP = ip
	ok, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Weight:      registerWeight(),
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		GroupName:   config.GlobalConfig.Nacos.Group,
		Metadata: map[string]string{
			"lang": "go",
		},
	})
	if err != nil || !ok {
		return fmt.Errorf("register %s@%s:%d: %w", serviceName, ip, port, err)
	}
	log.Printf("[nacos] registered %s at %s:%d group=%s", serviceName, ip, port, config.GlobalConfig.Nacos.Group)
	return nil
}

// DeregisterInstance removes this instance from Nacos. Safe to call multiple times.
func DeregisterInstance() {
	deregisterOnce.Do(deregisterInstance)
}

func deregisterInstance() {
	if namingClient == nil || config.GlobalConfig == nil || !config.GlobalConfig.Nacos.RegisterEnabled {
		return
	}
	serviceName := ServiceName()
	port, err := ServerPort()
	if err != nil {
		log.Printf("[WARN] nacos deregister port: %v", err)
		return
	}
	ip := registeredIP
	if ip == "" {
		ip = LocalIP()
	}
	_, err = namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   config.GlobalConfig.Nacos.Group,
		Ephemeral:   true,
	})
	if err != nil {
		log.Printf("[WARN] nacos deregister failed: %v", err)
		return
	}
	log.Printf("[nacos] deregistered %s at %s:%d", serviceName, ip, port)
}

func registerWeight() float64 {
	if v := os.Getenv("NACOS_REGISTER_WEIGHT"); v != "" {
		w, err := strconv.ParseFloat(v, 64)
		if err == nil && w > 0 {
			return w
		}
	}
	return 10
}

// ServiceName returns the Nacos service name (spring.application.name).
func ServiceName() string {
	if config.GlobalConfig != nil && config.GlobalConfig.Spring.Application.Name != "" {
		return config.GlobalConfig.Spring.Application.Name
	}
	return "ebike-device-worker"
}

// ServerPort parses server.port for Nacos registration.
func ServerPort() (uint64, error) {
	if config.GlobalConfig == nil {
		return 0, fmt.Errorf("config not loaded")
	}
	port, err := strconv.ParseUint(config.GlobalConfig.Server.Port, 10, 64)
	if err != nil || port == 0 {
		return 0, fmt.Errorf("invalid server.port %q", config.GlobalConfig.Server.Port)
	}
	return port, nil
}

// LocalIP returns pod/host IP for Nacos registration.
// Aligned with ebike-device-openapi-go / ebike-fence-go: POD_IP first, then site-local IPv4 on up interfaces.
func LocalIP() string {
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}
			if isSiteLocal(ip) {
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
