package clientconfig

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

func fetchFenceData(c *gin.Context, path string, body interface{}, client *dto.ClientDTO) (json.RawMessage, bool) {
	cmdCtx := middleware.CompleteCommandContext(c, client)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, path, body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return nil, false
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return nil, false
	}
	return result.Data, true
}

func filterIzOnArray(data json.RawMessage) json.RawMessage {
	if len(data) == 0 || string(data) == "null" {
		return data
	}
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return data
	}
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var probe struct {
			IzOn *bool `json:"izOn"`
		}
		if json.Unmarshal(item, &probe) == nil && probe.IzOn != nil && *probe.IzOn {
			out = append(out, item)
		}
	}
	b, _ := json.Marshal(out)
	return b
}

func findRawItemByLongId(items json.RawMessage, id int64) json.RawMessage {
	var list []json.RawMessage
	if json.Unmarshal(items, &list) != nil {
		return nil
	}
	for _, item := range list {
		var probe struct {
			Id *javacompat.LongStr `json:"id"`
		}
		if json.Unmarshal(item, &probe) == nil && probe.Id != nil && int64(*probe.Id) == id {
			return item
		}
	}
	return nil
}

func getFaqByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getFaqByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertFaqList(filterIzOnArray(raw)),
	)))
}

func getFaqById(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getFaqByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	if req.Id == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(faqCO{}))
		return
	}
	found := findRawItemByLongId(filterIzOnArray(raw), *req.Id)
	if found == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(faqCO{}))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertFaq(found))))
}

func getHomeScrollerMsgByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeScrollerMsgByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertHomeScrollerMsgList(filterIzOnArray(raw)),
	)))
}

func getHomeScrollerMsgByServiceIdV2(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeScrollerMsgByServiceId/v2", &req, &req.ClientDTO)
	if !ok {
		return
	}
	if javacompat.IsNullJSON(raw) {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	var probe struct {
		IzOn *bool `json:"izOn"`
	}
	if json.Unmarshal(raw, &probe) != nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	if probe.IzOn == nil || !*probe.IzOn {
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertHomeScrollerMsg(raw),
	)))
}

func getHomeScrollerMsgById(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeScrollerMsgByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	if req.Id == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(homeScrollerMsgCO{}))
		return
	}
	found := findRawItemByLongId(filterIzOnArray(raw), *req.Id)
	if found == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(homeScrollerMsgCO{}))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertHomeScrollerMsg(found))))
}

func getCustomerServiceByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getCustomerServiceByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertCustomerServiceCO(raw),
	)))
}

func getHomeActivityById(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeActivityById", &req, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertHomeActivityEntrance(raw),
	)))
}

func getGuidePageConfigByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getGuidePageConfigByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	filtered := filterIzOnArray(raw)
	if req.UserPin != nil && strings.TrimSpace(*req.UserPin) != "" {
		var okFilter bool
		filtered, okFilter = filterActivityItemsStrict(c, filtered, *req.UserPin, &req.ClientDTO)
		if !okFilter {
			return
		}
	}
	var list []json.RawMessage
	if json.Unmarshal(filtered, &list) != nil || len(list) == 0 {
		c.JSON(http.StatusOK, dto.NewSuccessResult(emptyGuidePageConfig()))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertGuidePageConfigCO(list[0]),
	)))
}

func getSpecialTipsByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getSpecialTipsByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	pin := ""
	if req.UserPin != nil {
		pin = *req.UserPin
	}
	items, ok := filterSpecialTipsItems(c, filterIzOnArray(raw), pin, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertSpecialTipsList(items))))
}

func getHomeActivityEntranceByServiceId(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeActivityEntranceByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	items := filterTimeWindow(filterIzOnArray(raw))
	if req.UserPin != nil && strings.TrimSpace(*req.UserPin) != "" {
		var okFilter bool
		items, okFilter = filterActivityItemsStrict(c, items, *req.UserPin, &req.ClientDTO)
		if !okFilter {
			return
		}
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(
		convertHomeActivityEntranceList(items),
	)))
}

func filterTimeWindow(data json.RawMessage) json.RawMessage {
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return data
	}
	now := time.Now()
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var probe struct {
			Unlimited *bool   `json:"unlimited"`
			StartTime *string `json:"startTime"`
			EndTime   *string `json:"endTime"`
		}
		if json.Unmarshal(item, &probe) != nil {
			continue
		}
		if probe.Unlimited != nil && *probe.Unlimited {
			out = append(out, item)
			continue
		}
		if probe.StartTime == nil || probe.EndTime == nil {
			continue
		}
		start, err1 := time.Parse("2006-01-02 15:04:05", *probe.StartTime)
		end, err2 := time.Parse("2006-01-02 15:04:05", *probe.EndTime)
		if err1 == nil && err2 == nil && now.After(start) && now.Before(end) {
			out = append(out, item)
		}
	}
	b, _ := json.Marshal(out)
	return b
}

type activityItem struct {
	ByRegister    bool    `json:"byRegister"`
	ByTags        bool    `json:"byTags"`
	VisibleRange  *int    `json:"visibleRange"`
	TagIds        *string `json:"tagIds"`
}

func filterActivityItemsStrict(c *gin.Context, data json.RawMessage, pin string, client *dto.ClientDTO) (json.RawMessage, bool) {
	isNewUser, userTags, ok := fetchActivityUserContext(c, pin, client)
	if !ok {
		return nil, false
	}
	return filterByActivityContext(data, isNewUser, userTags), true
}

func filterSpecialTipsItems(c *gin.Context, data json.RawMessage, pin string, client *dto.ClientDTO) (json.RawMessage, bool) {
	isNewUser, userTags, ok := fetchActivityUserContext(c, pin, client)
	if !ok {
		return nil, false
	}
	return filterByActivityContext(data, isNewUser, userTags), true
}

func filterByActivityContext(data json.RawMessage, isNewUser bool, userTags string) json.RawMessage {
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return data
	}
	activeTags := ""
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var ai activityItem
		if json.Unmarshal(item, &ai) != nil {
			continue
		}
		if ai.TagIds != nil {
			activeTags = *ai.TagIds
		} else {
			activeTags = ""
		}
		if isShowActivity(ai.ByRegister, ai.ByTags, ai.VisibleRange, isNewUser, userTags, activeTags) {
			out = append(out, item)
		}
	}
	b, _ := json.Marshal(out)
	return b
}

// fetchActivityUserContext mirrors HelpConfigServiceImpl orderGateway.lastOrderDetail
// + userGateway.getUserTags (tags RPC failure yields empty string).
func fetchActivityUserContext(c *gin.Context, pin string, client *dto.ClientDTO) (isNewUser bool, userTags string, ok bool) {
	cmdCtx := middleware.CompleteCommandContext(c, client)
	q := map[string]string{"userPin": pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOrder, "/order/detailLast", q, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return false, "", false
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return false, "", false
	}
	isNewUser = javacompat.IsNullJSON(result.Data)
	tagResult, err2 := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/tag/user/me", map[string]string{"pin": pin}, cmdCtx)
	if err2 == nil && tagResult != nil && tagResult.Success && !javacompat.IsNullJSON(tagResult.Data) {
		var tag struct {
			Tags *string `json:"tags"`
		}
		if json.Unmarshal(tagResult.Data, &tag) == nil && tag.Tags != nil {
			userTags = *tag.Tags
		}
	}
	return isNewUser, userTags, true
}

func getHomeNavByServiceId(c *gin.Context) {
	var req homeNavDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeNav", &req, &req.ClientDTO)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertHomeNavList(filterHomeNav(raw)))))
}

func getHomeNavById(c *gin.Context) {
	var req homeNavDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	raw, ok := fetchFenceData(c, "/helpConfig/getHomeNav", &req, &req.ClientDTO)
	if !ok {
		return
	}
	filtered := filterHomeNav(raw)
	var items []bHomeNavCO
	_ = json.Unmarshal(convertHomeNavList(filtered), &items)
	var found *bHomeNavCO
	if req.Id != nil {
		for i := range items {
			if items[i].Id != nil && int64(*items[i].Id) == *req.Id {
				found = &items[i]
				break
			}
		}
	}
	if found == nil {
		c.JSON(http.StatusOK, dto.NewSuccessResult(bHomeNavCO{}))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(found))
}

func filterHomeNav(data json.RawMessage) json.RawMessage {
	type navFilterRow struct {
		IzOn      *bool               `json:"izOn"`
		UpdatedAt *javacompat.DateTime `json:"updatedAt"`
	}
	var rows []navFilterRow
	if json.Unmarshal(data, &rows) != nil || len(rows) == 0 {
		return data
	}
	sorted := append([]navFilterRow(nil), rows...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].UpdatedAt == nil {
			return false
		}
		if sorted[j].UpdatedAt == nil {
			return true
		}
		return sorted[i].UpdatedAt.Time.After(sorted[j].UpdatedAt.Time)
	})
	if sorted[0].IzOn != nil && *sorted[0].IzOn {
		return data
	}
	b, _ := json.Marshal([]interface{}{})
	return b
}

func getIzMainPush(c *gin.Context) {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return
	}
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceDTOBindMsgs) {
		return
	}
	raw, ok := fetchFenceData(c, "/config/pushRidingCard/getByServiceId", &req, &req.ClientDTO)
	if !ok {
		return
	}
	var cfg struct {
		IzOpen *bool `json:"izOpen"`
	}
	_ = json.Unmarshal(raw, &cfg)
	out := map[string]interface{}{"izOn": false, "list": nil}
	if cfg.IzOpen != nil {
		out["izOn"] = *cfg.IzOpen
		if *cfg.IzOpen {
			cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
			body := map[string]*int64{"serviceId": req.ServiceId}
			mResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/ridingConfig/getMainPushList", body, cmdCtx)
			if err == nil && mResult != nil && mResult.Success {
				out["list"] = json.RawMessage(convertMainPushList(mResult.Data))
			}
		}
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}
