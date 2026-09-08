package model

import (
	"time"
)

type BaseEntity struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
	Version   int       `gorm:"column:version" json:"version"`
	IzDel     int       `gorm:"column:iz_del" json:"izDel"`
	TraceId   string    `gorm:"column:trace_id" json:"traceId"`
}

type CallRecord struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantId     string    `gorm:"column:tenant_id" json:"tenantId"`
	SupplierId   int64     `gorm:"column:supplier_id" json:"supplierId"`
	Name         string    `gorm:"column:name" json:"name"`
	IdentityNo   string    `gorm:"column:identity_no" json:"identityNo"`
	ImageUrl     string    `gorm:"column:image_url" json:"imageUrl"`
	CallAt       time.Time `gorm:"column:call_at" json:"callAt"`
	Status       int       `gorm:"column:status" json:"status"` // 0: success, 1: fail
	UniqueId     string    `gorm:"column:unique_id" json:"uniqueId"`
	Score        int       `gorm:"column:score" json:"score"`
	ResponseData string    `gorm:"column:response_data" json:"responseData"` // JSON string
	TraceId      string    `gorm:"column:trace_id" json:"traceId"`
}

// Note: CallRecord does NOT have a fixed TableName() because Java uses
// dynamic per-month tables: t_two_call_record_{yyyy_MM}, t_three_call_record_{yyyy_MM}.
// All queries use raw SQL with dynamic table names.

type ChargeRecord struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantId string `gorm:"column:tenant_id" json:"tenantId"`
	Amount   string `gorm:"column:amount" json:"amount"` // Use string to match Java BigDecimal
	Quantity int64  `gorm:"column:quantity" json:"quantity"`
	Type     int    `gorm:"column:type" json:"type"`
	BaseEntity
}

func (ChargeRecord) TableName() string {
	return "t_charge_record"
}

type Supplier struct {
	ID      int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Channel int    `gorm:"column:channel" json:"channel"`
	Config  string `gorm:"column:config" json:"config"` // JSON string
	BaseEntity
}

func (Supplier) TableName() string {
	return "t_supplier"
}

type TenantConfig struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantId   string `gorm:"column:tenant_id" json:"tenantId"`
	Type       int    `gorm:"column:type" json:"type"`       // 1: two-element, 2: three-element
	Pattern    int    `gorm:"column:pattern" json:"pattern"` // 1: prepay, 2: postpay
	SupplierId int64  `gorm:"column:supplier_id" json:"supplierId"`
	BaseEntity
}

func (TenantConfig) TableName() string {
	return "t_tenant_config"
}

// ShadowRecordList holds the dry-run operations for mutating requests like /auth and /charge
type ShadowRecordList struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RequestPath string    `gorm:"column:request_path" json:"requestPath"`
	Payload     string    `gorm:"column:payload" json:"payload"` // JSON string of the request
	TenantId    string    `gorm:"column:tenant_id" json:"tenantId"`
	TraceId     string    `gorm:"column:trace_id" json:"traceId"`
	Type        int       `gorm:"column:type" json:"type"` // 1=二要素, 2=三要素
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ShadowRecordList) TableName() string {
	return "t_shadow_record_list"
}
