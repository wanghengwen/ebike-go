package auth

import (
	"testing"

	"ebike-auth-go/internal/pkg/rpc"
)

func TestValidateAuthorizedGrantType(t *testing.T) {
	tenant := &rpc.TenantAuthCo{
		AuthorizedGrantTypes: "password,phone_code,refresh_token",
	}
	if err := validateAuthorizedGrantType(tenant, "password"); err != nil {
		t.Fatalf("expected password allowed: %v", err)
	}
	if err := validateAuthorizedGrantType(tenant, "ak"); err == nil {
		t.Fatal("expected ak to be rejected")
	}
}

func TestCheckBusinessPCPermission(t *testing.T) {
	userWithCode := &rpc.UserDO{Codes: []string{"saas-1"}}
	userWithoutCode := &rpc.UserDO{}

	if !checkBusinessPCPermission("pc", userWithCode) {
		t.Fatal("pc user with saas code should pass")
	}
	if checkBusinessPCPermission("pc", userWithoutCode) {
		t.Fatal("pc user without saas code should fail")
	}
	if !checkBusinessPCPermission("ios", userWithoutCode) {
		t.Fatal("non-pc platform should pass")
	}
}

func TestJavaClientPhoneSecretConstant(t *testing.T) {
	secret := javaClientPhoneSecret(testTenantID)
	if len(secret) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(secret))
	}
}
