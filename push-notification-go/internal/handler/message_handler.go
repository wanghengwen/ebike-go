package handler

import (
	"push-notification-go/internal/model"
	"push-notification-go/internal/pkg/config"
	"push-notification-go/internal/pkg/errcode"
	"push-notification-go/internal/pkg/logger"
	"push-notification-go/internal/pkg/response"
	"push-notification-go/internal/pkg/validator"
	"push-notification-go/internal/service"

	"github.com/gin-gonic/gin"
)

// MessageHandler handles SMS send requests.
type MessageHandler struct {
	messageService *service.MessageService
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// SendMessage handles POST /message/send.
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req model.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c, validator.FormatBindError(err, false))
		return
	}

	if !config.VerifySecret(req.Secret) {
		response.ErrorJSON(c, errcode.ErrSecretMistake)
		return
	}

	ctx := logger.WithTraceId(c.Request.Context(), req.TraceId)
	if err := h.messageService.SendMessage(ctx, &req); err != nil {
		if bizErr, ok := err.(*errcode.BizError); ok {
			response.ErrorJSON(c, bizErr)
			return
		}
		response.ErrorJSON(c, errcode.NewBizError("9999", err.Error()))
		return
	}

	response.SuccessJSON(c, "发送短信任务提交成功，异步发送中")
}
