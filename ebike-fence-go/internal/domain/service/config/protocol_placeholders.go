package configsvc

import (
	"context"
	"strings"

	infrarpc "ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

var managementRPC = infrarpc.NewManagementRPC()

func replaceProtocolPlaceholders(content, tenantName, companyName string) string {
	if content == "" {
		return content
	}
	replacements := []struct{ old, new string }{
		{"${tenantName}", tenantName},
		{"${companyName}", companyName},
	}
	out := content
	for _, r := range replacements {
		out = strings.ReplaceAll(out, r.old, r.new)
	}
	return out
}

func (s *ProtocolService) tenantNames(ctx context.Context, tenantID string) (tenantName, companyName string) {
	t, err := managementRPC.QueryTenant(ctx, tenantID)
	if err != nil || t == nil {
		return "", ""
	}
	return t.TenantName, t.CompanyName
}

func (s *ProtocolService) cloneDefaultWithPlaceholders(ctx context.Context, tenantID string, serviceID int64, typ int) (*model.ConfigProtocol, error) {
	def, err := s.defaultByType(ctx, serviceID, typ)
	if err != nil || def == nil {
		return def, err
	}
	cp := *def
	cp.ServiceID = serviceID
	tn, cn := s.tenantNames(ctx, tenantID)
	cp.Content = replaceProtocolPlaceholders(cp.Content, tn, cn)
	return &cp, nil
}
