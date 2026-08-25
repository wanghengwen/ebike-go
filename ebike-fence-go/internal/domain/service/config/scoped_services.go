package configsvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/mq"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/pkg/config"

	"gorm.io/gorm"
)

// --- Pay ---

type PayService struct{ repo *repo.ConfigRepository }

func NewPayService(r *repo.ConfigRepository) *PayService { return &PayService{repo: r} }

func (s *PayService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigPayCO, error) {
	key := rediskeys.PayConfig(tenantID, serviceID)
	var co dto.ConfigPayCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.ConfigPay
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.PayToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.ServiceId == nil {
		row := repo.NewDefaultPayRow(repo.NextConfigID(), serviceID)
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

func (s *PayService) Insert(ctx context.Context, tenantID, pin string, cmd dto.ConfigPayCmd) error {
	if cmd.ServiceId == nil {
		return errors.New("serviceId must not be null")
	}
	row := model.ConfigPay{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.PayConfig(tenantID, *cmd.ServiceId)
	return s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
}

// --- Use Car ---

type UseCarService struct{ repo *repo.ConfigRepository }

func NewUseCarService(r *repo.ConfigRepository) *UseCarService { return &UseCarService{repo: r} }

func (s *UseCarService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigUseCarCO, error) {
	co, fromRedis, err := s.loadUseCarCO(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if co != nil && co.ServiceId != nil {
		return finalizeUseCarCO(tenantID, co, fromRedis), nil
	}
	key := rediskeys.UseCarConfig(tenantID, serviceID)
	row := repo.NewDefaultUseCarRow(repo.NextConfigID(), serviceID)
	repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
	if err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, row)
	}); err != nil {
		fallback := &dto.ConfigUseCarCO{ServiceId: &serviceID}
		return finalizeUseCarCO(tenantID, fallback, false), nil
	}
	co, fromRedis, err = s.loadUseCarCO(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if co == nil || co.ServiceId == nil {
		fallback := &dto.ConfigUseCarCO{ServiceId: &serviceID}
		return finalizeUseCarCO(tenantID, fallback, false), nil
	}
	return finalizeUseCarCO(tenantID, co, fromRedis), nil
}

func applyUseCarNeedAuth(tenantID string, co *dto.ConfigUseCarCO) {
	if co == nil {
		return
	}
	needAuth := true
	for _, id := range config.GlobalConfig.Xyy.CancelAuthTenantIds {
		if id == tenantID {
			needAuth = false
			break
		}
	}
	co.IzNeedAuth = &needAuth
}

func (s *UseCarService) Insert(ctx context.Context, tenantID, pin string, cmd dto.ConfigUseCarCmd) error {
	if cmd.ServiceId == nil {
		return &service.BizError{Code: dto.CodeIllegalArgument, Msg: "serviceId must not be null"}
	}
	return s.insertRow(ctx, tenantID, pin, cmd)
}

func (s *UseCarService) UpdateRecharge(ctx context.Context, tenantID, pin string, cmd dto.ConfigUseCarCmd) error {
	if cmd.ServiceId == nil {
		return errors.New("serviceId must not be null")
	}
	// Java ConfigUseCarController.updateRecharge: load the full current config,
	// overwrite ONLY the recharge fields, copy back into the full cmd (id=null) and
	// insert a new version - preserving every other use-car config field.
	//
	// NOTE: the original Go code only carried serviceId + the 2 recharge fields into
	// the insert, which wiped out all other config. Preserved for reference:
	//
	//	merged := dto.ConfigUseCarCmd{ServiceId: cmd.ServiceId, RechargeBeforeUse: pre.RechargeBeforeUse, RechargeCost: pre.RechargeCost}
	//	return s.insertRow(ctx, tenantID, pin, merged)
	pre, err := s.Get(ctx, tenantID, *cmd.ServiceId)
	if err != nil {
		return err
	}
	pre.RechargeBeforeUse = cmd.RechargeBeforeUse
	pre.RechargeCost = cmd.RechargeCost
	// copyProperties(pre -> ConfigUseCarCmd) via JSON round-trip (keeps all fields).
	var merged dto.ConfigUseCarCmd
	mergeJSON(&merged, pre)
	merged.Id = nil
	merged.ServiceId = cmd.ServiceId
	return s.insertRow(ctx, tenantID, pin, merged)
}

func (s *UseCarService) insertRow(ctx context.Context, tenantID, pin string, cmd dto.ConfigUseCarCmd) error {
	row := model.ConfigUseCar{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.UseCarConfig(tenantID, *cmd.ServiceId)
	return s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
}

// --- Base Item ---

type BaseItemService struct{ repo *repo.ConfigRepository }

func NewBaseItemService(r *repo.ConfigRepository) *BaseItemService { return &BaseItemService{repo: r} }

func (s *BaseItemService) Get(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBaseItemCO, error) {
	key := rediskeys.BaseItemConfig(tenantID, serviceID)
	raw, err := s.repo.Store().GetJSON(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.ConfigBaseItem
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalBaseItemForRedis(&row)
	})
	if err != nil {
		return nil, err
	}
	co, err := convert.UnmarshalBaseItemCO(raw)
	if err != nil {
		return nil, err
	}
	if co.ServiceId == nil {
		row := &model.ConfigBaseItem{ID: repo.NextConfigID(), ServiceID: serviceID}
		repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
		if err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
			return s.repo.Create(ctx, row)
		}); err != nil {
			co.ServiceId = &serviceID
			return co, nil
		}
		return s.Get(ctx, tenantID, serviceID)
	}
	return co, nil
}

func (s *BaseItemService) Insert(ctx context.Context, tenantID, pin string, cmd dto.ConfigBaseItemCmd) error {
	if cmd.ServiceId == nil {
		return errors.New("serviceId must not be null")
	}
	row := model.ConfigBaseItem{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.BaseItemConfig(tenantID, *cmd.ServiceId)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
	if err != nil {
		return err
	}
	mq.SynchronizeConfig(ctx, mq.ConfigUpdateDTO{
		Config:    1,
		TenantId:  tenantID,
		ServiceId: *cmd.ServiceId,
	})
	return nil
}

func (s *BaseItemService) InsertSwapBatteryThreshold(ctx context.Context, tenantID, pin string, cmd dto.ConfigBaseItemCmd) error {
	return s.insertFromLatest(ctx, tenantID, pin, *cmd.ServiceId, func(row *model.ConfigBaseItem) {
		if cmd.SwapBatteryThreshold != nil {
			row.SwapBatteryThreshold = sql.NullInt32{Int32: int32(*cmd.SwapBatteryThreshold), Valid: true}
		}
		if cmd.IzAutoSwapBattery != nil {
			row.IzAutoSwapBattery = sql.NullBool{Bool: *cmd.IzAutoSwapBattery, Valid: true}
		}
	})
}

func (s *BaseItemService) InsertUserTicketPhotoWays(ctx context.Context, tenantID, pin string, cmd dto.ConfigBaseItemCmd) error {
	ways := "0"
	if len(cmd.UserTicketPhotoWays) > 0 {
		parts := make([]string, len(cmd.UserTicketPhotoWays))
		for i, v := range cmd.UserTicketPhotoWays {
			parts[i] = strconv.Itoa(v)
		}
		ways = strings.Join(parts, ",")
	}
	return s.insertFromLatest(ctx, tenantID, pin, *cmd.ServiceId, func(row *model.ConfigBaseItem) {
		row.UserTicketPhotoWays = sql.NullString{String: ways, Valid: true}
	})
}

func (s *BaseItemService) insertFromLatest(ctx context.Context, tenantID, pin string, serviceID int64, apply func(*model.ConfigBaseItem)) error {
	var latest model.ConfigBaseItem
	_ = s.repo.GetLatestByServiceID(ctx, tenantID, &latest, serviceID)
	row := latest
	row.ID = repo.NextConfigID()
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	apply(&row)
	key := rediskeys.BaseItemConfig(tenantID, serviceID)
	return s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
}

// --- Park Apply / Push Riding Card ---

type ParkApplyService struct{ repo *repo.ConfigRepository }

func NewParkApplyService(r *repo.ConfigRepository) *ParkApplyService { return &ParkApplyService{repo: r} }

func (s *ParkApplyService) GetByServiceID(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigParkApplyCO, error) {
	key := rediskeys.ParkApplyConfig(tenantID, serviceID)
	var co dto.ConfigParkApplyCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.ParkApplyConfig
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.ParkApplyToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.ServiceId == nil {
		row := repo.NewDefaultParkApplyRow(repo.NextConfigID(), serviceID)
		repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
		if err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
			return s.repo.Create(ctx, row)
		}); err != nil {
			co.ServiceId = &serviceID
			return &co, nil
		}
		return s.GetByServiceID(ctx, tenantID, serviceID)
	}
	return &co, nil
}

func (s *ParkApplyService) Insert(ctx context.Context, tenantID, pin string, cmd dto.ConfigParkApplyCmd) (bool, error) {
	if cmd.ServiceId == nil {
		return false, errors.New("serviceId must not be null")
	}
	row := model.ParkApplyConfig{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.ParkApplyConfig(tenantID, *cmd.ServiceId)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
	return err == nil, err
}

type PushRidingCardService struct{ repo *repo.ConfigRepository }

func NewPushRidingCardService(r *repo.ConfigRepository) *PushRidingCardService {
	return &PushRidingCardService{repo: r}
}

func (s *PushRidingCardService) GetByServiceID(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigPushRidingCardCO, error) {
	key := rediskeys.PushRidingCardConfig(tenantID, serviceID)
	var co dto.ConfigPushRidingCardCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.PushRidingCardConfig
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.PushRidingCardToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	// Java ConfigPushRidingCardServiceImpl: empty -> insert(serviceId only) -> re-fetch.
	if co.ServiceId == nil {
		row := &model.PushRidingCardConfig{
			ID:        repo.NextConfigID(),
			ServiceID: serviceID,
			IzOpen:    sql.NullBool{Bool: false, Valid: true},
		}
		repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
		if err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
			return s.repo.Create(ctx, row)
		}); err != nil {
			return nil, err
		}
		return s.GetByServiceID(ctx, tenantID, serviceID)
	}
	return &co, nil
}

func (s *PushRidingCardService) Insert(ctx context.Context, tenantID, pin string, cmd dto.ConfigPushRidingCardCmd) (bool, error) {
	if cmd.ServiceId == nil {
		return false, errors.New("serviceId must not be null")
	}
	row := model.PushRidingCardConfig{ID: repo.NextConfigID(), ServiceID: *cmd.ServiceId}
	mergeJSON(&row, cmd)
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.PushRidingCardConfig(tenantID, *cmd.ServiceId)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
	return err == nil, err
}

func mergeJSON(dst, src interface{}) {
	b, _ := json.Marshal(src)
	_ = json.Unmarshal(b, dst)
}
