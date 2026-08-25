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

// VoiceService handles voice notification sending with pool-based async execution.
type VoiceService struct {
	db        *gorm.DB
	voicePool *pool.WorkerPool
}

// NewVoiceService creates a new VoiceService.
func NewVoiceService(db *gorm.DB, voicePool *pool.WorkerPool) *VoiceService {
	return &VoiceService{
		db:        db,
		voicePool: voicePool,
	}
}

// SendVoice validates the request against cached config and submits the send
// task to the worker pool.
func (s *VoiceService) SendVoice(ctx context.Context, req *model.SendVoiceReq) error {
	// 1. Check tenant config (voice type)
	tenantCfg := cache.TenantConfigCache.Get(req.TenantId, cache.TypeVoice)
	if tenantCfg == nil {
		return errcode.ErrTenantConfigNotExist
	}

	// 2. Check supplier
	supplier := cache.SupplierCache.Get(tenantCfg.SupplierId)
	if supplier == nil {
		return errcode.ErrSupplierNotExist
	}

	// 3. Check template
	tmpl := cache.TemplateCache.Get(supplier.ID, req.TemplateCode)
	if tmpl == nil {
		return errcode.ErrTemplateNotExist
	}

	// 4. Check handler
	content := templateutil.RenderTemplate(tmpl.TemplateContent, req.Params)
	logger.Printf(ctx, "[VoiceService] processing voice tenantId=%s templateCode=%s phone=%s content=%s", req.TenantId, req.TemplateCode, req.Phone, content)
	handler, ok := sender.VoiceHandlers[supplier.Channel]
	if !ok {
		logger.Printf(ctx, "[VoiceService] channel handler not found, channel=%d", supplier.Channel)
		return errcode.ErrChannelNotExist
	}

	// 5. Submit task to pool
	asyncCtx := logger.ContextFrom(ctx)
	s.voicePool.Submit(func() {
		data := &sender.SendData{
			Ctx:                  asyncCtx,
			Phone:                req.Phone,
			SignName:             "",
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
			SignName:             "",
			SupplierTemplateCode: tmpl.SupplierTemplateCode,
			MessageContent:       content,
			Status:               status,
			Remark:               result.Message + " id=" + result.ID,
			SendDatetime:         time.Now(),
		}

		if err := model.InsertSendRecord(s.db, record); err != nil {
			logger.Printf(asyncCtx, "[VoiceService] failed to save send record phone=%s err=%v", req.Phone, err)
		}
	})

	return nil
}
