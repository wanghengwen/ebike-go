package controller

import (
	"encoding/json"
	"sync"

	"ebike-fence-go/internal/api/dto"
	helpconfigsvc "ebike-fence-go/internal/domain/service/helpconfig"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var (
	helpConfigOnce   sync.Once
	globalHelpConfig *helpconfigsvc.Service
)

func helpConfigService() *helpconfigsvc.Service {
	helpConfigOnce.Do(func() {
		globalHelpConfig = helpconfigsvc.NewService()
	})
	return globalHelpConfig
}

func helpConfigHandlers() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"/helpConfig/getHomeScrollerMsgByServiceId":      hcListHomeScrollerMsg,
		"/helpConfig/getHomeScrollerMsgByServiceId/v2":   hcFirstHomeScrollerMsg,
		"/helpConfig/addHomeScrollerMsg":                 hcAddHomeScrollerMsg,
		"/helpConfig/editHomeScrollerMsg":                hcEditHomeScrollerMsg,
		"/helpConfig/delHomeScrollerMsg":                 hcDelHomeScrollerMsg,
		"/helpConfig/getHomeScrollerMsgById":             hcGetHomeScrollerMsgById,
		"/helpConfig/getFaqByServiceId":                  hcListFaq,
		"/helpConfig/addFaq":                             hcAddFaq,
		"/helpConfig/editFaq":                            hcEditFaq,
		"/helpConfig/delFaq":                             hcDelFaq,
		"/helpConfig/getFaqById":                         hcGetFaqById,
		"/helpConfig/getGuidePageConfigByServiceId":      hcListGuidePage,
		"/helpConfig/getGuidePageConfigIzOn":             hcGuidePageIzOn,
		"/helpConfig/addGuidePageConfig":                 hcAddGuidePage,
		"/helpConfig/editGuidePage":                      hcEditGuidePage,
		"/helpConfig/delGuidePage":                       hcDelGuidePage,
		"/helpConfig/sortGuidePage":                      hcSortGuidePage,
		"/helpConfig/getHomeActivityEntranceByServiceId": hcListHomeActivity,
		"/helpConfig/getHomeActivityById":                hcGetHomeActivityById,
		"/helpConfig/addHomeActivityEntrance":            hcAddHomeActivity,
		"/helpConfig/editHomeActivityEntrance":           hcEditHomeActivity,
		"/helpConfig/delHomeActivityEntrance":            hcDelHomeActivity,
		"/helpConfig/getSpecialTipsByServiceId":          hcListSpecialTips,
		"/helpConfig/getSpecialTipsById":                 hcGetSpecialTipsById,
		"/helpConfig/addDefinedSpecialTips":              hcAddSpecialTips,
		"/helpConfig/addSpecialTips":                     hcAddSpecialTips, // business-go forward alias
		"/helpConfig/editSpecialTips":                    hcEditSpecialTips,
		"/helpConfig/delSpecialTips":                     hcDelSpecialTips,
		"/helpConfig/onOffSpecialTips":                   hcOnOffSpecialTips,
		"/helpConfig/editCustomerService":                hcEditCustomerService,
		"/helpConfig/editCustomerServiceBatch":           hcEditCustomerServiceBatch,
		"/helpConfig/getCustomerServiceByServiceId":      hcGetCustomerService,
		"/helpConfig/getHomeNav":                         hcListHomeNav,
		"/helpConfig/addHomeNav":                         hcAddHomeNav,
		"/helpConfig/isOpen":                             hcIsOpenHomeNav,
		"/helpConfig/editHomeNav":                        hcEditHomeNav,
		"/helpConfig/delHomeNav":                         hcDelHomeNav,
	}
}

type helpCtx struct {
	tenantID string
	pin      string
}

func bindHelpCtx(c *gin.Context, cmd *dto.Command) (helpCtx, bool) {
	cmdCtx, ok := middleware.ApplyCommand(c, cmd)
	if !ok {
		return helpCtx{}, false
	}
	return helpCtx{tenantID: cmdCtx.TenantId, pin: cmdCtx.Pin}, true
}

var helpConfigJSON = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	extra.RegisterFuzzyDecoders()
}

func cmdToModel[T any](src interface{}) (*T, error) {
	b, err := helpConfigJSON.Marshal(src)
	if err != nil {
		return nil, err
	}
	var out T
	if err := helpConfigJSON.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// guidePageCmdToModel maps GuidePageConfigCmd onto TConfigGuidePage without a JSON
// round-trip. cmdToModel cannot populate guide_pages (DTO array vs DB string) and
// encoding/json rejects Jackson-style string numbers such as "serviceId":"123".
func guidePageCmdToModel(req dto.GuidePageConfigCmd) *model.TConfigGuidePage {
	row := &model.TConfigGuidePage{
		TagIds:        req.TagIds,
		AllowSuperEsc: req.AllowSuperEsc,
		IzOn:          req.IzOn,
		ByRegister:    req.ByRegister,
		ByTags:        req.ByTags,
		GuidePages:    marshalGuidePages(req.GuidePages),
	}
	if req.Id != nil {
		row.ID = *req.Id
	}
	if req.ServiceId != nil {
		row.ServiceID = *req.ServiceId
	}
	if req.PageNumEsc != nil {
		row.PageNumEsc = *req.PageNumEsc
	}
	if req.VisibleRange != nil {
		row.VisibleRange = *req.VisibleRange
	}
	if req.Frequency != nil {
		row.Frequency = *req.Frequency
	}
	return row
}

// guidePageCmdToModelForAdd maps a create request. Java business @Null(CreateGroup)
// forbids client id; FenceKeyGenerator assigns FastId on insert.
func guidePageCmdToModelForAdd(req dto.GuidePageConfigCmd) *model.TConfigGuidePage {
	row := guidePageCmdToModel(req)
	row.ID = 0
	if row.OrderWeights == 0 {
		row.OrderWeights = 9
	}
	return row
}

// specialTipsCmdToModel maps SpecialTipsCmd onto TConfigSpecialTips without a JSON
// round-trip so int fields such as frequency=0 are preserved reliably.
func specialTipsCmdToModel(req dto.SpecialTipsCmd) *model.TConfigSpecialTips {
	row := &model.TConfigSpecialTips{
		BgUrl:            req.BgUrl,
		BgColor:          req.BgColor,
		Title:            req.Title,
		TitleColor:       req.TitleColor,
		Subtitle:         req.Subtitle,
		SubtitleColor:    req.SubtitleColor,
		Body:             req.Body,
		BodyColor:        req.BodyColor,
		ButtonText:       req.ButtonText,
		ButtonColor:      req.ButtonColor,
		ButtonTextColor:  req.ButtonTextColor,
		JumpPage:         req.JumpPage,
		CheckReadContent: req.CheckReadContent,
		TagIds:           req.TagIds,
		IzSubtitle:       req.IzSubtitle,
		IzButton:         req.IzButton,
		IzCheckRead:      req.IzCheckRead,
		IzOn:             req.IzOn,
		ByRegister:       req.ByRegister,
		ByTags:           req.ByTags,
	}
	if req.Id != nil {
		row.ID = *req.Id
	}
	if req.ServiceId != nil {
		row.ServiceID = *req.ServiceId
	}
	if req.PopUpType != nil {
		row.PopUpType = *req.PopUpType
	}
	if req.PopUpTime != nil {
		row.PopUpTime = *req.PopUpTime
	}
	if req.ClickEvent != nil {
		row.ClickEvent = *req.ClickEvent
	}
	if req.VisibleRange != nil {
		row.VisibleRange = *req.VisibleRange
	}
	if req.Frequency != nil {
		row.Frequency = *req.Frequency
	}
	if req.ClosePosition != nil {
		row.ClosePosition = *req.ClosePosition
	}
	return row
}

// specialTipsCmdToModelForAdd maps a create request. Java @Null(CreateGroup) forbids
// client id; FastId is assigned on insert.
func specialTipsCmdToModelForAdd(req dto.SpecialTipsCmd) *model.TConfigSpecialTips {
	row := specialTipsCmdToModel(req)
	row.ID = 0
	return row
}

// marshalGuidePages serializes the guidePages array into the JSON string stored in
// the t_config_guide_page.guide_pages column (Java Convertor.toGuidePageConfigDO).
// cmdToModel cannot do this because the DTO field is an array while the column is a
// string, so the array would otherwise be silently dropped.
func marshalGuidePages(pages []dto.JumpPage) string {
	if len(pages) == 0 {
		return ""
	}
	b, err := json.Marshal(pages)
	if err != nil {
		return ""
	}
	return string(b)
}

func respondOK(c *gin.Context, data interface{}) {
	result := dto.NewSuccessResult(data)
	web.RespondResult(c, &result, nil)
}

func writeSvcErr(c *gin.Context, err error) {
	if biz, ok := err.(*helpconfigsvc.BizError); ok {
		web.WriteBizError(c, biz.Code, biz.Msg)
		return
	}
	web.WriteException(c, err.Error())
}

func svcID(p *int64) int64 {
	if p != nil {
		return *p
	}
	return 0
}

func requireHelpServiceID(c *gin.Context, serviceID *int64) bool {
	if serviceID != nil && *serviceID > 0 {
		return true
	}
	web.WriteParamError(c, "serviceId 不能为null")
	return false
}

func requireHelpID(c *gin.Context, id *int64) bool {
	if id != nil && *id > 0 {
		return true
	}
	// Java business @NotNull(DeleteGroup|UpdateGroup) on GuidePageConfigDTO.id
	web.WriteParamError(c, "id 不能为null")
	return false
}

// bindGuidePageJSON binds GuidePageConfigCmd for edit/delete paths.
// Probes raw JSON first so null/"" id is rejected like Java @NotNull before
// jsoniter fuzzy decoders coerce empty strings to numeric zero.
func bindGuidePageJSON(c *gin.Context, req *dto.GuidePageConfigCmd, requireID bool) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	if requireID && !web.RequireJSONFieldNotNull(c, "id", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, map[string]string{"commandContext": "must not be null"})
}

func bindSpecialTipsJSON(c *gin.Context, req *dto.SpecialTipsCmd, requireID bool) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return false
	}
	if requireID && !web.RequireJSONFieldNotNull(c, "id", "不能为null") {
		return false
	}
	return web.BindJSON(c, req, map[string]string{"commandContext": "must not be null"})
}

func hcListHomeScrollerMsg(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListHomeScrollerMsg(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeScrollerMsgCOList(res))
}

func hcFirstHomeScrollerMsg(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().FirstHomeScrollerMsg(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeScrollerMsgCO(*res))
}

func hcAddHomeScrollerMsg(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeScrollMsg](req)
	if err := helpConfigService().AddHomeScrollerMsg(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditHomeScrollerMsg(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeScrollMsg](req)
	if err := helpConfigService().EditHomeScrollerMsg(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelHomeScrollerMsg(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	id := *req.Id
	if err := helpConfigService().DelHomeScrollerMsg(c.Request.Context(), hctx.tenantID, *req.ServiceId, id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcGetHomeScrollerMsgById(c *gin.Context) {
	var req dto.HomeScrollerMsgCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if _, ok := bindHelpCtx(c, &req.Command); !ok {
		return
	}
	res, err := helpConfigService().GetHomeScrollerMsgByID(c.Request.Context(), svcID(req.Id))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeScrollerMsgCO(*res))
}

func hcListFaq(c *gin.Context) {
	var req dto.FaqCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListFaq(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		res = []model.TConfigFaq{}
	}
	respondOK(c, helpconfigsvc.ToFaqCOList(res))
}

func hcAddFaq(c *gin.Context) {
	var req dto.FaqCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigFaq](req)
	if err := helpConfigService().AddFaq(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditFaq(c *gin.Context) {
	var req dto.FaqCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigFaq](req)
	if err := helpConfigService().EditFaq(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelFaq(c *gin.Context) {
	var req dto.FaqCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	if err := helpConfigService().DelFaq(c.Request.Context(), hctx.tenantID, *req.ServiceId, *req.Id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcGetFaqById(c *gin.Context) {
	var req dto.FaqCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if _, ok := bindHelpCtx(c, &req.Command); !ok {
		return
	}
	res, err := helpConfigService().GetFaqByID(c.Request.Context(), svcID(req.Id))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToFaqCO(*res))
}

func hcListGuidePage(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListGuidePage(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, helpconfigsvc.ToGuidePageCOList(res))
}

func hcGuidePageIzOn(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	cmdCtx := middleware.GetCommandContext(c)
	res, err := helpConfigService().GetGuidePageIzOn(c.Request.Context(), hctx.tenantID, hctx.pin, cmdCtx, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToGuidePageCO(*res))
}

func hcAddGuidePage(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row := guidePageCmdToModelForAdd(req)
	if err := helpConfigService().AddGuidePage(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditGuidePage(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !bindGuidePageJSON(c, &req, true) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	row := guidePageCmdToModel(req)
	if err := helpConfigService().EditGuidePage(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelGuidePage(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !bindGuidePageJSON(c, &req, true) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	if err := helpConfigService().DelGuidePage(c.Request.Context(), hctx.tenantID, *req.ServiceId, *req.Id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcSortGuidePage(c *gin.Context) {
	var req dto.GuidePageConfigCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	if err := helpConfigService().SortGuidePage(c.Request.Context(), hctx.tenantID, hctx.pin, *req.ServiceId, req.Ids); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcListHomeActivity(c *gin.Context) {
	var req dto.HomeActivityEntranceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListHomeActivity(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeActivityCOList(res))
}

func hcGetHomeActivityById(c *gin.Context) {
	var req dto.HomeActivityEntranceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if _, ok := bindHelpCtx(c, &req.Command); !ok {
		return
	}
	res, err := helpConfigService().GetHomeActivityByID(c.Request.Context(), svcID(req.Id))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeActivityCO(*res))
}

func hcAddHomeActivity(c *gin.Context) {
	var req dto.HomeActivityEntranceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeActivityEntrance](req)
	if err := helpConfigService().AddHomeActivity(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditHomeActivity(c *gin.Context) {
	var req dto.HomeActivityEntranceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeActivityEntrance](req)
	if err := helpConfigService().EditHomeActivity(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelHomeActivity(c *gin.Context) {
	var req dto.HomeActivityEntranceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	if err := helpConfigService().DelHomeActivity(c.Request.Context(), hctx.tenantID, *req.ServiceId, *req.Id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcListSpecialTips(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListSpecialTips(c.Request.Context(), hctx.tenantID, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, helpconfigsvc.ToSpecialTipsCOList(res))
}

func hcGetSpecialTipsById(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if _, ok := bindHelpCtx(c, &req.Command); !ok {
		return
	}
	res, err := helpConfigService().GetSpecialTipsByID(c.Request.Context(), svcID(req.Id))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	if res == nil {
		respondOK(c, nil)
		return
	}
	respondOK(c, helpconfigsvc.ToSpecialTipsCO(*res))
}

func hcAddSpecialTips(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !bindSpecialTipsJSON(c, &req, false) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row := specialTipsCmdToModelForAdd(req)
	if err := helpConfigService().AddSpecialTips(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditSpecialTips(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !bindSpecialTipsJSON(c, &req, true) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	row := specialTipsCmdToModel(req)
	if err := helpConfigService().EditSpecialTips(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelSpecialTips(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !bindSpecialTipsJSON(c, &req, true) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	if err := helpConfigService().DelSpecialTips(c.Request.Context(), hctx.tenantID, *req.ServiceId, *req.Id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcOnOffSpecialTips(c *gin.Context) {
	var req dto.SpecialTipsCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigSpecialTips](req)
	if err := helpConfigService().OnOffSpecialTips(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditCustomerService(c *gin.Context) {
	var req dto.CustomerServiceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigCustomerService](req)
	if err := helpConfigService().InsertCustomerService(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditCustomerServiceBatch(c *gin.Context) {
	var req dto.CustomerServiceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigCustomerService](req)
	if err := helpConfigService().InsertCustomerServiceBatch(c.Request.Context(), hctx.tenantID, hctx.pin, row, req.CopyServiceId); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcGetCustomerService(c *gin.Context) {
	var req dto.CustomerServiceCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	cmdCtx := middleware.GetCommandContext(c)
	res, err := helpConfigService().GetCustomerService(c.Request.Context(), hctx.tenantID, hctx.pin, cmdCtx, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, res)
}

func hcListHomeNav(c *gin.Context) {
	var req dto.HomeNavCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok {
		return
	}
	res, err := helpConfigService().ListHomeNav(c.Request.Context(), hctx.tenantID, hctx.pin, svcID(req.ServiceId))
	if err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, helpconfigsvc.ToHomeNavCOList(res))
}

func hcAddHomeNav(c *gin.Context) {
	var req dto.HomeNavCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeNav](req)
	if err := helpConfigService().AddHomeNav(c.Request.Context(), hctx.tenantID, hctx.pin, row); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcIsOpenHomeNav(c *gin.Context) {
	var req dto.HomeNavCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	izOn := req.IzOn != nil && *req.IzOn
	if err := helpConfigService().IsOpenHomeNav(c.Request.Context(), hctx.tenantID, hctx.pin, *req.ServiceId, izOn); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcEditHomeNav(c *gin.Context) {
	var req dto.HomeNavCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) {
		return
	}
	row, _ := cmdToModel[model.TConfigHomeNav](req)
	if err := helpConfigService().EditHomeNav(c.Request.Context(), hctx.tenantID, hctx.pin, row, req.Ids); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}

func hcDelHomeNav(c *gin.Context) {
	var req dto.HomeNavCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	hctx, ok := bindHelpCtx(c, &req.Command)
	if !ok || !requireHelpServiceID(c, req.ServiceId) || !requireHelpID(c, req.Id) {
		return
	}
	if err := helpConfigService().DelHomeNav(c.Request.Context(), hctx.tenantID, *req.ServiceId, *req.Id); err != nil {
		writeSvcErr(c, err)
		return
	}
	respondOK(c, nil)
}
