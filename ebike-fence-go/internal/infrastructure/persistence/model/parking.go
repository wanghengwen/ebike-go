package model

import "database/sql"

// TParking maps t_parking (Java ParkingDetailDO).
type TParking struct {
	ID             int64          `gorm:"column:id;primaryKey"`
	TenantID       string         `gorm:"column:tenant_id"`
	Imei           string         `gorm:"column:imei"`
	ServiceID      int64          `gorm:"column:service_id"`
	CarID          string         `gorm:"column:car_id"`
	ParkingID      int64          `gorm:"column:parking_id"`
	NoParkingID    int64          `gorm:"column:no_parking_id"`
	BanRidingID    int64          `gorm:"column:ban_riding_id"`
	MaintainAreaID int64          `gorm:"column:maintain_area_id"`
	NearParkingID  int64          `gorm:"column:near_parking_id"`
	FenceCustomID  int64          `gorm:"column:fence_custom_id"`
	IzDel          sql.NullBool   `gorm:"column:iz_del"`
	CreatedAt      sql.NullTime   `gorm:"column:created_at"`
	UpdatedAt      sql.NullTime   `gorm:"column:updated_at"`
	CreatedPin     sql.NullString `gorm:"column:created_pin"`
	UpdatedPin     sql.NullString `gorm:"column:updated_pin"`
	Version        sql.NullInt32  `gorm:"column:version"`
}

func (TParking) TableName() string { return "t_parking" }
