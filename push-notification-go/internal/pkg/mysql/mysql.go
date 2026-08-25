package mysql

import (
	"log"
	"push-notification-go/internal/pkg/config"
	"time"

	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database handle.
var DB *gorm.DB

// Init opens a new database connection using the DSN from GlobalConfig and
// configures the connection pool.
func Init() {
	dsn := config.GlobalConfig.MySQL.DSN
	if dsn == "" {
		log.Printf("[mysql] DSN is empty, skipping database initialization")
		return
	}

	var err error
	DB, err = gorm.Open(mysqldriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[mysql] failed to connect: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("[mysql] failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(config.GlobalConfig.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.GlobalConfig.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Printf("[mysql] connected, maxOpen=%d maxIdle=%d connMaxLife=30m",
		config.GlobalConfig.MySQL.MaxOpenConns, config.GlobalConfig.MySQL.MaxIdleConns)
}

// Reinit closes the existing connection and reconnects with the current DSN.
// This is intended to be called when the Nacos mysql config changes.
func Reinit() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("[mysql] error closing old connection: %v", err)
			}
		}
		DB = nil
		log.Printf("[mysql] closed old connection for reinit")
	}
	Init()
}
