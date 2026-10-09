package database

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// SessionPoolerURL switches a verified Supabase shared transaction-pooler URL
// to the same project's session-pooler port. Credentials and host are kept
// unchanged and are never logged by this helper.
func SessionPoolerURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.User == nil ||
		u.Port() != "6543" || !strings.HasSuffix(strings.ToLower(u.Hostname()), ".pooler.supabase.com") {
		return "", errors.New("a Supabase shared transaction pooler URL is required")
	}
	u.Host = net.JoinHostPort(u.Hostname(), "5432")
	return u.String(), nil
}
