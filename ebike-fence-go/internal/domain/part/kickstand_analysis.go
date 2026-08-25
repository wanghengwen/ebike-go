package part

import "ebike-fence-go/internal/infrastructure/rpc"

// KickstandReturnAnalysis mirrors Java KickstandEntity.returnAnalysisResult.
func KickstandReturnAnalysis(kick *rpc.KickstandEntity, izUsed bool) (name string, izExist, result *bool) {
	name = "kickstand"
	izExist = boolPtr(izUsed)
	if kick == nil {
		result = boolPtr(false)
		return name, izExist, result
	}
	ok := (kick.Type != nil && kick.MagState != nil && *kick.Type == 1 && *kick.MagState == 1) ||
		(kick.Type != nil && kick.RfState != nil && *kick.Type == 2 && *kick.RfState == 1) ||
		(kick.Type != nil && kick.MagState != nil && kick.RfState != nil && *kick.Type == 3 && *kick.MagState == 1 && *kick.RfState == 1) ||
		(kick.Type != nil && kick.RfState != nil && *kick.Type == 4 && *kick.RfState == 1)
	result = boolPtr(ok)
	return name, izExist, result
}
