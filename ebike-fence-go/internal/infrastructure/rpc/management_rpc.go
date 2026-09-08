package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/config"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"
	"ebike-fence-go/internal/pkg/shadow"

	lru "github.com/hashicorp/golang-lru/v2"
)

type ManagementRPC struct {
}

func NewManagementRPC() *ManagementRPC {
	return &ManagementRPC{}
}

type TenantCo struct {
	TenantName  string `json:"tenantName"`
	CompanyName string `json:"companyName"`
}

func (r *ManagementRPC) QueryTenant(ctx context.Context, tenantID string) (*TenantCo, error) {
	traceID := middleware.GetTraceIdFromCtx(ctx)
	body := map[string]interface{}{
		"tenantId": tenantID,
		"commandContext": map[string]interface{}{
			"tenantId": tenantID,
			"traceId":  traceID,
			"source":   config.GlobalConfig.Server.Name,
		},
	}
	var res Result[*TenantCo]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/tenant/query", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management tenant query error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

type CarInfo struct {
	CarId     string `json:"carId"`
	Imei      string `json:"imei"`
	ServiceId int64  `json:"serviceId"`
	Helmet    string `json:"helmet"`
	Kickstand string `json:"kickstand"`
	Camera    string `json:"camera"`
	Rfid      string `json:"rfid"`
	Direction string `json:"direction"`
	Beacon    string `json:"beacon"`
	Model     string `json:"model"`
}

type RepairModelCo struct {
	Id   int64 `json:"id"`
	Type int   `json:"type"`
}

type carInfoCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	CarId          string              `json:"carId,omitempty"`
	Imei           string              `json:"imei,omitempty"`
}

func (r *ManagementRPC) getCarDetail(ctx context.Context, cmdCtx *dto.CommandContext, carId, imei string) (*CarInfo, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	body := carInfoCmd{CommandContext: cmdCtx, CarId: carId, Imei: imei}
	var res Result[*CarInfo]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/car-info/detail", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

type cachedCarInfo struct {
	CarInfo   *CarInfo
	ExpiresAt time.Time
}

var (
	carInfoCache *lru.Cache[string, cachedCarInfo]
)

func init() {
	var err error
	carInfoCache, err = lru.New[string, cachedCarInfo](20000)
	if err != nil {
		panic(err)
	}
}

func (r *ManagementRPC) GetCarInfoByImei(ctx context.Context, imei string, cmdCtx *dto.CommandContext) (*CarInfo, error) {
	tenantID := ""
	if cmdCtx != nil {
		tenantID = cmdCtx.TenantId
	}
	cacheKey := fmt.Sprintf("car:%s:imei:%s", tenantID, imei)

	if val, ok := carInfoCache.Get(cacheKey); ok {
		if time.Now().Before(val.ExpiresAt) {
			val.ExpiresAt = time.Now().Add(5 * time.Minute)
			carInfoCache.Add(cacheKey, val)
			return val.CarInfo, nil
		}
		carInfoCache.Remove(cacheKey)
	}

	carInfo, err := r.getCarDetail(ctx, cmdCtx, "", imei)
	if err != nil {
		return nil, err
	}
	if carInfo != nil {
		carInfoCache.Add(cacheKey, cachedCarInfo{
			CarInfo:   carInfo,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		})
	}
	return carInfo, nil
}

type OpsUser struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	RoleId *int64 `json:"roleId"`
}

// RoleCO mirrors Java ManagementApiRpc.getRootRole result.
type RoleCO struct {
	Id   *int64 `json:"id"`
	Name string `json:"name"`
}

// GetRootRole mirrors Java ManagementApiRpc.getRootRole (ebike-management /role/getRootRole).
func (r *ManagementRPC) GetRootRole(ctx context.Context, cmdCtx *dto.CommandContext) (*RoleCO, error) {
	body := map[string]interface{}{"commandContext": cmdCtx}
	var res Result[*RoleCO]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/role/getRootRole", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// BatchSaveServiceRole mirrors Java ManagementApiRpc.batchSave (ebike-management
// /serviceRole/batchSave). Java swallows downstream errors (best-effort), so callers
// should treat a returned error as non-fatal.
func (r *ManagementRPC) BatchSaveServiceRole(ctx context.Context, cmdCtx *dto.CommandContext, roleID int64, serviceIDs []int64) error {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"roleId":         roleID,
		"serviceIds":     serviceIDs,
	}
	var res Result[json.RawMessage]
	return pkg_rpc.PostToService(ctx, "ebike-management", "/serviceRole/batchSave", body, &res)
}

// MsgReceivedUser mirrors Java rpcdataobject.ReceivedUser.
type MsgReceivedUser struct {
	ReceivePin     string            `json:"receivePin"`
	Phone          string            `json:"phone"`
	TemplateParams map[string]string `json:"templateParams"`
	AppType        *int              `json:"appType,omitempty"`
}

// MsgSendCmd mirrors Java management api MsgSendCmd.
type MsgSendCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Type           int                 `json:"type"`
	RemindTypes    []int               `json:"remindTypes"`
	ReceivedUser   []MsgReceivedUser   `json:"receivedUser"`
}

// SendMsg mirrors Java ManagementApiRpc.sendMsg (ebike-management /msg-template/sendMsg).
// Java returns early when remindTypes is empty.
func (r *ManagementRPC) SendMsg(ctx context.Context, msg MsgSendCmd) error {
	if len(msg.RemindTypes) == 0 {
		return nil
	}
	var res Result[bool]
	return pkg_rpc.PostToService(ctx, "ebike-management", "/msg-template/sendMsg", msg, &res)
}

// GetServiceIdsByRole mirrors Java ManagementApiRpc.getServiceIdsByRole.
func (r *ManagementRPC) GetServiceIdsByRole(ctx context.Context, cmdCtx *dto.CommandContext, roleIDs []int64) ([]int64, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"ids":            roleIDs,
	}
	var res Result[struct {
		ServiceIds []int64 `json:"serviceIds"`
	}]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/serviceRole/getServiceIdsByRole", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data.ServiceIds, nil
}

// IsRoot mirrors Java ManagementApiRpc.isRoot.
func (r *ManagementRPC) IsRoot(ctx context.Context, cmdCtx *dto.CommandContext, pin string) (bool, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"pin":            pin,
	}
	var res Result[bool]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/user/isRoot", body, &res); err != nil {
		return false, err
	}
	if !res.OK() {
		return false, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// GetRootTenantId mirrors Java ManagementApiRpc.getRootTenantId.
func (r *ManagementRPC) GetRootTenantId(ctx context.Context, cmdCtx *dto.CommandContext) (string, error) {
	body := map[string]interface{}{"commandContext": cmdCtx}
	var res Result[string]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/tenant/getRootId", body, &res); err != nil {
		return "", err
	}
	if !res.OK() {
		return "", fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// IgnoreTenant mirrors Java ManagementApiRpc.ignoreTenant.
func (r *ManagementRPC) IgnoreTenant(ctx context.Context, cmdCtx *dto.CommandContext) ([]string, error) {
	body := map[string]interface{}{"commandContext": cmdCtx}
	var res Result[[]string]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/tenant/ignore", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// GetUserByPin mirrors Java ManagementApiRpc.getUserByPin.
func (r *ManagementRPC) GetUserByPin(ctx context.Context, pin string, cmdCtx *dto.CommandContext) (*OpsUser, error) {
	body := map[string]interface{}{
		"commandContext": cmdCtx,
		"pin":            pin,
	}
	var res Result[*OpsUser]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/user/getUserByPin", body, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

type carRepairsByModelQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Model          string              `json:"model"`
	Type           int                 `json:"type"`
}

func (r *ManagementRPC) GetCarReapirsByModel(ctx context.Context, cmdCtx *dto.CommandContext, model string, repairType int) (*RepairModelCo, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[*RepairModelCo]
	body := carRepairsByModelQry{CommandContext: cmdCtx, Model: model, Type: repairType}
	err := pkg_rpc.PostToService(ctx, "ebike-management", "/car/getCarReapirsByModel", body, &res)
	if err != nil {
		return nil, err
	}

	if !res.OK() {
		return nil, fmt.Errorf("management api returned error: code=%d msg=%s", res.Code, res.Msg)
	}

	return res.Data, nil
}

type carTagRecordCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	ServiceId      int64               `json:"serviceId"`
	CarId          string              `json:"carId"`
	TypeIds        []int               `json:"typeIds"`
}

// CarTagRecordAdd mirrors Java ManagementApiRpcImpl.carTagRecordAdd (async).
func (r *ManagementRPC) CarTagRecordAdd(ctx context.Context, cmdCtx *dto.CommandContext, serviceID int64, carID string, typeIDs []int) {
	if shadow.IsShadowTest(ctx) {
		return
	}
	if len(typeIDs) == 0 {
		return
	}
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	cmd := carTagRecordCmd{
		CommandContext: cmdCtx,
		ServiceId:      serviceID,
		CarId:          carID,
		TypeIds:        typeIDs,
	}
	var res Result[json.RawMessage]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/carTag/record/addCarIdTypeIds/Async", cmd, &res); err != nil {
		return
	}
}

// CarTagRecordRemove mirrors Java ManagementApiRpcImpl.carTagRecordRemove (async).
func (r *ManagementRPC) CarTagRecordRemove(ctx context.Context, cmdCtx *dto.CommandContext, serviceID int64, carID string, typeIDs []int) {
	if shadow.IsShadowTest(ctx) {
		return
	}
	if len(typeIDs) == 0 {
		return
	}
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	cmd := carTagRecordCmd{
		CommandContext: cmdCtx,
		ServiceId:      serviceID,
		CarId:          carID,
		TypeIds:        typeIDs,
	}
	var res Result[int]
	if err := pkg_rpc.PostToService(ctx, "ebike-management", "/carTag/record/removeCarIdTypeIds/Async", cmd, &res); err != nil {
		return
	}
}

// GetCarInfoByCarId mirrors Java ParkingGatewayImpl.getCar / carInfoApi.getCar with full commandContext.
func (r *ManagementRPC) GetCarInfoByCarId(ctx context.Context, cmdCtx *dto.CommandContext, carId string) (*CarInfo, error) {
	return r.getCarDetail(ctx, cmdCtx, carId, "")
}
