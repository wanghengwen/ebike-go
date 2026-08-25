package db

import (
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestPgTableNoSuffix(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	if got := PgTable("ebike_gps"); got != "ebike_gps" {
		t.Fatalf("got %q", got)
	}
}

func TestPgTableWithSuffix(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.PersistConfig.TableSuffix = "_new"
	if got := PgTable("ebike_ping"); got != "ebike_ping_new" {
		t.Fatalf("got %q", got)
	}
}

func TestPgTableNilConfig(t *testing.T) {
	config.GlobalConfig = nil
	if got := PgTable("ebike_bms"); got != "ebike_bms" {
		t.Fatalf("got %q", got)
	}
}
