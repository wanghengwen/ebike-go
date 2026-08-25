package service

import (
	"context"
	"time"

	"push-notification-go/internal/cache"
	"push-notification-go/internal/model"
	"push-notification-go/internal/pkg/errcode"
	"push-notification-go/internal/pkg/logger"
	"push-notification-go/internal/pkg/templateutil"
	"push-notification-go/internal/pool"
	"push-notification-go/internal/sender"

	"gorm.io/gorm"
)

// MessageService handles SMS sending with pool-based async execution.
type MessageService struct {
	db      *gorm.DB
	msgPool *pool.WorkerPool
}

// NewMessageService creates a new MessageService.
func NewMessageService(db *gorm.DB, msgPool *pool.WorkerPool) *MessageService {
	return &MessageService{
		db:      db,
		msgPool: msgPool,
	}
}

// SendMessage validates the request against cached config and submits the send
// task to the worker pool.
func (s *MessageService) SendMessage(ctx context.Context, req *model.SendMessageReq) error {
	// 1. Check tenant config
	tenantCfg := cache.TenantConfigCache.Get(req.TenantId, cache.TypeMessage)
	if tenantCfg == nil {
		return errcode.ErrTenantConfigNotExist
	}

	// 2. Check supplier
	supplier := cache.SupplierCache.Get(tenantCfg.SupplierId)
	if supplier == nil {
		return errcode.ErrSupplierNotExist
	}

	// 3. Check sign name
	if !cache.MessageSignCache.Contains(req.TenantId, req.SignName) {
		return errcode.ErrSignNameNotExist
	}

	// 4. Check template
	tmpl := cache.TemplateCache.Get(supplier.ID, req.TemplateCode)
	if tmpl == nil {
		return errcode.ErrTemplateNotExist
	}

	content := templateutil.RenderTemplate(tmpl.TemplateContent, req.Params)
	logger.Printf(ctx, "[MessageService] processing message tenantId=%s templateCode=%s phone=%s content=%s", req.TenantId, req.TemplateCode, req.Phone, content)

	// 5. Check handler
	handler, ok := sender.MessageHandlers[supplier.Channel]
	if !ok {
		logger.Printf(ctx, "[MessageService] channel handler not found, channel=%d", supplier.Channel)
		return errcode.ErrChannelNotExist
	}

	// 6. Submit task to pool
	asyncCtx := logger.ContextFrom(ctx)
	s.msgPool.Submit(func() {
		data := &sender.SendData{
			Ctx:                  asyncCtx,
			Phone:                req.Phone,
			SignName:             req.SignName,
			TemplateContent:      tmpl.TemplateContent,
			SupplierTemplateCode: tmpl.SupplierTemplateCode,
			ConfigJson:           supplier.ConfigJson,
			TenantId:             req.TenantId,
			Params:               req.Params,
			Supplier:             supplier,
		}

		result := handler.Send(data)

		status := 0
		if result.Success {
			status = 1
		}

		record := &model.SendRecord{
			TenantId:             req.TenantId,
			SupplierId:           supplier.ID,
			Type:                 supplier.Type,
			Phone:                req.Phone,
			SignName:             req.SignName,
			SupplierTemplateCode: tmpl.SupplierTemplateCode,
			MessageContent:       content,
			Status:               status,
			Remark:               result.Message + " id=" + result.ID,
			SendDatetime:         time.Now(),
		}

		if err := model.InsertSendRecord(s.db, record); err != nil {
			logger.Printf(asyncCtx, "[MessageService] failed to save send record phone=%s err=%v", req.Phone, err)
		}
	})

	return nil
}
