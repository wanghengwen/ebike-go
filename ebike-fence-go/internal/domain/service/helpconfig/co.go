package helpconfig

import (
	"encoding/json"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"
)

func ToHomeScrollerMsgCO(m model.TConfigHomeScrollMsg) dto.HomeScrollerMsgCO {
	co := dto.HomeScrollerMsgCO{
		Id:          m.ID,
		ServiceId:   m.ServiceID,
		Content:     m.Content,
		Type:        m.Type,
		Appid:       m.Appid,
		SkipUrl:     m.SkipUrl,
		Params:      m.Params,
		Title:       m.Title,
		DetailTitle: m.DetailTitle,
		Detail:      m.Detail,
		IzOn:        m.IzOn,
		UpdatedPin:  m.UpdatedPin,
	}
	if !m.CreatedAt.IsZero() {
		co.CreatedAt = timefmt.FormatJavaLocalSpace(m.CreatedAt.Time)
	}
	if !m.UpdatedAt.IsZero() {
		co.UpdatedAt = timefmt.FormatJavaLocal(m.UpdatedAt.Time)
	}
	return co
}

func ToHomeScrollerMsgCOList(rows []model.TConfigHomeScrollMsg) []dto.HomeScrollerMsgCO {
	out := make([]dto.HomeScrollerMsgCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToHomeScrollerMsgCO(row))
	}
	return out
}

func ToFaqCO(m model.TConfigFaq) dto.FaqCO {
	co := dto.FaqCO{
		Id:          m.ID,
		ServiceId:   m.ServiceID,
		Title:       m.Title,
		DetailTitle: m.DetailTitle,
		Detail:      m.Detail,
		IzOn:        m.IzOn,
	}
	if !m.CreatedAt.IsZero() {
		at := timefmt.FormatJavaLocalSpace(m.CreatedAt.Time)
		co.CreatedAt = &at
	}
	return co
}

func ToFaqCOList(rows []model.TConfigFaq) []dto.FaqCO {
	out := make([]dto.FaqCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToFaqCO(row))
	}
	return out
}

func ToGuidePageCO(m model.TConfigGuidePage) dto.GuidePageConfigCO {
	co := dto.GuidePageConfigCO{
		Id:            m.ID,
		ServiceId:     m.ServiceID,
		AllowSuperEsc: m.AllowSuperEsc,
		PageNumEsc:    m.PageNumEsc,
		VisibleRange:  m.VisibleRange,
		Frequency:     m.Frequency,
		IzOn:          m.IzOn,
		ByRegister:    m.ByRegister,
		ByTags:        m.ByTags,
		TagIds:        m.TagIds,
	}
	if m.GuidePages != "" {
		var pages []dto.JumpPage
		if json.Unmarshal([]byte(m.GuidePages), &pages) == nil {
			co.GuidePages = pages
		}
	}
	// Mirror Java GuidePageConfigCO.getPageNumEsc(): empty pages / null / overflow → clamp.
	if len(co.GuidePages) == 0 {
		co.PageNumEsc = 1
	} else if co.PageNumEsc <= 0 {
		co.PageNumEsc = 1
	} else if co.PageNumEsc > len(co.GuidePages) {
		co.PageNumEsc = len(co.GuidePages)
	}
	return co
}

// ToHomeActivityCO mirrors Java ConvertorHelper HomeActivityEntranceDO → HomeActivityEntranceCO:
// no audit fields, no orderWeights, and times formatted as "yyyy-MM-dd HH:mm:ss".
func ToHomeActivityCO(m model.TConfigHomeActivityEntrance) dto.HomeActivityEntranceCO {
	co := dto.HomeActivityEntranceCO{
		Id:           m.ID,
		ServiceId:    m.ServiceID,
		ChainType:    m.ChainType,
		LinkUrl:      m.LinkUrl,
		PicUrl:       m.PicUrl,
		LinkTitle:    m.LinkTitle,
		AppId:        m.AppId,
		Param:        m.Param,
		IzOn:         m.IzOn,
		Position:     m.Position,
		OpDownOffset: m.OpDownOffset,
		VisibleRange: m.VisibleRange,
		Unlimited:    m.Unlimited,
		ByRegister:   m.ByRegister,
		ByTags:       m.ByTags,
		TagIds:       m.TagIds,
	}
	if m.StartTime != nil && !m.StartTime.IsZero() {
		at := timefmt.FormatJavaLocalSpace(*m.StartTime)
		co.StartTime = &at
	}
	if m.EndTime != nil && !m.EndTime.IsZero() {
		at := timefmt.FormatJavaLocalSpace(*m.EndTime)
		co.EndTime = &at
	}
	return co
}

func ToHomeActivityCOList(rows []model.TConfigHomeActivityEntrance) []dto.HomeActivityEntranceCO {
	out := make([]dto.HomeActivityEntranceCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToHomeActivityCO(row))
	}
	return out
}

func ToGuidePageCOList(rows []model.TConfigGuidePage) []dto.GuidePageConfigCO {
	out := make([]dto.GuidePageConfigCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToGuidePageCO(row))
	}
	return out
}

func ToCustomerServiceCO(m *model.TConfigCustomerService, serviceID int64) dto.CustomerServiceCO {
	if m == nil {
		return dto.CustomerServiceCO{ServiceId: serviceID}
	}
	sid := m.ServiceID
	if sid == 0 {
		sid = serviceID
	}
	startTime := normalizeTimeHHMMSS(m.StartTime)
	if startTime == "" {
		startTime = "09:00:00"
	}
	endTime := normalizeTimeHHMMSS(m.EndTime)
	if endTime == "" {
		endTime = "18:00:00"
	}
	izArtificial := m.IzArtificialEntrance
	if izArtificial == nil {
		f := false
		izArtificial = &f
	}
	co := dto.CustomerServiceCO{
		ServiceId:            sid,
		Tel:                  m.Tel,
		StartTime:            startTime,
		EndTime:              endTime,
		IzOnlineEntrance:     m.IzOnlineEntrance,
		IzArtificialEntrance: izArtificial,
		IzWorkTime:           m.IzWorkTime,
		Tips:                 m.Tips,
		UpdatedPin:           m.UpdatedPin,
	}
	if !m.UpdatedAt.IsZero() {
		co.UpdatedAt = timefmt.FormatJavaLocal(m.UpdatedAt.Time)
	}
	return co
}

func normalizeTimeHHMMSS(s string) string {
	if s == "" {
		return s
	}
	if strings.Count(s, ":") == 1 {
		return s + ":00"
	}
	return s
}

func ToSpecialTipsCO(m model.TConfigSpecialTips) dto.SpecialTipsCO {
	co := dto.SpecialTipsCO{
		Id:               m.ID,
		ServiceId:        m.ServiceID,
		PopUpType:        m.PopUpType,
		BgUrl:            m.BgUrl,
		BgColor:          m.BgColor,
		Title:            m.Title,
		TitleColor:       m.TitleColor,
		IzSubtitle:       m.IzSubtitle,
		Subtitle:         m.Subtitle,
		SubtitleColor:    m.SubtitleColor,
		Body:             m.Body,
		BodyColor:        m.BodyColor,
		PopUpTime:        m.PopUpTime,
		IzButton:         m.IzButton,
		ButtonText:       m.ButtonText,
		ButtonColor:      m.ButtonColor,
		ButtonTextColor:  m.ButtonTextColor,
		ClickEvent:       m.ClickEvent,
		JumpPage:         m.JumpPage,
		IzCheckRead:      m.IzCheckRead,
		CheckReadContent: m.CheckReadContent,
		VisibleRange:     m.VisibleRange,
		Frequency:        m.Frequency,
		ClosePosition:    m.ClosePosition,
		IzOn:             m.IzOn,
		ByRegister:       m.ByRegister,
		ByTags:           m.ByTags,
		TagIds:           m.TagIds,
	}
	return co
}

func ToSpecialTipsCOList(rows []model.TConfigSpecialTips) []dto.SpecialTipsCO {
	out := make([]dto.SpecialTipsCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToSpecialTipsCO(row))
	}
	return out
}

func ToHomeNavCO(m model.TConfigHomeNav) dto.HomeNavCO {
	co := dto.HomeNavCO{
		Id:         m.ID,
		ServiceId:  m.ServiceID,
		CarType:    m.CarType,
		Name:       m.Name,
		Icon:       m.Icon,
		JumpPage:   m.JumpPage,
		IzOn:       m.IzOn,
		UpdatedPin: m.UpdatedPin,
	}
	if !m.UpdatedAt.IsZero() {
		co.UpdatedAt = timefmt.FormatJavaLocal(m.UpdatedAt.Time)
	}
	return co
}

func ToHomeNavCOList(rows []model.TConfigHomeNav) []dto.HomeNavCO {
	out := make([]dto.HomeNavCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ToHomeNavCO(row))
	}
	return out
}
