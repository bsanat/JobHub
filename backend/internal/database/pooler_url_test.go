package database

import (
	"net/url"
	"testing"
)

func TestSessionPoolerURLPreservesIdentityAndQuery(t *testing.T) {
	raw := "postgresql://postgres.example:local-test@aws-0-test.pooler.supabase.com:6543/postgres?sslmode=require&search_path=public"
	converted, err := SessionPoolerURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(converted)
	if err != nil || u.Hostname() != "aws-0-test.pooler.supabase.com" || u.Port() != "5432" || u.User.Username() != "postgres.example" || u.Query().Get("sslmode") != "require" || u.Query().Get("search_path") != "public" {
		t.Fatalf("session mode changed the connection identity: %v", err)
	}
}

func TestSessionPoolerURLRejectsOtherTargets(t *testing.T) {
	for _, raw := range []string{
		"postgresql://user:password@localhost:6543/postgres",
		"postgresql://user:password@db.example.supabase.co:5432/postgres",
		"postgresql://user:password@aws-0-test.pooler.supabase.com:5432/postgres",
	} {
		if _, err := SessionPoolerURL(raw); err == nil {
			t.Fatalf("accepted an unverified transaction pooler: %s", raw)
		}
	}
}
