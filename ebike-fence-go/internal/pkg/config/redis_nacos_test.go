package config

import "testing"

func TestApplyRedisYAMLDatabaseField(t *testing.T) {
	content := `
redis:
  ebike_fence:
    host: 127.0.0.1
    port: 6379
    password: secret
    database: 6
`
	if err := applyRedisYAML(content); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.Redis.DB != 6 {
		t.Fatalf("expected db 6 from database field, got %d", GlobalConfig.Redis.DB)
	}
	if GlobalConfig.Redis.Password != "secret" {
		t.Fatalf("unexpected password %q", GlobalConfig.Redis.Password)
	}
}

func TestApplyRedisYAMLPrefersDbOverDatabase(t *testing.T) {
	content := `
redis:
  ebike_fence:
    host: 127.0.0.1
    db: 3
    database: 6
`
	if err := applyRedisYAML(content); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.Redis.DB != 3 {
		t.Fatalf("expected db 3 when both set, got %d", GlobalConfig.Redis.DB)
	}
}
