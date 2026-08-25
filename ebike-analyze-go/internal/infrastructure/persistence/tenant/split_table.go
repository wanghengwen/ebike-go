package tenant

import "strings"

// SplitTableNames is the confirmed list of tables that use dynamic table name splitting.
// Loaded from Nacos: xyy.tableSplit.tableNames
var SplitTableNames = map[string]bool{
	"t_car_statistics":                   true,
	"t_parking_site_statistics":          true,
	"t_user_user":                        true,
	"t_user_riding_qualification":        true,
	"t_user_riding_info":                 true,
	"t_user_third_login":                 true,
	"t_ebike_visual_wallet_detail":       true,
	"t_ebike_visual_riding_card_detail":  true,
	"t_ebike_visual_deposit_card_detail": true,
}

// Enabled mirrors Java's DynamicTableNameInnerInterceptor enable switch
// (xyy.table-split.enable). When false, NO table gets a tenant suffix,
// regardless of whether it appears in SplitTableNames.
var Enabled = true

// ResolveTableName resolves a table name with tenant suffix if applicable.
// Mirrors Java: enable && tableNames.contains(table) -> tableName + "_" + tenantId.
func ResolveTableName(tableName, tenantID string) string {
	if Enabled && tenantID != "" && SplitTableNames[tableName] {
		return tableName + "_" + tenantID
	}
	return tableName
}

// ResolveRawSQL replaces all split table names in a raw SQL string with tenant-suffixed versions.
// Used for cross-database JOIN queries (e.g., MemberStatisticMapper).
func ResolveRawSQL(sql, tenantID string) string {
	if !Enabled || tenantID == "" {
		return sql
	}
	for name := range SplitTableNames {
		sql = strings.ReplaceAll(sql, name, name+"_"+tenantID)
	}
	return sql
}

// InitFromConfig updates SplitTableNames and Enabled from configuration.
func InitFromConfig(enabled bool, tableNames []string) {
	Enabled = enabled
	if len(tableNames) > 0 {
		SplitTableNames = make(map[string]bool, len(tableNames))
		for _, name := range tableNames {
			SplitTableNames[name] = true
		}
	}
}
