package auth

import (
	"fmt"
	"strings"

	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/rpc"
)

func isGrantTypeAuthorized(tenantAuth *rpc.TenantAuthCo, grantType string) bool {
	if tenantAuth == nil || tenantAuth.AuthorizedGrantTypes == "" {
		return true
	}
	for _, allowed := range strings.Split(tenantAuth.AuthorizedGrantTypes, ",") {
		if strings.TrimSpace(allowed) == grantType {
			return true
		}
	}
	return false
}

func validateAuthorizedGrantType(tenantAuth *rpc.TenantAuthCo, grantType string) error {
	if isGrantTypeAuthorized(tenantAuth, grantType) {
		return nil
	}
	return fmt.Errorf("unsupported grant type: %s", grantType)
}

func isGrayTenant(tenantId string) bool {
	for _, id := range config.AppConfig.GrayTenantIds {
		if id == tenantId {
			return true
		}
	}
	return false
}

func checkBusinessPCPermission(platform string, user *rpc.UserDO) bool {
	if getPlatformName(platform) != "pc" {
		return true
	}
	return user != nil && len(user.Codes) > 0
}
