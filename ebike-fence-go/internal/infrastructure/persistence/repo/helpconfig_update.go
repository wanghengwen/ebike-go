package repo

import "gorm.io/gorm"

// helpConfigUpdate mirrors Java updateById: write zero-value ints/bools from the edit payload.
func helpConfigUpdate(db *gorm.DB, row interface{}, fields ...string) error {
	return db.Model(row).Select(fields).Updates(row).Error
}

var (
	helpConfigAuditFields = []string{"UpdatedPin", "UpdatedAt", "Version"}

	homeScrollMsgUpdateFields = []string{
		"Content", "Type", "Appid", "SkipUrl", "Params", "Title", "DetailTitle", "Detail", "IzOn",
	}
	faqUpdateFields       = []string{"Title", "DetailTitle", "Detail", "IzOn"}
	guidePageUpdateFields = []string{
		"GuidePages", "AllowSuperEsc", "PageNumEsc", "VisibleRange", "Frequency",
		"IzOn", "ByRegister", "ByTags", "TagIds",
	}
	homeActivityUpdateFields = []string{
		"ChainType", "LinkUrl", "PicUrl", "LinkTitle", "AppId", "Param", "IzOn",
		"Position", "OpDownOffset", "VisibleRange", "Unlimited", "ByRegister", "ByTags", "TagIds",
	}
	specialTipsUpdateFields = []string{
		"PopUpType", "BgUrl", "BgColor", "Title", "TitleColor", "IzSubtitle", "Subtitle", "SubtitleColor",
		"Body", "BodyColor", "PopUpTime", "IzButton", "ButtonText", "ButtonColor", "ButtonTextColor",
		"ClickEvent", "JumpPage", "IzCheckRead", "CheckReadContent", "VisibleRange", "Frequency",
		"ClosePosition", "IzOn", "ByRegister", "ByTags", "TagIds",
	}
	homeNavUpdateFields = []string{"CarType", "Name", "Icon", "JumpPage", "IzOn", "OrderWeights"}
)

func appendHelpConfigAuditFields(fields []string) []string {
	out := make([]string, 0, len(fields)+len(helpConfigAuditFields))
	out = append(out, fields...)
	out = append(out, helpConfigAuditFields...)
	return out
}
