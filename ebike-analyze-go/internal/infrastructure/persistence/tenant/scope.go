package tenant

import "gorm.io/gorm"

// TenantKey is the GORM setting key used by WithTenant / RegisterCallbacks.
// Mirrors Java TenantLineInnerInterceptor which reads tenantId from CommandContext.
const TenantKey = "xyy:tenant_id"

const tenantAppliedKey = "xyy:tenant_applied"

// TenantScope adds tenant_id filter to GORM queries.
// Aligns with Java UserTenantLineHandler:
//   - blank tenantId → WHERE tenant_id IS NULL
//   - non-blank       → WHERE tenant_id = ?
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return applyTenantWhere(db, tenantID)
	}
}

// WithTenant is the preferred repository entry point. It:
//
//  1. Stores tenantID on the session (for the Query callback safety net)
//
//  2. Eagerly applies TenantScope so builders like Table()/Count()/Find()
//     are filtered even if a callback path is skipped
//
//     tenant.WithTenant(db, tenantID).Table(...).Where(...)
func WithTenant(db *gorm.DB, tenantID string) *gorm.DB {
	if db == nil {
		return db
	}
	return db.
		Set(TenantKey, tenantID).
		Set(tenantAppliedKey, true).
		Scopes(TenantScope(tenantID))
}

func applyTenantWhere(db *gorm.DB, tenantID string) *gorm.DB {
	if tenantID == "" {
		return db.Where("tenant_id IS NULL")
	}
	return db.Where("tenant_id = ?", tenantID)
}

// RegisterCallbacks installs a Query callback that injects tenant_id when the
// session carries TenantKey but WithTenant's eager Scope was not applied
// (e.g. callers that only did db.Set(TenantKey, ...)). This is a safety net
// analogous to Java's TenantLineInnerInterceptor.
//
// NOTE: db.Raw(...) bypasses this callback — raw SQL must embed tenant_id
// conditions explicitly (enforced by scripts/check_tenant_filter.go).
func RegisterCallbacks(db *gorm.DB) {
	if db == nil {
		return
	}
	cb := db.Callback().Query().Before("gorm:query")
	// Use Replace for idempotent re-attach (mysql.Reinit). Do NOT Remove+Register:
	// GORM compile treats that as two pending ops for the same name and errors.
	if err := cb.Replace("xyy:tenant_filter", tenantQueryCallback); err != nil {
		_ = cb.Register("xyy:tenant_filter", tenantQueryCallback)
	}
}

func tenantQueryCallback(db *gorm.DB) {
	if db == nil {
		return
	}
	if _, applied := db.Get(tenantAppliedKey); applied {
		return
	}
	v, ok := db.Get(TenantKey)
	if !ok {
		return
	}
	tenantID, _ := v.(string)
	applyTenantWhere(db, tenantID)
}
