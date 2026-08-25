package cache

import (
	"encoding/json"

	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"

	"github.com/guregu/null/v5"
)

func modelConfigBaseToCache(b model.ConfigBaseDO) ConfigBase {
	cb := ConfigBase{
		TenantID:   b.TenantID,
		CreatedPin: b.CreatedPin,
		CreatedAt:  b.CreatedAt,
		UpdatedPin: b.UpdatedPin,
		UpdatedAt:  b.UpdatedAt,
		Version:    b.Version,
	}
	if b.IzDel.Valid {
		cb.IzDel = flexBool(null.BoolFrom(b.IzDel.Bool))
	}
	return cb
}

func MarshalHomeScrollMsgList(rows []model.TConfigHomeScrollMsg) (string, error) {
	out := make([]homeScrollMsg, len(rows))
	for i, r := range rows {
		out[i] = homeScrollMsg{
			ConfigBase:  modelConfigBaseToCache(r.ConfigBaseDO),
			ID:          r.ID,
			ServiceID:   r.ServiceID,
			Content:     r.Content,
			Type:        r.Type,
			Appid:       r.Appid,
			SkipUrl:     r.SkipUrl,
			Params:      r.Params,
			Title:       r.Title,
			DetailTitle: r.DetailTitle,
			Detail:      r.Detail,
			IzOn:        r.IzOn,
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MarshalFaqList(rows []model.TConfigFaq) (string, error) {
	out := make([]faq, len(rows))
	for i, r := range rows {
		out[i] = faq{
			ConfigBase:  modelConfigBaseToCache(r.ConfigBaseDO),
			ID:          r.ID,
			ServiceID:   r.ServiceID,
			Title:       r.Title,
			DetailTitle: r.DetailTitle,
			Detail:      r.Detail,
			IzOn:        r.IzOn,
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MarshalHomeActivityEntranceList(rows []model.TConfigHomeActivityEntrance) (string, error) {
	out := make([]homeActivityEntrance, len(rows))
	for i, r := range rows {
		item := homeActivityEntrance{
			ConfigBase:   modelConfigBaseToCache(r.ConfigBaseDO),
			ID:           r.ID,
			ServiceID:    r.ServiceID,
			ChainType:    r.ChainType,
			LinkUrl:      r.LinkUrl,
			PicUrl:       r.PicUrl,
			LinkTitle:    r.LinkTitle,
			AppId:        r.AppId,
			Param:        r.Param,
			IzOn:         r.IzOn,
			Position:     r.Position,
			OpDownOffset: r.OpDownOffset,
			VisibleRange: r.VisibleRange,
			Unlimited:    r.Unlimited,
			ByRegister:   r.ByRegister,
			ByTags:       r.ByTags,
			TagIds:       r.TagIds,
		}
		if r.StartTime != nil {
			item.StartTime = timefmt.FromTime(*r.StartTime)
		}
		if r.EndTime != nil {
			item.EndTime = timefmt.FromTime(*r.EndTime)
		}
		out[i] = item
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func modelToGuidePage(r model.TConfigGuidePage) guidePage {
	return guidePage{
		ConfigBase:    modelConfigBaseToCache(r.ConfigBaseDO),
		ID:            r.ID,
		ServiceID:     r.ServiceID,
		GuidePages:    r.GuidePages,
		AllowSuperEsc: r.AllowSuperEsc,
		PageNumEsc:    r.PageNumEsc,
		VisibleRange:  r.VisibleRange,
		Frequency:     r.Frequency,
		IzOn:          r.IzOn,
		OrderWeights:  r.OrderWeights,
		ByRegister:    r.ByRegister,
		ByTags:        r.ByTags,
		TagIds:        r.TagIds,
	}
}

func MarshalGuidePage(row model.TConfigGuidePage) (string, error) {
	b, err := json.Marshal(modelToGuidePage(row))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MarshalCustomerService(row model.TConfigCustomerService) (string, error) {
	out := customerService{
		ConfigBase:           modelConfigBaseToCache(row.ConfigBaseDO),
		ID:                   row.ID,
		ServiceID:            row.ServiceID,
		StartTime:            row.StartTime,
		EndTime:              row.EndTime,
		Tel:                  row.Tel,
		IzOnlineEntrance:     row.IzOnlineEntrance,
		IzArtificialEntrance: row.IzArtificialEntrance,
		IzWorkTime:           row.IzWorkTime,
		Tips:                 row.Tips,
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MarshalSpecialTipsList(rows []model.TConfigSpecialTips) (string, error) {
	out := make([]specialTips, len(rows))
	for i, r := range rows {
		out[i] = specialTips{
			ConfigBase:       modelConfigBaseToCache(r.ConfigBaseDO),
			ID:               r.ID,
			ServiceID:        r.ServiceID,
			PopUpType:        r.PopUpType,
			BgUrl:            r.BgUrl,
			BgColor:          r.BgColor,
			Title:            r.Title,
			TitleColor:       r.TitleColor,
			IzSubtitle:       r.IzSubtitle,
			Subtitle:         r.Subtitle,
			SubtitleColor:    r.SubtitleColor,
			Body:             r.Body,
			BodyColor:        r.BodyColor,
			PopUpTime:        r.PopUpTime,
			IzButton:         r.IzButton,
			ButtonText:       r.ButtonText,
			ButtonColor:      r.ButtonColor,
			ButtonTextColor:  r.ButtonTextColor,
			ClickEvent:       r.ClickEvent,
			JumpPage:         r.JumpPage,
			IzCheckRead:      r.IzCheckRead,
			CheckReadContent: r.CheckReadContent,
			VisibleRange:     r.VisibleRange,
			Frequency:        r.Frequency,
			ClosePosition:    r.ClosePosition,
			IzOn:             r.IzOn,
			ByRegister:       r.ByRegister,
			ByTags:           r.ByTags,
			TagIds:           r.TagIds,
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MarshalHomeNavList(rows []model.TConfigHomeNav) (string, error) {
	out := make([]homeNav, len(rows))
	for i, r := range rows {
		out[i] = homeNav{
			ConfigBase:   modelConfigBaseToCache(r.ConfigBaseDO),
			ID:           r.ID,
			ServiceID:    r.ServiceID,
			CarType:      r.CarType,
			Name:         r.Name,
			Icon:         r.Icon,
			JumpPage:     r.JumpPage,
			IzOn:         r.IzOn,
			OrderWeights: r.OrderWeights,
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
