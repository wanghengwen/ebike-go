package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence/cache"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/idgen"
	"ebike-fence-go/internal/pkg/timefmt"

	"gorm.io/gorm"
)

type HelpConfigRepo struct {
	store *gateway.ConfigStore
	db    *gorm.DB
}

func NewHelpConfigRepo(db *gorm.DB) *HelpConfigRepo {
	return &HelpConfigRepo{store: gateway.NewConfigStore(db), db: db}
}

func (r *HelpConfigRepo) DB() *gorm.DB { return r.db }

// assignHelpConfigInsertID mirrors Java FenceKeyGenerator on insert (fastId.nextId()).
func assignHelpConfigInsertID(rowID *int64) {
	if rowID != nil && *rowID <= 0 {
		*rowID = idgen.NextID()
	}
}

func applyAudit(base *model.ConfigBaseDO, tenantID, pin string, now time.Time) {
	base.TenantID = tenantID
	base.UpdatedPin = pin
	base.UpdatedAt = timefmt.FromTime(now)
	if base.CreatedAt.IsZero() {
		base.CreatedPin = pin
		base.CreatedAt = timefmt.FromTime(now)
	}
	if !base.IzDel.Valid {
		base.IzDel = sql.NullBool{Bool: false, Valid: true}
	}
}

func scopeHelpConfig(db *gorm.DB, tenantID string, serviceID int64) *gorm.DB {
	q := db
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	return q
}

// --- HomeScrollerMsg ---

func (r *HelpConfigRepo) ListHomeScrollMsg(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigHomeScrollMsg, error) {
	key := rediskeys.HomeScrollMsg(tenantID, serviceID)
	raw, err := r.store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		var rows []model.TConfigHomeScrollMsg
		if err := scopeHelpConfig(db, tenantID, serviceID).Order("updated_at DESC").Find(&rows).Error; err != nil {
			return "", err
		}
		return cache.MarshalHomeScrollMsgList(rows)
	})
	if err != nil || rediskeys.IsCacheMiss(raw) {
		return nil, err
	}
	var rows []model.TConfigHomeScrollMsg
	rows, err = cache.UnmarshalHomeScrollMsgList(raw)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *HelpConfigRepo) GetHomeScrollMsgByID(ctx context.Context, id int64) (*model.TConfigHomeScrollMsg, error) {
	var row model.TConfigHomeScrollMsg
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *HelpConfigRepo) InsertHomeScrollMsg(ctx context.Context, tenantID, pin string, row *model.TConfigHomeScrollMsg) error {
	key := rediskeys.HomeScrollMsg(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateHomeScrollMsg(ctx context.Context, tenantID, pin string, row *model.TConfigHomeScrollMsg) error {
	key := rediskeys.HomeScrollMsg(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields(homeScrollMsgUpdateFields)...)
	})
}

func (r *HelpConfigRepo) DeleteHomeScrollMsg(ctx context.Context, tenantID string, serviceID, id int64) error {
	key := rediskeys.HomeScrollMsg(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigHomeScrollMsg{}, id).Error
	})
}

// --- FAQ ---

func (r *HelpConfigRepo) ListFaq(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigFaq, error) {
	key := rediskeys.FAQ(tenantID, serviceID)
	raw, err := r.store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		var rows []model.TConfigFaq
		if err := scopeHelpConfig(db, tenantID, serviceID).Find(&rows).Error; err != nil {
			return "", err
		}
		return cache.MarshalFaqList(rows)
	})
	if err != nil {
		return nil, err
	}
	if rediskeys.IsCacheMiss(raw) {
		return []model.TConfigFaq{}, nil
	}
	var rows []model.TConfigFaq
	rows, err = cache.UnmarshalFaqList(raw)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *HelpConfigRepo) GetFaqByID(ctx context.Context, id int64) (*model.TConfigFaq, error) {
	var row model.TConfigFaq
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *HelpConfigRepo) InsertFaq(ctx context.Context, tenantID, pin string, row *model.TConfigFaq) error {
	key := rediskeys.FAQ(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateFaq(ctx context.Context, tenantID, pin string, row *model.TConfigFaq) error {
	key := rediskeys.FAQ(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields(faqUpdateFields)...)
	})
}

func (r *HelpConfigRepo) DeleteFaq(ctx context.Context, tenantID string, serviceID, id int64) error {
	key := rediskeys.FAQ(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigFaq{}, id).Error
	})
}

// --- GuidePage ---

// ListGuidePage mirrors Java GuidePageQueryImpl.getGuidePageConfigByServiceId,
// which filters only by service_id — izDel has no @TableLogic in Java's BaseDO,
// so it is never used as a soft-delete filter there (delete is a hard DELETE).
func (r *HelpConfigRepo) ListGuidePage(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigGuidePage, error) {
	var rows []model.TConfigGuidePage
	err := scopeHelpConfig(r.db.WithContext(ctx), tenantID, serviceID).
		Order("order_weights DESC, created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *HelpConfigRepo) GetGuidePageIzOnCached(ctx context.Context, tenantID string, serviceID int64) (*model.TConfigGuidePage, error) {
	key := rediskeys.GuidePage(tenantID, serviceID)
	raw, err := r.store.GetJSON(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.TConfigGuidePage
		err := scopeHelpConfig(db, tenantID, serviceID).Where("iz_on = ?", true).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "null", nil
		}
		if err != nil {
			return "", err
		}
		return cache.MarshalGuidePage(row)
	})
	if err != nil {
		return nil, err
	}
	if raw == "" || raw == "null" {
		return nil, nil
	}
	row, err := cache.UnmarshalGuidePage(raw)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (r *HelpConfigRepo) InsertGuidePage(ctx context.Context, tenantID, pin string, row *model.TConfigGuidePage) error {
	key := rediskeys.GuidePage(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		assignHelpConfigInsertID(&row.ID)
		if row.OrderWeights == 0 {
			row.OrderWeights = 9
		}
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateGuidePage(ctx context.Context, tenantID, pin string, row *model.TConfigGuidePage, disableOthers bool) error {
	key := rediskeys.GuidePage(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		if disableOthers && row.IzOn != nil && *row.IzOn {
			if err := scopeHelpConfig(db, tenantID, row.ServiceID).Model(&model.TConfigGuidePage{}).
				Where("id <> ? AND iz_on = ?", row.ID, true).
				Update("iz_on", false).Error; err != nil {
				return err
			}
		}
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields(guidePageUpdateFields)...)
	})
}

func (r *HelpConfigRepo) DeleteGuidePage(ctx context.Context, tenantID string, serviceID, id int64) error {
	key := rediskeys.GuidePage(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigGuidePage{}, id).Error
	})
}

func (r *HelpConfigRepo) SortGuidePage(ctx context.Context, tenantID, pin string, serviceID int64, ids []int64) error {
	key := rediskeys.GuidePage(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		order := 9
		now := gateway.Now()
		for _, id := range ids {
			if err := scopeHelpConfig(db, tenantID, serviceID).Model(&model.TConfigGuidePage{}).Where("id = ?", id).
				Updates(map[string]interface{}{
					"order_weights": order,
					"updated_pin":   pin,
					"updated_at":    now,
				}).Error; err != nil {
				return err
			}
			order--
		}
		return nil
	})
}

// --- HomeActivityEntrance ---

func (r *HelpConfigRepo) ListHomeActivityEntrance(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigHomeActivityEntrance, error) {
	key := rediskeys.HomeActivityEntrance(tenantID, serviceID)
	raw, err := r.store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		var rows []model.TConfigHomeActivityEntrance
		if err := scopeHelpConfig(db, tenantID, serviceID).Find(&rows).Error; err != nil {
			return "", err
		}
		return cache.MarshalHomeActivityEntranceList(rows)
	})
	if err != nil || rediskeys.IsCacheMiss(raw) {
		return nil, err
	}
	var rows []model.TConfigHomeActivityEntrance
	rows, err = cache.UnmarshalHomeActivityEntranceList(raw)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *HelpConfigRepo) GetHomeActivityEntranceByID(ctx context.Context, id int64) (*model.TConfigHomeActivityEntrance, error) {
	var row model.TConfigHomeActivityEntrance
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *HelpConfigRepo) InsertHomeActivityEntrance(ctx context.Context, tenantID, pin string, row *model.TConfigHomeActivityEntrance) error {
	key := rediskeys.HomeActivityEntrance(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateHomeActivityEntrance(ctx context.Context, tenantID, pin string, row *model.TConfigHomeActivityEntrance) error {
	key := rediskeys.HomeActivityEntrance(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		if err := helpConfigUpdate(db, row, appendHelpConfigAuditFields(homeActivityUpdateFields)...); err != nil {
			return err
		}
		// Java additionally force-sets start_time/end_time via UpdateWrapper.set(...),
		// so they must be written even when nil (clearing the activity window). A struct
		// Updates would skip nil pointers, so do an explicit map update here.
		return db.Model(&model.TConfigHomeActivityEntrance{}).
			Where("id = ?", row.ID).
			Updates(map[string]interface{}{
				"start_time": row.StartTime,
				"end_time":   row.EndTime,
			}).Error
	})
}

func (r *HelpConfigRepo) DeleteHomeActivityEntrance(ctx context.Context, tenantID string, serviceID, id int64) error {
	key := rediskeys.HomeActivityEntrance(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigHomeActivityEntrance{}, id).Error
	})
}

// --- SpecialTips ---

func (r *HelpConfigRepo) ListSpecialTips(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigSpecialTips, error) {
	key := rediskeys.SpecialTips(tenantID, serviceID)
	raw, err := r.store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		var rows []model.TConfigSpecialTips
		if err := scopeHelpConfig(db, tenantID, serviceID).Find(&rows).Error; err != nil {
			return "", err
		}
		return cache.MarshalSpecialTipsList(rows)
	})
	if err != nil || rediskeys.IsCacheMiss(raw) {
		return nil, err
	}
	var rows []model.TConfigSpecialTips
	rows, err = cache.UnmarshalSpecialTipsList(raw)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *HelpConfigRepo) GetSpecialTipsByID(ctx context.Context, id int64) (*model.TConfigSpecialTips, error) {
	var row model.TConfigSpecialTips
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *HelpConfigRepo) InsertSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	key := rediskeys.SpecialTips(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	key := rediskeys.SpecialTips(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields(specialTipsUpdateFields)...)
	})
}

func (r *HelpConfigRepo) DeleteSpecialTips(ctx context.Context, tenantID string, serviceID, id int64) error {
	key := rediskeys.SpecialTips(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigSpecialTips{}, id).Error
	})
}

func (r *HelpConfigRepo) OnOffSpecialTips(ctx context.Context, tenantID, pin string, row *model.TConfigSpecialTips) error {
	key := rediskeys.SpecialTips(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		if row.IzOn != nil && *row.IzOn {
			var others []model.TConfigSpecialTips
			if err := scopeHelpConfig(db, tenantID, row.ServiceID).Where("iz_on = ?", true).Find(&others).Error; err != nil {
				return err
			}
			now := gateway.Now()
			for _, o := range others {
				if o.PopUpTime == row.PopUpTime && o.IzOn != nil && *o.IzOn {
					if err := db.Model(&model.TConfigSpecialTips{}).Where("id = ?", o.ID).
						Updates(map[string]interface{}{"iz_on": false, "updated_pin": pin, "updated_at": now}).Error; err != nil {
						return err
					}
				}
			}
		}
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields([]string{"IzOn"})...)
	})
}

// --- CustomerService ---

// applyCustomerServiceInsertDefaults fills NOT NULL boolean columns before insert.
// MyBatis-Plus omits null fields so MySQL defaults apply; GORM writes explicit NULL otherwise.
func applyCustomerServiceInsertDefaults(row *model.TConfigCustomerService) {
	if row == nil {
		return
	}
	if row.IzOnlineEntrance == nil {
		v := false
		row.IzOnlineEntrance = &v
	}
	if row.IzArtificialEntrance == nil {
		v := false
		row.IzArtificialEntrance = &v
	}
	if row.IzWorkTime == nil {
		v := false
		row.IzWorkTime = &v
	}
}

func (r *HelpConfigRepo) GetCustomerService(ctx context.Context, tenantID string, serviceID int64) (*model.TConfigCustomerService, error) {
	key := rediskeys.CustomerService(tenantID, serviceID)
	raw, err := r.store.GetJSON(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.TConfigCustomerService
		err := scopeHelpConfig(db, tenantID, serviceID).Order("updated_at DESC").Limit(1).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "null", nil
		}
		if err != nil {
			return "", err
		}
		return cache.MarshalCustomerService(row)
	})
	if err != nil {
		return nil, err
	}
	if raw == "" || raw == "null" {
		return nil, nil
	}
	row, err := cache.UnmarshalCustomerService(raw)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (r *HelpConfigRepo) InsertCustomerService(ctx context.Context, tenantID, pin string, row *model.TConfigCustomerService) error {
	key := rediskeys.CustomerService(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		applyCustomerServiceInsertDefaults(row)
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

// --- HomeNav ---

type HomeNavDefault struct {
	ChainType int
	LinkUrl   string
	Icon      string
	Name      string
}

var homeNavDefaults = []HomeNavDefault{
	{0, "/pagesSub/protocol/protocol", "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/shortcut_protocol.png", "条款与协议"},
	{0, "/pagesSub/customerService/customerService", "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/shortcut_customerService.png", "联系客服"},
	{0, "/pagesSub/repair/repair", "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/shortcut_repair.png", "车辆报修"},
	{0, "/pagesSub/accountRules/accountRules", "https://feifan-frontend.oss-cn-shenzhen.aliyuncs.com/miniapp/project/feifan/shortcut_priceRules.png", "计费规则"},
}

func (r *HelpConfigRepo) ListHomeNav(ctx context.Context, tenantID, pin string, serviceID int64) ([]model.TConfigHomeNav, error) {
	key := rediskeys.HomeNav(tenantID, serviceID)
	raw, err := r.store.GetList(ctx, key, func(db *gorm.DB) (string, error) {
		rows, err := r.loadHomeNavRows(ctx, tenantID, pin, serviceID)
		if err != nil {
			return "", err
		}
		return cache.MarshalHomeNavList(rows)
	})
	if err != nil || rediskeys.IsCacheMiss(raw) {
		return nil, err
	}
	var rows []model.TConfigHomeNav
	rows, err = cache.UnmarshalHomeNavList(raw)
	if err != nil {
		return nil, err
	}
	if len(rows) < 4 {
		return r.ensureHomeNavDefaults(ctx, tenantID, pin, serviceID)
	}
	return rows, nil
}

func (r *HelpConfigRepo) fetchHomeNavRowsFromDB(ctx context.Context, tenantID string, serviceID int64) ([]model.TConfigHomeNav, error) {
	var rows []model.TConfigHomeNav
	err := scopeHelpConfig(r.db.WithContext(ctx), tenantID, serviceID).
		Order("order_weights DESC, created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *HelpConfigRepo) loadHomeNavRows(ctx context.Context, tenantID, pin string, serviceID int64) ([]model.TConfigHomeNav, error) {
	rows, err := r.fetchHomeNavRowsFromDB(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if len(rows) < 4 {
		return r.ensureHomeNavDefaults(ctx, tenantID, pin, serviceID)
	}
	return rows, nil
}

// ensureHomeNavDefaults inserts missing default nav items when total count is below 4.
// Always reads the DB first so stale/partial cache cannot duplicate existing defaults.
func (r *HelpConfigRepo) ensureHomeNavDefaults(ctx context.Context, tenantID, pin string, serviceID int64) ([]model.TConfigHomeNav, error) {
	dbList, err := r.fetchHomeNavRowsFromDB(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if pin == "" {
		pin = "@system"
	}
	nameList := make([]string, 0, len(dbList))
	for _, n := range dbList {
		nameList = append(nameList, n.Name)
	}
	count := 0
	if len(dbList) > 0 && len(dbList) < 4 {
		count = 4 - len(dbList)
	}
	if len(dbList) == 0 {
		count = 4
	}
	key := rediskeys.HomeNav(tenantID, serviceID)
	if count == 0 {
		r.store.Invalidate(ctx, key)
		return dbList, nil
	}
	now := gateway.Now()
	falseVal := false
	for i := 0; i < count; i++ {
		for _, def := range homeNavDefaults {
			if containsString(nameList, def.Name) {
				continue
			}
			jump := map[string]interface{}{
				"chainType": def.ChainType,
				"linkUrl":   def.LinkUrl,
			}
			jumpJSON, _ := json.Marshal(jump)
			row := model.TConfigHomeNav{
				ServiceID: serviceID,
				Icon:      def.Icon,
				Name:      def.Name,
				JumpPage:  string(jumpJSON),
				IzOn:      &falseVal,
			}
			applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
			assignHelpConfigInsertID(&row.ID)
			if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
				return nil, err
			}
			nameList = append(nameList, def.Name)
			break
		}
	}
	list, err := r.fetchHomeNavRowsFromDB(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	r.store.Invalidate(ctx, key)
	return list, nil
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// HomeNavHasSameName reports whether another nav item in the same service area has the given name.
// serviceId is resolved from the row identified by id (Java HomeNavQueryImpl.updateHaveSameName).
func (r *HelpConfigRepo) HomeNavHasSameName(ctx context.Context, tenantID string, id int64, name string) (bool, error) {
	var current model.TConfigHomeNav
	q := r.db.WithContext(ctx).Model(&model.TConfigHomeNav{})
	if tenantID != "" {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.First(&current, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	var count int64
	err := scopeHelpConfig(r.db.WithContext(ctx).Model(&model.TConfigHomeNav{}), tenantID, current.ServiceID).
		Where("id <> ? AND name = ?", id, name).
		Count(&count).Error
	return count > 0, err
}

func (r *HelpConfigRepo) CreateHomeNav(ctx context.Context, tenantID, pin string, row *model.TConfigHomeNav) error {
	key := rediskeys.HomeNav(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		assignHelpConfigInsertID(&row.ID)
		return db.Create(row).Error
	})
}

func (r *HelpConfigRepo) UpdateHomeNav(ctx context.Context, tenantID, pin string, row *model.TConfigHomeNav) error {
	key := rediskeys.HomeNav(tenantID, row.ServiceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		now := gateway.Now()
		applyAudit(&row.ConfigBaseDO, tenantID, pin, now)
		return helpConfigUpdate(db, row, appendHelpConfigAuditFields(homeNavUpdateFields)...)
	})
}

func (r *HelpConfigRepo) DeleteHomeNav(ctx context.Context, tenantID string, serviceID, id int64) error {
	var row model.TConfigHomeNav
	if err := scopeHelpConfig(r.db.WithContext(ctx), tenantID, serviceID).First(&row, id).Error; err != nil {
		return err
	}
	if row.IzOn != nil && *row.IzOn {
		return errors.New("CONFIG_DELETE_FAIL")
	}
	key := rediskeys.HomeNav(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return scopeHelpConfig(db, tenantID, serviceID).Delete(&model.TConfigHomeNav{}, id).Error
	})
}

func (r *HelpConfigRepo) HomeNavIsOpen(ctx context.Context, tenantID, pin string, serviceID int64, ids []int64, izOn bool) error {
	if izOn && len(ids) < 4 {
		return errors.New("DISABLE_OPEN")
	}
	key := rediskeys.HomeNav(tenantID, serviceID)
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		val := 0
		if izOn {
			val = 1
		}
		return db.Model(&model.TConfigHomeNav{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"iz_on":       val,
				"updated_pin": pin,
				"updated_at":  gateway.Now(),
			}).Error
	})
}

func (r *HelpConfigRepo) SortHomeNav(ctx context.Context, tenantID, pin string, serviceID int64, ids []int64) error {
	key := rediskeys.HomeNav(tenantID, serviceID)
	reversed := append([]int64(nil), ids...)
	sort.Slice(reversed, func(i, j int) bool { return i > j })
	return r.store.WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		weight := 0
		now := gateway.Now()
		for _, id := range reversed {
			if err := db.Model(&model.TConfigHomeNav{}).Where("id = ?", id).
				Updates(map[string]interface{}{
					"order_weights": weight,
					"updated_pin":   pin,
					"updated_at":    now,
				}).Error; err != nil {
				return err
			}
			weight++
		}
		return nil
	})
}

func boolTrue(v *bool) bool { return v != nil && *v }

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func tagsIntersect(active, user string) bool {
	a := splitTags(active)
	u := splitTags(user)
	if len(a) == 0 || len(u) == 0 {
		return false
	}
	for _, x := range a {
		for _, y := range u {
			if x == y {
				return true
			}
		}
	}
	return false
}
