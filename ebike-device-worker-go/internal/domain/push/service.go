package push

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/device"
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/infrastructure/kafka"
	"ebike-device-worker-go/internal/infrastructure/rpc"
	"ebike-device-worker-go/internal/pkg/cache"
	redispkg "ebike-device-worker-go/internal/pkg/redis"
)

type Service struct {
	cfg          *config.Config
	producer     *kafka.Producer
	paas         *rpc.PaasClient
	mappingCache *cache.TTLCache
	tenantCache  *cache.TTLCache
}

func NewService(cfg *config.Config, producer *kafka.Producer, paas *rpc.PaasClient) *Service {
	return &Service{
		cfg:          cfg,
		producer:     producer,
		paas:         paas,
		mappingCache: cache.NewTTLCache(cfg.Caffeine.DeviceTenantMappingCacheSize, cfg.Caffeine.TimeoutSec),
		tenantCache:  cache.NewTTLCache(cfg.Caffeine.TenantCacheSize, cfg.Caffeine.TimeoutSec),
	}
}

func (s *Service) PushData(messages []message.DeviceMessageDTO, handleMethod string) {
	for _, msg := range messages {
		expand, ok := s.getExpandData(msg)
		if !ok {
			continue
		}
		s.sendMessage(msg, expand, handleMethod)
	}
}

func (s *Service) HandleTenantChange(payload string) {
	if payload == "" {
		return
	}
	parts := strings.Split(payload, "_")
	if len(parts) < 3 {
		log.Printf("[push] invalid tenant change payload: %s", payload)
		return
	}
	imei := parts[0]
	tenantID := parts[2]
	log.Printf("[push] device tenant change imei=%s tenant=%s", imei, tenantID)
	s.mappingCache.Invalidate(imei)
	if tid, err := strconv.Atoi(tenantID); err == nil {
		s.mappingCache.Put(imei, tid)
	}
	if s.paas != nil {
		s.paas.Restart(imei, tenantID)
	}
}

func (s *Service) getExpandData(msg message.DeviceMessageDTO) (map[string]interface{}, bool) {
	bt := msg.BussinessType
	if bt == "" {
		return nil, false
	}
	deviceID := msg.Imei
	if deviceID == "" {
		return nil, false
	}
	data, err := message.ParseData(msg.Data)
	if err != nil {
		return nil, false
	}
	cmdVal, ok := data["cmd"].(float64)
	if !ok {
		return nil, false
	}
	deviceDataType := deviceDataTypeByCmd(int(cmdVal))
	if deviceDataType == "" {
		return nil, false
	}
	tenantID, ok := s.getDeviceTenantID(deviceID, bt)
	if !ok {
		return nil, false
	}
	s.mappingCache.Put(deviceID, tenantID)
	tenantMap, ok := s.getTenantMap(tenantID)
	if !ok || len(tenantMap) == 0 {
		return nil, false
	}
	s.tenantCache.Put(strconv.Itoa(tenantID), tenantMap)
	expand := map[string]interface{}{
		"appId":          tenantID,
		"appName":        tenantMap["tenantName"],
		"deviceDataType": deviceDataType,
	}
	return expand, true
}

func (s *Service) sendMessage(msg message.DeviceMessageDTO, expand map[string]interface{}, handleMethod string) {
	if s.producer == nil || s.cfg == nil || !s.cfg.Kafka.PushEnabled {
		return
	}
	data := s.buildSendData(msg, expand, handleMethod)
	topic := s.cfg.Kafka.Topics.ToSaasTopic
	if tenantID := intVal(data["appId"]); tenantID > 0 && s.cfg.IsGrayTenant(tenantID) {
		if s.cfg.Kafka.Topics.GrayToSaas != "" {
			topic = s.cfg.Kafka.Topics.GrayToSaas
		}
	}
	body, _ := json.Marshal(data)
	s.producer.Publish(topic, msg.Imei, string(body))
}

func (s *Service) buildSendData(msg message.DeviceMessageDTO, expand map[string]interface{}, handleMethod string) map[string]interface{} {
	delete(expand, "senderInfo")
	deviceDataType := expand["deviceDataType"]
	delete(expand, "deviceDataType")

	jsonData, err := message.ParseData(msg.Data)
	if err != nil {
		jsonData = make(map[string]interface{})
	}
	for k, v := range expand {
		jsonData[k] = v
	}
	if dt, ok := deviceDataType.(string); ok {
		jsonData["deviceDataType"] = dt
	}
	jsonData["imei"] = msg.Imei
	if handleMethod == "event" {
		jsonData["payload"] = rawToInterface(msg.Payload)
		jsonData["jobId"] = msg.MsgID
		delete(jsonData, "msgId")
	}
	device.SanitizeStringFields(jsonData)
	return jsonData
}

func (s *Service) getDeviceTenantID(deviceID, businessType string) (int, bool) {
	val, ok := s.mappingCache.Get(deviceID, func() (interface{}, bool) {
		key := deviceRedisKey(businessType, deviceID)
		if key == "" || redispkg.Client == nil {
			return 0, false
		}
		tidStr, err := redispkg.Client.HGet(redispkg.Ctx(), redispkg.GetKey(key), "tenantId").Result()
		if err != nil || tidStr == "" {
			return 0, false
		}
		tid, err := strconv.Atoi(tidStr)
		if err != nil {
			return 0, false
		}
		return tid, true
	})
	if !ok {
		return 0, false
	}
	tid, ok := val.(int)
	return tid, ok
}

func (s *Service) getTenantMap(tenantID int) (map[string]interface{}, bool) {
	key := strconv.Itoa(tenantID)
	val, ok := s.tenantCache.Get(key, func() (interface{}, bool) {
		if redispkg.Client == nil {
			return nil, false
		}
		m, err := redispkg.Client.HGetAll(redispkg.Ctx(), redispkg.GetKey(constants.RedisKeyHashTenant+strconv.Itoa(tenantID))).Result()
		if err != nil || len(m) == 0 {
			return nil, false
		}
		out := make(map[string]interface{}, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out, true
	})
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	return m, ok
}

func deviceRedisKey(businessType, deviceID string) string {
	switch strings.ToLower(businessType) {
	case "ebike":
		return constants.RedisKeyDeviceEbike + deviceID
	case "battery":
		return constants.RedisKeyDeviceBattery + deviceID
	case "bike":
		return constants.RedisKeyDeviceBike + deviceID
	default:
		return ""
	}
}

func deviceDataTypeByCmd(cmd int) string {
	switch cmd {
	case constants.CmdGPS1, constants.CmdGPSV6, constants.CmdGPSPack, constants.CmdBikeGPS:
		return "gps"
	case constants.CmdPing:
		return "ping"
	case constants.CmdShuaka:
		// 玉环项目：刷卡上报推送到 SaaS 时标记为 shuaka（对齐 Java DeviceDataTypeEnum.shuaka）。
		return "shuaka"
	case constants.CmdAlarm:
		return "alarm"
	case constants.CmdBMSInfo, constants.CmdBatteryBMS:
		return "bms"
	case constants.CmdLogin:
		return "login"
	case constants.CmdLogout:
		return "logout"
	case constants.CmdWild:
		return "cmd"
	case constants.CmdFault:
		return "fault"
	default:
		return ""
	}
}

func rawToInterface(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var v interface{}
	_ = json.Unmarshal(raw, &v)
	return v
}

func intVal(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return 0
	}
}
