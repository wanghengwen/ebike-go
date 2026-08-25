package service

import (
	"context"
	"sync"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	"ebike-fence-go/internal/infrastructure/rpc"
)

const codeCarNotFound = "13013"

// ComputeOutServiceDistance mirrors Java ServiceAreaServiceImpl.computeOutServiceDistance.
//
// 相较于 Java 原版的优化：
//  1. 入参校验前置：先校验 point 参数合法性，避免非法请求触发无用 RPC
//  2. 并行 I/O：GetServiceAreaById 与 GetUseCarConfig 之间无依赖关系，并行执行
//  3. 快速距离计算：使用 PointToPolygonDistanceFast 替代 PointToPolygonDistanceJava，
//     将 O(n×log(D)) 的二分查找 + JTS 几何库调用简化为 O(n) 的纯数学运算（详见 geoutils.go 注释）
func ComputeOutServiceDistance(ctx context.Context, tenantID string, cmd dto.ComputeDistanceCmd) (dto.ComputeDistanceCO, error) {
	// 1. 入参校验前置：避免非法 point 仍触发 RPC 调用
	loc, err := parseComputePoint(cmd)
	if err != nil {
		return dto.ComputeDistanceCO{}, err
	}

	// 2. 获取车辆信息（已有 LRU 缓存）
	mgmt := rpc.NewManagementRPC()
	car, err := mgmt.GetCarInfoByImei(ctx, cmd.Imei, dto.EnsureCommandContext(cmd.CommandContext, tenantID))
	if err != nil {
		return dto.ComputeDistanceCO{}, err
	}
	if car == nil || car.ServiceId == 0 {
		return dto.ComputeDistanceCO{}, newBizError(codeCarNotFound, "车辆未找到")
	}

	// 3. 并行获取服务区多边形 + 用车配置（两者仅依赖 serviceId，彼此无关）
	var serviceArea *gateway.FenceE
	var saErr error
	var nearLine int

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		serviceArea, saErr = GetServiceAreaById(ctx, tenantID, car.ServiceId)
	}()
	go func() {
		defer wg.Done()
		nearLine = 200
		cfgGw := gateway.NewConfigGateway()
		if useCarCfg, err := cfgGw.GetUseCarConfig(ctx, tenantID, car.ServiceId); err == nil && useCarCfg != nil && useCarCfg.NearLine != nil {
			nearLine = *useCarCfg.NearLine
		}
		if nearLine <= 0 {
			nearLine = 200
		}
	}()
	wg.Wait()

	if saErr != nil {
		return dto.ComputeDistanceCO{}, saErr
	}
	if serviceArea == nil {
		return dto.ComputeDistanceCO{}, newBizError("FENCE_SEARCH_FAIL", "fence not found")
	}

	// 4. 使用快速算法计算点到服务区边界的距离
	//    替代 PointToPolygonDistanceJava 的二分查找 + JTS Buffer + CoveredBy 拓扑检测。
	//    与 Java 版本最大偏差 < 1 米（Java 版整数截断 + 128 边形近似误差所致），
	//    对 nearLine（默认 200 米）的判断结果无影响。
	distance := geo.PointToPolygonDistanceFast(loc, serviceArea.ParsedPolygon)

	izCloseLine := distance <= float64(nearLine)
	return dto.ComputeDistanceCO{
		Distance:    distance,
		IzCloseLine: izCloseLine,
	}, nil
}

func parseComputePoint(cmd dto.ComputeDistanceCmd) (geo.Location, error) {
	if cmd.Point != nil {
		return geo.Location{Lng: cmd.Point.Lng, Lat: cmd.Point.Lat}, nil
	}
	if cmd.PointJSON != "" {
		loc, err := geo.ParsePointString(cmd.PointJSON)
		if err != nil {
			return geo.Location{}, newBizError("00004", "invalid point")
		}
		return loc, nil
	}
	return geo.Location{}, newBizError("00004", "point must not be null")
}
