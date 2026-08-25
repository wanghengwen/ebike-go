package config

import (
	"strings"
	"testing"
)

func TestToGoMySQLParams_JDBCConnectionParam(t *testing.T) {
	jdbc := "useSSL=false&useUnicode=true&characterEncoding=UTF-8&serverTimezone=GMT%2B8&zeroDateTimeBehavior=convertToNull"
	got := toGoMySQLParams(jdbc)

	for _, want := range []string{
		"parseTime=True",
		"loc=Asia%2FShanghai",
		"charset=utf8mb4",
		"tls=false",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
	for _, bad := range []string{"useSSL", "useUnicode", "characterEncoding", "serverTimezone", "zeroDateTimeBehavior"} {
		if strings.Contains(got, bad) {
			t.Fatalf("jdbc-only key %q should not appear in %q", bad, got)
		}
	}
}

func TestToGoMySQLParams_Empty(t *testing.T) {
	got := toGoMySQLParams("")
	if got == "" {
		t.Fatal("expected default params")
	}
	if !strings.Contains(got, "parseTime=True") || !strings.Contains(got, "tls=false") {
		t.Fatalf("unexpected defaults: %q", got)
	}
}

func TestToGoMySQLParams_PreservesGoNativeParams(t *testing.T) {
	jdbc := "useSSL=false&allowPublicKeyRetrieval=true"
	got := toGoMySQLParams(jdbc)
	if !strings.Contains(got, "allowPublicKeyRetrieval=true") {
		t.Fatalf("expected allowPublicKeyRetrieval preserved in %q", got)
	}
}
