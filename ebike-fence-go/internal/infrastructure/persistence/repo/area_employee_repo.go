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

type AreaEmployeeRepository struct {
	db *gorm.DB
}

func NewAreaEmployeeRepository(db *gorm.DB) *AreaEmployeeRepository {
	return &AreaEmployeeRepository{db: db}
}

func (r *AreaEmployeeRepository) ListByAreaID(ctx context.Context, areaID int64) ([]model.TAreaEmployee, error) {
	var rows []model.TAreaEmployee
	err := r.db.WithContext(ctx).
		Where("area_id = ? AND (iz_del = 0 OR iz_del IS NULL)", areaID).
		Find(&rows).Error
	return rows, err
}

func (r *AreaEmployeeRepository) ListByAreaIDs(ctx context.Context, areaIDs []int64) ([]model.TAreaEmployee, error) {
	if len(areaIDs) == 0 {
		return nil, nil
	}
	var rows []model.TAreaEmployee
	err := r.db.WithContext(ctx).
		Where("area_id IN ? AND (iz_del = 0 OR iz_del IS NULL)", areaIDs).
		Find(&rows).Error
	return rows, err
}

func (r *AreaEmployeeRepository) DeleteByAreaID(ctx context.Context, areaID int64) error {
	return r.db.WithContext(ctx).
		Where("area_id = ?", areaID).
		Delete(&model.TAreaEmployee{}).Error
}

func (r *AreaEmployeeRepository) InsertBatch(ctx context.Context, tenantID string, areaID int64, employees []model.TAreaEmployee) error {
	for i := range employees {
		employees[i].ID = idgen.NextID()
		employees[i].TenantID = tenantID
		employees[i].AreaID = areaID
		employees[i].IzDel = sql.NullBool{Bool: false, Valid: true}
		now := time.Now().UTC()
		employees[i].CreatedAt = sql.NullTime{Time: now, Valid: true}
		employees[i].UpdatedAt = sql.NullTime{Time: now, Valid: true}
		if err := r.db.WithContext(ctx).Create(&employees[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *AreaEmployeeRepository) GetByID(ctx context.Context, id int64) (*model.TAreaEmployee, error) {
	var row model.TAreaEmployee
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AreaEmployeeRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.TAreaEmployee{}, id).Error
}

func (r *AreaEmployeeRepository) ListByUserPin(ctx context.Context, tenantID, userPin string, serviceAreaID int64, typ *int) ([]model.TAreaEmployee, error) {
	var rows []model.TAreaEmployee
	q := r.db.WithContext(ctx).
		Where("user_pin = ? AND (iz_del = 0 OR iz_del IS NULL)", userPin)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if typ != nil {
		q = q.Where("type = ?", *typ)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	if serviceAreaID <= 0 {
		return rows, nil
	}
	// Filter by maintain areas belonging to service area via fence table join is done at service layer.
	return rows, nil
}
