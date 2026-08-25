package handler

import (
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/domain/push"
)

type DataHandler struct {
	deps *Deps
	push *push.Service
}

func NewDataHandler(deps *Deps, pushSvc *push.Service) *DataHandler {
	return &DataHandler{deps: deps, push: pushSvc}
}

func (h *DataHandler) HandleMessage(messages []message.DeviceMessageDTO) {
	h.push.PushData(messages, "data")
	h.saveData(messages)
}

func (h *DataHandler) saveData(messages []message.DeviceMessageDTO) {
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
