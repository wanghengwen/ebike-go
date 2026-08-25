package service

import (
	"errors"
	"testing"

	infrarpc "ebike-fence-go/internal/infrastructure/rpc"
)

func TestDeviceInfoBizErrorMapsAPI17012(t *testing.T) {
	err := deviceInfoBizError(&infrarpc.DeviceAPIError{
		Code: "17012",
		Msg:  "设备离线或操作失败: Post \"http://172.16.1.162:8080/ecu/wild\": context deadline exceeded",
	})
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatalf("expected BizError, got %T", err)
	}
	if biz.Code != "17012" {
		t.Fatalf("code=%q", biz.Code)
	}
	if biz.Msg == "" {
		t.Fatal("expected message")
	}
}

func TestDeviceInfoBizErrorMapsBlankImei(t *testing.T) {
	err := deviceInfoBizError(&infrarpc.DeviceAPIError{Code: "00004", Msg: "imei 不能为空"})
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatalf("expected BizError, got %T", err)
	}
	if biz.Code != "00004" || biz.Msg != "imei 不能为空" {
		t.Fatalf("got code=%q msg=%q", biz.Code, biz.Msg)
	}
}

func TestDeviceInfoBizErrorMapsTransportTimeout(t *testing.T) {
	err := deviceInfoBizError(errors.New(`RPC failed for http://172.16.1.9:8080/device/paas/deviceInfo: context deadline exceeded (Client.Timeout exceeded while awaiting headers)`))
	var biz *BizError
	if !errors.As(err, &biz) {
		t.Fatalf("expected BizError, got %T", err)
	}
	if biz.Code != codeDeviceTimeout {
		t.Fatalf("code=%q", biz.Code)
	}
}
