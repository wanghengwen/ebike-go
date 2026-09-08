package configsvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/persistence/cache"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/pkg/config"
	pkgredis "ebike-fence-go/internal/pkg/redis"
	"ebike-fence-go/internal/pkg/shadow"

	"gorm.io/gorm"
)

// --- Credit Score ---

type CreditScoreService struct{ repo *repo.ConfigRepository }

func NewCreditScoreService(r *repo.ConfigRepository) *CreditScoreService {
	return &CreditScoreService{repo: r}
}

func (s *CreditScoreService) Get(ctx context.Context, tenantID string) (*dto.CreditScoreConfigCO, error) {
	key := rediskeys.CreditScoreConfig(tenantID)
	var co dto.CreditScoreConfigCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.CreditScoreConfig
		if err := s.repo.GetLatestAny(ctx, &row); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.CreditScoreToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.Id == nil {
		f := false
		return &dto.CreditScoreConfigCO{IzCreditScore: &f}, nil
	}
	return &co, nil
}

func (s *CreditScoreService) Insert(ctx context.Context, tenantID, pin string, cmd dto.CreditScoreConfigCmd) (bool, error) {
	row := model.CreditScoreConfig{ID: repo.NextConfigID()}
	mergeJSON(&row, cmd)
	if cmd.IzCreditScore != nil && *cmd.IzCreditScore {
		if prev, _ := s.repo.GetLatestCreditScoreEnabled(ctx); prev != nil {
			// 防篡改：已有开启记录时，总分/告警分/禁行分沿用历史值。
			row.Score = prev.Score
			row.WarnScore = prev.WarnScore
			row.NoRiddingScore = prev.NoRiddingScore
		} else if !shadow.IsShadowTest(ctx) {
			// 首次开启信用分：初始化全量用户信用分（Java creditScoreApi.init）。
			// Java 在 init 抛异常时不会落库，这里同样在失败时中止。
			cc := dto.EnsureCommandContext(cmd.CommandContext, tenantID)
			if err := rpc.NewUserRPC().CreditScoreInit(ctx, cc, cmd.Score, ""); err != nil {
				return false, err
			}
		}
	}
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.CreditScoreConfig(tenantID)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
	return err == nil, err
}

// --- Big Screen ---

type BigScreenService struct{ repo *repo.ConfigRepository }

func NewBigScreenService(r *repo.ConfigRepository) *BigScreenService {
	return &BigScreenService{repo: r}
}

func (s *BigScreenService) Get(ctx context.Context, tenantID string) (*dto.ConfigBigScreenCO, error) {
	var row model.BigScreen
	if err := s.repo.GetLatestTenantRow(ctx, &row, tenantID); err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return &dto.ConfigBigScreenCO{}, nil
	}
	return convert.BigScreenToCO(&row), nil
}

func (s *BigScreenService) Update(ctx context.Context, tenantID, pin string, cmd dto.ConfigBigScreenCmd) (int, error) {
	row := model.BigScreen{}
	if cmd.Id != nil {
		row.ID = *cmd.Id
	}
	if cmd.DisplayCoefficient != nil {
		row.DisplayCoefficient = sql.NullFloat64{Float64: *cmd.DisplayCoefficient, Valid: true}
	}
	repo.StampBase(&row.BaseConfig, tenantID, pin, row.ID == 0)
	if row.ID == 0 {
		row.ID = repo.NextConfigID()
		if err := s.repo.Create(ctx, &row); err != nil {
			return 0, err
		}
		return 1, nil
	}
	if err := s.repo.Save(ctx, &row); err != nil {
		return 0, err
	}
	return 1, nil
}

// --- Ad Config ---

type AdConfigService struct{ repo *repo.ConfigRepository }

func NewAdConfigService(r *repo.ConfigRepository) *AdConfigService { return &AdConfigService{repo: r} }

func (s *AdConfigService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.AdConfigCO, error) {
	var svcRow model.AdConfig
	_ = s.repo.GetLatestByServiceID(ctx, tenantID, &svcRow, serviceID)
	nameOpen := map[string]bool{}
	if svcRow.ID != 0 && svcRow.IzOn != "" {
		parseAdNameOpen(svcRow.IzOn, nameOpen)
	}
	last, err := s.repo.GetAdConfigLastByTenant(ctx, tenantID)
	if err != nil || last == nil || last.ID == 0 || last.IzOn == "" {
		return nil, nil
	}
	merged := mergeAdIzOn(last.IzOn, nameOpen)
	id, sid := last.ID, serviceID
	return &dto.AdConfigCO{Id: &id, ServiceId: &sid, IzOn: merged}, nil
}

func (s *AdConfigService) Ins(ctx context.Context, tenantID, pin string, cmd dto.AdConfigCmd) error {
	row := model.AdConfig{ID: repo.NextConfigID()}
	if cmd.ServiceId != nil {
		row.ServiceID = *cmd.ServiceId
	}
	if cmd.IzOn != nil {
		b, _ := json.Marshal(cmd.IzOn)
		row.IzOn = string(b)
	}
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	return s.repo.Create(ctx, &row)
}

func (s *AdConfigService) Upd(ctx context.Context, tenantID, pin string, cmd dto.AdConfigCmd) (int, error) {
	if cmd.ServiceId == nil {
		return 0, errors.New("serviceId must not be null")
	}
	row := model.AdConfig{ServiceID: *cmd.ServiceId}
	if cmd.IzOn != nil {
		b, _ := json.Marshal(cmd.IzOn)
		row.IzOn = string(b)
	}
	n, _ := s.repo.CountAdConfigByService(ctx, *cmd.ServiceId)
	if n > 0 {
		if err := s.repo.UpdateAdConfigByService(ctx, *cmd.ServiceId, map[string]interface{}{"iz_on": row.IzOn}); err != nil {
			return 0, err
		}
		return 1, nil
	}
	row.ID = repo.NextConfigID()
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	if err := s.repo.Create(ctx, &row); err != nil {
		return 0, err
	}
	return 1, nil
}

func (s *AdConfigService) Detail(ctx context.Context, tenantID string, serviceID int64) (dto.AdsConfigCO, error) {
	key := rediskeys.AdConfig(tenantID, serviceID)
	rdb := pkgredis.GetClient()
	if rdb != nil {
		val, err := rdb.Get(ctx, key).Result()
		if err == nil && !rediskeys.IsCacheMiss(val) {
			var co dto.AdsConfigCO
			if json.Unmarshal([]byte(val), &co) == nil {
				return co, nil
			}
		}
	}
	co := defaultAdsConfig()
	if rdb != nil {
		b, _ := json.Marshal(co)
		_ = rdb.Set(ctx, key, string(b), 0).Err()
	}
	return co, nil
}

func (s *AdConfigService) SaveOrUpdate(ctx context.Context, tenantID string, serviceID int64, raw json.RawMessage) (bool, error) {
	key := rediskeys.AdConfig(tenantID, serviceID)
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return true, nil
	}
	return true, rdb.Set(ctx, key, string(raw), 0).Err()
}

func parseAdNameOpen(izOn string, out map[string]bool) {
	var arr []map[string]interface{}
	if json.Unmarshal([]byte(izOn), &arr) != nil {
		return
	}
	for _, item := range arr {
		name, _ := item["name"].(string)
		on, _ := item["on"].(bool)
		out[name] = on
	}
}

func mergeAdIzOn(template string, nameOpen map[string]bool) interface{} {
	var arr []map[string]interface{}
	if json.Unmarshal([]byte(template), &arr) != nil {
		return template
	}
	for i, item := range arr {
		name, _ := item["name"].(string)
		if v, ok := nameOpen[name]; ok {
			arr[i]["on"] = v
		} else {
			arr[i]["on"] = false
		}
	}
	return arr
}

func defaultAdsConfig() dto.AdsConfigCO {
	raw := `{"ygId":"","ygEmpId":"","coralAdId":"","wanjuyuanEnable":false,"ads":{"home":{"type":"video","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"pay":{"type":"banner","id":"","source":"coralAd","enable":false,"appId":null,"name":null,"path":null},"userInfo":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"charge":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"chargeList":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"order":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"precycling":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"wallet":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null},"withdraw":{"type":"banner","id":"","source":"wx","enable":false,"appId":null,"name":null,"path":null}},"halfMiniProgramAds":{"precycling":{"type":null,"id":null,"source":"wx","enable":false,"appId":"xxx","name":"微信原生半屏（饿了么）","path":"xxx"},"pay":{"type":null,"id":null,"source":"wx","enable":false,"appId":"xxx","name":"微信原生半屏（饿了么）","path":"xxx"},"afterPay":{"type":null,"id":null,"source":"ygEmp","enable":false,"appId":"","name":"杨梅路半屏（饿了么）","path":""}}}`
	var co dto.AdsConfigCO
	_ = json.Unmarshal([]byte(raw), &co)
	return co
}

// --- Alarm Contact ---

type AlarmContactService struct{ repo *repo.ConfigRepository }

func NewAlarmContactService(r *repo.ConfigRepository) *AlarmContactService {
	return &AlarmContactService{repo: r}
}

// alarmContactRowFromCmd maps only the provided (non-nil) cmd fields onto the model,
// leaving the rest as zero/NULL (so a partial update skips them).
func alarmContactRowFromCmd(cmd dto.AlarmContactCmd) model.AlarmContact {
	row := model.AlarmContact{}
	if cmd.Id != nil {
		row.ID = *cmd.Id
	}
	if cmd.ServiceId != nil {
		row.ServiceID = *cmd.ServiceId
	}
	if cmd.Type != nil {
		row.Type = sql.NullInt32{Int32: int32(*cmd.Type), Valid: true}
	}
	if cmd.Name != nil {
		row.Name = sql.NullString{String: *cmd.Name, Valid: true}
	}
	if cmd.Phone != nil {
		row.Phone = sql.NullString{String: *cmd.Phone, Valid: true}
	}
	if cmd.StartTime != nil {
		row.StartTime = sql.NullString{String: *cmd.StartTime, Valid: true}
	}
	if cmd.EndTime != nil {
		row.EndTime = sql.NullString{String: *cmd.EndTime, Valid: true}
	}
	if cmd.NotifyType != nil {
		row.NotifyType = sql.NullString{String: *cmd.NotifyType, Valid: true}
	}
	if cmd.TriggerArea != nil {
		row.TriggerArea = sql.NullInt32{Int32: int32(*cmd.TriggerArea), Valid: true}
	}
	return row
}

func (s *AlarmContactService) Add(ctx context.Context, tenantID, pin string, cmd dto.AlarmContactCmd) error {
	row := alarmContactRowFromCmd(cmd)
	row.ID = repo.NextConfigID()
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	return s.repo.Create(ctx, &row)
}

// Edit mirrors Java AlarmContactGatewayImpl.updateAlarmContact (MyBatis updateById):
// only the fields present in the request are updated; omitted fields are untouched.
//
// NOTE: the original Go code used mergeJSON + full Save, which wrote every column
// (zeroing out fields the caller did not send). Preserved for reference:
//
//	row := model.AlarmContact{ID: *cmd.Id}
//	mergeJSON(&row, cmd)
//	repo.StampBase(&row.BaseConfig, tenantID, pin, false)
//	return s.repo.Save(ctx, &row)
func (s *AlarmContactService) Edit(ctx context.Context, tenantID, pin string, cmd dto.AlarmContactCmd) error {
	if cmd.Id == nil {
		return errors.New("id must not be null")
	}
	row := alarmContactRowFromCmd(cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, false)
	return s.repo.UpdateNonZero(ctx, &row)
}

func (s *AlarmContactService) Del(ctx context.Context, cmd dto.AlarmContactCmd) error {
	if cmd.Id == nil {
		return errors.New("id must not be null")
	}
	return s.repo.DeleteByID(ctx, &model.AlarmContact{}, *cmd.Id)
}

func (s *AlarmContactService) Page(ctx context.Context, q dto.ContactQuery) (*dto.PageDTO[dto.AlarmContactCO], error) {
	if q.ServiceId == nil || q.Type == nil {
		return nil, errors.New("serviceId and type required")
	}
	rows, total, err := s.repo.PageAlarmContacts(ctx, *q.ServiceId, *q.Type, q.PageNum, q.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.AlarmContactCO, len(rows))
	for i, r := range rows {
		list[i] = convert.AlarmContactToCO(r)
	}
	return &dto.PageDTO[dto.AlarmContactCO]{
		List: list, Total: total, PageNum: q.PageNum, PageSize: q.PageSize,
	}, nil
}

func (s *AlarmContactService) Select(ctx context.Context, q dto.AlarmQuery) ([]dto.AlarmContactCO, error) {
	if q.ServiceId == nil || q.Type == nil {
		return nil, errors.New("serviceId and type required")
	}
	rows, err := s.repo.ListAlarmContacts(ctx, *q.ServiceId, *q.Type)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AlarmContactCO, len(rows))
	for i, r := range rows {
		out[i] = convert.AlarmContactToCO(r)
	}
	return out, nil
}

// --- Riding Permission ---

type RidingPermissionService struct{ repo *repo.ConfigRepository }

func NewRidingPermissionService(r *repo.ConfigRepository) *RidingPermissionService {
	return &RidingPermissionService{repo: r}
}

func (s *RidingPermissionService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.RidingPermissionCO, error) {
	key := rediskeys.RidingPermission(tenantID, serviceID)
	var co dto.RidingPermissionCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.RidingPermission
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.RidingPermissionToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.ServiceId == nil {
		row := &model.RidingPermission{ID: repo.NextConfigID(), ServiceID: serviceID}
		repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
		if err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
			return s.repo.Create(ctx, row)
		}); err != nil {
			co.ServiceId = &serviceID
			return &co, nil
		}
		return s.Get(ctx, tenantID, serviceID)
	}
	return &co, nil
}

func (s *RidingPermissionService) Insert(ctx context.Context, tenantID, pin string, cmd dto.RidingPermissionCmd) error {
	if cmd.ServiceId == nil {
		return errors.New("serviceId must not be null")
	}
	row := model.RidingPermission{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.RidingPermission(tenantID, *cmd.ServiceId)
	return s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
}

// --- Riding Car Config ---

type RidingCarConfigService struct {
	useCar *UseCarService
	pay    *PayService
}

func NewRidingCarConfigService() *RidingCarConfigService {
	return &RidingCarConfigService{}
}

func (s *RidingCarConfigService) bind(useCar *UseCarService, pay *PayService) {
	s.useCar = useCar
	s.pay = pay
}

func (s *RidingCarConfigService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.RidingCarConfigCO, error) {
	// Java RidingCarConfigServiceImpl calls useCarService.get() without commandContext;
	// ConfigUseCarServiceImpl NPEs when cancelAuthTenantIds is non-empty.
	if len(config.GlobalConfig.Xyy.CancelAuthTenantIds) > 0 {
		return nil, service.ErrRidingCarConfigNullContext
	}
	useCO, err := s.useCar.Get(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	payCO, err := s.pay.Get(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	return &dto.RidingCarConfigCO{UseCarCO: useCO, PayCO: payCO}, nil
}

// --- Protocol ---

type ProtocolService struct{ repo *repo.ConfigRepository }

func NewProtocolService(r *repo.ConfigRepository) *ProtocolService { return &ProtocolService{repo: r} }

func (s *ProtocolService) List(ctx context.Context, tenantID, pin string, serviceID int64) ([]dto.ConfigProtocolCO, error) {
	rows, err := s.repo.ListProtocolsByService(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		defaults, err := s.repo.DefaultProtocols(ctx)
		if err != nil {
			return nil, err
		}
		tn, cn := s.tenantNames(ctx, tenantID)
		for _, d := range defaults {
			row := d
			row.ID = repo.NextConfigID()
			row.ServiceID = serviceID
			row.Content = replaceProtocolPlaceholders(row.Content, tn, cn)
			repo.StampBase(&row.BaseConfig, tenantID, pin, true)
			_ = s.repo.Create(ctx, &row)
		}
		rows, err = s.repo.ListProtocolsByService(ctx, serviceID)
		if err != nil {
			return nil, err
		}
	}
	out := make([]dto.ConfigProtocolCO, 0, len(rows))
	for i := range rows {
		if co := convert.ProtocolToCO(&rows[i]); co != nil {
			out = append(out, *co)
		}
	}
	return out, nil
}

func (s *ProtocolService) Update(ctx context.Context, tenantID string, cmd dto.ConfigProtocolUpdateCmd) (bool, error) {
	if cmd.Id == nil || cmd.ServiceId == nil || cmd.Type == nil {
		return false, errors.New("id, serviceId, type required")
	}
	row := model.ConfigProtocol{ID: *cmd.Id, ServiceID: *cmd.ServiceId, Type: *cmd.Type}
	if cmd.Title != nil {
		row.Title = *cmd.Title
	}
	if cmd.Content != nil {
		row.Content = *cmd.Content
	}
	key := rediskeys.ProtocolConfigByType(tenantID, *cmd.ServiceId, strconv.Itoa(*cmd.Type))
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Save(ctx, &row)
	})
	return err == nil, err
}

func (s *ProtocolService) SetDefault(ctx context.Context, tenantID string, list []dto.ConfigProtocolUpdateCmd) (bool, error) {
	for _, item := range list {
		if item.ServiceId == nil || item.Type == nil {
			continue
		}
		def, err := s.cloneDefaultWithPlaceholders(ctx, tenantID, *item.ServiceId, *item.Type)
		if err != nil || def == nil {
			return false, errors.New("未配置此默认协议")
		}
		content := def.Content
		item.Content = &content
		if _, err := s.Update(ctx, tenantID, item); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *ProtocolService) ByType(ctx context.Context, tenantID string, serviceID int64, typ int) (*dto.ConfigProtocolCO, error) {
	key := rediskeys.ProtocolConfigByType(tenantID, serviceID, strconv.Itoa(typ))
	var co dto.ConfigProtocolCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		row, err := s.repo.GetProtocolByType(ctx, serviceID, typ)
		if err != nil {
			return "", err
		}
		if row == nil {
			def, _ := s.cloneDefaultWithPlaceholders(ctx, tenantID, serviceID, typ)
			if def == nil {
				return "", nil
			}
			return convert.MarshalRow(convert.ProtocolToCO(def))
		}
		return convert.MarshalRow(convert.ProtocolToCO(row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.Id == nil {
		def, _ := s.cloneDefaultWithPlaceholders(ctx, tenantID, serviceID, typ)
		return convert.ProtocolToCO(def), nil
	}
	return &co, nil
}

func (s *ProtocolService) DefaultByType(ctx context.Context, tenantID string, typ int) (*dto.ConfigProtocolCO, error) {
	def, err := s.cloneDefaultWithPlaceholders(ctx, tenantID, 0, typ)
	if err != nil {
		return nil, err
	}
	return convert.ProtocolToCO(def), nil
}

func (s *ProtocolService) defaultByType(ctx context.Context, serviceID int64, typ int) (*model.ConfigProtocol, error) {
	defaults, err := s.repo.DefaultProtocols(ctx)
	if err != nil || len(defaults) == 0 {
		return nil, errors.New("用户协议未初始化")
	}
	for _, d := range defaults {
		if d.Type == typ {
			cp := d
			cp.ServiceID = serviceID
			return &cp, nil
		}
	}
	if typ > 0 && typ <= len(defaults) {
		cp := defaults[typ-1]
		cp.ServiceID = serviceID
		return &cp, nil
	}
	return nil, nil
}

// --- Fence Tag ---

type FenceTagService struct{ repo *repo.ConfigRepository }

func NewFenceTagService(r *repo.ConfigRepository) *FenceTagService { return &FenceTagService{repo: r} }

func (s *FenceTagService) GetAll(ctx context.Context, tenantID string) ([]dto.FenceTagCO, error) {
	key := rediskeys.FenceTagListCache(tenantID)
	raw, err := s.repo.Store().GetList(ctx, key, func(db *gorm.DB) (string, error) {
		rows, err := s.repo.ListFenceTags(ctx)
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(rows)
		return string(b), err
	})
	if err != nil {
		return nil, err
	}
	if rediskeys.IsCacheMiss(raw) {
		return []dto.FenceTagCO{}, nil
	}
	var rows []model.FenceTag
	rows, err = cache.UnmarshalFenceTagList(raw)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceTagCO, len(rows))
	for i, r := range rows {
		out[i] = convert.FenceTagToCO(r)
	}
	return out, nil
}
