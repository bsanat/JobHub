package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jobhub-ai/backend/internal/auth"
)

type onboardingStub struct {
	authStub
	state   auth.Onboarding
	saveErr error
	userID  string
	input   auth.OnboardingInput
}

func (s *onboardingStub) GetOnboarding(_ context.Context, userID string) (auth.Onboarding, error) {
	s.userID = userID
	return s.state, nil
}
func (s *onboardingStub) SaveOnboarding(_ context.Context, userID string, input auth.OnboardingInput) (auth.Onboarding, error) {
	s.userID = userID
	s.input = input
	return s.state, s.saveErr
}

func TestOnboardingRoutes(t *testing.T) {
	for _, tc := range []struct {
		name, method, body, origin string
		authErr, csrfErr, saveErr  error
		want                       int
	}{
		{"read own account", "GET", "", "", nil, nil, nil, 200},
		{"save own draft", "PUT", `{"persona":"graduate","step":1,"version":0}`, "http://localhost:5173", nil, nil, nil, 200},
		{"guest rejected", "GET", "", "", auth.ErrInvalidSession, nil, nil, 401},
		{"mutation guest rejected", "PUT", `{}`, "", auth.ErrInvalidSession, nil, nil, 401},
		{"csrf required", "PUT", `{}`, "", nil, auth.ErrForbidden, nil, 403},
		{"untrusted origin", "PUT", `{}`, "https://evil.example", nil, nil, nil, 403},
		{"cannot name another user", "PUT", `{"user_id":"other"}`, "", nil, nil, nil, 400},
		{"unknown fields", "PUT", `{"completed_at":"today"}`, "", nil, nil, nil, 400},
		{"validation error", "PUT", `{}`, "", nil, nil, auth.ErrOnboardingValidation, 400},
		{"stale draft", "PUT", `{}`, "", nil, nil, auth.ErrOnboardingConflict, 409},
		{"database failure", "PUT", `{}`, "", nil, nil, io.ErrUnexpectedEOF, 500},
		{"oversized body", "PUT", `{"profession":"` + strings.Repeat("a", 9000) + `"}`, "", nil, nil, nil, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &onboardingStub{authStub: authStub{session: auth.Session{User: auth.User{ID: "own-account", Role: auth.RoleJobSeeker}}, err: tc.authErr, csrfErr: tc.csrfErr}, state: auth.Onboarding{Required: true, Step: 1}, saveErr: tc.saveErr}
			router := NewRouter(checker{}, slog.New(slog.NewTextHandler(io.Discard, nil)), "http://localhost:5173", stub)
			req := httptest.NewRequest(tc.method, "/api/v1/auth/onboarding", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tc.origin)
			req.Header.Set(auth.CSRFHeaderName, "fixture")
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "fixture"})
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("status=%d want=%d: %s", res.Code, tc.want, res.Body.String())
			}
			if tc.want == 200 {
				var state auth.Onboarding
				if err := json.Unmarshal(res.Body.Bytes(), &state); err != nil || !state.Required {
					t.Fatal("invalid response")
				}
				if stub.userID != "own-account" {
					t.Fatal("must use authenticated user ID")
				}
			}
		})
	}
}

func TestSessionIncludesOnboardingStatus(t *testing.T) {
	data, _ := json.Marshal(mapSessionResponse(auth.Session{OnboardingRequired: true}))
	if !strings.Contains(string(data), `"onboarding_required":true`) {
		t.Fatalf("missing onboarding state: %s", data)
	}
}
