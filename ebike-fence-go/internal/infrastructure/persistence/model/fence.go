package model

import (
	"database/sql"
	"time"
)

// TFence maps t_fence (Java FenceDO).
type TFence struct {
	ID                         int64           `gorm:"column:id;primaryKey"`
	TenantID                   string          `gorm:"column:tenant_id"`
	Type                       int             `gorm:"column:type"`
	CustomTypeID               sql.NullInt64   `gorm:"column:custom_type_id"`
	Name                       string          `gorm:"column:name"`
	ShapeType                  string          `gorm:"column:shape_type"`
	MaxParkingNumber           int             `gorm:"column:max_parking_number"`
	CenterLat                  float64         `gorm:"column:center_lat"`
	CenterLng                  float64         `gorm:"column:center_lng"`
	PointList                  string          `gorm:"column:point_list"`
	ServiceID                  int64           `gorm:"column:service_id"`
	Tbeacon                    sql.NullBool    `gorm:"column:tbeacon"`
	Directional                sql.NullBool    `gorm:"column:directional"`
	Direction                  sql.NullFloat64 `gorm:"column:direction"`
	FormulateDirection         sql.NullFloat64 `gorm:"column:formulate_direction"`
	Rfid                       sql.NullBool    `gorm:"column:rfid"`
	IzEnable                   sql.NullBool    `gorm:"column:iz_enable"`
	CoefficientOfDifficult     sql.NullFloat64 `gorm:"column:coefficient_of_difficult"`
	BufferDistance             sql.NullFloat64 `gorm:"column:buffer_distance"`
	Camera                     sql.NullBool    `gorm:"column:camera"`
	IzCameraDirectionalBackcar sql.NullBool    `gorm:"column:iz_camera_directional_backcar"`
	IzCameraPointBackcar       sql.NullBool    `gorm:"column:iz_camera_point_backcar"`
	Kickstand                  sql.NullBool    `gorm:"column:kickstand"`
	Pics                       sql.NullString  `gorm:"column:pics"`
	DataVersion                sql.NullInt64   `gorm:"column:data_version"`
	AreaSize                   sql.NullFloat64 `gorm:"column:area_size"`
	OpeningHoursBegin          sql.NullString  `gorm:"column:opening_hours_begin"`
	OpeningHoursEnd            sql.NullString  `gorm:"column:opening_hours_end"`
	IzFullPileNoStop           sql.NullInt32   `gorm:"column:iz_full_pile_no_stop"`
	Level                      sql.NullInt32   `gorm:"column:level"`
	IzOpenAllDay               sql.NullBool    `gorm:"column:iz_open_all_day"`
	ActivityID                 sql.NullInt64   `gorm:"column:activity_id"`
	MinAmount                  sql.NullInt32   `gorm:"column:min_amount"`
	MaxAmount                  sql.NullInt32   `gorm:"column:max_amount"`
	ExpirationTime             sql.NullTime    `gorm:"column:expiration_time"`
	CreatedPin                 sql.NullString  `gorm:"column:created_pin"`
	UpdatedPin                 sql.NullString  `gorm:"column:updated_pin"`
	Version                    sql.NullInt32   `gorm:"column:version"`
	IzDel                      sql.NullBool    `gorm:"column:iz_del"`
	CreatedAt                  sql.NullTime    `gorm:"column:created_at"`
	UpdatedAt                  sql.NullTime    `gorm:"column:updated_at"`
}

func (TFence) TableName() string { return "t_fence" }

// TFenceRfid maps t_fence_rfid.
type TFenceRfid struct {
	ID         int64          `gorm:"column:id;primaryKey"`
	TenantID   string         `gorm:"column:tenant_id"`
	FenceID    int64          `gorm:"column:fence_id"`
	RfidCode   string         `gorm:"column:rfid_code"`
	IzDel      sql.NullBool   `gorm:"column:iz_del"`
	CreatedAt  sql.NullTime   `gorm:"column:created_at"`
	UpdatedAt  sql.NullTime   `gorm:"column:updated_at"`
	CreatedPin sql.NullString `gorm:"column:created_pin"`
	UpdatedPin sql.NullString `gorm:"column:updated_pin"`
	Version    sql.NullInt32  `gorm:"column:version"`
}

func (TFenceRfid) TableName() string { return "t_fence_rfid" }

// TFenceCustomType maps t_fence_custom_type.
type TFenceCustomType struct {
	ID             int64          `gorm:"column:id;primaryKey"`
	TenantID       string         `gorm:"column:tenant_id"`
	Name           string         `gorm:"column:name"`
	Description    sql.NullString `gorm:"column:description"`
	Color          sql.NullString `gorm:"column:color"`
	IzCreateCarTag sql.NullBool   `gorm:"column:iz_create_car_tag"`
	CarTag         sql.NullString `gorm:"column:car_tag"`
	IzDel          sql.NullBool   `gorm:"column:iz_del"`
	CreatedAt      sql.NullTime   `gorm:"column:created_at"`
	UpdatedAt      sql.NullTime   `gorm:"column:updated_at"`
	CreatedPin     sql.NullString `gorm:"column:created_pin"`
	UpdatedPin     sql.NullString `gorm:"column:updated_pin"`
	Version        sql.NullInt32  `gorm:"column:version"`
}

func (TFenceCustomType) TableName() string { return "t_fence_custom_type" }

// TAreaEmployee maps t_area_employee.
type TAreaEmployee struct {
	ID         int64          `gorm:"column:id;primaryKey"`
	TenantID   string         `gorm:"column:tenant_id"`
	AreaID     int64          `gorm:"column:area_id"`
	UserPin    string         `gorm:"column:user_pin"`
	Name       sql.NullString `gorm:"column:name"`
	Phone      sql.NullString `gorm:"column:phone"`
	RoleName   sql.NullString `gorm:"column:role_name"`
	Type       sql.NullInt32  `gorm:"column:type"`
	IzDel      sql.NullBool   `gorm:"column:iz_del"`
	CreatedAt  sql.NullTime   `gorm:"column:created_at"`
	UpdatedAt  sql.NullTime   `gorm:"column:updated_at"`
	CreatedPin sql.NullString `gorm:"column:created_pin"`
	UpdatedPin sql.NullString `gorm:"column:updated_pin"`
	Version    sql.NullInt32  `gorm:"column:version"`
}

func (TAreaEmployee) TableName() string { return "t_area_employee" }

// TSiteApplication maps t_site_application.
type TSiteApplication struct {
	ID              int64           `gorm:"column:id;primaryKey"`
	TenantID        string          `gorm:"column:tenant_id"`
	ServiceID       int64           `gorm:"column:service_id"`
	Location        sql.NullString  `gorm:"column:location"`
	Lat             sql.NullFloat64 `gorm:"column:lat"`
	Lng             sql.NullFloat64 `gorm:"column:lng"`
	UserPin         sql.NullString  `gorm:"column:user_pin"`
	PhotoURL        sql.NullString  `gorm:"column:photo_url"`
	ApplicantRemark sql.NullString  `gorm:"column:applicant_remark"`
	State           sql.NullInt32   `gorm:"column:state"`
	OpManPin        sql.NullString  `gorm:"column:op_man_pin"`
	OpManRemark     sql.NullString  `gorm:"column:op_man_remark"`
	IzDel           sql.NullBool    `gorm:"column:iz_del"`
	CreatedAt       sql.NullTime    `gorm:"column:created_at"`
	UpdatedAt       sql.NullTime    `gorm:"column:updated_at"`
	CreatedPin      sql.NullString  `gorm:"column:created_pin"`
	UpdatedPin      sql.NullString  `gorm:"column:updated_pin"`
	Version         sql.NullInt32   `gorm:"column:version"`
}

func (TSiteApplication) TableName() string { return "t_site_application" }

func NowUTC() time.Time { return time.Now().UTC() }
