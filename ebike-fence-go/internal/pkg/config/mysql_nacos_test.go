package config

import "testing"

func TestBuildDSNFromNacosTemplate(t *testing.T) {
	ds := mysqlDataSource{
		Host:     "127.0.0.1",
		Port:     3306,
		Database: "ebike_fence",
		Username: "root",
		Password: "secret",
	}
	dsn := ds.buildDSN(defaultMySQLParams)
	want := "root:secret@tcp(127.0.0.1:3306)/ebike_fence?" + defaultMySQLParams
	if dsn != want {
		t.Fatalf("got %q want %q", dsn, want)
	}
}

func TestApplyMySQLYAML(t *testing.T) {
	yaml := `
mysql:
  connection-param: useSSL=false&parseTime=True
  ebike_fence:
    host: db.host
    port: 3306
    database: ebike_fence
    username: u
    password: p
`
	GlobalConfig.Nacos.Group = "xyy"
	if err := applyMySQLYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if GlobalConfig.MySQL.DSN == "" {
		t.Fatal("expected dsn to be set")
	}
}
