package clientconfig

import (
	"ebike-service-client-go/internal/pkg/javacompat"
	"encoding/json"
)

// guidePageConfigCO mirrors fence GuidePageConfigCO (null fields emitted).
type guidePageConfigCO struct {
	Id            *javacompat.LongStr `json:"id"`
	ServiceId     *javacompat.LongStr `json:"serviceId"`
	GuidePages    json.RawMessage     `json:"guidePages"`
	AllowSuperEsc *bool               `json:"allowSuperEsc"`
	PageNumEsc    *int                `json:"pageNumEsc"`
	VisibleRange  *int                `json:"visibleRange"`
	Frequency     *int                `json:"frequency"`
	IzOn          *bool               `json:"izOn"`
	ByRegister    *bool               `json:"byRegister"`
	ByTags        *bool               `json:"byTags"`
	TagIds        *string             `json:"tagIds"`
}

func emptyGuidePageConfig() guidePageConfigCO {
	one := 1
	return guidePageConfigCO{PageNumEsc: &one}
}

// finalizeGuidePageConfig mirrors GuidePageConfigCO.getPageNumEsc().
func finalizeGuidePageConfig(g guidePageConfigCO) guidePageConfigCO {
	one := 1
	if g.GuidePages == nil || string(g.GuidePages) == "null" {
		g.PageNumEsc = &one
		return g
	}
	var pages []interface{}
	if json.Unmarshal(g.GuidePages, &pages) != nil || len(pages) == 0 {
		g.PageNumEsc = &one
		return g
	}
	if g.PageNumEsc == nil {
		g.PageNumEsc = &one
		return g
	}
	if *g.PageNumEsc > len(pages) {
		n := len(pages)
		g.PageNumEsc = &n
	}
	return g
}

func convertGuidePageConfigCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co guidePageConfigCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	co = finalizeGuidePageConfig(co)
	b, err := javacompat.MarshalJSONNoHTMLEscape(co)
	if err != nil {
		return raw
	}
	return b
}

// bSpecialTipsCO mirrors client BSpecialTipsCO (Long fields as strings).
type bSpecialTipsCO struct {
	Id               *javacompat.LongStr `json:"id"`
	ServiceId        *javacompat.LongStr `json:"serviceId"`
	PopUpType        *int                `json:"popUpType"`
	BgUrl            json.RawMessage     `json:"bgUrl"`
	BgColor          *string             `json:"bgColor"`
	Title            *string             `json:"title"`
	TitleColor       *string             `json:"titleColor"`
	IzSubtitle       *bool               `json:"izSubtitle"`
	Subtitle         *string             `json:"subtitle"`
	SubtitleColor    *string             `json:"subtitleColor"`
	Body             *string             `json:"body"`
	BodyColor        *string             `json:"bodyColor"`
	PopUpTime        *int                `json:"popUpTime"`
	IzButton         *bool               `json:"izButton"`
	ButtonText       *string             `json:"buttonText"`
	ButtonColor      *string             `json:"buttonColor"`
	ButtonTextColor  *string             `json:"buttonTextColor"`
	ClickEvent       *int                `json:"clickEvent"`
	JumpPage         json.RawMessage     `json:"jumpPage"`
	IzCheckRead      *bool               `json:"izCheckRead"`
	CheckReadContent *string             `json:"checkReadContent"`
	VisibleRange     *int                `json:"visibleRange"`
	Frequency        *int                `json:"frequency"`
	ClosePosition    *int                `json:"closePosition"`
	IzOn             *bool               `json:"izOn"`
}

// bHomeNavCO mirrors client BHomeNavCO (Long fields as strings via LongStr).
type bHomeNavCO struct {
	Id        *javacompat.LongStr `json:"id"`
	ServiceId *javacompat.LongStr `json:"serviceId"`
	CarType   *int                `json:"carType"`
	Name      *string             `json:"name"`
	Icon      *string             `json:"icon"`
	JumpPage  json.RawMessage     `json:"jumpPage"`
	IzOn      *bool               `json:"izOn"`
}

type mainPushRidingConfigCO struct {
	CardId         *javacompat.LongStr `json:"cardId"`
	RidingCardName *string             `json:"ridingCardName"`
	DeductionType  *int                `json:"deductionType"`
	CurCost        *int                `json:"curCost"`
	OriginCost     *int                `json:"originCost"`
	ExpiryDate     *int                `json:"expiryDate"`
	TotalTimes     *int                `json:"totalTimes"`
	FreeMoney      *int                `json:"freeMoney"`
	OpenStartTime  *javacompat.LongStr `json:"openStartTime"`
	OpenEndTime    *javacompat.LongStr `json:"openEndTime"`
	State          *int                `json:"state"`
	CreatedAt      *javacompat.LongStr `json:"createdAt"`
	DeductionRules *int                `json:"deductionRules"`
	DescriptionTag *string             `json:"descriptionTag"`
	PromotionTag   *string             `json:"promotionTag"`
	BackOfCardUrl  *string             `json:"backOfCardUrl"`
	IzMainPush     *bool               `json:"izMainPush"`
	DetailInfo     *string             `json:"detailInfo"`
}

// parseCOJSONField mirrors BSpecialTipsCO setBgUrl/setJumpPage: fence stores JSON
// as a string; client CO emits parsed array/object.
func parseCOJSONField(raw interface{}) json.RawMessage {
	if raw == nil {
		return nil
	}
	if s, ok := raw.(string); ok {
		if s == "" {
			return nil
		}
		var v interface{}
		if json.Unmarshal([]byte(s), &v) == nil {
			b, _ := json.Marshal(v)
			return b
		}
		return json.RawMessage(s)
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	return b
}

func convertSpecialTipsItem(raw json.RawMessage) bSpecialTipsCO {
	var row map[string]json.RawMessage
	if json.Unmarshal(raw, &row) != nil {
		return bSpecialTipsCO{}
	}
	b, _ := json.Marshal(row)
	var co bSpecialTipsCO
	if json.Unmarshal(b, &co) != nil {
		return bSpecialTipsCO{}
	}
	if v, ok := row["bgUrl"]; ok {
		co.BgUrl = parseCOJSONFieldRaw(v)
	}
	if v, ok := row["jumpPage"]; ok {
		co.JumpPage = parseCOJSONFieldRaw(v)
	}
	return co
}

func convertSpecialTipsList(data json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(data) {
		return data
	}
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return data
	}
	out := make([]bSpecialTipsCO, 0, len(items))
	for _, item := range items {
		out = append(out, convertSpecialTipsItem(item))
	}
	b, err := javacompat.MarshalJSONNoHTMLEscape(out)
	if err != nil {
		return data
	}
	return b
}

func convertMainPushList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil {
		var out []mainPushRidingConfigCO
		if json.Unmarshal(raw, &out) != nil {
			return raw
		}
		b, _ := json.Marshal(out)
		return b
	}
	out := make([]mainPushRidingConfigCO, 0, len(items))
	for _, item := range items {
		var co mainPushRidingConfigCO
		if json.Unmarshal(item, &co) == nil {
			out = append(out, co)
		}
	}
	b, _ := json.Marshal(out)
	return b
}

func convertHomeNavList(data json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(data) {
		return data
	}
	var rows []map[string]json.RawMessage
	if json.Unmarshal(data, &rows) != nil {
		return data
	}
	out := make([]bHomeNavCO, 0, len(rows))
	for _, row := range rows {
		b, _ := json.Marshal(row)
		var co bHomeNavCO
		if json.Unmarshal(b, &co) != nil {
			continue
		}
		if v, ok := row["jumpPage"]; ok {
			co.JumpPage = parseCOJSONFieldRaw(v)
		}
		out = append(out, co)
	}
	b, _ := json.Marshal(out)
	return b
}

func parseCOJSONFieldRaw(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return nil
	}
	var v interface{}
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return parseCOJSONField(v)
}

type faqCO struct {
	Id          *javacompat.LongStr  `json:"id"`
	ServiceId   *javacompat.LongStr  `json:"serviceId"`
	Title       *string              `json:"title"`
	DetailTitle *string              `json:"detailTitle"`
	Detail      *string              `json:"detail"`
	IzOn        *bool                `json:"izOn"`
	CreatedAt   *javacompat.DateTime `json:"createdAt"`
}

func convertFaqList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []faqCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertFaq(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co faqCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type homeScrollerMsgCO struct {
	Id          *javacompat.LongStr  `json:"id"`
	ServiceId   *javacompat.LongStr  `json:"serviceId"`
	Content     *string              `json:"content"`
	Type        *int                 `json:"type"`
	Appid       *string              `json:"appid"`
	SkipUrl     *string              `json:"skipUrl"`
	Params      *string              `json:"params"`
	Title       *string              `json:"title"`
	DetailTitle *string              `json:"detailTitle"`
	Detail      *string              `json:"detail"`
	IzOn        *bool                `json:"izOn"`
	CreatedAt   *javacompat.DateTime `json:"createdAt"`
	UpdatedPin  *string              `json:"updatedPin"`
	UpdatedAt   *javacompat.DateTime `json:"updatedAt"`
	UpdateName  *string              `json:"updateName"`
}

func convertHomeScrollerMsgList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []homeScrollerMsgCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertHomeScrollerMsg(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co homeScrollerMsgCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type customerServiceCO struct {
	ServiceId            *javacompat.LongStr  `json:"serviceId"`
	Tel                  *string              `json:"tel"`
	StartTime            *string              `json:"startTime"`
	EndTime              *string              `json:"endTime"`
	IzOnlineEntrance     *bool                `json:"izOnlineEntrance"`
	IzArtificialEntrance *bool                `json:"izArtificialEntrance"`
	IzWorkTime           *bool                `json:"izWorkTime"`
	Tips                 *string              `json:"tips"`
	UpdatedName          *string              `json:"updatedName"`
	UpdatedPin           *string              `json:"updatedPin"`
	UpdatedAt            *javacompat.DateTime `json:"updatedAt"`
}

func convertCustomerServiceCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co customerServiceCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type homeActivityEntranceCO struct {
	Id           *javacompat.LongStr  `json:"id"`
	ServiceId    *javacompat.LongStr  `json:"serviceId"`
	ChainType    *int                 `json:"chainType"`
	LinkUrl      *string              `json:"linkUrl"`
	PicUrl       *string              `json:"picUrl"`
	LinkTitle    *string              `json:"linkTitle"`
	AppId        *string              `json:"appId"`
	Param        *string              `json:"param"`
	IzOn         *bool                `json:"izOn"`
	Position     *int                 `json:"position"`
	OpDownOffset *int                 `json:"opDownOffset"`
	VisibleRange *int                 `json:"visibleRange"`
	StartTime    *javacompat.DateTime `json:"startTime"`
	EndTime      *javacompat.DateTime `json:"endTime"`
	Unlimited    *bool                `json:"unlimited"`
	ByRegister   *bool                `json:"byRegister"`
	ByTags       *bool                `json:"byTags"`
	TagIds       *string              `json:"tagIds"`
}

func convertHomeActivityEntranceList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []homeActivityEntranceCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertHomeActivityEntrance(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co homeActivityEntranceCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}
