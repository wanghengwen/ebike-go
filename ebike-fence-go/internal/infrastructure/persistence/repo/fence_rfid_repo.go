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

type FenceRfidRepository struct {
	db *gorm.DB
}

func NewFenceRfidRepository(db *gorm.DB) *FenceRfidRepository {
	return &FenceRfidRepository{db: db}
}

func (r *FenceRfidRepository) ListByFenceID(ctx context.Context, fenceID int64) ([]model.TFenceRfid, error) {
	var rows []model.TFenceRfid
	err := r.db.WithContext(ctx).
		Where("fence_id = ? AND (iz_del = 0 OR iz_del IS NULL)", fenceID).
		Find(&rows).Error
	return rows, err
}

func (r *FenceRfidRepository) GetByRfidCode(ctx context.Context, tenantID, rfidCode string) (*model.TFenceRfid, error) {
	var row model.TFenceRfid
	q := r.db.WithContext(ctx).
		Where("rfid_code = ? AND (iz_del = 0 OR iz_del IS NULL)", rfidCode)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ErrRfidNewBindExist mirrors Java FenceMsgCode.FENCE_RFID_BIND_EXIST (13044):
// the new rfid code is already bound to a station.
var ErrRfidNewBindExist = errors.New("新rfid已绑定站点,请更换其他rfid编号")

// checkHasEnable mirrors Java FenceRfidGatewayImpl.checkHasEnable: count of enabled
// (iz_del=0) rows with this rfid code.
func (r *FenceRfidRepository) checkHasEnable(ctx context.Context, tenantID, rfidCode string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.TFenceRfid{}).
		Where("rfid_code = ? AND (iz_del = 0 OR iz_del IS NULL)", rfidCode)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// updateDel mirrors Java FenceRfidGatewayImpl.updateDel: soft-delete ALL rows with
// the given rfid code (Java's wrapper filters only by rfid_code).
func (r *FenceRfidRepository) updateDel(ctx context.Context, tenantID, pin, rfidCode string) error {
	q := r.db.WithContext(ctx).Model(&model.TFenceRfid{}).Where("rfid_code = ?", rfidCode)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"iz_del":      true,
		"updated_at":  now,
		"updated_pin": pin,
	}
	return q.Updates(updates).Error
}

// insertRfid inserts a fresh enabled binding row.
func (r *FenceRfidRepository) insertRfid(ctx context.Context, tenantID, pin string, fenceID int64, rfidCode string) (int64, error) {
	now := time.Now().UTC()
	row := model.TFenceRfid{
		ID:         idgen.NextID(),
		TenantID:   tenantID,
		FenceID:    fenceID,
		RfidCode:   rfidCode,
		IzDel:      sql.NullBool{Bool: false, Valid: true},
		CreatedAt:  sql.NullTime{Time: now, Valid: true},
		UpdatedAt:  sql.NullTime{Time: now, Valid: true},
		CreatedPin: sql.NullString{String: pin, Valid: pin != ""},
		UpdatedPin: sql.NullString{String: pin, Valid: pin != ""},
		Version:    sql.NullInt32{Int32: 1, Valid: true},
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

// SaveOrUpdate mirrors Java FenceRfidGatewayImpl.saveOrUpdate: if the rfid code is
// currently bound, soft-delete ALL rows with that code, then insert a brand-new
// enabled row. It never updates a row in place.
//
// NOTE: the original Go behavior updated the existing row's fence_id in place.
// Preserved here for reference:
//
//	existing, err := r.GetByRfidCode(ctx, tenantID, rfidCode)
//	if existing != nil { existing.FenceID = fenceID; ... r.db.Save(existing) ... return existing.ID }
//	... else create new row ...
func (r *FenceRfidRepository) SaveOrUpdate(ctx context.Context, tenantID, pin string, fenceID int64, rfidCode string) (int64, error) {
	cnt, err := r.checkHasEnable(ctx, tenantID, rfidCode)
	if err != nil {
		return 0, err
	}
	if cnt > 0 {
		if err := r.updateDel(ctx, tenantID, pin, rfidCode); err != nil {
			return 0, err
		}
	}
	return r.insertRfid(ctx, tenantID, pin, fenceID, rfidCode)
}

// ChangeRfid mirrors Java FenceRfidGatewayImpl.change: validate the old binding exists
// (else effectNum 0), reject if the new code is already bound (ErrRfidNewBindExist),
// soft-delete all rows with the old code, then insert a new enabled row with the new
// code under the old code's fence id. Returns the affected (inserted) row count.
//
// NOTE: the original Go behavior soft-deleted only the old row by id then called
// SaveOrUpdate(newCode), missing the new-code conflict check. Preserved for reference:
//
//	old := GetByRfidCode(oldCode); if old == nil { return nil }
//	... Update iz_del where id = old.ID ...
//	_, err = r.SaveOrUpdate(ctx, tenantID, old.FenceID, newCode); return err
func (r *FenceRfidRepository) ChangeRfid(ctx context.Context, tenantID, pin, oldCode, newCode string) (int, error) {
	old, err := r.GetByRfidCode(ctx, tenantID, oldCode)
	if err != nil {
		return 0, err
	}
	if old == nil || old.RfidCode == "" {
		return 0, nil
	}
	newInfo, err := r.GetByRfidCode(ctx, tenantID, newCode)
	if err != nil {
		return 0, err
	}
	if newInfo != nil {
		return 0, ErrRfidNewBindExist
	}
	fenceID := old.FenceID
	if err := r.updateDel(ctx, tenantID, pin, oldCode); err != nil {
		return 0, err
	}
	if _, err := r.insertRfid(ctx, tenantID, pin, fenceID, newCode); err != nil {
		return 0, err
	}
	return 1, nil
}

func (r *FenceRfidRepository) ListByRfidCodes(ctx context.Context, tenantID string, codes []string) ([]model.TFenceRfid, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var rows []model.TFenceRfid
	q := r.db.WithContext(ctx).
		Where("rfid_code IN ? AND (iz_del = 0 OR iz_del IS NULL)", codes)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Find(&rows).Error
	return rows, err
}
