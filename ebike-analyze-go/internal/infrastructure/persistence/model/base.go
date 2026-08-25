package model

import (
	"database/sql"
	"time"
)

// BaseDO contains common fields for all database objects.
// Aligns with Java BaseDO.
type BaseDO struct {
	ID         int64         `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID   string        `gorm:"column:tenant_id"`
	CreatedPin string        `gorm:"column:created_pin"`
	CreatedAt  time.Time     `gorm:"column:created_at;autoCreateTime"`
	UpdatedPin string        `gorm:"column:updated_pin"`
	UpdatedAt  time.Time     `gorm:"column:updated_at;autoUpdateTime"`
	Version    sql.NullInt32 `gorm:"column:version"`
	IzDel      bool          `gorm:"column:iz_del"`
}
