package part

import domainconfig "ebike-fence-go/internal/domain/config"

const (
	PartOpeFix         = 0
	PartOpeTakeOutAuto = 1
	PartOpeWearAuto    = 2
	PartOpeHandle      = 3
)

// HelmetAnalysisInput mirrors Java HelmetEntity fields used by returnAnalysisResult.
type HelmetAnalysisInput struct {
	OpenHelmet       string
	HelmetAudit      bool
	HelmetLock       *int
	Helmet6React     *int
	Helmet6Lock      *int
	IzExist          bool
	FixTicketNo      *int64
	BackcarConfig    *domainconfig.ConfigBackcarCO
	HelmetAuditWhite []string
	TenantId         string
}

// HelmetAnalysisOutput mirrors PartAnalysisResultEntity from helmet analysis.
type HelmetAnalysisOutput struct {
	Name    string
	IzExist *bool
	Result  *bool
	CanUse  *bool
	UseType *int
}

func boolPtr(v bool) *bool {
	b := v
	return &b
}

func intPtr(v int) *int {
	i := v
	return &i
}

// HelmetReturnAnalysis mirrors Java HelmetEntity.returnAnalysisResult.
func HelmetReturnAnalysis(in HelmetAnalysisInput) HelmetAnalysisOutput {
	out := HelmetAnalysisOutput{Name: "helmet", IzExist: boolPtr(in.IzExist)}
	if in.OpenHelmet != "" && !in.HelmetAudit {
		ok := HelmetReturnResult(in.OpenHelmet, false, in.HelmetLock, in.Helmet6React, in.Helmet6Lock, in.HelmetAuditWhite, in.TenantId)
		out.Result = boolPtr(ok)
	} else {
		out.Result = boolPtr(true)
	}
	canUse := CanHelmetUnlock(in.IzExist, in.FixTicketNo)
	out.CanUse = &canUse
	useType := HelmetUnlockType(in.IzExist, in.BackcarConfig)
	out.UseType = &useType
	return out
}

// CanHelmetUnlock mirrors Java HelmetEntity.canHelmetUnlock.
func CanHelmetUnlock(izExist bool, fixTicketNo *int64) bool {
	return izExist && fixTicketNo == nil
}

// HelmetUnlockType mirrors Java HelmetEntity.helmetUnlockType.
func HelmetUnlockType(izExist bool, cfg *domainconfig.ConfigBackcarCO) int {
	if !izExist {
		return PartOpeFix
	}
	if cfg != nil && cfg.GetIzHelmetWearDetection() {
		return PartOpeWearAuto
	}
	if cfg != nil && cfg.GetIzHelmetRemovalDetection() {
		return PartOpeTakeOutAuto
	}
	return PartOpeHandle
}
