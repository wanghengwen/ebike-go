package model

import (
	"time"

	"ebike-device-worker-go/internal/pkg/db"
)

// EBikeGpsDO maps to ebike_gps table
type EBikeGpsDO struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	Imei       string    `gorm:"column:imei"`
	Geometry   string    `gorm:"column:geometry"` // Store as WKT or EWKT string depending on query
	Metric     string    `gorm:"column:metric;type:jsonb"`
	Timestamp  time.Time `gorm:"column:timestamp"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (EBikeGpsDO) TableName() string {
	return db.PgTable("ebike_gps")
}

// EBikeItineraryDO maps to ebike_itinerary table
type EBikeItineraryDO struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	OrderId    string    `gorm:"column:order_id"`
	Imei       string    `gorm:"column:imei"`
	StartTime  time.Time `gorm:"column:start_time"`
	EndTime    time.Time `gorm:"column:end_time"`
	Geometry   string    `gorm:"column:geometry"` // Store as WKT or EWKT string
	Metric     string    `gorm:"column:metric;type:jsonb"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (EBikeItineraryDO) TableName() string {
	return db.PgTable("ebike_itinerary")
}
