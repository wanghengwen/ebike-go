package mysql

import (
	"log"
	"time"

	"ebike-analyze-go/internal/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the primary connection to ebike_analyze database.
var DB *gorm.DB

// VisualDB is the connection to ebike_visual database.
var VisualDB *gorm.DB

// Init opens MySQL pools for both AnalyzeDSN and VisualDSN when configured.
func Init() {
	if err := openAll(); err != nil {
		log.Fatalf("Failed to connect MySQL: %v", err)
	}
}

// Reinit closes existing pools and reconnects using the latest GlobalConfig.MySQL DSNs.
func Reinit() error {
	analyzeDSN := config.GlobalConfig.MySQL.AnalyzeDSN
	visualDSN := config.GlobalConfig.MySQL.VisualDSN

	var newDB, newVisualDB *gorm.DB
	var err error

	if analyzeDSN != "" {
		newDB, err = openDBWithDSN(analyzeDSN)
		if err != nil {
			return err
		}
	}

	if visualDSN != "" {
		newVisualDB, err = openDBWithDSN(visualDSN)
		if err != nil {
			// Close the already-opened analyze DB to avoid leak
			if newDB != nil {
				if sqlDB, e := newDB.DB(); e == nil && sqlDB != nil {
					_ = sqlDB.Close()
				}
			}
			return err
		}
	}

	closeAll()

	DB = newDB
	VisualDB = newVisualDB

	if analyzeDSN == "" {
		log.Println("MySQL AnalyzeDSN not configured; primary DB disabled")
	} else {
		log.Println("MySQL (analyze) reconnected successfully")
	}
	if visualDSN == "" {
		log.Println("MySQL VisualDSN not configured; visual DB disabled")
	} else {
		log.Println("MySQL (visual) reconnected successfully")
	}
	return nil
}

// Close shuts down both MySQL connection pools.
func Close() {
	closeAll()
}

func openAll() error {
	analyzeDSN := config.GlobalConfig.MySQL.AnalyzeDSN
	visualDSN := config.GlobalConfig.MySQL.VisualDSN

	if analyzeDSN == "" {
		log.Println("MySQL AnalyzeDSN not configured; primary DB disabled")
	} else {
		db, err := openDBWithDSN(analyzeDSN)
		if err != nil {
			return err
		}
		DB = db
		log.Println("MySQL (analyze) connected successfully")
	}

	if visualDSN == "" {
		log.Println("MySQL VisualDSN not configured; visual DB disabled")
	} else {
		db, err := openDBWithDSN(visualDSN)
		if err != nil {
			return err
		}
		VisualDB = db
		log.Println("MySQL (visual) connected successfully")
	}

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
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	return db, nil
}

func closeAll() {
	closeDB(&DB)
	closeDB(&VisualDB)
}

func closeDB(dbPtr **gorm.DB) {
	if *dbPtr != nil {
		if sqlDB, err := (*dbPtr).DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		*dbPtr = nil
	}
}
