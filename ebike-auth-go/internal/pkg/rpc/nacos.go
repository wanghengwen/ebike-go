package rpc

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

var (
	namingClient naming_client.INamingClient
)

func InitNacosNaming() error {
	nc := config.AppConfig.Nacos
	if nc.ServerAddr == "" {
		return nil
	}

	parts := strings.Split(nc.ServerAddr, ":")
	ip := parts[0]
	portStr := "8848"
	if len(parts) > 1 {
		portStr = parts[1]
	}
	port, _ := strconv.ParseUint(portStr, 10, 64)

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(ip, port, constant.WithContextPath("/nacos")),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(nc.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("./logs/nacos"),
		constant.WithCacheDir("./logs/nacos/cache"),
		constant.WithLogLevel("error"),
	)

	var err error
	namingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return fmt.Errorf("create nacos naming client error: %w", err)
	}

	// Register service only if explicitly enabled
	if nc.RegisterEnabled {
		serviceName := "ebike-auth-" + config.AppConfig.AuthMode
		registerIP := getRegisterIP()
		success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          registerIP,
			Port:        uint64(config.AppConfig.Port),
			ServiceName: serviceName,
			Weight:      10,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			GroupName:   nc.Group,
		})
		if err != nil || !success {
			logger.Log.Error("Failed to register service to Nacos", zap.Error(err), zap.String("service", serviceName))
		} else {
			logger.Log.Info("Successfully registered service to Nacos", zap.String("service", serviceName))
		}
	}

	return nil
}

// ResolveServiceAddress resolves healthy server instance from Nacos, with fallback to local env settings
func ResolveServiceAddress(serviceName string) (string, error) {
	// If nacos naming is disabled or naming client is nil, we fall back to direct config
	if namingClient == nil {
		return "", fmt.Errorf("nacos naming client is not initialized")
	}

	instances, err := namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   config.AppConfig.Nacos.Group,
		HealthyOnly: true,
	})
	if err != nil {
		return "", err
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	// Simple random load balancing
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(len(instances))
	inst := instances[idx]
	return fmt.Sprintf("http://%s:%d", inst.Ip, inst.Port), nil
}

func DeregisterInstance() {
	if namingClient != nil && config.AppConfig.Nacos.RegisterEnabled {
		serviceName := "ebike-auth-" + config.AppConfig.AuthMode
		success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          getRegisterIP(),
			Port:        uint64(config.AppConfig.Port),
			ServiceName: serviceName,
			GroupName:   config.AppConfig.Nacos.Group,
			Ephemeral:   true,
		})
		if err != nil || !success {
			logger.Log.Error("Failed to deregister from Nacos", zap.Error(err))
		} else {
			logger.Log.Info("Successfully deregistered from Nacos")
		}
	}
}

func LogServiceURLs() {
	mURL := getManagementBaseURL()
	uURL := getUserBaseURL()
	logger.Log.Info("RPC service endpoints resolved",
		zap.String("ebike-management", mURL),
		zap.String("ebike-user", uURL),
	)
}

func getRegisterIP() string {
	if ip := os.Getenv("POD_IP"); ip != "" {
		return ip
	}
	if ip := os.Getenv("NACOS_REGISTER_IP"); ip != "" {
		return ip
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && addr.IP != nil {
			return addr.IP.String()
		}
	}
	return "127.0.0.1"
}

