package fenceadmin

import (
	"context"
	"database/sql"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	domainsvc "ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

type MaintainAreaAdmin struct{}

func NewMaintainAreaAdmin() *MaintainAreaAdmin { return &MaintainAreaAdmin{} }

func (m *MaintainAreaAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.MaintainAreaCmd) (int64, error) {
	fe := cmdToFenceE(fence.TypeMaintainArea, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.MaxParkingNumber = cmd.MaxParkingNumber
	fe.AreaSize = cmd.AreaSize
	sid := fe.ServiceId
	return saveFenceRejectDuplicate(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheMaintainArea, serviceID: &sid})
}

func (m *MaintainAreaAdmin) Update(ctx context.Context, tenantID, pin string, cmd dto.MaintainAreaCmd) error {
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	fe := cmdToFenceE(fence.TypeMaintainArea, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.Id = cmd.Id
	fe.MaxParkingNumber = cmd.MaxParkingNumber
	fe.AreaSize = cmd.AreaSize
	sid := fe.ServiceId
	return updateFenceRejectDuplicate(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheMaintainArea, serviceID: &sid})
}

func (m *MaintainAreaAdmin) Delete(ctx context.Context, tenantID string, id int64) error {
	return deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheMaintainArea})
}

func (m *MaintainAreaAdmin) DeleteBatch(ctx context.Context, tenantID string, ids []int64) error {
	return deleteFences(ctx, tenantID, ids, writeOpts{kind: gateway.CacheMaintainArea})
}

func (m *MaintainAreaAdmin) ListByServiceID(ctx context.Context, tenantID string, serviceID int64) ([]dto.MaintainAreaCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeMaintainArea, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MaintainAreaCO, 0, len(list))
	for _, fe := range list {
		out = append(out, toMaintainAreaCO(fe))
	}
	return out, nil
}

func (m *MaintainAreaAdmin) PageListByServiceID(ctx context.Context, tenantID string, q dto.MaintainAreaPageQuery) (dto.PageDTO[dto.MaintainAreaCO], error) {
	r, err := requireFenceRepo()
	if err != nil {
		return dto.PageDTO[dto.MaintainAreaCO]{}, err
	}
	list, total, err := r.PageByTypeAndServiceID(ctx, tenantID, fence.TypeMaintainArea, q.ServiceId, "", nil, nil, q.PageNum, q.PageSize)
	if err != nil {
		return dto.PageDTO[dto.MaintainAreaCO]{}, err
	}
	out := make([]dto.MaintainAreaCO, 0, len(list))
	for _, fe := range list {
		co := toMaintainAreaCO(fe)
		if persistence.AreaEmployee != nil {
			emps, _ := persistence.AreaEmployee.ListByAreaID(ctx, fe.Id)
			co.MaintainNum = len(emps)
		}
		out = append(out, co)
	}
	return dto.PageDTO[dto.MaintainAreaCO]{List: out, Total: total, PageNum: q.PageNum, PageSize: q.PageSize}, nil
}

func (m *MaintainAreaAdmin) PersonnelManagementList(ctx context.Context, tenantID string, q dto.MaintainAreaPageQuery) (dto.PageDTO[dto.MaintainAreaCO], error) {
	page, err := m.PageListByServiceID(ctx, tenantID, q)
	if err != nil || len(page.List) == 0 || persistence.AreaEmployee == nil {
		return page, err
	}
	ids := make([]int64, len(page.List))
	for i, co := range page.List {
		ids[i] = co.Id
	}
	emps, err := persistence.AreaEmployee.ListByAreaIDs(ctx, ids)
	if err != nil {
		return page, err
	}
	empMap := groupEmployeesByAreaID(emps)
	for i := range page.List {
		list := empMap[page.List[i].Id]
		page.List[i].MaintainNum = len(list)
		page.List[i].LeaderName = leaderNamesFromEmployees(list)
	}
	return page, nil
}

func groupEmployeesByAreaID(rows []model.TAreaEmployee) map[int64][]model.TAreaEmployee {
	out := make(map[int64][]model.TAreaEmployee)
	for _, row := range rows {
		out[row.AreaID] = append(out[row.AreaID], row)
	}
	return out
}

func leaderNamesFromEmployees(rows []model.TAreaEmployee) string {
	names := make([]string, 0)
	for _, row := range rows {
		if row.Type.Valid && row.Type.Int32 == 1 && row.Name.Valid && row.Name.String != "" {
			names = append(names, row.Name.String)
		}
	}
	return strings.Join(names, ",")
}

func (m *MaintainAreaAdmin) PersonnelDetail(ctx context.Context, areaID int64) (dto.PersonnelDetailCO, error) {
	if persistence.AreaEmployee == nil {
		return dto.PersonnelDetailCO{}, nil
	}
	rows, err := persistence.AreaEmployee.ListByAreaID(ctx, areaID)
	if err != nil {
		return dto.PersonnelDetailCO{}, err
	}
	var out dto.PersonnelDetailCO
	for _, row := range rows {
		vo := areaEmployeeToVO(row)
		switch {
		case row.Type.Valid && row.Type.Int32 == 1:
			out.LeaderList = append(out.LeaderList, vo)
		case row.Type.Valid && row.Type.Int32 == 2:
			out.CheckerList = append(out.CheckerList, vo)
		case row.Type.Valid && row.Type.Int32 == 3:
			out.MaintainerList = append(out.MaintainerList, vo)
		case row.Type.Valid && row.Type.Int32 == 4:
			out.ChangePowerPersonList = append(out.ChangePowerPersonList, vo)
		case row.Type.Valid && row.Type.Int32 == 5:
			out.MoveCarPersonList = append(out.MoveCarPersonList, vo)
		}
	}
	return out, nil
}

func (m *MaintainAreaAdmin) PersonnelEdit(ctx context.Context, tenantID, pin string, cmd dto.PersonnelEditCMD) error {
	if persistence.AreaEmployee == nil {
		return nil
	}
	if err := persistence.AreaEmployee.DeleteByAreaID(ctx, cmd.AreaId); err != nil {
		return err
	}
	rows := make([]model.TAreaEmployee, 0, len(cmd.List))
	for _, vo := range cmd.List {
		rows = append(rows, model.TAreaEmployee{
			UserPin:    vo.UserPin,
			Name:       sql.NullString{String: vo.Name, Valid: vo.Name != ""},
			Phone:      sql.NullString{String: vo.Phone, Valid: vo.Phone != ""},
			RoleName:   sql.NullString{String: vo.RoleName, Valid: vo.RoleName != ""},
			Type:       sql.NullInt32{Int32: int32(vo.Type), Valid: true},
			CreatedPin: sql.NullString{String: pin, Valid: pin != ""},
			UpdatedPin: sql.NullString{String: pin, Valid: pin != ""},
			Version:    sql.NullInt32{Int32: 1, Valid: true},
		})
	}
	if len(cmd.List) == 0 {
		return touchMaintainAreaUpdated(ctx, cmd.AreaId, pin)
	}
	if err := persistence.AreaEmployee.InsertBatch(ctx, tenantID, cmd.AreaId, rows); err != nil {
		return err
	}
	return touchMaintainAreaUpdated(ctx, cmd.AreaId, pin)
}

func (m *MaintainAreaAdmin) PersonnelDelete(ctx context.Context, pin string, employeeID int64) error {
	if persistence.AreaEmployee == nil {
		return nil
	}
	row, err := persistence.AreaEmployee.GetByID(ctx, employeeID)
	if err != nil || row == nil {
		return err
	}
	if err := persistence.AreaEmployee.DeleteByID(ctx, employeeID); err != nil {
		return err
	}
	return touchMaintainAreaUpdated(ctx, row.AreaID, pin)
}

func touchMaintainAreaUpdated(ctx context.Context, areaID int64, pin string) error {
	if persistence.Fence == nil {
		return nil
	}
	return persistence.Fence.TouchUpdatedAt(ctx, areaID, pin)
}

func (m *MaintainAreaAdmin) GetByUserPin(ctx context.Context, tenantID string, cmd dto.AreaEmployeeCmd) ([]dto.MaintainAreaCO, error) {
	if persistence.AreaEmployee == nil {
		return nil, nil
	}
	rows, err := persistence.AreaEmployee.ListByUserPin(ctx, tenantID, cmd.UserPin, cmd.ServiceAreaId, cmd.Type)
	if err != nil {
		return nil, err
	}
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	out := make([]dto.MaintainAreaCO, 0)
	for _, row := range rows {
		fe, err := r.GetByID(ctx, row.AreaID)
		if err != nil || fe == nil {
			continue
		}
		if cmd.ServiceAreaId > 0 && fe.ServiceId != cmd.ServiceAreaId {
			continue
		}
		out = append(out, toMaintainAreaCO(*fe))
	}
	return out, nil
}

func areaEmployeeToVO(row model.TAreaEmployee) dto.AreaEmployeeVO {
	vo := dto.AreaEmployeeVO{UserPin: row.UserPin}
	if row.Name.Valid {
		vo.Name = row.Name.String
	}
	if row.Phone.Valid {
		vo.Phone = row.Phone.String
	}
	if row.RoleName.Valid {
		vo.RoleName = row.RoleName.String
	}
	if row.Type.Valid {
		vo.Type = int(row.Type.Int32)
	}
	return vo
}

func (m *MaintainAreaAdmin) FindNear(ctx context.Context, tenantID string, lat, lng float64) (*gateway.FenceE, error) {
	return domainsvc.FindMaintainAreaNear(ctx, tenantID, geo.Location{Lat: lat, Lng: lng}, 1000, 20)
}
