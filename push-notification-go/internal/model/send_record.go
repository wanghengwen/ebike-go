package model

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SendRecord represents the t_send_record table (with monthly partitioning).
type SendRecord struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantId             string    `gorm:"column:tenant_id" json:"tenantId"`
	SupplierId           int64     `gorm:"column:supplier_id" json:"supplierId"`
	Type                 int       `gorm:"column:type" json:"type"`
	Phone                string    `gorm:"column:phone" json:"phone"`
	SignName             string    `gorm:"column:sign_name" json:"signName"`
	SupplierTemplateCode string    `gorm:"column:supplier_template_code" json:"supplierTemplateCode"`
	MessageContent       string    `gorm:"column:message_content" json:"messageContent"`
	Status               int       `gorm:"column:status" json:"status"` // 0=失败 1=成功
	Remark               string    `gorm:"column:remark" json:"remark"`
	SendDatetime         time.Time `gorm:"column:send_datetime" json:"sendDatetime"`
}

// TableName returns the base table name (used as template for monthly partitions).
func (SendRecord) TableName() string {
	return "t_send_record"
}

// DynamicTableName returns the monthly partition table name: t_send_record_2026_06
// Corresponds to Java DynamicTableNameInnerInterceptor in MybatisPlusConfig.
func DynamicTableName() string {
	return "t_send_record_" + time.Now().Format("2006_01")
}

// --- Auto table creation ---

var (
	createdTables = make(map[string]struct{})
	tableMu       sync.Mutex
)

// EnsureTableExists creates the monthly partition table if it doesn't exist.
// Uses CREATE TABLE IF NOT EXISTS ... LIKE t_send_record to copy structure.
// Tracks created tables in-memory to avoid repeated DDL calls.
func EnsureTableExists(db *gorm.DB, tableName string) {
	tableMu.Lock()
	defer tableMu.Unlock()

	if _, ok := createdTables[tableName]; ok {
		return // Already created in this process lifetime
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` LIKE `t_send_record`", tableName)
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("WARNING: auto-create table %s failed: %v", tableName, err)
		return
	}

	createdTables[tableName] = struct{}{}
	log.Printf("Auto-created table: %s", tableName)
}

// InsertSendRecord inserts a send record into the dynamic monthly table.
func InsertSendRecord(db *gorm.DB, record *SendRecord) error {
	tableName := DynamicTableName()
	EnsureTableExists(db, tableName)
	return db.Table(tableName).Create(record).Error
}
