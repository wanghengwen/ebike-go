package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/service"
	configsvc "ebike-fence-go/internal/domain/service/config"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

var (
	configInitOnce sync.Once
)

func initConfigServices() {
	configInitOnce.Do(func() {
		if persistence.Config != nil {
			configsvc.Init(persistence.Config, gateway.NewConfigGateway())
		}
	})
}

func configHandlers() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"/config/backcar/createConfig":             handleConfigBackcarCreate,
		"/config/backcar/getConfigById":            handleConfigBackcarGetByID,
		"/config/backcar/getConfigByServiceId":     handleConfigBackcarGetByServiceID,
		"/config/backcar/updateConfig":             handleConfigBackcarUpdate,
		"/config/base/getConfigByServiceId":        handleConfigBaseGet,
		"/config/base/insertSwapBatteryThreshold":  handleConfigBaseInsertSwapBattery,
		"/config/base/insertUserTicketPhotoWays":   handleConfigBaseInsertPhotoWays,
		"/config/base/updateConfig":                handleConfigBaseUpdate,
		"/config/pay/getConfigByServiceId":         handleConfigPayGet,
		"/config/pay/updateConfig":                 handleConfigPayUpdate,
		"/config/usecar/getConfigByServiceId":      handleConfigUseCarGet,
		"/config/usecar/updateConfig":              handleConfigUseCarUpdate,
		"/config/usecar/recharge":                  handleConfigUseCarRecharge,
		"/config/pushRidingCard/getByServiceId":    handleConfigPushRidingCardGet,
		"/config/pushRidingCard/update":            handleConfigPushRidingCardUpdate,
		"/config/parkApply/getByServiceId":         handleConfigParkApplyGet,
		"/config/parkApply/update":                 handleConfigParkApplyUpdate,
		"/config/creditScore/getConfigByServiceId": handleCreditScoreGet,
		"/config/creditScore/updateConfig":         handleCreditScoreUpdate,
		"/config/bigSceen/getConfigByServiceId":    handleBigScreenGet,
		"/config/bigSceen/updateConfig":            handleBigScreenUpdate,
		"/config/getAd":                            handleAdConfigGet,
		"/config/insAd":                            handleAdConfigIns,
		"/config/updAd":                            handleAdConfigUpd,
		"/ad_config/detail":                        handleAdConfigDetail,
		"/ad_config/saveOrUpdate":                  handleAdConfigSaveOrUpdate,
		"/alarmContact/addAlarmContact":            handleAlarmContactAdd,
		"/alarmContact/delAlarmContact":            handleAlarmContactDel,
		"/alarmContact/editAlarmContact":           handleAlarmContactEdit,
		"/alarmContact/pageAlarmContact":           handleAlarmContactPage,
		"/alarmContact/selectAlarmContact":         handleAlarmContactSelect,
		"/ridingPermission/get":                    handleRidingPermissionGet,
		"/ridingPermission/edit":                   handleRidingPermissionEdit,
		"/config/getRidingCarConfig":               handleRidingCarConfigGet,
		"/config/protocol/list":                    handleConfigProtocolList,
		"/config/protocol/update":                  handleConfigProtocolUpdate,
		"/config/protocol/setDefault":              handleConfigProtocolSetDefault,
		"/config/protocol/byType":                  handleConfigProtocolByType,
		"/config/protocol/default":                 handleConfigProtocolDefault,
		"/fence/tags/getAll":                       handleFenceTagsGetAll,
	}
}

func respondConfig(c *gin.Context, data interface{}, err error) {
	if err != nil {
		if biz, ok := err.(*service.BizError); ok {
			web.WriteBizError(c, biz.Code, biz.Msg)
			return
		}
		if exc, ok := err.(*service.ExceptionError); ok {
			web.WriteException(c, exc.Error())
			return
		}
		web.WriteException(c, err.Error())
		return
	}
	result := dto.NewSuccessResult(data)
	web.RespondResult(c, &result, nil)
}

func handleConfigBackcarGetByServiceID(c *gin.Context) {
	initConfigServices()
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := configsvc.Services.Backcar.GetByServiceID(c.Request.Context(), tenantID, *req.Id)
	respondConfig(c, res, err)
}

func handleConfigBackcarGetByID(c *gin.Context) {
	initConfigServices()
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := configsvc.Services.Backcar.GetByID(c.Request.Context(), tenantID, *req.Id)
	respondConfig(c, res, err)
}

func handleConfigBackcarCreate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBackcarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.Backcar.Create(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigBackcarUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBackcarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.Backcar.Update(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigPayGet(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigPayCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.Pay.Get(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleConfigPayUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigPayCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.Pay.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigBaseGet(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBaseItemCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.BaseItem.Get(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleConfigBaseUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBaseItemCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.BaseItem.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigBaseInsertSwapBattery(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBaseItemCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.BaseItem.InsertSwapBatteryThreshold(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigBaseInsertPhotoWays(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBaseItemCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.BaseItem.InsertUserTicketPhotoWays(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigUseCarGet(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigUseCarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.UseCar.Get(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleConfigUseCarUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigUseCarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.UseCar.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, 1, err)
}

func handleConfigUseCarRecharge(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigUseCarCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.UseCar.UpdateRecharge(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, 1, err)
}

func handleConfigPushRidingCardGet(c *gin.Context) {
	initConfigServices()
	var req dto.ServiceIdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "serviceId": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := configsvc.Services.PushRidingCard.GetByServiceID(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleConfigPushRidingCardUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigPushRidingCardCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.PushRidingCard.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleConfigParkApplyGet(c *gin.Context) {
	initConfigServices()
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := configsvc.Services.ParkApply.GetByServiceID(c.Request.Context(), tenantID, *req.Id)
	respondConfig(c, res, err)
}

func handleConfigParkApplyUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigParkApplyCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.ParkApply.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleCreditScoreGet(c *gin.Context) {
	initConfigServices()
	var req dto.CreditScoreConfigCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := configsvc.Services.CreditScore.Get(c.Request.Context(), tenantID)
	respondConfig(c, res, err)
}

func handleCreditScoreUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.CreditScoreConfigCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	okVal, err := configsvc.Services.CreditScore.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, okVal, err)
}

func handleBigScreenGet(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBigScreenCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := configsvc.Services.BigScreen.Get(c.Request.Context(), tenantID)
	respondConfig(c, res, err)
}

func handleBigScreenUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigBigScreenCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.BigScreen.Update(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleAdConfigGet(c *gin.Context) {
	initConfigServices()
	var req dto.AdConfigCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.AdConfig.Get(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleAdConfigIns(c *gin.Context) {
	initConfigServices()
	var req dto.AdConfigCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	co, _ := configsvc.Services.AdConfig.Get(c.Request.Context(), tenantID, derefInt64(req.ServiceId))
	if co == nil {
		err := configsvc.Services.AdConfig.Ins(c.Request.Context(), tenantID, repoPin(c), req)
		respondConfig(c, nil, err)
		return
	}
	respondConfig(c, nil, nil)
}

func handleAdConfigUpd(c *gin.Context) {
	initConfigServices()
	var req dto.AdConfigCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	n, err := configsvc.Services.AdConfig.Upd(c.Request.Context(), tenantID, repoPin(c), req)
	msg := "fail"
	if n > 0 {
		msg = "success"
	}
	respondConfig(c, msg, err)
}

func handleAdConfigDetail(c *gin.Context) {
	initConfigServices()
	var req dto.ServiceIdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "serviceId": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := configsvc.Services.AdConfig.Detail(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleAdConfigSaveOrUpdate(c *gin.Context) {
	initConfigServices()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	var envelope struct {
		dto.Command
		ServiceId *int64 `json:"serviceId"`
	}
	if !bindJSONCmd(c, &envelope, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	delete(raw, "commandContext")
	payload, _ := json.Marshal(raw)
	tenantID, ok := applyCmd(c, &envelope.Command)
	if !ok || !requireServiceID(c, envelope.ServiceId) {
		return
	}
	okVal, err := configsvc.Services.AdConfig.SaveOrUpdate(c.Request.Context(), tenantID, *envelope.ServiceId, payload)
	respondConfig(c, okVal, err)
}

func handleAlarmContactAdd(c *gin.Context) {
	initConfigServices()
	var req dto.AlarmContactCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.AlarmContact.Add(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleAlarmContactDel(c *gin.Context) {
	initConfigServices()
	var req dto.AlarmContactCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	err := configsvc.Services.AlarmContact.Del(c.Request.Context(), req)
	respondConfig(c, nil, err)
}

func handleAlarmContactEdit(c *gin.Context) {
	initConfigServices()
	var req dto.AlarmContactCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.AlarmContact.Edit(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, nil, err)
}

func handleAlarmContactPage(c *gin.Context) {
	initConfigServices()
	var req dto.ContactQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := configsvc.Services.AlarmContact.Page(c.Request.Context(), req)
	respondConfig(c, res, err)
}

func handleAlarmContactSelect(c *gin.Context) {
	initConfigServices()
	var req dto.AlarmQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := configsvc.Services.AlarmContact.Select(c.Request.Context(), req)
	respondConfig(c, res, err)
}

func handleRidingPermissionGet(c *gin.Context) {
	initConfigServices()
	var req dto.RidingPermissionCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.RidingPermission.Get(c.Request.Context(), tenantID, *req.ServiceId)
	respondConfig(c, res, err)
}

func handleRidingPermissionEdit(c *gin.Context) {
	initConfigServices()
	var req dto.RidingPermissionCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.RidingPermission.Insert(c.Request.Context(), tenantID, repoPin(c), req)
	respondConfig(c, 1, err)
}

func handleRidingCarConfigGet(c *gin.Context) {
	initConfigServices()
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := configsvc.Services.RidingCarConfig.Get(c.Request.Context(), tenantID, *req.Id)
	respondConfig(c, res, err)
}

func handleConfigProtocolList(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigProtocolCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireServiceID(c, req.ServiceId) {
		return
	}
	res, err := configsvc.Services.Protocol.List(c.Request.Context(), tenantID, repoPin(c), *req.ServiceId)
	respondConfig(c, res, err)
}

func handleConfigProtocolUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigProtocolUpdateCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.Protocol.Update(c.Request.Context(), tenantID, req)
	respondConfig(c, nil, err)
}

func handleConfigProtocolSetDefault(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigProtocolBatchUpdateCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := configsvc.Services.Protocol.SetDefault(c.Request.Context(), tenantID, req.List)
	respondConfig(c, nil, err)
}

func handleConfigProtocolByType(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigProtocolCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	if !requireServiceID(c, req.ServiceId) {
		return
	}
	if req.Type == nil {
		web.WriteParamError(c, "type must not be null")
		return
	}
	res, err := configsvc.Services.Protocol.ByType(c.Request.Context(), tenantID, *req.ServiceId, *req.Type)
	respondConfig(c, res, err)
}

func handleConfigProtocolDefault(c *gin.Context) {
	initConfigServices()
	var req dto.ConfigProtocolCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || req.Type == nil {
		return
	}
	res, err := configsvc.Services.Protocol.DefaultByType(c.Request.Context(), tenantID, *req.Type)
	respondConfig(c, res, err)
}

func handleFenceTagsGetAll(c *gin.Context) {
	initConfigServices()
	var req dto.Command
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req)
	if !ok {
		return
	}
	res, err := configsvc.Services.FenceTag.GetAll(c.Request.Context(), tenantID)
	if err != nil {
		respondConfig(c, nil, err)
		return
	}
	respondRaw(c, res, nil)
}

func repoPin(c *gin.Context) string {
	return repo.PinFromContext(c)
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
