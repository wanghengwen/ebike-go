package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/idgen"

	"gorm.io/gorm"
)

type FenceCustomTypeRepository struct {
	db *gorm.DB
}

func NewFenceCustomTypeRepository(db *gorm.DB) *FenceCustomTypeRepository {
	return &FenceCustomTypeRepository{db: db}
}

func (r *FenceCustomTypeRepository) GetAll(ctx context.Context, tenantID string) ([]model.TFenceCustomType, error) {
	var rows []model.TFenceCustomType
	q := r.db.WithContext(ctx).Where("(iz_del = 0 OR iz_del IS NULL)")
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *FenceCustomTypeRepository) GetByID(ctx context.Context, id int64) (*model.TFenceCustomType, error) {
	var row model.TFenceCustomType
	err := r.db.WithContext(ctx).Where("id = ? AND (iz_del = 0 OR iz_del IS NULL)", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FenceCustomTypeRepository) GetByIDs(ctx context.Context, tenantID string, ids []int64) ([]model.TFenceCustomType, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []model.TFenceCustomType
	q := r.db.WithContext(ctx).Where("id IN ? AND (iz_del = 0 OR iz_del IS NULL)", ids)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *FenceCustomTypeRepository) GetByName(ctx context.Context, tenantID, name string) (*model.TFenceCustomType, error) {
	var row model.TFenceCustomType
	err := r.db.WithContext(ctx).
		Where("name = ? AND tenant_id = ? AND (iz_del = 0 OR iz_del IS NULL)", name, tenantID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *FenceCustomTypeRepository) SaveOrUpdate(ctx context.Context, tenantID string, row *model.TFenceCustomType) (int64, error) {
	if row == nil {
		return 0, nil
	}
	row.TenantID = tenantID
	now := time.Now().UTC()
	if row.ID == 0 {
		row.ID = idgen.NextID()
		row.CreatedAt = sql.NullTime{Time: now, Valid: true}
		row.UpdatedAt = sql.NullTime{Time: now, Valid: true}
		row.IzDel = sql.NullBool{Bool: false, Valid: true}
		if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
			return 0, err
		}
		return row.ID, nil
	}
	updates := map[string]interface{}{
		"name":              row.Name,
		"description":       row.Description,
		"color":             row.Color,
		"iz_create_car_tag": row.IzCreateCarTag,
		"car_tag":           row.CarTag,
		"updated_at":        now,
		"updated_pin":       row.UpdatedPin,
	}
	if err := r.db.WithContext(ctx).Model(&model.TFenceCustomType{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *FenceCustomTypeRepository) Delete(ctx context.Context, tenantID string, id int64) error {
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	return q.Delete(&model.TFenceCustomType{}).Error
}
