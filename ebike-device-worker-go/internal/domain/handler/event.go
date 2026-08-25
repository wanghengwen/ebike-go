package handler

import (
	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/domain/push"
)

type EventHandler struct {
	deps *Deps
	push *push.Service
}

func NewEventHandler(deps *Deps, pushSvc *push.Service) *EventHandler {
	return &EventHandler{deps: deps, push: pushSvc}
}

func (h *EventHandler) HandleMessage(messages []message.DeviceMessageDTO) {
	h.push.PushData(messages, "event")
	h.saveData(messages)
}

func (h *EventHandler) saveData(messages []message.DeviceMessageDTO) {
	jsonList, structList := getPersistMessages(messages)
	keys, vals := collectEventRedisBatch(jsonList, structList)
	if len(keys) > 0 {
		hSetPipelined(keys, vals)
	}
	writeData(h.deps, structList)
}

// collectEventRedisBatch mirrors EventMessageHandler redis loop (logout breaks without hSet).
func collectEventRedisBatch(jsonList []map[string]interface{}, structList []message.MessageInfoDTO) (keys []string, vals []map[string]interface{}) {
	if len(jsonList) == 0 {
		return nil, nil
	}
	keys = make([]string, 0, len(jsonList))
	vals = make([]map[string]interface{}, 0, len(jsonList))
	for i, data := range jsonList {
		if cmd, ok := data["cmd"].(float64); ok && int(cmd) == constants.CmdLogout {
			data["isOnline"] = 0
			break
		}
		keys = append(keys, redisKey(structList[i].BussinessType, structList[i].DeviceID))
		vals = append(vals, data)
	}
	return keys, vals
}

// ExportCollectEventRedisBatch exposes parsed event redis batching for tests.
func ExportCollectEventRedisBatch(jsonList []map[string]interface{}, structList []message.MessageInfoDTO) (keys []string, vals []map[string]interface{}) {
	return collectEventRedisBatch(jsonList, structList)
}

// ExportEventRedisBatch exposes event redis batching for tests.
func ExportEventRedisBatch(messages []message.DeviceMessageDTO) (keys []string, vals []map[string]interface{}, structList []message.MessageInfoDTO) {
	jsonList, structList := getPersistMessages(messages)
	keys, vals = collectEventRedisBatch(jsonList, structList)
	return keys, vals, structList
}
