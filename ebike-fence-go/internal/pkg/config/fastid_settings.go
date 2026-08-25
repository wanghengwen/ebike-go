package config

import (
	"os"
	"strconv"
	"strings"

	"ebike-fence-go/internal/pkg/fastid"
)

// FastIDSettings mirrors spring.xyy.fastid in Java Nacos fast_id.yaml.
type FastIDSettings struct {
	Enabled                  bool   `yaml:"enabled"`
	ServerAddr               string `yaml:"server-addr"`
	URL                      string `yaml:"url"`
	Namespace                string `yaml:"namespace"`
	GroupID                  string `yaml:"groupId"`
	Secret                   string `yaml:"secret"`
	AppName                  string `yaml:"app-name"`
	Port                     int    `yaml:"port"`
	InstanceNoLocalDirectory string `yaml:"instance-no-local-directory"`
	MachineUUID              string `yaml:"machine-uuid"`
	DriftTime                int    `yaml:"drift-time"`
	UseHTTPS                 bool   `yaml:"use-https"`
	UseHTTPSExplicit         bool   `yaml:"-"`
}

// FastIDEnabled reports whether FastId client should start.
func FastIDEnabled() bool {
	_, enabled := resolvedFastIDSettings()
	return enabled
}

// ResolvedFastID builds fastid.Config from spring.xyy.fastid (Nacos-aligned).
func ResolvedFastID() fastid.Config {
	cfg, _ := resolvedFastIDSettings()
	return cfg
}

func resolvedFastIDSettings() (fastid.Config, bool) {
	src := GlobalConfig.Spring.Xyy.Fastid
	if !src.Enabled && GlobalConfig.FastID.Enabled {
		src = GlobalConfig.FastID
	}
	if src.ServerAddr == "" && src.URL == "" && GlobalConfig.FastID.URL != "" {
		src = GlobalConfig.FastID
	}
	if !src.Enabled && !GlobalConfig.FastID.Enabled && (src.ServerAddr != "" || GlobalConfig.FastID.URL != "") {
		src.Enabled = true
	}

	url := firstNonEmpty(src.ServerAddr, src.URL, GlobalConfig.FastID.ServerAddr, GlobalConfig.FastID.URL)
	appName := resolveFastIDAppName(src.AppName, GlobalConfig.Spring.Application.Name, GlobalConfig.Server.Name)
	namespace := firstNonEmpty(src.Namespace, GlobalConfig.FastID.Namespace, GlobalConfig.Nacos.Namespace)
	groupID := firstNonEmpty(src.GroupID, GlobalConfig.FastID.GroupID, "xyy")
	secret := firstNonEmpty(os.Getenv("SPRING_XYY_FASTID_SECRET"), src.Secret, GlobalConfig.FastID.Secret)

	port := src.Port
	if port == 0 {
		port = GlobalConfig.FastID.Port
	}
	if port == 0 {
		port = GlobalConfig.Server.Port
	}
	if port == 0 {
		port = 8080
	}

	localDir := firstNonEmpty(src.InstanceNoLocalDirectory, GlobalConfig.FastID.InstanceNoLocalDirectory)
	machineUUID := firstNonEmpty(src.MachineUUID, GlobalConfig.FastID.MachineUUID, os.Getenv("FASTID_MACHINE_UUID"))
	drift := src.DriftTime
	if drift == 0 {
		drift = GlobalConfig.FastID.DriftTime
	}
	if drift == 0 {
		drift = 10
	}

	enabled := src.Enabled || GlobalConfig.FastID.Enabled
	if v := os.Getenv("SPRING_XYY_FASTID_ENABLED"); v != "" {
		enabled, _ = strconv.ParseBool(v)
	}

	return fastid.Config{
		URL:                      url,
		Namespace:                namespace,
		GroupID:                  groupID,
		AppName:                  appName,
		Port:                     port,
		Secret:                   secret,
		InstanceNoLocalDirectory: localDir,
		MachineUUID:              machineUUID,
		DriftTime:                drift,
		UseHTTPS:                 src.UseHTTPS,
		UseHTTPSExplicit:         src.UseHTTPSExplicit,
	}, enabled
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func isUnresolvedPlaceholder(value string) bool {
	return strings.Contains(value, "${")
}

func resolveFastIDAppName(fastidAppName, springAppName, serverName string) string {
	appName := strings.TrimSpace(springAppName)
	if appName == "" || isUnresolvedPlaceholder(appName) {
		appName = strings.TrimSpace(serverName)
	}
	if appName == "" {
		appName = "ebike-fence"
	}
	if fastidAppName == "" || isUnresolvedPlaceholder(fastidAppName) {
		return appName
	}
	resolved := strings.ReplaceAll(fastidAppName, "${spring.application.name}", appName)
	if isUnresolvedPlaceholder(resolved) {
		return appName
	}
	return resolved
}
