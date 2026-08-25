package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/shadow"

	"gorm.io/gorm"
)

type ParkingRepository struct {
	db *gorm.DB
}

func NewParkingRepository(db *gorm.DB) *ParkingRepository {
	return &ParkingRepository{db: db}
}

func (r *ParkingRepository) activeQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Where("(iz_del = 0 OR iz_del IS NULL)")
}

func rowToDetail(row *model.TParking) *gateway.ParkingDetailE {
	if row == nil {
		return nil
	}
	return &gateway.ParkingDetailE{
		Id:             row.ID,
		Imei:           row.Imei,
		ServiceId:      row.ServiceID,
		CarId:          row.CarID,
		ParkingId:      row.ParkingID,
		NoParkingId:    row.NoParkingID,
		BanRidingId:    row.BanRidingID,
		MaintainAreaId: row.MaintainAreaID,
		NearParkingId:  row.NearParkingID,
		FenceCustomId:  row.FenceCustomID,
	}
}

func (r *ParkingRepository) GetByCarID(ctx context.Context, tenantID, carID string) (*gateway.ParkingDetailE, error) {
	var row model.TParking
	q := r.activeQuery(ctx).Where("car_id = ?", carID)
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
	return rowToDetail(&row), nil
}

// CountGroupedByParkingIDByService mirrors Java ParkingDetailMapper.selectParkingCountByServiceId.
func (r *ParkingRepository) CountGroupedByParkingIDByService(ctx context.Context, serviceID int64) (map[int64]int64, error) {
	type row struct {
		ParkingID int64 `gorm:"column:parking_id"`
		Cou       int64 `gorm:"column:cou"`
	}
	var rows []row
	err := r.activeQuery(ctx).Model(&model.TParking{}).
		Select("count(parking_id) as cou, parking_id").
		Where("service_id = ? AND parking_id != 0", serviceID).
		Group("parking_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, row := range rows {
		out[row.ParkingID] = row.Cou
	}
	return out, nil
}

// CountByParkingID mirrors Java ParkingDetailGatewayImpl.countByParkingId.
func (r *ParkingRepository) CountByParkingID(ctx context.Context, tenantID string, parkingID int64) (int64, error) {
	var count int64
	q := r.activeQuery(ctx).Model(&model.TParking{}).Where("parking_id = ?", parkingID)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByFenceCustomIDList mirrors Java ParkingDetailGatewayImpl.countByFenceCustomIdList.
func (r *ParkingRepository) CountByFenceCustomIDList(ctx context.Context, tenantID string, fenceCustomIDs []int64) (map[int64]int64, error) {
	if len(fenceCustomIDs) == 0 {
		return nil, nil
	}
	type row struct {
		FenceCustomID int64 `gorm:"column:fence_custom_id"`
		Count         int64 `gorm:"column:count"`
	}
	var rows []row
	q := r.activeQuery(ctx).Model(&model.TParking{}).
		Select("fence_custom_id, count(1) as count").
		Where("fence_custom_id IN ?", fenceCustomIDs).
		Group("fence_custom_id")
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, row := range rows {
		out[row.FenceCustomID] = row.Count
	}
	return out, nil
}

// SaveOrUpdateV1 mirrors Java ParkingDetailGatewayImpl.saveOrUpdateV1.
func (r *ParkingRepository) SaveOrUpdateV1(ctx context.Context, tenantID, pin string, detail *gateway.ParkingDetailE) error {
	if shadow.IsShadowTest(ctx) {
		return nil
	}
	if detail == nil {
		return nil
	}
	now := time.Now().UTC()
	row := model.TParking{
		ID:             detail.Id,
		TenantID:       tenantID,
		Imei:           detail.Imei,
		ServiceID:      detail.ServiceId,
		CarID:          detail.CarId,
		ParkingID:      detail.ParkingId,
		NoParkingID:    detail.NoParkingId,
		BanRidingID:    detail.BanRidingId,
		MaintainAreaID: detail.MaintainAreaId,
		NearParkingID:  detail.NearParkingId,
		FenceCustomID:  detail.FenceCustomId,
		IzDel:          sql.NullBool{Bool: false, Valid: true},
	}
	if row.ID == 0 {
		row.CreatedAt = sql.NullTime{Time: now, Valid: true}
		row.UpdatedAt = sql.NullTime{Time: now, Valid: true}
		row.CreatedPin = sql.NullString{String: pin, Valid: pin != ""}
		row.UpdatedPin = sql.NullString{String: pin, Valid: pin != ""}
		row.Version = sql.NullInt32{Int32: 1, Valid: true}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
		detail.Id = row.ID
		return nil
	}
	var existing model.TParking
	if err := r.db.WithContext(ctx).Where("id = ?", row.ID).First(&existing).Error; err != nil {
		return err
	}
	row.CreatedAt = existing.CreatedAt
	row.CreatedPin = existing.CreatedPin
	row.UpdatedPin = existing.UpdatedPin
	row.Version = existing.Version
	row.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	return r.db.WithContext(ctx).Save(&row).Error
}

// UnBindParking mirrors Java ParkingDetailGatewayImpl.unBindParking.
func (r *ParkingRepository) UnBindParking(ctx context.Context, tenantID, carID string, parkingID int64) (bool, error) {
	var row model.TParking
	q := r.activeQuery(ctx).Where("car_id = ? AND parking_id = ?", carID, parkingID)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	row.ParkingID = 0
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return false, err
	}
	detail := rowToDetail(&row)
	_ = gateway.SetDetailCache(ctx, tenantID, detail)
	return true, nil
}

// RemoveParkingID mirrors Java ParkingDetailGatewayImpl.removeParkingId:
// when a fence is deleted, clear the matching binding column in t_parking
// (parking_id for FORPARK=2, no_parking_id for NOPARKING=4, ban_riding_id for NOCRAWLPARKING=9).
func (r *ParkingRepository) RemoveParkingID(ctx context.Context, tenantID string, id int64, fenceType int) error {
	if shadow.IsShadowTest(ctx) {
		return nil
	}
	var column string
	switch fenceType {
	case 2: // FenceTypeEnum.FORPARK
		column = "parking_id"
	case 4: // FenceTypeEnum.NOPARKING
		column = "no_parking_id"
	case 9: // FenceTypeEnum.NOCRAWLPARKING
		column = "ban_riding_id"
	default:
		return nil
	}
	q := r.db.WithContext(ctx).Model(&model.TParking{}).Where(column+" = ?", id)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	return q.Update(column, 0).Error
}

// PageByParkingID mirrors Java parkingDetailByStationId.
func (r *ParkingRepository) PageByParkingID(ctx context.Context, tenantID string, parkingID int64, pageNum, pageSize int) (dto.PageDTO[gateway.ParkingDetailE], error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.activeQuery(ctx).Model(&model.TParking{}).Where("parking_id = ?", parkingID)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return dto.PageDTO[gateway.ParkingDetailE]{}, err
	}
	var rows []model.TParking
	offset := (pageNum - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return dto.PageDTO[gateway.ParkingDetailE]{}, err
	}
	out := make([]gateway.ParkingDetailE, 0, len(rows))
	for i := range rows {
		if d := rowToDetail(&rows[i]); d != nil {
			out = append(out, *d)
		}
	}
	return dto.PageDTO[gateway.ParkingDetailE]{
		List:     out,
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
	}, nil
}

// ListByMaintainAreaID mirrors Java parkingDetailByMaintainAreaId.
func (r *ParkingRepository) ListByMaintainAreaID(ctx context.Context, tenantID string, maintainAreaID int64) ([]gateway.ParkingDetailE, error) {
	var rows []model.TParking
	q := r.activeQuery(ctx).Where("maintain_area_id = ?", maintainAreaID)
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]gateway.ParkingDetailE, 0, len(rows))
	for i := range rows {
		if d := rowToDetail(&rows[i]); d != nil {
			out = append(out, *d)
		}
	}
	return out, nil
}
