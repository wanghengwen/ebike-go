package configsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	pkgredis "ebike-fence-go/internal/pkg/redis"
	"ebike-fence-go/internal/pkg/timefmt"

	"gorm.io/gorm"
)

type ResourceManagementService struct{ repo *repo.ConfigRepository }

func NewResourceManagementService(r *repo.ConfigRepository) *ResourceManagementService {
	return &ResourceManagementService{repo: r}
}

func (s *ResourceManagementService) Create(ctx context.Context, tenantID, pin string, cmd dto.ResourceManagementCmd) error {
	if cmd.ServiceId == nil {
		return errors.New("serviceId must not be null")
	}
	var n int64
	_ = s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).
		Where("service_id = ? AND status IN ?", *cmd.ServiceId, []int{resourceStatusOccupy, resourceStatusFree}).
		Count(&n).Error
	if n > 5 {
		// Align Java FenceMsgCode.RESOURCE_BIT_COUNT_HAS_UPPER_MAX_5
		return &service.BizError{
			Code: "13021",
			Msg:  "当前资源位已全部启用，如需新增，请先下架至少一个资源位。",
		}
	}
	row := model.ResourceManagement{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	applyResourceManagementCmd(&row, cmd)
	expire := resourceStatusExpire
	row.Status = sqlNullInt32(&expire)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	return s.repo.Create(ctx, &row)
}

func (s *ResourceManagementService) Update(ctx context.Context, tenantID, pin string, cmd dto.ResourceManagementCmd) error {
	if cmd.Id == nil {
		return errors.New("id must not be null")
	}
	var existing model.ResourceManagement
	if err := s.repo.GetByID(ctx, tenantID, &existing, *cmd.Id); err != nil || existing.ID == 0 {
		return &service.BizError{Code: dto.CodeException, Msg: "资源位不存在"}
	}
	if existing.Status.Valid && existing.Status.Int32 != int32(resourceStatusExpire) {
		return &service.BizError{Code: dto.CodeException, Msg: "请先下架资源"}
	}
	row := existing
	applyResourceManagementCmd(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, false)
	return s.repo.Save(ctx, &row)
}

func (s *ResourceManagementService) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	var n int64
	_ = s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).
		Where("id IN ? AND status IN ?", ids, []int{resourceStatusFree, resourceStatusOccupy}).
		Count(&n).Error
	if n > 0 {
		return &service.BizError{Code: dto.CodeException, Msg: "请先下架资源再删除"}
	}
	return s.repo.DB().WithContext(ctx).Where("id IN ?", ids).Delete(&model.ResourceManagement{}).Error
}

func (s *ResourceManagementService) List(ctx context.Context, cmd dto.ResourceManagementCmd) ([]dto.ResourceManagementCO, error) {
	if cmd.ServiceId == nil {
		return nil, &service.BizError{Code: dto.CodeException, Msg: "服务区id不能为空"}
	}
	q := s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).
		Where("service_id = ?", *cmd.ServiceId)
	if cmd.PageCode != nil && *cmd.PageCode != 0 {
		q = q.Where("page_code = ?", *cmd.PageCode)
	}
	if cmd.Status != nil {
		now := time.Now()
		switch *cmd.Status {
		case resourceStatusExpire:
			q = q.Where("iz_limit_time = ? AND end_time < ?", true, now)
		case resourceStatusFree:
			q = q.Where("status = ?", *cmd.Status)
		default:
			q = q.Where("status = ?", *cmd.Status).
				Where("(start_time < ? AND end_time > ?) OR iz_limit_time = ?", now, now, false)
		}
	}
	if cmd.Type != nil && *cmd.Type != 0 {
		q = q.Where("type = ?", *cmd.Type)
	}
	var rows []model.ResourceManagement
	if err := q.Order("sort ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dto.ResourceManagementCO, len(rows))
	for i, r := range rows {
		out[i] = convert.ResourceToCO(r)
	}
	return out, nil
}

func (s *ResourceManagementService) AppList(ctx context.Context, cmd dto.ResourceManagementCmd) ([]dto.ResourceManagementCO, error) {
	if cmd.ServiceId == nil {
		return nil, nil
	}
	pageCode := 1
	if cmd.PageCode != nil {
		pageCode = *cmd.PageCode
	}
	now := time.Now()
	var rows []model.ResourceManagement
	err := s.repo.DB().WithContext(ctx).
		Where("service_id = ? AND page_code = ? AND status = ?", *cmd.ServiceId, pageCode, resourceStatusOccupy).
		Where("(start_time < ? AND end_time > ?) OR iz_limit_time = ?", now, now, false).
		Order("sort ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]dto.ResourceManagementCO, len(rows))
	for i, r := range rows {
		out[i] = convert.ResourceToCO(r)
	}
	return out, nil
}

func (s *ResourceManagementService) Sort(ctx context.Context, pin string, ids []int64) error {
	now := time.Now().UTC()
	for i, id := range ids {
		sortVal := i + 1
		updates := map[string]interface{}{
			"sort":        sortVal,
			"updated_pin": pin,
			"updated_at":  now,
		}
		if err := s.repo.UpdateColumns(ctx, &model.ResourceManagement{ID: id}, updates); err != nil {
			return err
		}
	}
	return nil
}

func (s *ResourceManagementService) BatchOffline(ctx context.Context, ids []int64) error {
	return s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).
		Where("id IN ?", ids).
		Update("status", resourceStatusExpire).Error
}

func (s *ResourceManagementService) BatchOnline(ctx context.Context, ids []int64) error {
	list, err := s.onlineCheck(ctx, ids)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, info := range list {
		status := resourceStatusFree
		if info.StartTime.Valid && now.After(info.StartTime.Time) {
			status = resourceStatusOccupy
		}
		if err := s.repo.UpdateColumns(ctx, &model.ResourceManagement{ID: info.ID}, map[string]interface{}{"status": status}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ResourceManagementService) onlineCheck(ctx context.Context, ids []int64) ([]model.ResourceManagement, error) {
	if len(ids) > 5 {
		return nil, &service.BizError{Code: dto.CodeException, Msg: "最多只能选择5个资源位同时上架"}
	}
	var list []model.ResourceManagement
	if err := s.repo.DB().WithContext(ctx).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, info := range list {
		if info.Status.Valid && info.Status.Int32 != int32(resourceStatusExpire) {
			return nil, &service.BizError{Code: dto.CodeException, Msg: "请勿重复上架"}
		}
	}
	sortSet := map[int32]struct{}{}
	for _, info := range list {
		if info.Sort.Valid {
			if _, ok := sortSet[info.Sort.Int32]; ok {
				return nil, &service.BizError{Code: dto.CodeException, Msg: "同一资源位只能上架一条资源"}
			}
			sortSet[info.Sort.Int32] = struct{}{}
		}
	}
	if len(list) == 0 {
		return nil, &service.BizError{Code: dto.CodeException, Msg: "资源不存在"}
	}
	entity := list[0]
	var online []model.ResourceManagement
	_ = s.repo.DB().WithContext(ctx).
		Where("service_id = ? AND status IN ?", entity.ServiceID, []int{resourceStatusOccupy, resourceStatusFree}).
		Find(&online).Error
	onlineSort := map[int32]struct{}{}
	for _, o := range online {
		if o.Sort.Valid {
			onlineSort[o.Sort.Int32] = struct{}{}
		}
	}
	for _, info := range list {
		if info.Sort.Valid {
			if _, ok := onlineSort[info.Sort.Int32]; ok {
				return nil, &service.BizError{Code: dto.CodeException, Msg: fmt.Sprintf("当前资源位-%d 已被占用，请调整资源位后上架", info.Sort.Int32)}
			}
		}
	}
	count := int64(len(online))
	if count+int64(len(ids)) > 5 {
		return nil, &service.BizError{Code: dto.CodeException, Msg: fmt.Sprintf("当前已上架%d个资源位，只能选择最多%d个资源位上架", count, 5-count)}
	}
	now := time.Now()
	for _, info := range list {
		if info.EndTime.Valid && info.EndTime.Time.Before(now) {
			name := ""
			if info.Name.Valid {
				name = info.Name.String
			}
			return nil, &service.BizError{Code: dto.CodeException, Msg: fmt.Sprintf("资源“%s”展示时间已过期，请调整后重新上架", name)}
		}
	}
	return list, nil
}

func (s *ResourceManagementService) AddExposure(ctx context.Context, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	err := s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).
		Where("id IN ?", ids).
		UpdateColumn("exposure_count", gorm.Expr("exposure_count + 1")).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *ResourceManagementService) AddClick(ctx context.Context, pin string, id int64) (bool, error) {
	key := rediskeys.ResourceClickPerson(pin, id)
	rdb := pkgredis.GetClient()
	seen := false
	if rdb != nil {
		if val, err := rdb.Get(ctx, key).Result(); err == nil && val != "" {
			seen = true
		}
	}
	q := s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).Where("id = ?", id)
	if err := q.UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error; err != nil {
		return false, err
	}
	if !seen {
		_ = q.UpdateColumn("click_person", gorm.Expr("click_person + 1")).Error
		if rdb != nil {
			_ = rdb.Set(ctx, key, "1", 0).Err()
		}
	}
	return true, nil
}

func (s *ResourceManagementService) Page(ctx context.Context, q dto.ResourceManagementPageQuery) (*dto.PageDTO[dto.ResourceManagementCO], error) {
	if q.ServiceId == nil {
		return nil, errors.New("serviceId must not be null")
	}
	pageNum, pageSize := q.PageNum, q.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var total int64
	base := s.repo.DB().WithContext(ctx).Model(&model.ResourceManagement{}).Where("service_id = ?", *q.ServiceId)
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.ResourceManagement
	if err := base.Order("sort ASC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]dto.ResourceManagementCO, len(rows))
	for i, r := range rows {
		list[i] = convert.ResourceToCO(r)
	}
	return &dto.PageDTO[dto.ResourceManagementCO]{
		List: list, Total: total, PageNum: pageNum, PageSize: pageSize,
	}, nil
}

func sqlNullInt32(v *int) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*v), Valid: true}
}

// applyResourceManagementCmd maps the DTO onto the model. The model uses sql.Null*
// column types, which the previous JSON round-trip (mergeJSON) could not populate
// from the DTO's scalar JSON values, so start/end time and the other fields were
// silently dropped. This mirrors Java's copyProperties + updateById (skip-null)
// semantics: only fields present in the command are written.
func applyResourceManagementCmd(row *model.ResourceManagement, cmd dto.ResourceManagementCmd) {
	if cmd.PageCode != nil {
		row.PageCode = sql.NullInt32{Int32: int32(*cmd.PageCode), Valid: true}
	}
	if cmd.Type != nil {
		row.Type = sql.NullInt32{Int32: int32(*cmd.Type), Valid: true}
	}
	if cmd.Status != nil {
		row.Status = sql.NullInt32{Int32: int32(*cmd.Status), Valid: true}
	}
	if cmd.Name != nil {
		row.Name = sql.NullString{String: *cmd.Name, Valid: true}
	}
	if cmd.StartTime != nil {
		if t, err := timefmt.ParseJavaLocalString(*cmd.StartTime); err == nil && !t.IsZero() {
			row.StartTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if cmd.EndTime != nil {
		if t, err := timefmt.ParseJavaLocalString(*cmd.EndTime); err == nil && !t.IsZero() {
			row.EndTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if cmd.IzLimitTime != nil {
		row.IzLimitTime = sql.NullBool{Bool: *cmd.IzLimitTime, Valid: true}
	}
	if cmd.AdvId != nil {
		row.AdvID = sql.NullString{String: *cmd.AdvId, Valid: true}
	}
	if cmd.Title != nil {
		row.Title = sql.NullString{String: *cmd.Title, Valid: true}
	}
	if cmd.Appid != nil {
		row.AppID = sql.NullString{String: *cmd.Appid, Valid: true}
	}
	if cmd.SkipUrl != nil {
		row.SkipURL = sql.NullString{String: *cmd.SkipUrl, Valid: true}
	}
	if cmd.Params != nil {
		row.Params = sql.NullString{String: *cmd.Params, Valid: true}
	}
	if cmd.ImgUrl != nil {
		row.ImgURL = sql.NullString{String: *cmd.ImgUrl, Valid: true}
	}
	if cmd.Sort != nil {
		row.Sort = sql.NullInt32{Int32: int32(*cmd.Sort), Valid: true}
	}
}
