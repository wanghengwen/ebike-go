package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"ebike-fence-go/internal/pkg/timefmt"

	jsoniter "github.com/json-iterator/go"
)

// ConfigBaseDO mirrors Java BaseDO fields on helpconfig tables.
type ConfigBaseDO struct {
	TenantID   string                `gorm:"column:tenant_id" json:"tenantId,omitempty"`
	CreatedPin string                `gorm:"column:created_pin" json:"createdPin,omitempty"`
	CreatedAt  timefmt.JavaLocalTime `gorm:"column:created_at" json:"createdAt,omitempty"`
	UpdatedPin string                `gorm:"column:updated_pin" json:"updatedPin,omitempty"`
	UpdatedAt  timefmt.JavaLocalTime `gorm:"column:updated_at" json:"updatedAt,omitempty"`
	Version    int                   `gorm:"column:version" json:"version,omitempty"`
	IzDel      sql.NullBool          `gorm:"column:iz_del" json:"izDel,omitempty"`
}

// TConfigHomeScrollMsg maps t_config_home_scroll_msg.
type TConfigHomeScrollMsg struct {
	ConfigBaseDO
	ID          int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID   int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	Content     string `gorm:"column:content" json:"content,omitempty"`
	Type        int    `gorm:"column:type" json:"type,omitempty"`
	Appid       string `gorm:"column:appid" json:"appid,omitempty"`
	SkipUrl     string `gorm:"column:skip_url" json:"skipUrl,omitempty"`
	Params      string `gorm:"column:params" json:"params,omitempty"`
	Title       string `gorm:"column:title" json:"title,omitempty"`
	DetailTitle string `gorm:"column:detail_title" json:"detailTitle,omitempty"`
	Detail      string `gorm:"column:detail" json:"detail,omitempty"`
	IzOn        *bool  `gorm:"column:iz_on" json:"izOn,omitempty"`
}

func (TConfigHomeScrollMsg) TableName() string { return "t_config_home_scroll_msg" }

// TConfigFaq maps t_config_faq.
type TConfigFaq struct {
	ConfigBaseDO
	ID          int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID   int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	Title       string `gorm:"column:title" json:"title,omitempty"`
	DetailTitle string `gorm:"column:detail_title" json:"detailTitle,omitempty"`
	Detail      string `gorm:"column:detail" json:"detail,omitempty"`
	IzOn        *bool  `gorm:"column:iz_on" json:"izOn,omitempty"`
}

func (TConfigFaq) TableName() string { return "t_config_faq" }

// TConfigGuidePage maps t_config_guide_page.
type TConfigGuidePage struct {
	ConfigBaseDO
	ID            int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID     int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	GuidePages    string `gorm:"column:guide_pages" json:"guidePages,omitempty"`
	AllowSuperEsc *bool  `gorm:"column:allow_super_esc" json:"allowSuperEsc,omitempty"`
	PageNumEsc    int    `gorm:"column:page_num_esc" json:"pageNumEsc,omitempty"`
	VisibleRange  int    `gorm:"column:visible_range" json:"visibleRange,omitempty"`
	Frequency     int    `gorm:"column:frequency" json:"frequency,omitempty"`
	IzOn          *bool  `gorm:"column:iz_on" json:"izOn,omitempty"`
	OrderWeights  int    `gorm:"column:order_weights" json:"orderWeights,omitempty"`
	ByRegister    *bool  `gorm:"column:by_register" json:"byRegister,omitempty"`
	ByTags        *bool  `gorm:"column:by_tags" json:"byTags,omitempty"`
	TagIds        string `gorm:"column:tag_ids" json:"tagIds,omitempty"`
}

func (TConfigGuidePage) TableName() string { return "t_config_guide_page" }

// UnmarshalJSON accepts guidePages as a JSON string (DB/Redis) and ignores arrays
// (DTO cmdToModel round-trip) so other fields still map correctly.
func (g *TConfigGuidePage) UnmarshalJSON(data []byte) error {
	type Alias TConfigGuidePage
	aux := &struct {
		GuidePages json.RawMessage `json:"guidePages,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.GuidePages) > 0 && aux.GuidePages[0] == '"' {
		if err := json.Unmarshal(aux.GuidePages, &g.GuidePages); err != nil {
			return err
		}
	}
	return nil
}

// TConfigHomeActivityEntrance maps t_config_home_activity_entrance.
type TConfigHomeActivityEntrance struct {
	ConfigBaseDO
	ID           int64      `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID    int64      `gorm:"column:service_id" json:"serviceId,omitempty"`
	ChainType    int        `gorm:"column:chain_type" json:"chainType,omitempty"`
	LinkUrl      string     `gorm:"column:link_url" json:"linkUrl,omitempty"`
	PicUrl       string     `gorm:"column:pic_url" json:"picUrl,omitempty"`
	LinkTitle    string     `gorm:"column:link_title" json:"linkTitle,omitempty"`
	AppId        string     `gorm:"column:app_id" json:"appId,omitempty"`
	Param        string     `gorm:"column:param" json:"param,omitempty"`
	IzOn         *bool      `gorm:"column:iz_on" json:"izOn,omitempty"`
	Position     int        `gorm:"column:position" json:"position,omitempty"`
	OpDownOffset int        `gorm:"column:op_down_offset" json:"opDownOffset,omitempty"`
	VisibleRange int        `gorm:"column:visible_range" json:"visibleRange,omitempty"`
	StartTime    *time.Time `gorm:"column:start_time" json:"startTime,omitempty"`
	EndTime      *time.Time `gorm:"column:end_time" json:"endTime,omitempty"`
	Unlimited    *bool      `gorm:"column:unlimited" json:"unlimited,omitempty"`
	ByRegister   *bool      `gorm:"column:by_register" json:"byRegister,omitempty"`
	ByTags       *bool      `gorm:"column:by_tags" json:"byTags,omitempty"`
	TagIds       string     `gorm:"column:tag_ids" json:"tagIds,omitempty"`
}

func (TConfigHomeActivityEntrance) TableName() string { return "t_config_home_activity_entrance" }

// TConfigSpecialTips maps t_config_special_tips.
type TConfigSpecialTips struct {
	ConfigBaseDO
	ID               int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID        int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	PopUpType        int    `gorm:"column:pop_up_type" json:"popUpType,omitempty"`
	BgUrl            string `gorm:"column:bg_url" json:"bgUrl,omitempty"`
	BgColor          string `gorm:"column:bg_color" json:"bgColor,omitempty"`
	Title            string `gorm:"column:title" json:"title,omitempty"`
	TitleColor       string `gorm:"column:title_color" json:"titleColor,omitempty"`
	IzSubtitle       *bool  `gorm:"column:iz_subtitle" json:"izSubtitle,omitempty"`
	Subtitle         string `gorm:"column:subtitle" json:"subtitle,omitempty"`
	SubtitleColor    string `gorm:"column:subtitle_color" json:"subtitleColor,omitempty"`
	Body             string `gorm:"column:body" json:"body,omitempty"`
	BodyColor        string `gorm:"column:body_color" json:"bodyColor,omitempty"`
	PopUpTime        int    `gorm:"column:pop_up_time" json:"popUpTime,omitempty"`
	IzButton         *bool  `gorm:"column:iz_button" json:"izButton,omitempty"`
	ButtonText       string `gorm:"column:button_text" json:"buttonText,omitempty"`
	ButtonColor      string `gorm:"column:button_color" json:"buttonColor,omitempty"`
	ButtonTextColor  string `gorm:"column:button_text_color" json:"buttonTextColor,omitempty"`
	ClickEvent       int    `gorm:"column:click_event" json:"clickEvent,omitempty"`
	JumpPage         string `gorm:"column:jump_page" json:"jumpPage,omitempty"`
	IzCheckRead      *bool  `gorm:"column:iz_check_read" json:"izCheckRead,omitempty"`
	CheckReadContent string `gorm:"column:check_read_content" json:"checkReadContent,omitempty"`
	VisibleRange     int    `gorm:"column:visible_range" json:"visibleRange,omitempty"`
	Frequency        int    `gorm:"column:frequency" json:"frequency,omitempty"`
	ClosePosition    int    `gorm:"column:close_position" json:"closePosition,omitempty"`
	IzOn             *bool  `gorm:"column:iz_on" json:"izOn,omitempty"`
	ByRegister       *bool  `gorm:"column:by_register" json:"byRegister,omitempty"`
	ByTags           *bool  `gorm:"column:by_tags" json:"byTags,omitempty"`
	TagIds           string `gorm:"column:tag_ids" json:"tagIds,omitempty"`
}

func (TConfigSpecialTips) TableName() string { return "t_config_special_tips" }

// TConfigCustomerService maps t_config_customer_service.
type TConfigCustomerService struct {
	ConfigBaseDO
	ID                   int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID            int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	StartTime            string `gorm:"column:start_time;type:time" json:"startTime,omitempty"`
	EndTime              string `gorm:"column:end_time;type:time" json:"endTime,omitempty"`
	Tel                  string `gorm:"column:tel" json:"tel,omitempty"`
	IzOnlineEntrance     *bool  `gorm:"column:iz_online_entrance" json:"izOnlineEntrance,omitempty"`
	IzArtificialEntrance *bool  `gorm:"column:iz_artificial_entrance" json:"izArtificialEntrance,omitempty"`
	IzWorkTime           *bool  `gorm:"column:iz_work_time" json:"izWorkTime,omitempty"`
	Tips                 string `gorm:"column:tips" json:"tips,omitempty"`
}

func (TConfigCustomerService) TableName() string { return "t_config_customer_service" }

// TConfigHomeNav maps t_config_home_nav.
type TConfigHomeNav struct {
	ConfigBaseDO
	ID           int64  `gorm:"column:id;primaryKey" json:"id,omitempty"`
	ServiceID    int64  `gorm:"column:service_id" json:"serviceId,omitempty"`
	CarType      int    `gorm:"column:car_type" json:"carType,omitempty"`
	Name         string `gorm:"column:name" json:"name,omitempty"`
	Icon         string `gorm:"column:icon" json:"icon,omitempty"`
	JumpPage     string `gorm:"column:jump_page" json:"jumpPage,omitempty"`
	IzOn         *bool  `gorm:"column:iz_on" json:"izOn,omitempty"`
	OrderWeights int    `gorm:"column:order_weights" json:"orderWeights,omitempty"`
}

func (TConfigHomeNav) TableName() string { return "t_config_home_nav" }
