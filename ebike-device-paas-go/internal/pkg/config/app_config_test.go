package config

import "testing"

// realAppYAML is the actual ebike-device-paas.yml stored in Nacos (group=xyy).
// It mixes Spring/Java-only sections (server.tomcat, spring.*, feign.*, logging.*)
// — including spring.redis.port as a "${...}" placeholder string — with the
// business fields we care about. The test guards that this real document merges
// cleanly and the business fields are captured.
const realAppYAML = `
server:
  tomcat:
    accept-count: 200
    max-connections: 8192
    connection-timeout: 5000
    threads:
      min-spare: 20
      max: 200

spring:
  jackson:
    date-format: yyyy-MM-dd HH:mm:ss
    time-zone: GMT+8
  redis:
    host: ${redis.ebike_device_paas.host}
    port: ${redis.ebike_device_paas.port}
    password: ${redis.ebike_device_paas.password}
    database: ${redis.ebike_device_paas.database}
    connect-timeout: 1000ms
    timeout: 1000ms
    lettuce:
      pool:
        max-active: 8
        max-idle: 8
        min-idle: 0
        max-wait: 500ms

feign:
  hystrix:
    enabled: false
  client:
    refresh-enabled: true
    config:
      default:
        connectTimeout: 10000
        readTimeout: 18000
      ebike-device-worker:
        DeviceTrajectoryApiFeign#getTrajectoryDistance(TrajectoryRealTimeQryDto):
          connectTimeout: 2000
          readTimeout: 3000
    auth-url-regex:
      - /ebike/.*

anvelink:
  openapi:
    url: openapi.luopingtech.com
    aes-key: 13dade85494f4b2ca4f49b21cd058625

ecu:
  debug:
    isEnable: false
  kafka: 
    parent-topic: saas_0

coordinate: 
  type: 1  ## 0:wgs84坐标系 1:GCJ-02坐标系

logging:
  level:
    com.xyy.ebike.device.paas.infrastructure.dataApiImpl.rpc: info
`

func TestMergeNacosAppConfigRealYAML(t *testing.T) {
	// Isolate the package-level GlobalConfig and start from a zero Config so that
	// passing assertions can only result from a successful unmarshal+merge.
	saved := GlobalConfig
	t.Cleanup(func() { GlobalConfig = saved })
	GlobalConfig = &Config{}

	MergeNacosAppConfig(realAppYAML)

	if got := GlobalConfig.Coordinate.Type; got != 1 {
		t.Errorf("coordinate.type = %d, want 1", got)
	}
	if got := GlobalConfig.Ecu.Kafka.ParentTopic; got != "saas_0" {
		t.Errorf("ecu.kafka.parent-topic = %q, want %q", got, "saas_0")
	}
	if got := GlobalConfig.Anvelink.Openapi.URL; got != "openapi.luopingtech.com" {
		t.Errorf("anvelink.openapi.url = %q, want %q", got, "openapi.luopingtech.com")
	}
	if got := GlobalConfig.Anvelink.Openapi.AesKey; got != "13dade85494f4b2ca4f49b21cd058625" {
		t.Errorf("anvelink.openapi.aes-key = %q, want %q", got, "13dade85494f4b2ca4f49b21cd058625")
	}

	// spring.* is unmapped, so the "${...}" placeholder must NOT have leaked into
	// our top-level Redis config (which is fed only by redis.yaml / env).
	if GlobalConfig.Redis.Host != "" {
		t.Errorf("Redis.Host should stay empty (spring.redis is ignored), got %q", GlobalConfig.Redis.Host)
	}
}
