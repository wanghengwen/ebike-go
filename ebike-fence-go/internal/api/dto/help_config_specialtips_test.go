package dto

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

const specialTipsPCJSON = `{
    "serviceId": "364848372922716182",
    "popUpTime": 1,
    "jumpPage": {"chainType": "", "linkUrl": "", "linkTitle": ""},
    "bgUrl": ["https://luoping-upload.oss-cn-shanghai.aliyuncs.com/download/backend/1007/type/20260707/6ee40312-efbb-42ba-954b-a64ae09bed19.png"],
    "title": "案说法"
}`

func TestSpecialTipsCmdUnmarshalObjectJumpPageAndArrayBgUrl(t *testing.T) {
	var req SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(specialTipsPCJSON), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.ServiceId == nil || *req.ServiceId != 364848372922716182 {
		t.Fatalf("serviceId=%v", req.ServiceId)
	}
	if req.JumpPage == "" {
		t.Fatal("jumpPage should be JSON string")
	}
	if req.BgUrl == "" {
		t.Fatal("bgUrl should be JSON string")
	}
}
