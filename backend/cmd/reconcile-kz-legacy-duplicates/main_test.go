package main

import (
	"strings"
	"testing"
)

func TestReconciliationRejectsUnlistedOrUngatedSources(t *testing.T) {
	if err := run([]string{"--source", "halyk", "--dry-run"}); err == nil || !strings.Contains(err.Error(), "unsupported legacy source") {
		t.Fatalf("unlisted source was not rejected: %v", err)
	}
	if err := run([]string{"--source", "kcell", "--once"}); err == nil || !strings.Contains(err.Error(), "KZ_IMPORT_WRITE_ENABLED") {
		t.Fatalf("write without explicit gate was not rejected: %v", err)
	}
}

func TestReconciliationChecksTargetIdentityWithoutExposingURL(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("KZ_IMPORT_TARGET", "production")
	t.Setenv("KZ_IMPORT_DATABASE_HOST", "expected.example")
	t.Setenv("KZ_IMPORT_DATABASE_USER", "expected")
	t.Setenv("KZ_IMPORT_READ_DATABASE_URL", "postgresql://other:password@different.example/postgres")
	_, err := checkedDatabaseURL("KZ_IMPORT_READ_DATABASE_URL")
	if err == nil || strings.Contains(err.Error(), "password") {
		t.Fatalf("database identity was accepted or credentials leaked: %v", err)
	}
}
