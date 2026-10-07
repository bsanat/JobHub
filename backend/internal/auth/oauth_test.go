package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSupabaseEmailVerifierRequiresConfirmedEmail(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantEmail string
		wantErr   error
	}{
		{name: "confirmed", status: http.StatusOK, body: `{"email":"a@example.com","email_confirmed_at":"2026-10-07T10:00:00Z"}`, wantEmail: "a@example.com"},
		{name: "unconfirmed", status: http.StatusOK, body: `{"email":"a@example.com","email_confirmed_at":""}`, wantErr: ErrEmailVerificationInvalid},
		{name: "invalid token", status: http.StatusUnauthorized, body: `{}`, wantErr: ErrEmailVerificationInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/auth/v1/user" || r.Header.Get("Authorization") != "Bearer token" || r.Header.Get("apikey") != "secret" {
					t.Fatalf("unexpected request: %s %#v", r.URL.Path, r.Header)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			verifier, err := NewSupabaseOAuthVerifier(server.URL, "secret", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			identity, err := verifier.VerifyEmail(context.Background(), "token")
			if !errors.Is(err, tc.wantErr) || identity.Email != tc.wantEmail {
				t.Fatalf("identity=%#v err=%v", identity, err)
			}
		})
	}
}
