package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOAuthUnavailable = errors.New("oauth unavailable")

type OAuthIdentity struct {
	Email    string
	FullName string
}

type OAuthVerifier interface {
	VerifyGoogle(context.Context, string) (OAuthIdentity, error)
}

type EmailVerifier interface {
	VerifyEmail(context.Context, string) (OAuthIdentity, error)
}

type NeonOAuthVerifier struct {
	keys   keyfunc.Keyfunc
	issuer string
	pool   *pgxpool.Pool
}

func NewNeonOAuthVerifier(ctx context.Context, baseURL string, pool *pgxpool.Pool) (*NeonOAuthVerifier, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("NEON_AUTH_BASE_URL must be an HTTPS URL")
	}
	keys, err := keyfunc.NewDefaultOverrideCtx(ctx, []string{parsed.String() + "/.well-known/jwks.json"}, keyfunc.Override{HTTPTimeout: 5 * time.Second, RateLimitWaitMax: time.Second})
	if err != nil {
		return nil, fmt.Errorf("configure Neon token verification: %w", err)
	}
	return &NeonOAuthVerifier{keys: keys, issuer: parsed.Scheme + "://" + parsed.Host, pool: pool}, nil
}

type neonClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"emailVerified"`
	Banned        bool   `json:"banned"`
}

func (v *NeonOAuthVerifier) verify(ctx context.Context, raw string) (neonClaims, error) {
	var claims neonClaims
	_, err := jwt.ParseWithClaims(raw, &claims, v.keys.KeyfuncCtx(ctx), jwt.WithIssuer(v.issuer), jwt.WithAudience(v.issuer), jwt.WithValidMethods([]string{"EdDSA"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || claims.Subject == "" || claims.Email == "" || !claims.EmailVerified || claims.Banned {
		return neonClaims{}, ErrInvalidCredentials
	}
	return claims, nil
}

func (v *NeonOAuthVerifier) VerifyEmail(ctx context.Context, raw string) (OAuthIdentity, error) {
	claims, err := v.verify(ctx, raw)
	if err != nil {
		return OAuthIdentity{}, ErrEmailVerificationInvalid
	}
	return OAuthIdentity{Email: claims.Email, FullName: claims.Name}, nil
}

func (v *NeonOAuthVerifier) VerifyGoogle(ctx context.Context, raw string) (OAuthIdentity, error) {
	claims, err := v.verify(ctx, raw)
	if err != nil {
		return OAuthIdentity{}, err
	}
	// Neon JWTs do not identify the login provider; verify the linked account in the branch.
	var linked bool
	err = v.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM neon_auth.account a JOIN neon_auth.user u ON u.id = a."userId" WHERE a."userId" = $1 AND a."providerId" = 'google' AND u."emailVerified" AND NOT COALESCE(u.banned, false) AND lower(u.email) = lower($2))`, claims.Subject, claims.Email).Scan(&linked)
	if err != nil {
		return OAuthIdentity{}, ErrOAuthUnavailable
	}
	if !linked {
		return OAuthIdentity{}, ErrInvalidCredentials
	}
	return OAuthIdentity{Email: claims.Email, FullName: claims.Name}, nil
}
