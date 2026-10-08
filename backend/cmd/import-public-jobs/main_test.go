package main

import (
	"net/url"
	"testing"
)

func TestMatchesDatabaseUser(t *testing.T) {
	production, err := url.Parse("postgresql://postgres.production:password@pooler.example.com:6543/postgres")
	if err != nil {
		t.Fatal(err)
	}
	staging, err := url.Parse("postgresql://postgres.staging:password@pooler.example.com:6543/postgres")
	if err != nil {
		t.Fatal(err)
	}
	if !matchesDatabaseUser(production, "postgres.production") {
		t.Fatal("expected the reviewed production user to match")
	}
	if matchesDatabaseUser(staging, "postgres.production") {
		t.Fatal("a staging URL on the same pooler must not match")
	}
	if matchesDatabaseUser(production, "") {
		t.Fatal("missing reviewed user must not match")
	}
}

func TestParseDatabaseURLRejectsSecretFormatting(t *testing.T) {
	for _, raw := range []string{
		"'postgresql://user:password@pooler.example.com:6543/postgres'",
		"postgresql://user:[YOUR-PASSWORD]@pooler.example.com:6543/postgres",
		"not-a-database-url",
	} {
		if _, err := parseDatabaseURL(raw); err == nil {
			t.Fatalf("expected malformed URL to be rejected")
		}
	}
	if _, err := parseDatabaseURL("  postgresql://user:password@pooler.example.com:6543/postgres\n"); err != nil {
		t.Fatal(err)
	}
}
