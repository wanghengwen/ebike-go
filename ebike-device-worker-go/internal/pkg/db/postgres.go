package db

import (
	"log"
	"time"

	"ebike-device-worker-go/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	if config.GlobalConfig.Database.Dsn == "" {
		log.Println("Warning: Database DSN is empty, skipping DB initialization")
		return
	}

	db, err := gorm.Open(postgres.Open(config.GlobalConfig.Database.Dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	log.Println("Successfully connected to PostgreSQL")

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Warning: failed to get sql.DB: %v", err)
		return
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	log.Println("Database connection pool limits configured (MaxIdle: 10, MaxOpen: 100)")
}
