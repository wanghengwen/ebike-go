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

type SiteApplicationRepository struct {
	db *gorm.DB
}

func NewSiteApplicationRepository(db *gorm.DB) *SiteApplicationRepository {
	return &SiteApplicationRepository{db: db}
}

func (r *SiteApplicationRepository) Insert(ctx context.Context, tenantID, pin string, row *model.TSiteApplication) (int64, error) {
	if row == nil {
		return 0, nil
	}
	row.ID = idgen.NextID()
	row.TenantID = tenantID
	row.UserPin = sql.NullString{String: pin, Valid: pin != ""}
	now := time.Now().UTC()
	row.CreatedAt = sql.NullTime{Time: now, Valid: true}
	row.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	row.IzDel = sql.NullBool{Bool: false, Valid: true}
	row.CreatedPin = sql.NullString{String: pin, Valid: pin != ""}
	row.UpdatedPin = sql.NullString{String: pin, Valid: pin != ""}
	row.Version = sql.NullInt32{Int32: 1, Valid: true}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *SiteApplicationRepository) Update(ctx context.Context, row *model.TSiteApplication) error {
	if row == nil || row.ID == 0 {
		return nil
	}
	row.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *SiteApplicationRepository) GetByID(ctx context.Context, id int64) (*model.TSiteApplication, error) {
	var row model.TSiteApplication
	err := r.db.WithContext(ctx).Where("id = ? AND (iz_del = 0 OR iz_del IS NULL)", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type SiteApplicationPageQuery struct {
	ServiceID        int64
	Phone            string
	State            *int
	OpManPhone       string
	CreatedTimeStart *time.Time
	CreatedTimeEnd   *time.Time
	PageNum          int
	PageSize         int
}

func (r *SiteApplicationRepository) Page(ctx context.Context, tenantID string, q SiteApplicationPageQuery) ([]model.TSiteApplication, int64, error) {
	if q.PageNum <= 0 {
		q.PageNum = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	dbq := r.db.WithContext(ctx).Where("(iz_del = 0 OR iz_del IS NULL)")
	if tenantID != "" {
		dbq = dbq.Where("tenant_id = ?", tenantID)
	}
	if q.ServiceID > 0 {
		dbq = dbq.Where("service_id = ?", q.ServiceID)
	}
	if q.State != nil {
		dbq = dbq.Where("state = ?", *q.State)
	}
	if q.CreatedTimeStart != nil {
		dbq = dbq.Where("created_at >= ?", *q.CreatedTimeStart)
	}
	if q.CreatedTimeEnd != nil {
		dbq = dbq.Where("created_at <= ?", *q.CreatedTimeEnd)
	}
	var total int64
	if err := dbq.Model(&model.TSiteApplication{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.TSiteApplication
	offset := (q.PageNum - 1) * q.PageSize
	if err := dbq.Offset(offset).Limit(q.PageSize).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
