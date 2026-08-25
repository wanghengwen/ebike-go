package middleware

import (
	"strings"
	"sync"

	"ebike-service-client-go/internal/pkg/config"
)

var (
	authExcludeOnce     sync.Once
	authExcludePrefixes []string
)

// InitAuthExclude loads xyy.gatewayfilter.exclude from config. Call after LoadConfig.
func InitAuthExclude() {
	authExcludeOnce = sync.Once{}
	authExcludePrefixes = nil
	authExcludeOnce.Do(loadAuthExclude)
}

func loadAuthExclude() {
	raw := config.GlobalConfig.Xyy.GatewayFilter.Exclude
	if raw == "" {
		return
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			authExcludePrefixes = append(authExcludePrefixes, part)
		}
	}
}

func excludedPrefixes() []string {
	authExcludeOnce.Do(loadAuthExclude)
	return authExcludePrefixes
}

// normalizeRequestPath mirrors FilterAutoAssemble: strip trailing slashes.
func normalizeRequestPath(path string) string {
	return strings.TrimRight(path, "/")
}

func isAuthExcluded(path string) bool {
	path = normalizeRequestPath(path)
	for _, ex := range excludedPrefixes() {
		if strings.HasPrefix(path, ex) {
			return true
		}
	}
	return false
}
