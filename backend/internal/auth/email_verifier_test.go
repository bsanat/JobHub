package auth

import "context"

type testEmailVerifier struct{}

func (testEmailVerifier) VerifyEmail(_ context.Context, token string) (OAuthIdentity, error) {
	if token == "" {
		return OAuthIdentity{}, ErrEmailVerificationInvalid
	}
	return OAuthIdentity{Email: token}, nil
}
