package model

import (
	"gorm.io/gorm"
)

// InsertShadowRecord inserts a dry-run request into the shadow record list
func InsertShadowRecord(db *gorm.DB, record *ShadowRecordList) error {
	return db.Create(record).Error
}

// InsertCallRecord inserts a real call record
func InsertCallRecord(db *gorm.DB, record *CallRecord) error {
	return db.Create(record).Error
}

// InsertChargeRecord inserts a real charge record
func InsertChargeRecord(db *gorm.DB, record *ChargeRecord) error {
	return db.Create(record).Error
}

// GetTenantConfig retrieves the tenant config
func GetTenantConfig(db *gorm.DB, tenantId string) (*TenantConfig, error) {
	var config TenantConfig
	err := db.Where("tenant_id = ? AND iz_del = 0", tenantId).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetSupplier retrieves supplier by ID
func GetSupplier(db *gorm.DB, id int64) (*Supplier, error) {
	var supplier Supplier
	err := db.Where("id = ? AND iz_del = 0", id).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}
