package mysql

import (
	"log"
	"time"

	"ebike-fence-go/internal/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init opens a MySQL pool when dsn is configured (typically from Nacos mysql.yaml).
// Empty dsn skips DB and keeps RPC fallback.
func Init() {
	if err := openDB(); err != nil {
		log.Fatalf("Failed to connect MySQL: %v", err)
	}
}

// Reinit closes the current pool and reconnects using the latest GlobalConfig.MySQL.DSN.
// On failure the previous pool is kept when possible.
func Reinit() error {
	dsn := config.GlobalConfig.MySQL.DSN
	if dsn == "" {
		closeDB()
		log.Println("MySQL dsn not configured; DB read path disabled")
		return nil
	}

	db, err := openDBWithDSN(dsn)
	if err != nil {
		return err
	}

	closeDB()
	DB = db
	log.Println("MySQL reconnected successfully")
	return nil
}

func openDB() error {
	dsn := config.GlobalConfig.MySQL.DSN
	if dsn == "" {
		log.Println("MySQL dsn not configured; DB read path disabled")
		return nil
	}
	db, err := openDBWithDSN(dsn)
	if err != nil {
		return err
	}
	DB = db
	log.Println("MySQL connected successfully")
	return nil
}

func openDBWithDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	maxOpen := config.GlobalConfig.MySQL.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 20
	}
	maxIdle := config.GlobalConfig.MySQL.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func closeDB() {
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		DB = nil
	}
}
