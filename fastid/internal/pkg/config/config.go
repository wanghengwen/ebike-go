package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// NacosConfig holds Nacos connection parameters.
type NacosConfig struct {
	ServerAddr string            `mapstructure:"server-addr"`
	Namespace  string            `mapstructure:"namespace"`
	Group      string            `mapstructure:"group"`
	AppName    string            `mapstructure:"app-name"`
	Discovery  NacosDiscovery    `mapstructure:"discovery"`
	Config     NacosConfigSwitch `mapstructure:"config"`
}

// NacosDiscovery controls service registration.
type NacosDiscovery struct {
	Enabled bool `mapstructure:"enabled"`
}

// NacosConfigSwitch controls pulling config from Nacos.
type NacosConfigSwitch struct {
	Enabled bool `mapstructure:"enabled"`
}

// DatabaseConfig holds MySQL connection parameters.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Params   string `mapstructure:"params"`
}

// FastIdConfig holds application-level settings such as API secrets.
type FastIdConfig struct {
	Secrets []string `mapstructure:"secrets"`
}

// AppConfig is the root configuration struct.
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
	Database DatabaseConfig `mapstructure:"database"`
	FastId   FastIdConfig   `mapstructure:"fastid"`
}

// DSN returns a Go MySQL driver compatible data source name.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		d.Username, d.Password, d.Host, d.Port, d.Name, d.Params)
}

// LoadConfig reads conf/application.yml and optionally pulls overrides from Nacos.
func LoadConfig() (*AppConfig, error) {
	v := viper.New()
	v.SetConfigName("application")
	v.SetConfigType("yaml")
	v.AddConfigPath("conf")
	v.AddConfigPath(".")
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		v.SetConfigFile(path)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read application config: %w", err)
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables for seamless migration in K8S
	overrideWithEnv(cfg)

	// Default server port
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	// Try loading from Nacos if enabled
	if cfg.Nacos.Config.Enabled {
		if err := loadFromNacos(cfg); err != nil {
			log.Printf("[WARN] Failed to load config from Nacos, using local config: %v", err)
		}
	}

	return cfg, nil
}

// overrideWithEnv maps Spring-style environment variables to config fields.
func overrideWithEnv(cfg *AppConfig) {
	// 1. Server Port
	if val := getEnv("SERVER_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = p
		}
	}

	// 2. Nacos Server Address
	if val := getEnv("SPRING_NACOS_SERVER_ADDR", "SPRING_CLOUD_NACOS_DISCOVERY_SERVER_ADDR", "SPRING_CLOUD_NACOS_CONFIG_SERVER_ADDR", "NACOS_SERVER_ADDR"); val != "" {
		cfg.Nacos.ServerAddr = val
	}

	// 3. Nacos Namespace
	if val := getEnv("SPRING_NACOS_NAMESPACE", "SPRING_CLOUD_NACOS_DISCOVERY_NAMESPACE", "SPRING_CLOUD_NACOS_CONFIG_NAMESPACE", "NACOS_NAMESPACE"); val != "" {
		cfg.Nacos.Namespace = val
	}

	// 4. Nacos Group
	if val := getEnv("SPRING_NACOS_GROUP", "SPRING_CLOUD_NACOS_DISCOVERY_GROUP", "SPRING_CLOUD_NACOS_CONFIG_GROUP", "NACOS_GROUP"); val != "" {
		cfg.Nacos.Group = val
	}

	// 5. App Name (Data ID)
	if val := getEnv("SPRING_APPLICATION_NAME", "NACOS_APP_NAME", "APP_NAME"); val != "" {
		cfg.Nacos.AppName = val
	}

	// Auto-enable Nacos config and discovery if server address is specified via env
	if cfg.Nacos.ServerAddr != "" {
		cfg.Nacos.Config.Enabled = true
		cfg.Nacos.Discovery.Enabled = true
	}

	// 6. Discovery Enabled / Register Enabled
	if val := getEnv("SPRING_NACOS_REGISTER_ENABLED", "SPRING_CLOUD_NACOS_DISCOVERY_REGISTER_ENABLED"); val != "" {
		valLower := strings.ToLower(val)
		if valLower == "false" || valLower == "0" {
			cfg.Nacos.Discovery.Enabled = false
		} else if valLower == "true" || valLower == "1" {
			cfg.Nacos.Discovery.Enabled = true
		}
	}
}

func getEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// nacosClientParams builds the common Nacos client configuration.
func nacosClientParams(cfg *AppConfig) (vo.NacosClientParam, error) {
	host, port, err := parseAddr(cfg.Nacos.ServerAddr)
	if err != nil {
		return vo.NacosClientParam{}, err
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(host, port),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.Nacos.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)

	return vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	}, nil
}

// loadFromNacos pulls mysql.yaml and app config from the Nacos config center.
func loadFromNacos(cfg *AppConfig) error {
	params, err := nacosClientParams(cfg)
	if err != nil {
		return err
	}

	configClient, err := clients.NewConfigClient(params)
	if err != nil {
		return fmt.Errorf("failed to create Nacos config client: %w", err)
	}

	// Pull mysql.yaml from <group>_ops
	opsGroup := cfg.Nacos.Group + "_ops"
	mysqlContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "mysql.yaml",
		Group:  opsGroup,
	})
	if err != nil {
		log.Printf("[WARN] Failed to get mysql.yaml from Nacos group=%s: %v", opsGroup, err)
	} else if mysqlContent != "" {
		parseMySQLConfig(mysqlContent, cfg)
	}

	// Pull app-specific config (xyy-fastid.yml)
	appDataID := cfg.Nacos.AppName + ".yml"
	appContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: appDataID,
		Group:  cfg.Nacos.Group,
	})

	// Fallback to xyy-fastid.yml if the configured name returns empty or fails
	if err != nil || appContent == "" {
		fallbackDataID := "xyy-fastid.yml"
		if appDataID != fallbackDataID {
			log.Printf("[INFO] Failed or empty config for %s, trying fallback %s", appDataID, fallbackDataID)
			appContentFallback, errFallback := configClient.GetConfig(vo.ConfigParam{
				DataId: fallbackDataID,
				Group:  cfg.Nacos.Group,
			})
			if errFallback == nil && appContentFallback != "" {
				appContent = appContentFallback
				appDataID = fallbackDataID
			}
		}
	}

	if appContent != "" {
		parseAppConfig(appContent, cfg)
	}

	return nil
}

// RegisterService registers this application with Nacos service discovery.
// Returns a naming client that can be used for deregistration on shutdown.
func RegisterService(cfg *AppConfig) (naming_client.INamingClient, error) {
	if !cfg.Nacos.Discovery.Enabled {
		log.Println("[INFO] Nacos service discovery is disabled, skipping registration")
		return nil, nil
	}

	params, err := nacosClientParams(cfg)
	if err != nil {
		return nil, err
	}

	namingClient, err := clients.NewNamingClient(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nacos naming client: %w", err)
	}

	localIP := getLocalIP()
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          localIP,
		Port:        uint64(cfg.Server.Port),
		ServiceName: cfg.Nacos.AppName,
		GroupName:   cfg.Nacos.Group,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"preserved.register.source": "GO",
		},
	})
	if err != nil {
		return namingClient, fmt.Errorf("failed to register instance in Nacos: %w", err)
	}

	log.Printf("[INFO] Nacos service registered: service=%s ip=%s port=%d success=%v",
		cfg.Nacos.AppName, localIP, cfg.Server.Port, success)
	return namingClient, nil
}

// DeregisterService removes this application from Nacos service discovery.
func DeregisterService(namingClient naming_client.INamingClient, cfg *AppConfig) {
	if namingClient == nil {
		return
	}
	localIP := getLocalIP()
	_, _ = namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          localIP,
		Port:        uint64(cfg.Server.Port),
		ServiceName: cfg.Nacos.AppName,
		GroupName:   cfg.Nacos.Group,
		Ephemeral:   true,
	})
	log.Printf("[INFO] Nacos service deregistered: service=%s", cfg.Nacos.AppName)
}

// ---- Nacos config parsing helpers ----

var (
	rawConnectionParam string
	placeholderRegex   = regexp.MustCompile(`\$\{([^}]+)\}`)
)

// mysqlYamlRoot models the structure of the shared mysql.yaml in Nacos.
type mysqlYamlRoot struct {
	MySQL struct {
		ConnectionParam string `yaml:"connection-param"`
		FastID          struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Database string `yaml:"database"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"fastid"`
	} `yaml:"mysql"`
}

func parseMySQLConfig(content string, cfg *AppConfig) {
	var mc mysqlYamlRoot
	if err := yaml.Unmarshal([]byte(content), &mc); err != nil {
		log.Printf("[WARN] Failed to parse mysql.yaml from Nacos: %v", err)
		return
	}

	rawConnectionParam = mc.MySQL.ConnectionParam

	f := mc.MySQL.FastID

	// Resolve placeholder values from environments if present
	host := resolvePlaceholders(f.Host, cfg, rawConnectionParam)
	username := resolvePlaceholders(f.Username, cfg, rawConnectionParam)
	password := resolvePlaceholders(f.Password, cfg, rawConnectionParam)

	if host != "" {
		cfg.Database.Host = host
	}
	if f.Port != 0 {
		cfg.Database.Port = f.Port
	}
	if f.Database != "" {
		cfg.Database.Name = f.Database
	}
	if username != "" {
		cfg.Database.Username = username
	}
	if password != "" {
		cfg.Database.Password = password
	}
	if mc.MySQL.ConnectionParam != "" {
		cfg.Database.Params = toGoMySQLParams(resolvePlaceholders(mc.MySQL.ConnectionParam, cfg, ""))
	}
	log.Printf("[INFO] MySQL config loaded from Nacos: host=%s port=%d db=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}

// appYamlRoot models the structure of the luopingtech-fastid.yml or xyy-fastid.yml config in Nacos.
type appYamlRoot struct {
	Spring struct {
		XYY struct {
			FastID struct {
				Secrets []string `yaml:"secrets"`
			} `yaml:"fastid"`
		} `yaml:"xyy"`
		LuopingTech struct {
			FastID struct {
				Secrets []string `yaml:"secrets"`
			} `yaml:"fastid"`
		} `yaml:"luopingtech"`
		Datasource struct {
			URL             string `yaml:"url"`
			Username        string `yaml:"username"`
			Password        string `yaml:"password"`
			DriverClassName string `yaml:"driverClassName"`
		} `yaml:"datasource"`
	} `yaml:"spring"`
}

func parseAppConfig(content string, cfg *AppConfig) {
	var ac appYamlRoot
	if err := yaml.Unmarshal([]byte(content), &ac); err != nil {
		log.Printf("[WARN] Failed to parse app config from Nacos: %v", err)
		return
	}

	// 1. Secrets (try both xyy and luopingtech prefixes, not reading xiaoantech)
	if secrets := ac.Spring.XYY.FastID.Secrets; len(secrets) > 0 {
		cfg.FastId.Secrets = secrets
		log.Printf("[INFO] FastId secrets loaded from Nacos (xyy): count=%d", len(secrets))
	} else if secrets := ac.Spring.LuopingTech.FastID.Secrets; len(secrets) > 0 {
		cfg.FastId.Secrets = secrets
		log.Printf("[INFO] FastId secrets loaded from Nacos (luopingtech): count=%d", len(secrets))
	}

	// 2. Datasource (URL, username, password)
	if ac.Spring.Datasource.URL != "" {
		resolvedURL := resolvePlaceholders(ac.Spring.Datasource.URL, cfg, rawConnectionParam)
		log.Printf("[INFO] Parsing JDBC URL from app config: %s", resolvedURL)
		parseJDBCURL(resolvedURL, cfg)
	}

	if ac.Spring.Datasource.Username != "" {
		resolvedUsername := resolvePlaceholders(ac.Spring.Datasource.Username, cfg, rawConnectionParam)
		if resolvedUsername != "" {
			cfg.Database.Username = resolvedUsername
		}
	}

	if ac.Spring.Datasource.Password != "" {
		resolvedPassword := resolvePlaceholders(ac.Spring.Datasource.Password, cfg, rawConnectionParam)
		if resolvedPassword != "" {
			cfg.Database.Password = resolvedPassword
		}
	}
}

// resolvePlaceholders generic placeholder resolver using environment and loaded config.
func resolvePlaceholders(input string, cfg *AppConfig, rawParam string) string {
	if input == "" {
		return ""
	}
	return placeholderRegex.ReplaceAllStringFunc(input, func(match string) string {
		key := match[2 : len(match)-1] // Extract key name between ${ and }

		// Split by colon to handle default values, e.g. ${SOME_KEY:default_val}
		parts := strings.SplitN(key, ":", 2)
		name := parts[0]
		defaultVal := ""
		if len(parts) > 1 {
			defaultVal = parts[1]
		}

		// Check known config properties
		switch name {
		case "mysql.fastid.host":
			if cfg.Database.Host != "" {
				return cfg.Database.Host
			}
		case "mysql.fastid.port":
			if cfg.Database.Port != 0 {
				return strconv.Itoa(cfg.Database.Port)
			}
		case "mysql.fastid.username":
			if cfg.Database.Username != "" {
				return cfg.Database.Username
			}
		case "mysql.fastid.password":
			if cfg.Database.Password != "" {
				return cfg.Database.Password
			}
		case "mysql.connection-param":
			if rawParam != "" {
				return rawParam
			}
		}

		// Check environment variables
		if val := os.Getenv(name); val != "" {
			return val
		}

		// Fallback to default value if provided
		if len(parts) > 1 {
			return defaultVal
		}

		return match // Keep as is if unresolved
	})
}

// parseJDBCURL parses a standard JDBC URL and extracts host, port, database and query parameters.
func parseJDBCURL(jdbcURL string, cfg *AppConfig) {
	if !strings.HasPrefix(jdbcURL, "jdbc:mysql://") {
		return
	}
	url := strings.TrimPrefix(jdbcURL, "jdbc:mysql://")

	// Separate query parameters
	parts := strings.SplitN(url, "?", 2)
	hostPath := parts[0]
	var rawParams string
	if len(parts) > 1 {
		rawParams = parts[1]
	}

	// Separate host:port and database name
	hostDbParts := strings.SplitN(hostPath, "/", 2)
	hostPort := hostDbParts[0]
	if len(hostDbParts) > 1 && hostDbParts[1] != "" {
		cfg.Database.Name = hostDbParts[1]
	}

	// Separate host and port
	host, portStr, err := net.SplitHostPort(hostPort)
	if err == nil {
		cfg.Database.Host = host
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Database.Port = port
		}
	} else {
		// If port is missing, hostPort is just the host
		cfg.Database.Host = hostPort
	}

	if rawParams != "" {
		cfg.Database.Params = toGoMySQLParams(rawParams)
	}
}

var jdbcParamsBlacklist = map[string]bool{
	"usessl":                   true,
	"useunicode":               true,
	"characterencoding":        true,
	"servertimezone":           true,
	"zerodatetimebehavior":     true,
	"useaffectedrows":          true,
	"rewritebatchedstatements": true,
	"allowmultiqueries":        true,
	"autoreconnect":            true,
	"tinyint1isbit":            true,
	"transformedbitisboolean":  true,
	"useconfigs":               true,
	"enablepacketdebug":        true,
	"cacheprepstmts":           true,
	"prepstmtcachesize":        true,
	"prepstmtcachesqllimit":    true,
	"useserverprepstmts":       true,
}

// toGoMySQLParams converts JDBC-style connection params to Go MySQL driver params.
func toGoMySQLParams(jdbcParams string) string {
	parts := strings.Split(jdbcParams, "&")
	var validParts []string

	hasLoc := false
	hasParseTime := false
	hasCharset := false

	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		key := strings.ToLower(kv[0])

		// Skip blacklisted JDBC parameters
		if jdbcParamsBlacklist[key] {
			if key == "usessl" && len(kv) > 1 {
				val := strings.ToLower(kv[1])
				if val == "true" {
					validParts = append(validParts, "tls=true")
				} else if val == "false" {
					validParts = append(validParts, "tls=false")
				}
			}
			continue
		}

		if key == "loc" {
			hasLoc = true
		}
		if key == "parsetime" {
			hasParseTime = true
		}
		if key == "charset" {
			hasCharset = true
		}

		validParts = append(validParts, part)
	}

	if !hasParseTime {
		validParts = append(validParts, "parseTime=True")
	}
	if !hasLoc {
		validParts = append(validParts, "loc=Local")
	}
	if !hasCharset {
		validParts = append(validParts, "charset=utf8mb4")
	}

	return strings.Join(validParts, "&")
}

// ---- Network helpers ----

func parseAddr(addr string) (string, uint64, error) {
	parts := strings.Split(addr, ":")
	host := parts[0]
	port := uint64(8848)
	if len(parts) > 1 {
		p, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return "", 0, fmt.Errorf("invalid Nacos port in address %q: %w", addr, err)
		}
		port = p
	}
	return host, port, nil
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
