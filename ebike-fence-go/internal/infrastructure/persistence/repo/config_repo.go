package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/middleware"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigRepository provides generic CRUD for service-scoped config rows.
type ConfigRepository struct {
	db    *gorm.DB
	store *gateway.ConfigStore
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db, store: gateway.NewConfigStore(db)}
}

func (r *ConfigRepository) DB() *gorm.DB                { return r.db }
func (r *ConfigRepository) Store() *gateway.ConfigStore { return r.store }

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("iz_del = 0 OR iz_del IS NULL")
}

// GetLatestByServiceID loads the newest row for a service-scoped config table.
func (r *ConfigRepository) GetLatestByServiceID(ctx context.Context, tenantID string, dest interface{}, serviceID int64) error {
	q := notDeleted(r.db.WithContext(ctx)).Where("service_id = ?", serviceID)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	// Java Config*QueryImpl: orderByDesc(updatedAt).last("limit 1") only.
	err := q.Order("updated_at DESC").
		Limit(1).
		First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// GetByID loads a row by primary key, optionally scoped by tenant.
func (r *ConfigRepository) GetByID(ctx context.Context, tenantID string, dest interface{}, id int64) error {
	q := notDeleted(r.db.WithContext(ctx)).Where("id = ?", id)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// GetLatestTenantRow loads the newest row for tenant-scoped tables (credit score, big screen).
func (r *ConfigRepository) GetLatestTenantRow(ctx context.Context, dest interface{}, tenantID string) error {
	err := notDeleted(r.db.WithContext(ctx)).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC, updated_at DESC").
		Limit(1).
		First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// GetLatestAny loads the newest row ignoring tenant (big screen Java behavior).
func (r *ConfigRepository) GetLatestAny(ctx context.Context, dest interface{}) error {
	err := notDeleted(r.db.WithContext(ctx)).
		Order("created_at DESC").
		Limit(1).
		First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// Create inserts a config row with audit fields.
func (r *ConfigRepository) Create(ctx context.Context, row interface{}) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// Save updates a row by primary key.
func (r *ConfigRepository) Save(ctx context.Context, row interface{}) error {
	return r.db.WithContext(ctx).Save(row).Error
}

// UpdateColumns updates selected columns on a model.
func (r *ConfigRepository) UpdateColumns(ctx context.Context, row interface{}, cols map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(row).Updates(cols).Error
}

// UpdateNonZero updates only the non-zero fields of row (matches MyBatis-Plus
// updateById, which skips null fields), using row's primary key as the condition.
func (r *ConfigRepository) UpdateNonZero(ctx context.Context, row interface{}) error {
	return r.db.WithContext(ctx).Model(row).Updates(row).Error
}

// DeleteByID hard-deletes by id.
func (r *ConfigRepository) DeleteByID(ctx context.Context, model interface{}, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(model).Error
}

// InsertVersioned copies the latest service-scoped row, applies mutate, clears ID, inserts.
func (r *ConfigRepository) InsertVersioned(ctx context.Context, tenantID, pin string, serviceID int64, sample interface{}, loadLatest func() (interface{}, error), apply func(newRow interface{})) error {
	latest, err := loadLatest()
	if err != nil {
		return err
	}
	newRow := cloneConfigRow(sample, latest)
	if newRow == nil {
		newRow = sample
	}
	apply(newRow)
	stampAudit(newRow, tenantID, pin, true)
	clearPrimaryKey(newRow)
	return r.Create(ctx, newRow)
}

// StampFromGin fills tenant/pin/timestamp on BaseConfig from gin context.
func StampFromGin(c *gin.Context, base *model.BaseConfig, tenantID, pin string, isNew bool) {
	now := time.Now()
	base.TenantID = tenantID
	if pin == "" {
		pin = "ebike_fence"
	}
	base.UpdatedPin = pin
	base.UpdatedAt = now
	if isNew {
		base.CreatedPin = pin
		base.CreatedAt = now
	}
}

// PinFromContext returns pin from command context.
func PinFromContext(c *gin.Context) string {
	if ctx := middleware.GetCommandContext(c); ctx != nil && ctx.Pin != "" {
		return ctx.Pin
	}
	return "ebike_fence"
}

// NextConfigID generates a time-based id compatible with Java snowflake-style ids.
func NextConfigID() int64 {
	return time.Now().UnixNano() / 1000
}

func stampAudit(row interface{}, tenantID, pin string, isNew bool) {
	switch v := row.(type) {
	case *model.ConfigBackcar:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ConfigPay:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ConfigBaseItem:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ConfigUseCar:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.CreditScoreConfig:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ParkApplyConfig:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.PushRidingCardConfig:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.AdConfig:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.AlarmContact:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.BigScreen:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.RidingPermission:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ConfigProtocol:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	case *model.ResourceManagement:
		StampBase(&v.BaseConfig, tenantID, pin, isNew)
	}
}

func StampBase(base *model.BaseConfig, tenantID, pin string, isNew bool) {
	now := time.Now()
	base.TenantID = tenantID
	if pin == "" {
		pin = "ebike_fence"
	}
	base.UpdatedPin = pin
	base.UpdatedAt = now
	if isNew {
		base.CreatedPin = pin
		base.CreatedAt = now
	}
	if !base.Version.Valid {
		base.Version = sql.NullInt32{Int32: 0, Valid: true}
	}
	if !base.IzDel.Valid {
		base.IzDel = sql.NullBool{Bool: false, Valid: true}
	}
}

func clearPrimaryKey(row interface{}) {
	switch v := row.(type) {
	case *model.ConfigBackcar:
		v.ID = 0
	case *model.ConfigPay:
		v.ID = 0
	case *model.ConfigBaseItem:
		v.ID = 0
	case *model.ConfigUseCar:
		v.ID = 0
	case *model.CreditScoreConfig:
		v.ID = 0
	case *model.ParkApplyConfig:
		v.ID = 0
	case *model.PushRidingCardConfig:
		v.ID = 0
	case *model.AdConfig:
		v.ID = 0
	case *model.AlarmContact:
		v.ID = 0
	case *model.BigScreen:
		v.ID = 0
	case *model.RidingPermission:
		v.ID = 0
	case *model.ConfigProtocol:
		v.ID = 0
	case *model.ResourceManagement:
		v.ID = 0
	}
}

func cloneConfigRow(sample, latest interface{}) interface{} {
	if latest == nil {
		return nil
	}
	switch sample.(type) {
	case *model.ConfigBaseItem:
		if src, ok := latest.(*model.ConfigBaseItem); ok && src != nil {
			cp := *src
			return &cp
		}
	case *model.ConfigUseCar:
		if src, ok := latest.(*model.ConfigUseCar); ok && src != nil {
			cp := *src
			return &cp
		}
	case *model.ConfigPay:
		if src, ok := latest.(*model.ConfigPay); ok && src != nil {
			cp := *src
			return &cp
		}
	}
	return nil
}

// PageAlarmContacts pages alarm contacts by service and type.
func (r *ConfigRepository) PageAlarmContacts(ctx context.Context, serviceID int64, contactType int, pageNum, pageSize int) ([]model.AlarmContact, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := notDeleted(r.db.WithContext(ctx).Model(&model.AlarmContact{})).
		Where("service_id = ? AND type = ?", serviceID, contactType)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AlarmContact
	err := q.Order("id DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

// ListAlarmContacts lists contacts for a service/type.
func (r *ConfigRepository) ListAlarmContacts(ctx context.Context, serviceID int64, contactType int) ([]model.AlarmContact, error) {
	var rows []model.AlarmContact
	err := notDeleted(r.db.WithContext(ctx)).
		Where("service_id = ? AND type = ?", serviceID, contactType).
		Find(&rows).Error
	return rows, err
}

// ListProtocolsByService lists protocol rows for a service.
func (r *ConfigRepository) ListProtocolsByService(ctx context.Context, serviceID int64) ([]model.ConfigProtocol, error) {
	var rows []model.ConfigProtocol
	err := notDeleted(r.db.WithContext(ctx)).
		Where("service_id = ?", serviceID).
		Find(&rows).Error
	return rows, err
}

// GetProtocolByType loads one protocol row.
func (r *ConfigRepository) GetProtocolByType(ctx context.Context, serviceID int64, typ int) (*model.ConfigProtocol, error) {
	var row model.ConfigProtocol
	err := notDeleted(r.db.WithContext(ctx)).
		Where("service_id = ? AND type = ?", serviceID, typ).
		Order("id DESC").
		Limit(1).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// DefaultProtocols loads tenant_id='0' template protocols (Redis cache first, mirrors Java).
func (r *ConfigRepository) DefaultProtocols(ctx context.Context) ([]model.ConfigProtocol, error) {
	key := rediskeys.ProtocolConfigDefaultList()
	if rdb := pkgredis.GetClient(); rdb != nil {
		val, err := rdb.Get(ctx, key).Result()
		if err == nil && !rediskeys.IsCacheMiss(val) {
			var rows []model.ConfigProtocol
			if json.Unmarshal([]byte(val), &rows) == nil && len(rows) > 0 {
				return rows, nil
			}
		}
	}
	var rows []model.ConfigProtocol
	err := r.db.WithContext(ctx).
		Select("title, type, content").
		Where("tenant_id = ?", "0").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		if rdb := pkgredis.GetClient(); rdb != nil {
			if b, err := json.Marshal(rows); err == nil {
				_ = rdb.Set(ctx, key, string(b), 0).Err()
			}
		}
	}
	return rows, err
}

// GetAdConfigLastByTenant returns the latest ad config row for a tenant (Java getLastConfig).
func (r *ConfigRepository) GetAdConfigLastByTenant(ctx context.Context, tenantID string) (*model.AdConfig, error) {
	var row model.AdConfig
	err := notDeleted(r.db.WithContext(ctx)).
		Where("tenant_id = ?", tenantID).
		Order("updated_at DESC").
		Limit(1).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListFenceTags loads all fence tags (tenant filter via MySQL plugin if any).
func (r *ConfigRepository) ListFenceTags(ctx context.Context) ([]model.FenceTag, error) {
	var rows []model.FenceTag
	err := notDeleted(r.db.WithContext(ctx)).Find(&rows).Error
	return rows, err
}

// FilterFenceIDsByTags mirrors Java FenceTagMapper.queryHasTagsFenceByTag.
func (r *ConfigRepository) FilterFenceIDsByTags(ctx context.Context, tagIDs []int, fenceIDs []int64, fenceType int) ([]int64, error) {
	if len(tagIDs) == 0 || len(fenceIDs) == 0 {
		return nil, nil
	}
	type row struct {
		FenceID int64 `gorm:"column:fence_id"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT ref.fence_id
		FROM t_fence_tag t
		LEFT JOIN t_fence_tag_ref ref ON ref.tag_id = t.id
		WHERE ref.fence_type = ? AND ref.tag_id IN ? AND ref.fence_id IN ?
		  AND (ref.iz_del = 0 OR ref.iz_del IS NULL)
		ORDER BY ref.created_at DESC
	`, fenceType, tagIDs, fenceIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[int64]struct{}, len(rows))
	out := make([]int64, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.FenceID]; ok {
			continue
		}
		seen[row.FenceID] = struct{}{}
		out = append(out, row.FenceID)
	}
	return out, nil
}

// ListFenceRefTagsByFenceID mirrors Java FenceTagMapper.queryFenceRefTagsById.
func (r *ConfigRepository) ListFenceRefTagsByFenceID(ctx context.Context, fenceID int64, fenceType int) ([]model.FenceTag, error) {
	// version/created_at/created_pin come from the ref row, not the tag row, matching Java.
	var rows []model.FenceTag
	err := r.db.WithContext(ctx).Raw(`
		SELECT t.id, t.tag_name, t.iz_enable, ref.created_at, ref.version, ref.created_pin
		FROM t_fence_tag t
		LEFT JOIN t_fence_tag_ref ref ON ref.tag_id = t.id
		WHERE ref.fence_id = ? AND ref.fence_type = ? AND (ref.iz_del = 0 OR ref.iz_del IS NULL)
		ORDER BY ref.created_at DESC
	`, fenceID, fenceType).Scan(&rows).Error
	return rows, err
}

// CountAdConfigByService counts rows for service.
func (r *ConfigRepository) CountAdConfigByService(ctx context.Context, serviceID int64) (int64, error) {
	var n int64
	err := notDeleted(r.db.WithContext(ctx).Model(&model.AdConfig{})).
		Where("service_id = ?", serviceID).
		Count(&n).Error
	return n, err
}

// UpdateAdConfigByService updates ad config for service.
func (r *ConfigRepository) UpdateAdConfigByService(ctx context.Context, serviceID int64, cols map[string]interface{}) error {
	return notDeleted(r.db.WithContext(ctx).Model(&model.AdConfig{})).
		Where("service_id = ?", serviceID).
		Updates(cols).Error
}

// GetLatestCreditScoreEnabled returns latest row where iz_credit_score=true.
func (r *ConfigRepository) GetLatestCreditScoreEnabled(ctx context.Context) (*model.CreditScoreConfig, error) {
	var row model.CreditScoreConfig
	err := notDeleted(r.db.WithContext(ctx)).
		Where("iz_credit_score = ?", true).
		Order("created_at DESC").
		Limit(1).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
