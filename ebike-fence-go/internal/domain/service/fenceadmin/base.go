package fenceadmin

import (
	"context"
	"fmt"
	"strconv"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
)

func requireFenceRepo() (*repo.FenceRepository, error) {
	if persistence.Fence == nil {
		return nil, fmt.Errorf("mysql fence repository not initialized")
	}
	return persistence.Fence, nil
}

func dedupeName(ctx context.Context, r *repo.FenceRepository, tenantID, baseName string, fenceType int, serviceID *int64) (string, error) {
	exists, err := r.ExistsByName(ctx, tenantID, baseName, fenceType, serviceID)
	if err != nil {
		return baseName, err
	}
	if !exists {
		return baseName, nil
	}
	suffixRows, err := r.ListNamesBySuffixRange(ctx, tenantID, baseName, fenceType, serviceID, 2, 20)
	if err != nil {
		return baseName, err
	}
	return nextDedupedName(baseName, suffixRows), nil
}

// nextDedupedName mirrors Java NoParkingGatewayImpl/ParkingGatewayImpl/ServiceAreaGatewayImpl
// save: walk suffix rows in DB order and pick the first gap in baseName+2, baseName+3, …
func nextDedupedName(baseName string, suffixRows []string) string {
	suffix := 2
	for _, name := range suffixRows {
		if name != baseName+strconv.Itoa(suffix) {
			break
		}
		suffix++
	}
	return baseName + strconv.Itoa(suffix)
}

type writeOpts struct {
	kind         gateway.FenceCacheKind
	customTypeID int64
	serviceID    *int64
}

func saveFence(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) (int64, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return 0, err
	}
	name, err := dedupeName(ctx, r, tenantID, fe.Name, fe.Type, opt.serviceID)
	if err != nil {
		return 0, err
	}
	fe.Name = name
	return saveFenceNoDedupe(ctx, tenantID, pin, fe, opt)
}

// saveFenceRejectDuplicate mirrors Java MaintainAreaGatewayImpl.save: reject duplicate
// names with 13003 instead of auto-suffixing.
func saveFenceRejectDuplicate(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) (int64, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return 0, err
	}
	exists, err := r.ExistsByName(ctx, tenantID, fe.Name, fe.Type, opt.serviceID)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, newBizError("13003", "新增失败")
	}
	return saveFenceNoDedupe(ctx, tenantID, pin, fe, opt)
}

// saveFenceNoDedupe mirrors Java FenceGatewayImpl.save (plain insert + cache sync,
// no name de-duplication). Used by parking copy which preserves the source name.
func saveFenceNoDedupe(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) (int64, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return 0, err
	}
	id, err := r.Save(ctx, tenantID, pin, fe)
	if err != nil {
		return 0, err
	}
	saved, err := r.GetByID(ctx, id)
	if err != nil {
		return id, err
	}
	if saved != nil {
		_ = gateway.SyncFenceCache(ctx, opt.kind, tenantID, saved, opt.customTypeID)
	}
	return id, nil
}

func updateFence(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) error {
	return updateFenceCore(ctx, tenantID, pin, fe, opt, true, false)
}

// updateFenceNoDedupe mirrors Java FenceCustomGatewayImpl.saveOrUpdate on update (no rename dedupe).
func updateFenceNoDedupe(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) error {
	return updateFenceCore(ctx, tenantID, pin, fe, opt, false, false)
}

// updateFenceRejectDuplicate mirrors Java MaintainAreaGatewayImpl.update name conflict check.
func updateFenceRejectDuplicate(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts) error {
	return updateFenceCore(ctx, tenantID, pin, fe, opt, false, true)
}

func updateFenceCore(ctx context.Context, tenantID, pin string, fe *gateway.FenceE, opt writeOpts, dedupeOnRename, rejectDuplicateOnRename bool) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	old, err := r.GetByID(ctx, fe.Id)
	if err != nil {
		return err
	}
	if old != nil && old.Name != fe.Name {
		if rejectDuplicateOnRename {
			exists, err := r.ExistsByName(ctx, tenantID, fe.Name, fe.Type, opt.serviceID)
			if err != nil {
				return err
			}
			if exists {
				return newBizError("13003", "新增失败")
			}
		} else if dedupeOnRename {
			name, err := dedupeName(ctx, r, tenantID, fe.Name, fe.Type, opt.serviceID)
			if err != nil {
				return err
			}
			fe.Name = name
		}
	}
	_ = gateway.InvalidateFenceCache(ctx, opt.kind, tenantID, fe.Id)
	n, err := r.Update(ctx, tenantID, pin, fe)
	if err != nil {
		return err
	}
	if n <= 0 {
		return newBizError("13004", "修改失败")
	}
	updated, err := r.GetByID(ctx, fe.Id)
	if err != nil {
		return err
	}
	if updated != nil {
		_ = gateway.SyncFenceCache(ctx, opt.kind, tenantID, updated, opt.customTypeID)
	}
	return nil
}

func deleteFence(ctx context.Context, tenantID string, id int64, opt writeOpts) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	n, err := r.Delete(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if n <= 0 {
		return newBizError("13005", "删除失败")
	}
	return gateway.DeleteFenceCache(ctx, opt.kind, tenantID, id, opt.customTypeID)
}

// removeParkingBinding mirrors Java parkingDetailGateway.removeParkingId(id, type),
// called after a fence delete to clear the binding column in t_parking.
// NOTE: for noParking/banRiding, Java clears the binding only on single deletes; their
// batch delete paths (NoParkingGatewayImpl/BanRidingGatewayImpl.deleteByIds) do NOT.
// Parking is the exception: ParkingGatewayImpl.deleteByIds clears the binding per id on
// batch delete too, so ParkingAdmin.DeleteBatch calls this while the other batch deletes do not.
func removeParkingBinding(ctx context.Context, tenantID string, id int64, fenceType int) error {
	if persistence.Parking == nil {
		return nil
	}
	return persistence.Parking.RemoveParkingID(ctx, tenantID, id, fenceType)
}

func deleteFences(ctx context.Context, tenantID string, ids []int64, opt writeOpts) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	if err := r.DeleteByIDs(ctx, ids); err != nil {
		return err
	}
	for _, id := range ids {
		_ = gateway.DeleteFenceCache(ctx, opt.kind, tenantID, id, opt.customTypeID)
	}
	return nil
}

func getFenceOrErr(ctx context.Context, id int64) (*gateway.FenceE, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	fe, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if fe == nil {
		return nil, newBizError("13002", "数据查询失败，未查到数据")
	}
	return fe, nil
}

func newBizError(code, msg string) error {
	return &bizError{code: code, msg: msg}
}

type bizError struct {
	code string
	msg  string
}

func (e *bizError) Error() string { return e.msg }
func (e *bizError) Code() string  { return e.code }
