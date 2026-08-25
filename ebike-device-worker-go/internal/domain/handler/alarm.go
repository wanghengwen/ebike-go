package handler

import (
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/domain/push"
)

type AlarmHandler struct {
	deps *Deps
	push *push.Service
}

func NewAlarmHandler(deps *Deps, pushSvc *push.Service) *AlarmHandler {
	return &AlarmHandler{deps: deps, push: pushSvc}
}

func (h *AlarmHandler) HandleMessage(messages []message.DeviceMessageDTO) {
	h.push.PushData(messages, "alarm")
	h.saveData(messages)
}

func (h *AlarmHandler) saveData(messages []message.DeviceMessageDTO) {
	jsonList, structList := getPersistMessages(messages)
	if len(jsonList) > 0 {
		keys := make([]string, 0, len(jsonList))
		vals := make([]map[string]interface{}, 0, len(jsonList))
		for i, data := range jsonList {
			keys = append(keys, redisKey(structList[i].BussinessType, structList[i].DeviceID))
			vals = append(vals, data)
		}
		hSetPipelined(keys, vals)
	}
	writeData(h.deps, structList)
}
