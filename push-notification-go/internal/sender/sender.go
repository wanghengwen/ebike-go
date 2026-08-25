package sender

import (
	"context"

	"push-notification-go/internal/model"
)

// SendHandler defines the interface for sending messages/voice via different suppliers.
type SendHandler interface {
	Send(data *SendData) *model.SendResult
}

// SendData holds all parameters needed to perform a send operation.
type SendData struct {
	Ctx                  context.Context
	Phone                string
	SignName             string
	TemplateContent      string
	SupplierTemplateCode string
	ConfigJson           string
	TenantId             string
	Params               map[string]string
	Supplier             *model.Supplier
}

// MessageHandlers maps supplier channel ID to the corresponding SMS handler.
var MessageHandlers map[int]SendHandler

// VoiceHandlers maps supplier channel ID to the corresponding voice handler.
var VoiceHandlers map[int]SendHandler

// Register initializes the handler maps and registers all known handlers.
func Register() {
	MessageHandlers = make(map[int]SendHandler)
	VoiceHandlers = make(map[int]SendHandler)

	MessageHandlers[1] = &AliyunSMSHandler{}
	MessageHandlers[2] = &ChuanglanSMSHandler{}
	VoiceHandlers[1] = &AliyunVoiceHandler{}
}
