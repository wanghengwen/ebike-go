package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/idgen"

	"gorm.io/gorm"
)

type FenceRepository struct {
	db *gorm.DB
}

func NewFenceRepository(db *gorm.DB) *FenceRepository {
	return &FenceRepository{db: db}
}

func (r *FenceRepository) activeQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Where("(iz_del = 0 OR iz_del IS NULL)")
}

func (r *FenceRepository) GetByID(ctx context.Context, id int64) (*gateway.FenceE, error) {
	var row model.TFence
	err := r.activeQuery(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fe := convert.FenceToEntity(row)
	return &fe, nil
}

func (r *FenceRepository) GetByIDs(ctx context.Context, ids []int64) (map[int64]gateway.FenceE, error) {
	return r.GetByIDsAndType(ctx, ids, 0, "")
}

func (r *FenceRepository) GetByIDsAndType(ctx context.Context, ids []int64, fenceType int, tenantID string) (map[int64]gateway.FenceE, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []model.TFence
	q := r.activeQuery(ctx).Where("id IN ?", ids)
	if fenceType > 0 {
		q = q.Where("type = ?", fenceType)
	}
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]gateway.FenceE, len(rows))
	for _, row := range rows {
		out[row.ID] = convert.FenceToEntity(row)
	}
	return out, nil
}

func (r *FenceRepository) Save(ctx context.Context, tenantID, pin string, fe *gateway.FenceE) (int64, error) {
	if fe == nil {
		return 0, fmt.Errorf("fence must not be null")
	}
	row := convert.EntityToModel(tenantID, pin, *fe)
	if row.ID == 0 {
		row.ID = idgen.NextID()
	}
	now := time.Now().UTC()
	row.CreatedAt = sql.NullTime{Time: now, Valid: true}
	row.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	row.CreatedPin = sql.NullString{String: pin, Valid: pin != ""}
	row.IzDel = sql.NullBool{Bool: false, Valid: true}
	row.Version = sql.NullInt32{Int32: 1, Valid: true}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	fe.Id = row.ID
	return row.ID, nil
}

func (r *FenceRepository) Update(ctx context.Context, tenantID, pin string, fe *gateway.FenceE) (int64, error) {
	if fe == nil || fe.Id == 0 {
		return 0, fmt.Errorf("fence id must not be null")
	}
	row := convert.EntityToModel(tenantID, pin, *fe)
	res := r.db.WithContext(ctx).Model(&model.TFence{}).Where("id = ?", fe.Id).Updates(&row)
	return res.RowsAffected, res.Error
}

// TouchUpdatedAt mirrors Java MaintainAreaGatewayImpl.updatePinAndTime (MP updateById with only id set).
func (r *FenceRepository) TouchUpdatedAt(ctx context.Context, id int64, pin string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"updated_at":  now,
		"updated_pin": pin,
	}
	return r.db.WithContext(ctx).Model(&model.TFence{}).Where("id = ?", id).
		Updates(updates).Error
}

func (r *FenceRepository) Delete(ctx context.Context, tenantID string, id int64) (int64, error) {
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	res := q.Delete(&model.TFence{})
	return res.RowsAffected, res.Error
}

func (r *FenceRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.TFence{}).Error
}

func (r *FenceRepository) UpdateByIDs(ctx context.Context, ids []int64, updates map[string]interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.TFence{}).Where("id IN ?", ids).Updates(updates).Error
}

func (r *FenceRepository) ListByTypeAndServiceID(ctx context.Context, tenantID string, fenceType int, serviceID int64) ([]gateway.FenceE, error) {
	var rows []model.TFence
	q := r.activeQuery(ctx).Where("type = ?", fenceType)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rowsToEntities(rows), nil
}

func (r *FenceRepository) ListByType(ctx context.Context, tenantID string, fenceType int) ([]gateway.FenceE, error) {
	return r.ListByTypeAndServiceID(ctx, tenantID, fenceType, 0)
}

func (r *FenceRepository) ListAllByType(ctx context.Context, fenceType int) ([]gateway.FenceE, error) {
	var rows []model.TFence
	if err := r.activeQuery(ctx).Where("type = ?", fenceType).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rowsToEntities(rows), nil
}

// ParkingPageFilter mirrors Java FenceQueryImpl.getFenceListByTypeAndTagIds query body.
type ParkingPageFilter struct {
	TenantID  string
	ServiceID int64
	FenceIDs  []int64
	Name      string
	AreaSize  *float64
	FieldList []string
	IzEnable  *int
	PageNum   int
	PageSize  int
}

func applyParkingFieldList(q *gorm.DB, fieldList []string) *gorm.DB {
	for _, s := range fieldList {
		switch s {
		case "iz_enable":
			q = q.Where("iz_enable = ?", false)
		case "directional":
			q = q.Where("directional = ?", true)
		case "rfid":
			q = q.Where("rfid = ?", true)
		case "kickstand":
			q = q.Where("kickstand = ?", true)
		case "tbeacon":
			q = q.Where("tbeacon = ?", true)
		case "camera":
			q = q.Where("camera = ?", true)
		case "izFullPileNoStop":
			q = q.Where("iz_full_pile_no_stop = ?", 1)
		}
	}
	return q
}

func (r *FenceRepository) PageParkingByFilter(ctx context.Context, filter ParkingPageFilter) ([]gateway.FenceE, int64, error) {
	if len(filter.FenceIDs) == 0 {
		return nil, 0, nil
	}
	pageNum := filter.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.activeQuery(ctx).
		Where("type = ?", fence.TypeParking).
		Where("service_id = ?", filter.ServiceID).
		Where("id IN ?", filter.FenceIDs)
	if filter.TenantID != "" {
		q = q.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Name != "" {
		q = q.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.AreaSize != nil {
		q = q.Where("area_size >= ?", *filter.AreaSize)
	}
	if filter.IzEnable != nil {
		q = q.Where("iz_enable = ?", *filter.IzEnable != 0)
	}
	q = applyParkingFieldList(q, filter.FieldList)

	var total int64
	if err := q.Model(&model.TFence{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.TFence
	offset := (pageNum - 1) * pageSize
	if err := q.Order("iz_enable DESC").Order("updated_at DESC").
		Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rowsToEntities(rows), total, nil
}

func (r *FenceRepository) PageByTypeAndServiceID(ctx context.Context, tenantID string, fenceType int, serviceID int64, name string, areaSize *float64, fieldList []string, pageNum, pageSize int) ([]gateway.FenceE, int64, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.activeQuery(ctx).Where("type = ?", fenceType)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if areaSize != nil {
		q = q.Where("area_size >= ?", *areaSize)
	}
	if len(fieldList) > 0 {
		for _, s := range fieldList {
			switch s {
			case "iz_enable":
				q = q.Where("iz_enable = ?", false)
			case "directional":
				q = q.Where("directional = ?", true)
			case "rfid":
				q = q.Where("rfid = ?", true)
			case "kickstand":
				q = q.Where("kickstand = ?", true)
			case "tbeacon":
				q = q.Where("tbeacon = ?", true)
			case "camera":
				q = q.Where("camera = ?", true)
			case "izFullPileNoStop":
				q = q.Where("iz_full_pile_no_stop = ?", 1)
			}
		}
	}
	var total int64
	if err := q.Model(&model.TFence{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.TFence
	offset := (pageNum - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rowsToEntities(rows), total, nil
}

func (r *FenceRepository) ExistsByName(ctx context.Context, tenantID, name string, fenceType int, serviceID *int64) (bool, error) {
	q := r.activeQuery(ctx).Where("name = ? AND type = ?", name, fenceType)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID != nil {
		q = q.Where("service_id = ?", *serviceID)
	}
	var count int64
	if err := q.Model(&model.TFence{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListNamesBySuffixRange mirrors Java FenceGatewayImpl.getByNameCount: fences whose
// names are baseName+from … baseName+to (inclusive), returned in DB row order.
func (r *FenceRepository) ListNamesBySuffixRange(ctx context.Context, tenantID, baseName string, fenceType int, serviceID *int64, from, to int) ([]string, error) {
	if from > to {
		return nil, nil
	}
	candidates := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		candidates = append(candidates, baseName+strconv.Itoa(i))
	}
	q := r.activeQuery(ctx).Select("name").Where("name IN ? AND type = ?", candidates, fenceType)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID != nil {
		q = q.Where("service_id = ?", *serviceID)
	}
	var rows []model.TFence
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names, nil
}

func (r *FenceRepository) ListByCustomType(ctx context.Context, tenantID string, serviceID, customTypeID int64) ([]gateway.FenceE, error) {
	q := r.activeQuery(ctx).Where("type = 10")
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	if customTypeID > 0 {
		q = q.Where("custom_type_id = ?", customTypeID)
	}
	var rows []model.TFence
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rowsToEntities(rows), nil
}

func (r *FenceRepository) CountByCustomType(ctx context.Context, tenantID string, customTypeID int64) (int64, error) {
	var count int64
	q := r.activeQuery(ctx).Where("custom_type_id = ? AND type = ?", customTypeID, fence.TypeCustom)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Model(&model.TFence{}).Count(&count).Error
	return count, err
}

func rowsToEntities(rows []model.TFence) []gateway.FenceE {
	out := make([]gateway.FenceE, 0, len(rows))
	for _, row := range rows {
		out = append(out, convert.FenceToEntity(row))
	}
	return out
}

// UpdateRadPacketParking mirrors Java ParkingGatewayImpl.createRadPacketParking.
func (r *FenceRepository) UpdateRadPacketParking(ctx context.Context, tenantID string, parkingID int64, minAmount, maxAmount int, activityID int64, expiration sql.NullTime) error {
	var row model.TFence
	err := r.activeQuery(ctx).Where("id = ? AND type = ?", parkingID, 2).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if row.ExpirationTime.Valid && row.ExpirationTime.Time.After(time.Now()) {
		return nil
	}
	updates := map[string]interface{}{
		"min_amount":      minAmount,
		"max_amount":      maxAmount,
		"activity_id":     activityID,
		"expiration_time": expiration,
	}
	if err := r.db.WithContext(ctx).Model(&model.TFence{}).Where("id = ?", parkingID).Updates(updates).Error; err != nil {
		return err
	}
	_ = gateway.InvalidateFenceCache(ctx, gateway.CacheParking, tenantID, parkingID)
	return nil
}
