package fenceadmin

import (
	"context"
	"database/sql"
	"fmt"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

type FenceCustomAdmin struct{}

func NewFenceCustomAdmin() *FenceCustomAdmin { return &FenceCustomAdmin{} }

func (f *FenceCustomAdmin) CreateType(ctx context.Context, tenantID, pin string, cmd dto.FenceCustomTypeCmd) error {
	if persistence.FenceCustomType == nil {
		return fmt.Errorf("mysql not initialized")
	}
	if cmd.Name == "" {
		return newBizError("00004", "name must not be null")
	}
	if existing, err := persistence.FenceCustomType.GetByName(ctx, tenantID, cmd.Name); err != nil {
		return err
	} else if existing != nil {
		return newBizError("13041", "自定义围栏类型重复")
	}
	row := &model.TFenceCustomType{
		Name:           cmd.Name,
		Description:    sql.NullString{String: cmd.Description, Valid: cmd.Description != ""},
		Color:          sql.NullString{String: cmd.Color, Valid: cmd.Color != ""},
		IzCreateCarTag: sql.NullBool{Bool: cmd.IzCreateCarTag, Valid: true},
		CarTag:         sql.NullString{String: cmd.CarTagList, Valid: cmd.CarTagList != ""},
		CreatedPin:     sql.NullString{String: pin, Valid: pin != ""},
		UpdatedPin:     sql.NullString{String: pin, Valid: pin != ""},
		Version:        sql.NullInt32{Int32: 1, Valid: true},
	}
	_, err := persistence.FenceCustomType.SaveOrUpdate(ctx, tenantID, row)
	return err
}

func (f *FenceCustomAdmin) UpdateType(ctx context.Context, tenantID, pin string, cmd dto.FenceCustomTypeCmd) error {
	if persistence.FenceCustomType == nil {
		return fmt.Errorf("mysql not initialized")
	}
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	if cmd.Name == "" {
		return newBizError("00004", "name must not be null")
	}
	row := &model.TFenceCustomType{
		ID:             cmd.Id,
		Name:           cmd.Name,
		Description:    sql.NullString{String: cmd.Description, Valid: cmd.Description != ""},
		Color:          sql.NullString{String: cmd.Color, Valid: cmd.Color != ""},
		IzCreateCarTag: sql.NullBool{Bool: cmd.IzCreateCarTag, Valid: true},
		CarTag:         sql.NullString{String: cmd.CarTagList, Valid: cmd.CarTagList != ""},
		UpdatedPin:     sql.NullString{String: pin, Valid: pin != ""},
	}
	_, err := persistence.FenceCustomType.SaveOrUpdate(ctx, tenantID, row)
	return err
}

func (f *FenceCustomAdmin) DeleteType(ctx context.Context, tenantID string, id int64) error {
	if persistence.FenceCustomType == nil || persistence.Fence == nil {
		return fmt.Errorf("mysql not initialized")
	}
	count, err := persistence.Fence.CountByCustomType(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return newBizError("13042", "请先删除围栏区域")
	}
	return persistence.FenceCustomType.Delete(ctx, tenantID, id)
}

func (f *FenceCustomAdmin) ListTypes(ctx context.Context, tenantID string) ([]dto.FenceCustomTypeCO, error) {
	if persistence.FenceCustomType == nil {
		return nil, fmt.Errorf("mysql not initialized")
	}
	rows, err := persistence.FenceCustomType.GetAll(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceCustomTypeCO, 0, len(rows))
	for _, row := range rows {
		co := dto.FenceCustomTypeCO{Id: row.ID, Name: row.Name}
		if row.Color.Valid {
			co.Color = row.Color.String
		}
		if row.Description.Valid {
			co.Description = row.Description.String
		}
		if row.IzCreateCarTag.Valid {
			co.IzCreateCarTag = row.IzCreateCarTag.Bool
		}
		if row.CarTag.Valid {
			co.CarTagList = row.CarTag.String
		}
		out = append(out, co)
	}
	return out, nil
}

func (f *FenceCustomAdmin) CreateFence(ctx context.Context, tenantID, pin string, cmd dto.FenceCustomCmd) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	sid := cmd.ServiceId
	exists, err := r.ExistsByName(ctx, tenantID, cmd.Name, fence.TypeCustom, &sid)
	if err != nil {
		return err
	}
	if exists {
		return newBizError("13043", "名称重复")
	}
	fe := cmdToFenceE(fence.TypeCustom, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.CustomTypeId = cmd.CustomTypeId
	_, err = saveFenceNoDedupe(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheCustom, customTypeID: cmd.CustomTypeId, serviceID: &sid})
	return err
}

func (f *FenceCustomAdmin) UpdateFence(ctx context.Context, tenantID, pin string, cmd dto.FenceCustomCmd) error {
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	fe := cmdToFenceE(fence.TypeCustom, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.Id = cmd.Id
	fe.CustomTypeId = cmd.CustomTypeId
	sid := fe.ServiceId
	return updateFenceNoDedupe(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheCustom, customTypeID: cmd.CustomTypeId, serviceID: &sid})
}

func (f *FenceCustomAdmin) DeleteFence(ctx context.Context, tenantID string, id, customTypeID int64) error {
	return deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheCustom, customTypeID: customTypeID})
}

func (f *FenceCustomAdmin) ListFences(ctx context.Context, tenantID string, q dto.CustomFenceListQry) ([]dto.FenceCustomCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByCustomType(ctx, tenantID, q.ServiceId, q.CustomTypeId)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(list))
	for i, fe := range list {
		ids[i] = fe.Id
	}
	carCounts := map[int64]int64{}
	if persistence.Parking != nil && len(ids) > 0 {
		carCounts, _ = persistence.Parking.CountByFenceCustomIDList(ctx, tenantID, ids)
	}
	out := make([]dto.FenceCustomCO, 0, len(list))
	for _, fe := range list {
		co := toFenceCustomCO(fe)
		co.CarCount = carCounts[fe.Id]
		out = append(out, co)
	}
	return out, nil
}

func (f *FenceCustomAdmin) AllTypeGeo(ctx context.Context, tenantID string, q dto.CustomFenceListQry) ([]dto.FenceCustomCO, error) {
	list, err := f.ListFences(ctx, tenantID, q)
	if err != nil {
		return nil, err
	}
	if persistence.FenceCustomType == nil {
		return list, nil
	}
	typeIDs := make([]int64, 0)
	for _, co := range list {
		if co.CustomTypeId > 0 {
			typeIDs = append(typeIDs, co.CustomTypeId)
		}
	}
	types, _ := persistence.FenceCustomType.GetByIDs(ctx, tenantID, typeIDs)
	colorMap := map[int64]string{}
	for _, t := range types {
		if t.Color.Valid {
			colorMap[t.ID] = t.Color.String
		}
	}
	for i := range list {
		list[i].Color = colorMap[list[i].CustomTypeId]
	}
	return list, nil
}
