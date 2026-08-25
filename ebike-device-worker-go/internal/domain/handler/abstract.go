package handler

import (
	"log"
	"strings"
	"time"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/domain/persistence"
	"ebike-device-worker-go/internal/pkg/redis"
)

var validBusinessTypes = map[string]struct{}{
	"ebike": {}, "bike": {}, "battery": {}, "cabinet": {},
}

type Deps struct {
	JDBC *persistence.JDBC
	Cfg  *config.Config
}

// ExportGetPersistMessages exposes parsing for tests.
func ExportGetPersistMessages(messages []message.DeviceMessageDTO) ([]map[string]interface{}, []message.MessageInfoDTO) {
	return getPersistMessages(messages)
}

func getPersistMessages(messages []message.DeviceMessageDTO) (jsonList []map[string]interface{}, structList []message.MessageInfoDTO) {
	nowMs := time.Now().UnixMilli()
	for _, msg := range messages {
		imei := msg.Imei
		bt := msg.BussinessType
		if imei == "" {
			log.Printf("[handler] missing imei: %+v", msg)
			continue
		}
		if _, ok := validBusinessTypes[strings.ToLower(bt)]; !ok {
			log.Printf("[handler] invalid bussinessType imei=%s type=%s", imei, bt)
			continue
		}
		originalData, err := message.ParseData(msg.Data)
		if err != nil {
			continue
		}
		cmdVal, ok := originalData["cmd"].(float64)
		if !ok {
			continue
		}
		cmd := int(cmdVal)

		// Redis uses enriched payload; PG metric mirrors Java message.jsonData() (original).
		redisData := make(map[string]interface{}, len(originalData)+3)
		for k, v := range originalData {
			redisData[k] = v
		}
		redisData["lastOnlineTime"] = nowMs
		redisData["isOnline"] = 1
		redisData["deviceType"] = bt

		receiveTime := time.Unix(time.Now().Unix(), 0)
		if msg.ReceiveDataTime != nil {
			receiveTime = msg.ReceiveDataTime.TruncateToSecond()
		}

		jsonList = append(jsonList, redisData)
		structList = append(structList, message.MessageInfoDTO{
			DeviceID: imei, Cmd: cmd, BussinessType: bt, MsgType: msg.MsgType,
			Data: originalData, ReceiveDataTime: receiveTime,
		})
	}
	return jsonList, structList
}

func redisKey(businessType, imei string) string {
	switch strings.ToLower(businessType) {
	case "ebike":
		return constants.RedisKeyDeviceEbike + imei
	case "battery":
		return constants.RedisKeyDeviceBattery + imei
	case "cabinet":
		return constants.RedisKeyDeviceCabinet + imei
	case "bike":
		return constants.RedisKeyDeviceBike + imei
	default:
		return ""
	}
}

func hSetPipelined(keys []string, values []map[string]interface{}) {
	if redis.Client == nil || len(keys) == 0 {
		return
	}
	if err := redis.HSetPipelined(keys, values); err != nil {
		log.Printf("[handler] redis hSetPipelined failed: %v", err)
	}
}

func writeData(deps *Deps, structs []message.MessageInfoDTO) {
	if deps == nil || deps.JDBC == nil || len(structs) == 0 {
		return
	}
	deps.JDBC.Persist(structs)
}
