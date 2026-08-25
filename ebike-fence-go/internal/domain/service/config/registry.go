package configsvc

import (
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
)

// Registry holds all config domain services.
type Registry struct {
	Backcar            *BackcarService
	Pay                *PayService
	BaseItem           *BaseItemService
	UseCar             *UseCarService
	ParkApply          *ParkApplyService
	PushRidingCard     *PushRidingCardService
	CreditScore        *CreditScoreService
	BigScreen          *BigScreenService
	AdConfig           *AdConfigService
	AlarmContact       *AlarmContactService
	RidingPermission   *RidingPermissionService
	RidingCarConfig    *RidingCarConfigService
	Protocol           *ProtocolService
	FenceTag           *FenceTagService
	ResourceManagement *ResourceManagementService
}

var Services *Registry

func Init(r *repo.ConfigRepository, gw *gateway.ConfigGateway) {
	if r == nil {
		return
	}
	Services = &Registry{
		Backcar:            NewBackcarService(r),
		Pay:                NewPayService(r),
		BaseItem:           NewBaseItemService(r),
		UseCar:             NewUseCarService(r),
		ParkApply:          NewParkApplyService(r),
		PushRidingCard:     NewPushRidingCardService(r),
		CreditScore:        NewCreditScoreService(r),
		BigScreen:          NewBigScreenService(r),
		AdConfig:           NewAdConfigService(r),
		AlarmContact:       NewAlarmContactService(r),
		RidingPermission:   NewRidingPermissionService(r),
		RidingCarConfig:    NewRidingCarConfigService(),
		Protocol:           NewProtocolService(r),
		FenceTag:           NewFenceTagService(r),
		ResourceManagement: NewResourceManagementService(r),
	}
	Services.RidingCarConfig.bind(Services.UseCar, Services.Pay)
	_ = gw
}
