package env

import (
	"os"
	"strings"
)

// DryRun enables shadow dual-run: Go executes for comparison, client receives Java response
// on default (non-liveList, non-recordList) paths. Does not affect Nacos registration.
func DryRun() bool {
	return strings.EqualFold(os.Getenv("DRY_RUN"), "true")
}

// NacosRegisterEnabled controls whether this instance registers to Nacos.
// Independent of DRY_RUN so shadow diff can run while the pod is discoverable.
// Set NACOS_REGISTER_ENABLED=false to skip registration (local dev).
func NacosRegisterEnabled() bool {
	v := strings.TrimSpace(os.Getenv("NACOS_REGISTER_ENABLED"))
	if v == "" {
		return true
	}
	return strings.EqualFold(v, "true")
}
