package cache

import (
	"database/sql"
	"encoding/json"
	"time"

	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"
)

// ConfigBase mirrors Java BaseDO JSON in Redis (Jackson), separate from GORM model.
type ConfigBase struct {
	TenantID   string                `json:"tenantId"`
	CreatedPin string                `json:"createdPin"`
	CreatedAt  timefmt.JavaLocalTime `json:"createdAt"`
	UpdatedPin string                `json:"updatedPin"`
	UpdatedAt  timefmt.JavaLocalTime `json:"updatedAt"`
	Version    int                   `json:"version"`
	IzDel      flexBool              `json:"izDel"`
}

func (b ConfigBase) toModel() model.ConfigBaseDO {
	var izDel sql.NullBool
	if nb := b.IzDel.toNull(); nb.Valid {
		izDel = sql.NullBool{Bool: nb.Bool, Valid: true}
	}
	return model.ConfigBaseDO{
		TenantID:   b.TenantID,
		CreatedPin: b.CreatedPin,
		CreatedAt:  b.CreatedAt,
		UpdatedPin: b.UpdatedPin,
		UpdatedAt:  b.UpdatedAt,
		Version:    b.Version,
		IzDel:      izDel,
	}
}

func (b ConfigBase) toBaseConfig() model.BaseConfig {
	var version sql.NullInt32
	if b.Version != 0 {
		version = sql.NullInt32{Int32: int32(b.Version), Valid: true}
	}
	var izDel sql.NullBool
	if nb := b.IzDel.toNull(); nb.Valid {
		izDel = sql.NullBool{Bool: nb.Bool, Valid: true}
	}
	return model.BaseConfig{
		TenantID:   b.TenantID,
		CreatedPin: b.CreatedPin,
		CreatedAt:  b.CreatedAt.Time,
		UpdatedPin: b.UpdatedPin,
		UpdatedAt:  b.UpdatedAt.Time,
		Version:    version,
		IzDel:      izDel,
	}
}

type homeScrollMsg struct {
	ConfigBase
	ID          int64  `json:"id"`
	ServiceID   int64  `json:"serviceId"`
	Content     string `json:"content"`
	Type        int    `json:"type"`
	Appid       string `json:"appid"`
	SkipUrl     string `json:"skipUrl"`
	Params      string `json:"params"`
	Title       string `json:"title"`
	DetailTitle string `json:"detailTitle"`
	Detail      string `json:"detail"`
	IzOn        *bool  `json:"izOn"`
}

func (c homeScrollMsg) toModel() model.TConfigHomeScrollMsg {
	return model.TConfigHomeScrollMsg{
		ConfigBaseDO: c.ConfigBase.toModel(),
		ID:           c.ID,
		ServiceID:    c.ServiceID,
		Content:      c.Content,
		Type:         c.Type,
		Appid:        c.Appid,
		SkipUrl:      c.SkipUrl,
		Params:       c.Params,
		Title:        c.Title,
		DetailTitle:  c.DetailTitle,
		Detail:       c.Detail,
		IzOn:         c.IzOn,
	}
}

type faq struct {
	ConfigBase
	ID          int64  `json:"id"`
	ServiceID   int64  `json:"serviceId"`
	Title       string `json:"title"`
	DetailTitle string `json:"detailTitle"`
	Detail      string `json:"detail"`
	IzOn        *bool  `json:"izOn"`
}

func (c faq) toModel() model.TConfigFaq {
	return model.TConfigFaq{
		ConfigBaseDO: c.ConfigBase.toModel(),
		ID:           c.ID,
		ServiceID:    c.ServiceID,
		Title:        c.Title,
		DetailTitle:  c.DetailTitle,
		Detail:       c.Detail,
		IzOn:         c.IzOn,
	}
}

type guidePage struct {
	ConfigBase
	ID            int64  `json:"id"`
	ServiceID     int64  `json:"serviceId"`
	GuidePages    string `json:"guidePages"`
	AllowSuperEsc *bool  `json:"allowSuperEsc"`
	PageNumEsc    int    `json:"pageNumEsc"`
	VisibleRange  int    `json:"visibleRange"`
	Frequency     int    `json:"frequency"`
	IzOn          *bool  `json:"izOn"`
	OrderWeights  int    `json:"orderWeights"`
	ByRegister    *bool  `json:"byRegister"`
	ByTags        *bool  `json:"byTags"`
	TagIds        string `json:"tagIds"`
}

func (c guidePage) toModel() model.TConfigGuidePage {
	return model.TConfigGuidePage{
		ConfigBaseDO:  c.ConfigBase.toModel(),
		ID:            c.ID,
		ServiceID:     c.ServiceID,
		GuidePages:    c.GuidePages,
		AllowSuperEsc: c.AllowSuperEsc,
		PageNumEsc:    c.PageNumEsc,
		VisibleRange:  c.VisibleRange,
		Frequency:     c.Frequency,
		IzOn:          c.IzOn,
		OrderWeights:  c.OrderWeights,
		ByRegister:    c.ByRegister,
		ByTags:        c.ByTags,
		TagIds:        c.TagIds,
	}
}

type homeActivityEntrance struct {
	ConfigBase
	ID           int64                `json:"id"`
	ServiceID    int64                `json:"serviceId"`
	ChainType    int                  `json:"chainType"`
	LinkUrl      string               `json:"linkUrl"`
	PicUrl       string               `json:"picUrl"`
	LinkTitle    string               `json:"linkTitle"`
	AppId        string               `json:"appId"`
	Param        string               `json:"param"`
	IzOn         *bool                `json:"izOn"`
	Position     int                  `json:"position"`
	OpDownOffset int                  `json:"opDownOffset"`
	VisibleRange int                  `json:"visibleRange"`
	StartTime    timefmt.JavaLocalTime `json:"startTime"`
	EndTime      timefmt.JavaLocalTime `json:"endTime"`
	Unlimited    *bool                `json:"unlimited"`
	ByRegister   *bool                `json:"byRegister"`
	ByTags       *bool                `json:"byTags"`
	TagIds       string               `json:"tagIds"`
}

func (c homeActivityEntrance) toModel() model.TConfigHomeActivityEntrance {
	var start, end *time.Time
	if !c.StartTime.IsZero() {
		t := c.StartTime.Time
		start = &t
	}
	if !c.EndTime.IsZero() {
		t := c.EndTime.Time
		end = &t
	}
	return model.TConfigHomeActivityEntrance{
		ConfigBaseDO: c.ConfigBase.toModel(),
		ID:           c.ID,
		ServiceID:    c.ServiceID,
		ChainType:    c.ChainType,
		LinkUrl:      c.LinkUrl,
		PicUrl:       c.PicUrl,
		LinkTitle:    c.LinkTitle,
		AppId:        c.AppId,
		Param:        c.Param,
		IzOn:         c.IzOn,
		Position:     c.Position,
		OpDownOffset: c.OpDownOffset,
		VisibleRange: c.VisibleRange,
		StartTime:    start,
		EndTime:      end,
		Unlimited:    c.Unlimited,
		ByRegister:   c.ByRegister,
		ByTags:       c.ByTags,
		TagIds:       c.TagIds,
	}
}

type specialTips struct {
	ConfigBase
	ID               int64  `json:"id"`
	ServiceID        int64  `json:"serviceId"`
	PopUpType        int    `json:"popUpType"`
	BgUrl            string `json:"bgUrl"`
	BgColor          string `json:"bgColor"`
	Title            string `json:"title"`
	TitleColor       string `json:"titleColor"`
	IzSubtitle       *bool  `json:"izSubtitle"`
	Subtitle         string `json:"subtitle"`
	SubtitleColor    string `json:"subtitleColor"`
	Body             string `json:"body"`
	BodyColor        string `json:"bodyColor"`
	PopUpTime        int    `json:"popUpTime"`
	IzButton         *bool  `json:"izButton"`
	ButtonText       string `json:"buttonText"`
	ButtonColor      string `json:"buttonColor"`
	ButtonTextColor  string `json:"buttonTextColor"`
	ClickEvent       int    `json:"clickEvent"`
	JumpPage         string `json:"jumpPage"`
	IzCheckRead      *bool  `json:"izCheckRead"`
	CheckReadContent string `json:"checkReadContent"`
	VisibleRange     int    `json:"visibleRange"`
	Frequency        int    `json:"frequency"`
	ClosePosition    int    `json:"closePosition"`
	IzOn             *bool  `json:"izOn"`
	ByRegister       *bool  `json:"byRegister"`
	ByTags           *bool  `json:"byTags"`
	TagIds           string `json:"tagIds"`
}

func (c specialTips) toModel() model.TConfigSpecialTips {
	return model.TConfigSpecialTips{
		ConfigBaseDO:     c.ConfigBase.toModel(),
		ID:               c.ID,
		ServiceID:        c.ServiceID,
		PopUpType:        c.PopUpType,
		BgUrl:            c.BgUrl,
		BgColor:          c.BgColor,
		Title:            c.Title,
		TitleColor:       c.TitleColor,
		IzSubtitle:       c.IzSubtitle,
		Subtitle:         c.Subtitle,
		SubtitleColor:    c.SubtitleColor,
		Body:             c.Body,
		BodyColor:        c.BodyColor,
		PopUpTime:        c.PopUpTime,
		IzButton:         c.IzButton,
		ButtonText:       c.ButtonText,
		ButtonColor:      c.ButtonColor,
		ButtonTextColor:  c.ButtonTextColor,
		ClickEvent:       c.ClickEvent,
		JumpPage:         c.JumpPage,
		IzCheckRead:      c.IzCheckRead,
		CheckReadContent: c.CheckReadContent,
		VisibleRange:     c.VisibleRange,
		Frequency:        c.Frequency,
		ClosePosition:    c.ClosePosition,
		IzOn:             c.IzOn,
		ByRegister:       c.ByRegister,
		ByTags:           c.ByTags,
		TagIds:           c.TagIds,
	}
}

type customerService struct {
	ConfigBase
	ID                   int64  `json:"id"`
	ServiceID            int64  `json:"serviceId"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	Tel                  string `json:"tel"`
	IzOnlineEntrance     *bool  `json:"izOnlineEntrance"`
	IzArtificialEntrance *bool  `json:"izArtificialEntrance"`
	IzWorkTime           *bool  `json:"izWorkTime"`
	Tips                 string `json:"tips"`
}

func (c customerService) toModel() model.TConfigCustomerService {
	return model.TConfigCustomerService{
		ConfigBaseDO:         c.ConfigBase.toModel(),
		ID:                   c.ID,
		ServiceID:            c.ServiceID,
		StartTime:            c.StartTime,
		EndTime:              c.EndTime,
		Tel:                  c.Tel,
		IzOnlineEntrance:     c.IzOnlineEntrance,
		IzArtificialEntrance: c.IzArtificialEntrance,
		IzWorkTime:           c.IzWorkTime,
		Tips:                 c.Tips,
	}
}

type homeNav struct {
	ConfigBase
	ID           int64  `json:"id"`
	ServiceID    int64  `json:"serviceId"`
	CarType      int    `json:"carType"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	JumpPage     string `json:"jumpPage"`
	IzOn         *bool  `json:"izOn"`
	OrderWeights int    `json:"orderWeights"`
}

func (c homeNav) toModel() model.TConfigHomeNav {
	return model.TConfigHomeNav{
		ConfigBaseDO: c.ConfigBase.toModel(),
		ID:           c.ID,
		ServiceID:    c.ServiceID,
		CarType:      c.CarType,
		Name:         c.Name,
		Icon:         c.Icon,
		JumpPage:     c.JumpPage,
		IzOn:         c.IzOn,
		OrderWeights: c.OrderWeights,
	}
}

func UnmarshalHomeScrollMsgList(raw string) ([]model.TConfigHomeScrollMsg, error) {
	return unmarshalList(raw, func(c homeScrollMsg) model.TConfigHomeScrollMsg { return c.toModel() })
}

func UnmarshalFaqList(raw string) ([]model.TConfigFaq, error) {
	return unmarshalList(raw, func(c faq) model.TConfigFaq { return c.toModel() })
}

func UnmarshalGuidePage(raw string) (*model.TConfigGuidePage, error) {
	return unmarshalOne(raw, func(c guidePage) model.TConfigGuidePage { return c.toModel() })
}

func UnmarshalHomeActivityEntranceList(raw string) ([]model.TConfigHomeActivityEntrance, error) {
	return unmarshalList(raw, func(c homeActivityEntrance) model.TConfigHomeActivityEntrance { return c.toModel() })
}

func UnmarshalSpecialTipsList(raw string) ([]model.TConfigSpecialTips, error) {
	return unmarshalList(raw, func(c specialTips) model.TConfigSpecialTips { return c.toModel() })
}

func UnmarshalCustomerService(raw string) (*model.TConfigCustomerService, error) {
	return unmarshalOne(raw, func(c customerService) model.TConfigCustomerService { return c.toModel() })
}

func UnmarshalHomeNavList(raw string) ([]model.TConfigHomeNav, error) {
	return unmarshalList(raw, func(c homeNav) model.TConfigHomeNav { return c.toModel() })
}

func unmarshalList[C any, M any](raw string, convert func(C) M) ([]M, error) {
	var cached []C
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		return nil, err
	}
	out := make([]M, len(cached))
	for i, c := range cached {
		out[i] = convert(c)
	}
	return out, nil
}

func unmarshalOne[C any, M any](raw string, convert func(C) M) (*M, error) {
	var cached C
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		return nil, err
	}
	out := convert(cached)
	return &out, nil
}
