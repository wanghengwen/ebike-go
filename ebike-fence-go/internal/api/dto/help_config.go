package dto

import (
	"encoding/json"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func init() {
	extra.RegisterFuzzyDecoders()
}

var helpConfigDTOJSON = jsoniter.ConfigCompatibleWithStandardLibrary

// JumpPage matches Java helpConfig.JumpPage.
type JumpPage struct {
	ChainType int    `json:"chainType"`
	LinkUrl   string `json:"linkUrl"`
	PicUrl    string `json:"picUrl"`
	LinkTitle string `json:"linkTitle"`
	AppId     string `json:"appId"`
	Param     string `json:"param"`
}

// HomeScrollerMsgCmd matches Java HomeScrollerMsgCmd.
type HomeScrollerMsgCmd struct {
	Command
	Id          *int64 `json:"id,omitempty"`
	ServiceId   *int64 `json:"serviceId,omitempty"`
	Title       string `json:"title,omitempty"`
	Content     string `json:"content,omitempty"`
	Type        *int   `json:"type,omitempty"`
	Appid       string `json:"appid,omitempty"`
	SkipUrl     string `json:"skipUrl,omitempty"`
	Params      string `json:"params,omitempty"`
	DetailTitle string `json:"detailTitle,omitempty"`
	Detail      string `json:"detail,omitempty"`
	IzOn        *bool  `json:"izOn,omitempty"`
}

// HomeScrollerMsgCO matches Java HomeScrollerMsgCO.
type HomeScrollerMsgCO struct {
	Id          int64  `json:"id"`
	ServiceId   int64  `json:"serviceId"`
	Content     string `json:"content"`
	Type        int    `json:"type"`
	Appid       string `json:"appid"`
	SkipUrl     string `json:"skipUrl"`
	Params      string `json:"params"`
	Title       string `json:"title"`
	DetailTitle string `json:"detailTitle"`
	Detail      string `json:"detail"`
	IzOn        *bool  `json:"izOn"`
	CreatedAt   string `json:"createdAt"`
	UpdatedPin  string `json:"updatedPin"`
	UpdatedAt   string `json:"updatedAt"`
	// updateName is not populated from DO by Java ConvertorHelper → always JSON null.
	UpdateName *string `json:"updateName"`
}

// FaqCmd matches Java FaqCmd.
type FaqCmd struct {
	Command
	Id          *int64 `json:"id,omitempty"`
	ServiceId   *int64 `json:"serviceId,omitempty"`
	Title       string `json:"title,omitempty"`
	DetailTitle string `json:"detailTitle,omitempty"`
	Detail      string `json:"detail,omitempty"`
	IzOn        *bool  `json:"izOn,omitempty"`
}

// FaqCO matches Java FaqCO. createdAt uses @JsonFormat("yyyy-MM-dd HH:mm:ss").
type FaqCO struct {
	Id          int64   `json:"id"`
	ServiceId   int64   `json:"serviceId"`
	Title       string  `json:"title"`
	DetailTitle string  `json:"detailTitle"`
	Detail      string  `json:"detail"`
	IzOn        *bool   `json:"izOn"`
	CreatedAt   *string `json:"createdAt"`
}

// GuidePageConfigCmd matches Java GuidePageConfigCmd.
type GuidePageConfigCmd struct {
	Command
	Id            *int64     `json:"id,omitempty"`
	ServiceId     *int64     `json:"serviceId,omitempty"`
	GuidePages    []JumpPage `json:"guidePages,omitempty"`
	AllowSuperEsc *bool      `json:"allowSuperEsc,omitempty"`
	PageNumEsc    *int       `json:"pageNumEsc,omitempty"`
	VisibleRange  *int       `json:"visibleRange,omitempty"`
	Frequency     *int       `json:"frequency,omitempty"`
	IzOn          *bool      `json:"izOn,omitempty"`
	Ids           []int64    `json:"ids,omitempty"`
	ByRegister    *bool      `json:"byRegister,omitempty"`
	ByTags        *bool      `json:"byTags,omitempty"`
	TagIds        string     `json:"tagIds,omitempty"`
}

// GuidePageConfigCO matches Java GuidePageConfigCO.
type GuidePageConfigCO struct {
	Id            int64      `json:"id"`
	ServiceId     int64      `json:"serviceId"`
	GuidePages    []JumpPage `json:"guidePages"`
	AllowSuperEsc *bool      `json:"allowSuperEsc"`
	PageNumEsc    int        `json:"pageNumEsc"`
	VisibleRange  int        `json:"visibleRange"`
	Frequency     int        `json:"frequency"`
	IzOn          *bool      `json:"izOn"`
	ByRegister    *bool      `json:"byRegister"`
	ByTags        *bool      `json:"byTags"`
	TagIds        string     `json:"tagIds"`
}

// HomeActivityEntranceCmd matches Java HomeActivityEntranceCmd.
type HomeActivityEntranceCmd struct {
	Command
	Id           *int64     `json:"id,omitempty"`
	ServiceId    *int64     `json:"serviceId,omitempty"`
	ChainType    *int       `json:"chainType,omitempty"`
	LinkUrl      string     `json:"linkUrl,omitempty"`
	PicUrl       string     `json:"picUrl,omitempty"`
	LinkTitle    string     `json:"linkTitle,omitempty"`
	AppId        string     `json:"appId,omitempty"`
	Param        string     `json:"param,omitempty"`
	IzOn         *bool      `json:"izOn,omitempty"`
	Position     *int       `json:"position,omitempty"`
	OpDownOffset *int       `json:"opDownOffset,omitempty"`
	VisibleRange *int       `json:"visibleRange,omitempty"`
	StartTime    *time.Time `json:"startTime,omitempty"`
	EndTime      *time.Time `json:"endTime,omitempty"`
	Unlimited    *bool      `json:"unlimited,omitempty"`
	ByRegister   *bool      `json:"byRegister,omitempty"`
	ByTags       *bool      `json:"byTags,omitempty"`
	TagIds       string     `json:"tagIds,omitempty"`
}

// HomeActivityEntranceCO matches Java HomeActivityEntranceCO.
// startTime/endTime use @JsonFormat("yyyy-MM-dd HH:mm:ss"), not RFC3339.
type HomeActivityEntranceCO struct {
	Id           int64   `json:"id"`
	ServiceId    int64   `json:"serviceId"`
	ChainType    int     `json:"chainType"`
	LinkUrl      string  `json:"linkUrl"`
	PicUrl       string  `json:"picUrl"`
	LinkTitle    string  `json:"linkTitle"`
	AppId        string  `json:"appId"`
	Param        string  `json:"param"`
	IzOn         *bool   `json:"izOn"`
	Position     int     `json:"position"`
	OpDownOffset int     `json:"opDownOffset"`
	VisibleRange int     `json:"visibleRange"`
	StartTime    *string `json:"startTime"`
	EndTime      *string `json:"endTime"`
	Unlimited    *bool   `json:"unlimited"`
	ByRegister   *bool   `json:"byRegister"`
	ByTags       *bool   `json:"byTags"`
	TagIds       string  `json:"tagIds"`
}

// SpecialTipsCmd matches Java SpecialTipsCmd.
type SpecialTipsCmd struct {
	Command
	Id               *int64 `json:"id,omitempty"`
	ServiceId        *int64 `json:"serviceId,omitempty"`
	PopUpType        *int   `json:"popUpType,omitempty"`
	BgUrl            string `json:"bgUrl,omitempty"`
	BgColor          string `json:"bgColor,omitempty"`
	Title            string `json:"title,omitempty"`
	TitleColor       string `json:"titleColor,omitempty"`
	IzSubtitle       *bool  `json:"izSubtitle,omitempty"`
	Subtitle         string `json:"subtitle,omitempty"`
	SubtitleColor    string `json:"subtitleColor,omitempty"`
	Body             string `json:"body,omitempty"`
	BodyColor        string `json:"bodyColor,omitempty"`
	PopUpTime        *int   `json:"popUpTime,omitempty"`
	IzButton         *bool  `json:"izButton,omitempty"`
	ButtonText       string `json:"buttonText,omitempty"`
	ButtonColor      string `json:"buttonColor,omitempty"`
	ButtonTextColor  string `json:"buttonTextColor,omitempty"`
	ClickEvent       *int   `json:"clickEvent,omitempty"`
	JumpPage         string `json:"jumpPage,omitempty"`
	IzCheckRead      *bool  `json:"izCheckRead,omitempty"`
	CheckReadContent string `json:"checkReadContent,omitempty"`
	VisibleRange     *int   `json:"visibleRange,omitempty"`
	Frequency        *int   `json:"frequency,omitempty"`
	ClosePosition    *int   `json:"closePosition,omitempty"`
	IzOn             *bool  `json:"izOn,omitempty"`
	ByRegister       *bool  `json:"byRegister,omitempty"`
	ByTags           *bool  `json:"byTags,omitempty"`
	TagIds           string `json:"tagIds,omitempty"`
}

// UnmarshalJSON accepts jumpPage/bgUrl as JSON strings (fence Cmd / business Feign) or as
// object/array (PC payload forwarded by business-go), mirroring Java SpecialTipsDTO setters.
func (s *SpecialTipsCmd) UnmarshalJSON(data []byte) error {
	type Alias SpecialTipsCmd
	aux := &struct {
		JumpPage json.RawMessage `json:"jumpPage,omitempty"`
		BgUrl    json.RawMessage `json:"bgUrl,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := helpConfigDTOJSON.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.JumpPage = normalizeHelpConfigJSONString(aux.JumpPage)
	s.BgUrl = normalizeHelpConfigJSONString(aux.BgUrl)
	return nil
}

func normalizeHelpConfigJSONString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	if trimmed[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str); err != nil {
			return ""
		}
		return str
	}
	return trimmed
}

// SpecialTipsCO matches Java SpecialTipsCO.
type SpecialTipsCO struct {
	Id               int64  `json:"id"`
	ServiceId        int64  `json:"serviceId"`
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

// CustomerServiceCmd matches Java CustomerServiceCmd.
type CustomerServiceCmd struct {
	Command
	Id                   *int64  `json:"id,omitempty"`
	ServiceId            *int64  `json:"serviceId,omitempty"`
	CopyServiceId        []int64 `json:"copyServiceId,omitempty"`
	Tel                  string  `json:"tel,omitempty"`
	IzOnlineEntrance     *bool   `json:"izOnlineEntrance,omitempty"`
	IzArtificialEntrance *bool   `json:"izArtificialEntrance,omitempty"`
	IzWorkTime           *bool   `json:"izWorkTime,omitempty"`
	StartTime            string  `json:"startTime,omitempty"`
	EndTime              string  `json:"endTime,omitempty"`
	Tips                 string  `json:"tips,omitempty"`
}

// CustomerServiceCO matches Java CustomerServiceCO.
type CustomerServiceCO struct {
	ServiceId            int64   `json:"serviceId"`
	Tel                  string  `json:"tel"`
	StartTime            string  `json:"startTime"`
	EndTime              string  `json:"endTime"`
	IzOnlineEntrance     *bool   `json:"izOnlineEntrance"`
	IzArtificialEntrance *bool   `json:"izArtificialEntrance"`
	IzWorkTime           *bool   `json:"izWorkTime"`
	Tips                 string  `json:"tips"`
	UpdatedName          *string `json:"updatedName"`
	UpdatedPin           string  `json:"updatedPin"`
	UpdatedAt            string  `json:"updatedAt"`
}

// HomeNavCmd matches Java HomeNavCmd.
type HomeNavCmd struct {
	Command
	Id        *int64  `json:"id,omitempty"`
	ServiceId *int64  `json:"serviceId,omitempty"`
	CarType   *int    `json:"carType,omitempty"`
	Name      string  `json:"name,omitempty"`
	Icon      string  `json:"icon,omitempty"`
	JumpPage  string  `json:"jumpPage,omitempty"`
	IzOn      *bool   `json:"izOn,omitempty"`
	Ids       []int64 `json:"ids,omitempty"`
}

// HomeNavCO matches Java HomeNavCO. ids/updatedName have no source in HomeNavDO,
// so ConvertorHelper leaves them null.
type HomeNavCO struct {
	Id          int64   `json:"id"`
	ServiceId   int64   `json:"serviceId"`
	CarType     int     `json:"carType"`
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	JumpPage    string  `json:"jumpPage"`
	IzOn        *bool   `json:"izOn"`
	Ids         []int64 `json:"ids"`
	UpdatedPin  string  `json:"updatedPin"`
	UpdatedAt   string  `json:"updatedAt"`
	UpdatedName *string `json:"updatedName"`
}
