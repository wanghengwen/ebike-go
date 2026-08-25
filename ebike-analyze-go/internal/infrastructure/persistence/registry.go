package persistence

import (
	"ebike-analyze-go/internal/infrastructure/persistence/tenant"
	"ebike-analyze-go/internal/pkg/config"
	"ebike-analyze-go/internal/pkg/mysql"
)

// Init initializes persistence config and registers GORM tenant callbacks
// on both analyze and visual DB pools (mirrors Java TenantLineInnerInterceptor).
// Called after mysql.Init() in main.go.
func Init() {
	tenant.InitFromConfig(config.GlobalConfig.Xyy.TableSplit.Enable, config.GlobalConfig.Xyy.TableSplit.TableNames)
	tenant.RegisterCallbacks(mysql.DB)
	tenant.RegisterCallbacks(mysql.VisualDB)
}
