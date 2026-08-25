package configsvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/infrastructure/persistence/repo"

	"gorm.io/gorm"
)

type BackcarService struct{ repo *repo.ConfigRepository }

func NewBackcarService(r *repo.ConfigRepository) *BackcarService { return &BackcarService{repo: r} }

func (s *BackcarService) GetByServiceID(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBackcarCO, error) {
	key := rediskeys.BackCarConfig(tenantID, serviceID)
	var co dto.ConfigBackcarCO
	err := s.repo.Store().GetObject(ctx, key, func(db *gorm.DB) (string, error) {
		var row model.ConfigBackcar
		if err := s.repo.GetLatestByServiceID(ctx, tenantID, &row, serviceID); err != nil {
			return "", err
		}
		if row.ID == 0 {
			return "", nil
		}
		return convert.MarshalRow(convert.BackcarToCO(&row))
	}, &co)
	if err != nil {
		return nil, err
	}
	if co.ServiceId == nil {
		if err := s.ensureDefault(ctx, tenantID, serviceID); err != nil {
			return nil, err
		}
		return s.GetByServiceID(ctx, tenantID, serviceID)
	}
	convert.ApplyBackcarDefaults(&co)
	return &co, nil
}

func (s *BackcarService) ensureDefault(ctx context.Context, tenantID string, serviceID int64) error {
	row := &model.ConfigBackcar{
		ID:             repo.NextConfigID(),
		ServiceID:      serviceID,
		BufferDistance: sql.NullFloat64{Float64: 10, Valid: true},
	}
	repo.StampBase(&row.BaseConfig, tenantID, "ebike_fence", true)
	key := rediskeys.BackCarConfig(tenantID, serviceID)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, row)
	})
	if err == nil {
		gateway.NewConfigGateway().InvalidateConfigByServiceId(tenantID, serviceID)
	}
	return err
}

func (s *BackcarService) GetByID(ctx context.Context, tenantID string, id int64) (*dto.ConfigBackcarCO, error) {
	var row model.ConfigBackcar
	if err := s.repo.GetByID(ctx, tenantID, &row, id); err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return convert.BackcarToCO(&row), nil
}

func (s *BackcarService) Create(ctx context.Context, tenantID, pin string, cmd dto.ConfigBackcarCmd) (bool, error) {
	if cmd.ServiceId == nil {
		return false, errors.New("serviceId must not be null")
	}
	row := backcarFromCmd(cmd)
	row.ID = repo.NextConfigID()
	repo.StampBase(&row.BaseConfig, tenantID, pin, true)
	key := rediskeys.BackCarConfig(tenantID, *cmd.ServiceId)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Create(ctx, &row)
	})
	if err == nil {
		gateway.NewConfigGateway().InvalidateConfigByServiceId(tenantID, *cmd.ServiceId)
	}
	return err == nil, err
}

func (s *BackcarService) Update(ctx context.Context, tenantID, pin string, cmd dto.ConfigBackcarCmd) (bool, error) {
	if cmd.Id == nil {
		return false, errors.New("id must not be null")
	}
	row := backcarFromCmd(cmd)
	row.ID = *cmd.Id
	repo.StampBase(&row.BaseConfig, tenantID, pin, false)
	key := rediskeys.BackCarConfig(tenantID, row.ServiceID)
	err := s.repo.Store().WriteWithInvalidate(ctx, key, func(db *gorm.DB) error {
		return s.repo.Save(ctx, &row)
	})
	if err == nil {
		gateway.NewConfigGateway().InvalidateConfigByServiceId(tenantID, row.ServiceID)
	}
	return err == nil, err
}

func backcarFromCmd(cmd dto.ConfigBackcarCmd) model.ConfigBackcar {
	b, _ := json.Marshal(cmd)
	var row model.ConfigBackcar
	_ = json.Unmarshal(b, &row)
	if cmd.ServiceId != nil {
		row.ServiceID = *cmd.ServiceId
	}
	return row
}
