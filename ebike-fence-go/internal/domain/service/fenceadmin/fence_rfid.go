package fenceadmin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	pkgredis "ebike-fence-go/internal/pkg/redis"
)

type FenceRfidAdmin struct{}

func NewFenceRfidAdmin() *FenceRfidAdmin { return &FenceRfidAdmin{} }

func (f *FenceRfidAdmin) QueryRFIDsByFenceID(ctx context.Context, tenantID string, fenceID int64) ([]dto.FenceRfidCO, error) {
	if persistence.FenceRfid == nil {
		return nil, fmt.Errorf("mysql not initialized")
	}
	rows, err := persistence.FenceRfid.ListByFenceID(ctx, fenceID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceRfidCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.FenceRfidCO{Id: row.ID, FenceId: row.FenceID, RfidCode: row.RfidCode})
	}
	return out, nil
}

func (f *FenceRfidAdmin) QueryFenceIDByRfidCache(ctx context.Context, tenantID, rfidCode string) (*dto.FenceRfidCO, error) {
	rdb := pkgredis.GetClient()
	if rdb != nil {
		key := rediskeys.RfidFenceBind(tenantID, rfidCode)
		val, err := rdb.Get(ctx, key).Result()
		if err == nil && val != "" {
			id, _ := strconv.ParseInt(val, 10, 64)
			return &dto.FenceRfidCO{FenceId: id}, nil
		}
	}
	if persistence.FenceRfid == nil {
		return nil, nil
	}
	row, err := persistence.FenceRfid.GetByRfidCode(ctx, tenantID, rfidCode)
	if err != nil || row == nil {
		return nil, err
	}
	f.syncRfidCache(ctx, tenantID, rfidCode, row.FenceID)
	return &dto.FenceRfidCO{FenceId: row.FenceID}, nil
}

func (f *FenceRfidAdmin) QueryFenceByRfids(ctx context.Context, tenantID string, codes []string) ([]dto.FenceInfoByRfidsCO, error) {
	if persistence.FenceRfid == nil {
		return nil, fmt.Errorf("mysql not initialized")
	}
	rows, err := persistence.FenceRfid.ListByRfidCodes(ctx, tenantID, codes)
	if err != nil {
		return nil, err
	}
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceInfoByRfidsCO, 0, len(rows))
	for _, row := range rows {
		fe, _ := r.GetByID(ctx, row.FenceID)
		item := dto.FenceInfoByRfidsCO{FenceId: row.FenceID, RfidCode: row.RfidCode}
		if fe != nil {
			item.FenceInfo = map[string]interface{}{
				"id": fe.Id, "name": fe.Name, "type": fe.Type,
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (f *FenceRfidAdmin) SaveOrUpdate(ctx context.Context, tenantID, pin string, cmd dto.FenceRfidSaveCmd) (int, error) {
	if persistence.FenceRfid == nil {
		return 0, fmt.Errorf("mysql not initialized")
	}
	codes := cmd.RfidCodes
	if cmd.RfidCode != "" {
		codes = append(codes, cmd.RfidCode)
	}
	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, err := persistence.FenceRfid.SaveOrUpdate(ctx, tenantID, pin, cmd.FenceId, code); err != nil {
			return 0, err
		}
		f.delRfidCache(ctx, tenantID, code)
		f.syncRfidCache(ctx, tenantID, code, cmd.FenceId)
	}
	return len(codes), nil
}

func (f *FenceRfidAdmin) Change(ctx context.Context, tenantID, pin string, cmd dto.FenceRfidChangeCmd) (int, error) {
	if persistence.FenceRfid == nil {
		return 0, fmt.Errorf("mysql not initialized")
	}
	effectNum, err := persistence.FenceRfid.ChangeRfid(ctx, tenantID, pin, cmd.OldRfidCode, cmd.NewRfidCode)
	if err != nil {
		if errors.Is(err, repo.ErrRfidNewBindExist) {
			return 0, newBizError("13044", err.Error())
		}
		return 0, err
	}
	// Java deletes both old and new rfid station caches after change.
	f.delRfidCache(ctx, tenantID, cmd.OldRfidCode)
	f.delRfidCache(ctx, tenantID, cmd.NewRfidCode)
	return effectNum, nil
}

func (f *FenceRfidAdmin) BatchSaveOrUpdate(ctx context.Context, tenantID, pin string, cmd dto.FenceRfidSaveCmd) (int, error) {
	return f.SaveOrUpdate(ctx, tenantID, pin, cmd)
}

func (f *FenceRfidAdmin) syncRfidCache(ctx context.Context, tenantID, rfidCode string, fenceID int64) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	key := rediskeys.RfidFenceBind(tenantID, rfidCode)
	_ = rdb.Set(ctx, key, strconv.FormatInt(fenceID, 10), 48*time.Hour).Err()
}

func (f *FenceRfidAdmin) delRfidCache(ctx context.Context, tenantID, rfidCode string) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	_ = rdb.Del(ctx, rediskeys.RfidFenceBind(tenantID, rfidCode)).Err()
}
