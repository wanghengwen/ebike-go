package service

import (
	"context"
	"encoding/json"
	"log"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/rpc"
)

// syncCarTagOnCustomFenceChange mirrors Java FenceCustomServiceImpl.createCarTag.
func syncCarTagOnCustomFenceChange(ctx context.Context, cmdCtx *dto.CommandContext, tenantID string, carID string, serviceID, oldFenceCustomID int64, newFence *gateway.FenceE) {
	var newID int64
	if newFence != nil {
		newID = newFence.Id
	}
	haveOld := oldFenceCustomID != 0
	if !haveOld && newFence == nil {
		return
	}
	if haveOld && newID == oldFenceCustomID {
		return
	}

	cmdCtx = dto.EnsureCommandContext(cmdCtx, tenantID)
	mgmt := rpc.NewManagementRPC()
	if haveOld {
		typeIDs := carTagTypeIDs(ctx, oldFenceCustomID)
		mgmt.CarTagRecordRemove(ctx, cmdCtx, serviceID, carID, typeIDs)
	}
	if newFence != nil {
		typeIDs := carTagTypeIDsByCustomType(ctx, newFence.CustomTypeId)
		mgmt.CarTagRecordAdd(ctx, cmdCtx, serviceID, carID, typeIDs)
	}
}

func carTagTypeIDs(ctx context.Context, fenceCustomID int64) []int {
	if persistence.Fence == nil {
		return nil
	}
	fe, err := persistence.Fence.GetByID(ctx, fenceCustomID)
	if err != nil || fe == nil {
		log.Printf("carTag: load fence custom %d: %v", fenceCustomID, err)
		return nil
	}
	return carTagTypeIDsByCustomType(ctx, fe.CustomTypeId)
}

func carTagTypeIDsByCustomType(ctx context.Context, customTypeID int64) []int {
	if customTypeID == 0 || persistence.FenceCustomType == nil {
		return nil
	}
	row, err := persistence.FenceCustomType.GetByID(ctx, customTypeID)
	if err != nil || row == nil {
		log.Printf("carTag: load custom type %d: %v", customTypeID, err)
		return nil
	}
	if !row.CarTag.Valid || row.CarTag.String == "" {
		return nil
	}
	var typeIDs []int
	if err := json.Unmarshal([]byte(row.CarTag.String), &typeIDs); err != nil {
		log.Printf("carTag: parse car_tag json for type %d: %v", customTypeID, err)
		return nil
	}
	return typeIDs
}
