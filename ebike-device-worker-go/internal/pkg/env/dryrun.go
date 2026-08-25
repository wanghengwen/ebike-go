package env

import (
	"os"
	"strings"
)

// IsDryRun reports whether shadow comparison mode is enabled.
// Accepts true/1/yes/on; trims spaces and surrounding quotes (Windows cmd: set DRY_RUN="true").
func IsDryRun() bool {
	v := strings.TrimSpace(os.Getenv("DRY_RUN"))
	v = strings.Trim(v, `"'`)
	switch strings.ToLower(v) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
