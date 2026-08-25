package model

import "time"

// BaseDO mirrors Java BaseDO common columns.
type BaseDO struct {
	TenantID   string     `gorm:"column:tenant_id"`
	CreatedPin string     `gorm:"column:created_pin"`
	CreatedAt  *time.Time `gorm:"column:created_at"`
	UpdatedPin string     `gorm:"column:updated_pin"`
	UpdatedAt  *time.Time `gorm:"column:updated_at"`
	Version    *int       `gorm:"column:version"`
	IzDel      *bool      `gorm:"column:iz_del"`
}
