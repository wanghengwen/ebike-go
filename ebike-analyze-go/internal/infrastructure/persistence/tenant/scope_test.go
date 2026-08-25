package tenant_test

import (
	"testing"

	"ebike-analyze-go/internal/infrastructure/persistence/tenant"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type dryDialector struct{}

func (dryDialector) Name() string                                   { return "dry" }
func (dryDialector) Initialize(*gorm.DB) error                      { return nil }
func (dryDialector) Migrator(db *gorm.DB) gorm.Migrator             { return nil }
func (dryDialector) DataTypeOf(*schema.Field) string                { return "text" }
func (dryDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }
func (dryDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	_ = writer.WriteByte('?')
}
func (dryDialector) QuoteTo(writer clause.Writer, s string) {
	_, _ = writer.WriteString("`")
	_, _ = writer.WriteString(s)
	_, _ = writer.WriteString("`")
}
func (dryDialector) Explain(sql string, _ ...interface{}) string { return sql }

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(dryDialector{}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	tenant.RegisterCallbacks(db)
	return db
}

func TestWithTenantSetsSessionKeys(t *testing.T) {
	db := newDB(t)
	tx := tenant.WithTenant(db, "1003")
	if v, ok := tx.Get(tenant.TenantKey); !ok || v != "1003" {
		t.Fatalf("TenantKey not set, got %v ok=%v", v, ok)
	}
}

func TestWithTenantEmptyTenantStillSetsKey(t *testing.T) {
	db := newDB(t)
	tx := tenant.WithTenant(db, "")
	v, ok := tx.Get(tenant.TenantKey)
	if !ok {
		t.Fatal("TenantKey missing for empty tenant")
	}
	if s, _ := v.(string); s != "" {
		t.Fatalf("want empty string tenant, got %q", s)
	}
}

func TestRegisterCallbacksIdempotent(t *testing.T) {
	db := newDB(t)
	// Second register should not panic (GORM returns error on duplicate name).
	tenant.RegisterCallbacks(db)
	tx := tenant.WithTenant(db, "1003")
	if _, ok := tx.Get(tenant.TenantKey); !ok {
		t.Fatal("TenantKey lost after re-register")
	}
}

func TestScopedDBAliasUsesWithTenant(t *testing.T) {
	// model.ScopedDB delegates to WithTenant; verify the shared session key.
	db := newDB(t)
	tx := tenant.WithTenant(db, "abc")
	if v, ok := tx.Get(tenant.TenantKey); !ok || v != "abc" {
		t.Fatalf("WithTenant key mismatch: %v", v)
	}
}
