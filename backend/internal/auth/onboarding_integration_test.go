package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"jobhub-ai/backend/internal/database"
)

func TestOnboardingIntegration(t *testing.T) {
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Skip("onboarding integration requires an isolated local database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Connect(ctx, raw, "cache_statement")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := NewService(pool)
	service.SetEmailVerifier(testEmailVerifier{})
	for _, role := range []string{RoleJobSeeker, RoleEmployer} {
		t.Run(role, func(t *testing.T) {
			email := fmt.Sprintf("onboarding-%s-%d@example.com", role, time.Now().UnixNano())
			session, token, err := service.Register(ctx, RegisterInput{FullName: "Onboarding Fixture", Email: email, Password: "isolated-test-password", Role: role, CompanyName: fmt.Sprintf("Fixture Company %d", time.Now().UnixNano()), EmailVerificationToken: email})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_, _ = pool.Exec(context.Background(), `DELETE FROM jobhub.users WHERE id=$1`, session.User.ID)
				if session.Company != nil {
					_, _ = pool.Exec(context.Background(), `DELETE FROM jobhub.companies WHERE id=$1`, session.Company.ID)
				}
			}()
			if !session.OnboardingRequired {
				t.Fatal("new registration must require onboarding")
			}
			state, err := service.GetOnboarding(ctx, session.User.ID)
			if err != nil || !state.Required || state.Step != 1 || state.Version != 0 {
				t.Fatalf("initial state: %#v %v", state, err)
			}
			draft := OnboardingInput{Persona: "other", PersonaOther: "Өзгеріс", Step: 2, Version: 0}
			state, err = service.SaveOnboarding(ctx, session.User.ID, draft)
			if err != nil || state.PersonaOther != "Өзгеріс" || state.Step != 2 || state.Version != 1 {
				t.Fatalf("draft save: %#v %v", state, err)
			}
			if _, err := service.SaveOnboarding(ctx, session.User.ID, draft); !errors.Is(err, ErrOnboardingConflict) {
				t.Fatalf("stale write accepted: %v", err)
			}
			second, secondToken, err := service.Login(ctx, LoginInput{Email: email, Password: "isolated-test-password"})
			if err != nil || !second.OnboardingRequired {
				t.Fatalf("resume on login: %v", err)
			}
			resumed, err := service.GetOnboarding(ctx, second.User.ID)
			if err != nil || resumed.Step != 2 || resumed.PersonaOther != "Өзгеріс" {
				t.Fatalf("lost draft: %#v %v", resumed, err)
			}
			complete := OnboardingInput{Persona: "other", PersonaOther: "Өзгеріс", Goal: "internship", CityID: "almaty", Profession: "Разработка", Step: 5, Version: state.Version, Complete: true}
			state, err = service.SaveOnboarding(ctx, session.User.ID, complete)
			if err != nil || state.Required || state.CompletedAt == nil || state.Referral != "skip" {
				t.Fatalf("completion: %#v %v", state, err)
			}
			complete.Version = state.Version
			if _, err := service.SaveOnboarding(ctx, session.User.ID, complete); !errors.Is(err, ErrOnboardingConflict) {
				t.Fatalf("completion was overwritten: %v", err)
			}
			for _, activeToken := range []string{token, secondToken} {
				me, err := service.Authenticate(ctx, activeToken)
				if err != nil || me.OnboardingRequired {
					t.Fatalf("completion not shared between devices: %v", err)
				}
			}
			third, _, err := service.Login(ctx, LoginInput{Email: email, Password: "isolated-test-password"})
			if err != nil || third.OnboardingRequired {
				t.Fatalf("completed onboarding reappeared: %v", err)
			}
			_, err = pool.Exec(ctx, `DELETE FROM jobhub.account_onboarding WHERE user_id=$1`, session.User.ID)
			if err != nil {
				t.Fatal(err)
			}
			legacy, err := service.Authenticate(ctx, token)
			if err != nil || legacy.OnboardingRequired {
				t.Fatal("legacy account was enrolled")
			}
			if _, err := service.SaveOnboarding(ctx, session.User.ID, OnboardingInput{Step: 1}); !errors.Is(err, ErrOnboardingConflict) {
				t.Fatal("legacy account unexpectedly enrolled by save")
			}
		})
	}
	state, err := service.GetOnboarding(ctx, "00000000-0000-0000-0000-000000000001")
	if err != nil || state.Required || state.Persona != "" {
		t.Fatal("answers leaked to another account")
	}
}
