package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ebike-fence-go/internal/api/dto"
	customRedis "ebike-fence-go/internal/pkg/redis"
	pkg_rpc "ebike-fence-go/internal/pkg/rpc"

	"github.com/go-redis/redis/v8"
)

// Java DeviceApiRpcImpl uses Feign clients on Nacos service "ebike-device-paas"
// (EcuQueryApi, DeviceInfoApi) — not "ebike-device".
const devicePaasService = "ebike-device-paas"

// DeviceRPC handles HTTP calls to ebike-device-paas (mirrors Java DeviceApiRpcImpl).
type DeviceRPC struct{}

func NewDeviceRPC() *DeviceRPC {
	return &DeviceRPC{}
}

type RFIDEntity struct {
	Event  *int   `json:"event"`
	CardID string `json:"cardID"`
}

type CameraEntity struct {
	Event       *int `json:"event"`
	AngleDet    *int `json:"angleDet"`
	CameraAngle *int `json:"cameraAngle"`
}

type CameraCacheEntity struct {
	Event     *int  `json:"event"`
	AngleDet  *int  `json:"angleDet"`
	Timestamp int64 `json:"timestamp"`
}

type KickstandEntity struct {
	Event    *int `json:"event"`
	Type     *int `json:"type"`
	MagState *int `json:"magState"`
	RfState  *int `json:"rfState"`
}

type HeadingEntity struct {
	HeadingAngle *float64 `json:"headingAngle"`
}

type BluetoothBeaconEntity struct {
	Event       *int   `json:"event"`
	TBeaconAddr string `json:"tBeaconAddr"`
}

type DeviceInfoEntity struct {
	Lat            float64          `json:"lat"`
	Lng            float64          `json:"lng"`
	HelmetLock     *int             `json:"helmetLock"`
	Helmet6React   *int             `json:"helmet6React"`
	Helmet6Lock    *int             `json:"helmet6Lock"`
	BleHelmetState *int             `json:"bleHelmetState"`
	Rfid           *RFIDEntity      `json:"rfid"`
	Camera         *CameraEntity    `json:"camera"`
	KickStand      *KickstandEntity `json:"kickStand"`
	Heading        *HeadingEntity   `json:"heading"`
}

type DeviceDetailEntity struct {
	RfidCarId     string `json:"rfidCarId"`
	RfidTimestamp int64  `json:"rfidTimestamp"`
	HelmetReact   *int   `json:"helmetReact"`
}

type deviceInfoQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
	IsCameraEnable *int                `json:"isCameraEnable,omitempty"`
}

type blueTBeaconQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
}

type imeiQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
}

type deviceDetailQry struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
}

type deviceGpsCo struct {
	Lat *float64 `json:"lat"`
	Lng *float64 `json:"lng"`
}

// deviceInfoCo mirrors Java DeviceInfoCo (EcuQueryApi.getDeviceInfo payload).
type deviceInfoCo struct {
	HelmetLock     *int             `json:"helmetLock"`
	Helmet6React   *int             `json:"helmet6React"`
	Helmet6Lock    *int             `json:"helmet6Lock"`
	BleHelmetState *int             `json:"bleHelmetState"`
	Gps            *deviceGpsCo     `json:"gps"`
	Rfid           *RFIDEntity      `json:"RFID"`
	Camera         *CameraEntity    `json:"camera"`
	KickStand      *KickstandEntity `json:"kickStand"`
	Heading        *HeadingEntity   `json:"heading"`
}

type commandResultWrap[T any] struct {
	Result *T `json:"result"`
}

func deviceInfoFromCo(co *deviceInfoCo) *DeviceInfoEntity {
	if co == nil {
		return nil
	}
	out := &DeviceInfoEntity{
		HelmetLock:     co.HelmetLock,
		Helmet6React:   co.Helmet6React,
		Helmet6Lock:    co.Helmet6Lock,
		BleHelmetState: co.BleHelmetState,
		Rfid:           co.Rfid,
		Camera:         co.Camera,
		KickStand:      co.KickStand,
		Heading:        co.Heading,
	}
	if co.Gps != nil {
		if co.Gps.Lat != nil {
			out.Lat = *co.Gps.Lat
		}
		if co.Gps.Lng != nil {
			out.Lng = *co.Gps.Lng
		}
	}
	return out
}

// GetDeviceInfo mirrors Java DeviceApiRpcImpl.getDeviceInfo -> EcuQueryApi.getDeviceInfo.
func (r *DeviceRPC) GetDeviceInfo(ctx context.Context, cmdCtx *dto.CommandContext, imei string) (*DeviceInfoEntity, error) {
	if strings.TrimSpace(imei) == "" {
		return nil, &DeviceAPIError{Code: "00004", Msg: "imei 不能为空"}
	}
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[*commandResultWrap[deviceInfoCo]]
	body := deviceInfoQry{CommandContext: cmdCtx, Imei: imei}
	err := pkg_rpc.PostToDevicePaasService(ctx, devicePaasService, "/device/paas/deviceInfo", body, &res)
	if err != nil {
		if pkg_rpc.IsServiceDiscoveryError(err) {
			return nil, nil
		}
		return nil, err
	}
	if !res.OK() {
		return nil, newDeviceAPIError(int(res.Code), res.Msg)
	}
	if res.Data == nil {
		return nil, nil
	}
	return deviceInfoFromCo(res.Data.Result), nil
}

// GetDeviceDetail mirrors Java DeviceApiRpcImpl.getDeviceDetail -> DeviceInfoApi.getDeviceDetail.
func (r *DeviceRPC) GetDeviceDetail(ctx context.Context, cmdCtx *dto.CommandContext, imei string) (*DeviceDetailEntity, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[*DeviceDetailEntity]
	body := deviceDetailQry{CommandContext: cmdCtx, Imei: imei}
	err := pkg_rpc.PostToService(ctx, devicePaasService, "/device/paas/device/detail", body, &res)
	if err != nil {
		if pkg_rpc.IsServiceDiscoveryError(err) {
			return &DeviceDetailEntity{}, nil
		}
		return nil, err
	}
	if !res.OK() {
		return &DeviceDetailEntity{}, nil
	}
	if res.Data == nil {
		return &DeviceDetailEntity{}, nil
	}
	return res.Data, nil
}

// GetBlueTBeacon mirrors Java DeviceApiRpcImpl.getBlueTBeacon -> EcuQueryApi.getBlueTBeacon.
func (r *DeviceRPC) GetBlueTBeacon(ctx context.Context, cmdCtx *dto.CommandContext, tenantId, imei string) (*BluetoothBeaconEntity, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{TenantId: tenantId}
	} else if tenantId != "" && cmdCtx.TenantId == "" {
		cp := *cmdCtx
		cp.TenantId = tenantId
		cmdCtx = &cp
	}
	if rdb := customRedis.GetClient(); rdb != nil {
		key := fmt.Sprintf("bluetooth_beacon_info_%s_%s", tenantId, imei)
		val, err := rdb.Get(ctx, key).Result()
		if err == nil && val != "" && val != "null" {
			var beacon BluetoothBeaconEntity
			if json.Unmarshal([]byte(val), &beacon) == nil {
				return &beacon, nil
			}
		} else if err != nil && err != redis.Nil {
			return nil, err
		}
	}

	var res Result[*commandResultWrap[BluetoothBeaconEntity]]
	qry := blueTBeaconQry{CommandContext: cmdCtx, Imei: imei}
	if err := pkg_rpc.PostToService(ctx, devicePaasService, "/device/paas/blueTBeacon", qry, &res); err != nil {
		return &BluetoothBeaconEntity{}, nil
	}
	if !res.OK() || res.Data == nil || res.Data.Result == nil {
		return &BluetoothBeaconEntity{}, nil
	}
	beacon := *res.Data.Result
	// Java caches successful beacon reads for 1 minute.
	if rdb := customRedis.GetClient(); rdb != nil && tenantId != "" {
		key := fmt.Sprintf("bluetooth_beacon_info_%s_%s", tenantId, imei)
		if raw, err := json.Marshal(beacon); err == nil {
			_ = rdb.Set(ctx, key, raw, time.Minute).Err()
		}
	}
	return &beacon, nil
}

// QueryCameraState mirrors Java DeviceApiRpcImpl.queryCameraState -> DeviceInfoApi.queryCameraState.
func (r *DeviceRPC) QueryCameraState(ctx context.Context, cmdCtx *dto.CommandContext, imei string) (*CameraCacheEntity, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	var res Result[*CameraCacheEntity]
	body := imeiQry{CommandContext: cmdCtx, Imei: imei}
	if err := pkg_rpc.PostToService(ctx, devicePaasService, "/device/paas/device/queryCameraState", body, &res); err != nil {
		if pkg_rpc.IsServiceDiscoveryError(err) {
			return nil, nil
		}
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("queryCameraState error: code=%d msg=%s", res.Code, res.Msg)
	}
	return res.Data, nil
}

// CameraCacheValid returns true when cache event=1 and report within 15 seconds.
func CameraCacheValid(cache *CameraCacheEntity) bool {
	if cache == nil || cache.Event == nil || *cache.Event != 1 {
		return false
	}
	return time.Now().UnixMilli()-cache.Timestamp < 15*1000
}

type HelmetLockCmd struct {
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
	Imei           string              `json:"imei"`
	CarId          string              `json:"carId"`
	Sw             int                 `json:"sw"`
}

type HelmetCommandResult struct {
	JobId   string `json:"jobId"`
	Async   *bool  `json:"async"`
	EcuCode string `json:"ecuCode"`
	Payload string `json:"payload"`
}

// HelmetCommand mirrors Java HelmetGatewayImp unlock/lock via ecuCommandApi.helmetCommand.
func (r *DeviceRPC) HelmetCommand(ctx context.Context, cmdCtx *dto.CommandContext, imei, carId string, sw int) (*HelmetCommandResult, error) {
	if cmdCtx == nil {
		cmdCtx = &dto.CommandContext{}
	}
	cmd := HelmetLockCmd{CommandContext: cmdCtx, Imei: imei, CarId: carId, Sw: sw}
	var res Result[*HelmetCommandResult]
	if err := pkg_rpc.PostToService(ctx, devicePaasService, "/device/paas/helmetLock", cmd, &res); err != nil {
		return nil, err
	}
	if !res.OK() {
		return nil, fmt.Errorf("helmetCommand error: code=%d msg=%s", res.Code, res.Msg)
	}
	if res.Data == nil {
		return &HelmetCommandResult{}, nil
	}
	return res.Data, nil
}

// GetBlueTBeaconInfoFromCache reads tenant-scoped beacon cache from Redis.
func (r *DeviceRPC) GetBlueTBeaconInfoFromCache(ctx context.Context, tenantId, imei string) (string, error) {
	if rdb := customRedis.GetClient(); rdb != nil {
		key := fmt.Sprintf("bluetooth_beacon_info_%s_%s", tenantId, imei)
		val, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			return "", nil
		}
		return val, err
	}
	return "", nil
}
