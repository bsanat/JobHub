package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestNeonEmailVerification(t *testing.T) {
	pub, private, _ := ed25519.GenerateKey(rand.Reader)
	jwks := []byte(fmt.Sprintf(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"test","alg":"EdDSA","use":"sig","x":"%s"}]}`, base64.RawURLEncoding.EncodeToString(pub)))
	keys, err := keyfunc.NewJWKSetJSON(jwks)
	if err != nil {
		t.Fatal(err)
	}
	v := &NeonOAuthVerifier{keys: keys, issuer: "https://auth.example"}
	for _, name := range []string{"valid", "unconfirmed", "expired", "issuer", "audience", "missing-expiry", "banned", "empty-subject", "forged", "wrong-algorithm"} {
		t.Run(name, func(t *testing.T) {
			claims := neonClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: v.issuer, Audience: jwt.ClaimStrings{v.issuer}, Subject: "test-user", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)), IssuedAt: jwt.NewNumericDate(time.Now())}, Email: "a@example.com", Name: "Асан", EmailVerified: true}
			method := jwt.SigningMethodEdDSA
			var signingKey any = private
			switch name {
			case "unconfirmed":
				claims.EmailVerified = false
			case "expired":
				claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
			case "issuer":
				claims.Issuer = "https://other.example"
			case "audience":
				claims.Audience = jwt.ClaimStrings{"https://other.example"}
			case "missing-expiry":
				claims.ExpiresAt = nil
			case "banned":
				claims.Banned = true
			case "empty-subject":
				claims.Subject = ""
			case "forged":
				_, signingKey, _ = ed25519.GenerateKey(rand.Reader)
			}
			token := jwt.NewWithClaims(method, claims)
			token.Header["kid"] = "test"
			if name == "wrong-algorithm" {
				token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				token.Header["kid"] = "test"
				signingKey = []byte("not-a-real-secret")
			}
			raw, err := token.SignedString(signingKey)
			if err != nil {
				t.Fatal(err)
			}
			identity, err := v.VerifyEmail(context.Background(), raw)
			if name == "valid" {
				if err != nil || identity.Email != claims.Email || identity.FullName != claims.Name {
					t.Fatalf("valid token rejected: %v", err)
				}
			} else if err != ErrEmailVerificationInvalid {
				t.Fatalf("unsafe token accepted: %s", name)
			}
		})
	}
}

func TestNeonAuthRejectsInvalidConfiguration(t *testing.T) {
	for _, u := range []string{"", "http://example.com/auth", "https://user:secret@example.com/auth", "https://example.com/auth?key=secret"} {
		if _, err := NewNeonOAuthVerifier(context.Background(), u, nil); err == nil {
			t.Fatal("invalid URL accepted")
		}
	}
}
