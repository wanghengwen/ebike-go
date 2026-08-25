package config

import "testing"

func TestApplyMySQLYAML_JavaLayout(t *testing.T) {
	yaml := `
mysql:
  driver-class-name: com.mysql.cj.jdbc.Driver
  type: com.zaxxer.hikari.HikariDataSource
  connection-param: useSSL=false&useUnicode=true&characterEncoding=UTF-8&serverTimezone=GMT%2B8
  ebike_analyze:
    host: db.analyze
    port: 3306
    database: ebike_analyze
    username: analyze_user
    password: analyze_pass
  ebike_visual:
    host: db.visual
    port: 3306
    database: ebike_visual
    username: visual_user
    password: visual_pass
`
	GlobalConfig.Nacos.Group = "xyy"
	if err := applyMySQLYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.MySQL.AnalyzeDSN == "" {
		t.Fatal("expected analyzeDSN")
	}
	if GlobalConfig.MySQL.VisualDSN == "" {
		t.Fatal("expected visualDSN")
	}
}

func TestApplyRedisYAML_EbikeOrderKey(t *testing.T) {
	yaml := `
redis:
  ebike_order:
    host: 127.0.0.1
    port: 6379
    password: order-secret
`
	GlobalConfig.Nacos.Group = "xyy"
	if err := applyRedisYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.Redis.Password != "order-secret" {
		t.Fatalf("password=%q", GlobalConfig.Redis.Password)
	}
	if GlobalConfig.Redis.DB != 13 {
		t.Fatalf("expected default db 13 for ebike_order, got %d", GlobalConfig.Redis.DB)
	}
}

func TestApplyRedisYAML_EbikeAnalyzeHyphenKey(t *testing.T) {
	yaml := `
redis:
  ebike-analyze:
    host: 127.0.0.1
    port: 6379
    password: secret
    database: 2
`
	GlobalConfig.Nacos.Group = "xyy"
	if err := applyRedisYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.Redis.Password != "secret" {
		t.Fatalf("password=%q", GlobalConfig.Redis.Password)
	}
	if GlobalConfig.Redis.DB != 2 {
		t.Fatalf("db=%d", GlobalConfig.Redis.DB)
	}
}

func TestApplyRedisYAML_EbikeAnalyzeSnakeKey(t *testing.T) {
	yaml := `
redis:
  ebike_analyze:
    host: 127.0.0.1
    port: 6379
    password: snake
    db: 1
`
	if err := applyRedisYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.Redis.Password != "snake" || GlobalConfig.Redis.DB != 1 {
		t.Fatalf("addr=%s password=%q db=%d", GlobalConfig.Redis.Addr, GlobalConfig.Redis.Password, GlobalConfig.Redis.DB)
	}
}
