package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrOAuthUnavailable = errors.New("oauth unavailable")

type OAuthIdentity struct {
	Email    string
	FullName string
}

type OAuthVerifier interface {
	VerifyGoogle(context.Context, string) (OAuthIdentity, error)
}

type SupabaseOAuthVerifier struct {
	projectURL string
	apiKey     string
	client     *http.Client
}

func NewSupabaseOAuthVerifier(projectURL, apiKey string, client *http.Client) (*SupabaseOAuthVerifier, error) {
	projectURL = strings.TrimRight(strings.TrimSpace(projectURL), "/")
	apiKey = strings.TrimSpace(apiKey)
	parsed, err := url.Parse(projectURL)
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return nil, fmt.Errorf("SUPABASE_URL must be a project HTTPS origin")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_SECRET_KEY is required for OAuth verification")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &SupabaseOAuthVerifier{projectURL: projectURL, apiKey: apiKey, client: client}, nil
}

func (v *SupabaseOAuthVerifier) VerifyGoogle(ctx context.Context, accessToken string) (OAuthIdentity, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return OAuthIdentity{}, ErrInvalidCredentials
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.projectURL+"/auth/v1/user", nil)
	if err != nil {
		return OAuthIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", v.apiKey)
	res, err := v.client.Do(req)
	if err != nil {
		return OAuthIdentity{}, fmt.Errorf("%w: verify google token", ErrOAuthUnavailable)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return OAuthIdentity{}, ErrInvalidCredentials
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return OAuthIdentity{}, fmt.Errorf("%w: supabase auth returned HTTP %d", ErrOAuthUnavailable, res.StatusCode)
	}
	var payload struct {
		Email        string         `json:"email"`
		AppMetadata  map[string]any `json:"app_metadata"`
		UserMetadata struct {
			FullName string `json:"full_name"`
			Name     string `json:"name"`
		} `json:"user_metadata"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return OAuthIdentity{}, fmt.Errorf("%w: decode supabase user", ErrOAuthUnavailable)
	}
	provider, _ := payload.AppMetadata["provider"].(string)
	if provider != "" && provider != "google" {
		return OAuthIdentity{}, ErrInvalidCredentials
	}
	name := strings.TrimSpace(payload.UserMetadata.FullName)
	if name == "" {
		name = strings.TrimSpace(payload.UserMetadata.Name)
	}
	return OAuthIdentity{Email: payload.Email, FullName: name}, nil
}
