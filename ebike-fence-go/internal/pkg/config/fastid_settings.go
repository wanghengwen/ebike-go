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
	// ServerURL overrides ServerAddr for Go clients; Java ignores this field.
	ServerURL                string `yaml:"server-url"`
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
	ApplyFastIDEnvOverrides()
	applyFastIDSchemeFromAddr(&GlobalConfig.Spring.Xyy.Fastid)

	src := GlobalConfig.Spring.Xyy.Fastid
	if !src.Enabled && GlobalConfig.FastID.Enabled {
		src = GlobalConfig.FastID
	}
	if !hasFastIDEndpoint(src) && hasFastIDEndpoint(GlobalConfig.FastID) {
		src = GlobalConfig.FastID
	}
	if !src.Enabled && !GlobalConfig.FastID.Enabled && hasFastIDEndpoint(src) {
		src.Enabled = true
	}

	// Endpoint precedence within merged settings: server-url > server-addr > url
	url := firstNonEmpty(
		src.ServerURL, src.ServerAddr, src.URL,
		GlobalConfig.FastID.ServerURL, GlobalConfig.FastID.ServerAddr, GlobalConfig.FastID.URL,
	)
	appName := resolveFastIDAppName(src.AppName, GlobalConfig.Spring.Application.Name, GlobalConfig.Server.Name)
	namespace := firstNonEmpty(src.Namespace, GlobalConfig.FastID.Namespace, GlobalConfig.Nacos.Namespace)
	groupID := firstNonEmpty(src.GroupID, GlobalConfig.FastID.GroupID, "xyy")
	secret := firstNonEmpty(src.Secret, GlobalConfig.FastID.Secret)

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

func hasFastIDEndpoint(s FastIDSettings) bool {
	return s.ServerURL != "" || s.ServerAddr != "" || s.URL != ""
}

// ApplyFastIDEnvOverrides lets SPRING_XYY_FASTID_* override Nacos/local FastID settings.
// Precedence: env > nacos > local. Within endpoint fields: server-url > server-addr > url.
func ApplyFastIDEnvOverrides() {
	dst := &GlobalConfig.Spring.Xyy.Fastid
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_SECRET")); v != "" {
		dst.Secret = v
	}
	serverURLEnv := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_SERVER_URL"))
	if serverURLEnv == "" {
		serverURLEnv = strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_URL"))
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_SERVER_ADDR")); v != "" {
		dst.ServerAddr = v
		// Env server-addr should beat Nacos server-url unless env also sets server-url.
		if serverURLEnv == "" {
			dst.ServerURL = ""
		}
		applyFastIDSchemeFromAddr(dst)
	}
	if serverURLEnv != "" {
		dst.ServerURL = serverURLEnv
		applyFastIDSchemeFromAddr(dst)
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_NAMESPACE")); v != "" {
		dst.Namespace = v
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_GROUPID")); v != "" {
		dst.GroupID = v
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_APP_NAME")); v != "" {
		dst.AppName = v
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_PORT")); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			dst.Port = port
		}
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_ENABLED")); v != "" {
		dst.Enabled, _ = strconv.ParseBool(v)
	}
	if v := strings.TrimSpace(os.Getenv("SPRING_XYY_FASTID_USE_HTTPS")); v != "" {
		dst.UseHTTPS, _ = strconv.ParseBool(v)
		dst.UseHTTPSExplicit = true
	}
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
