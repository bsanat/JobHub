package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"jobhub-ai/backend/internal/locations"
)

var (
	ErrOnboardingValidation = errors.New("invalid onboarding answers")
	ErrOnboardingConflict   = errors.New("onboarding changed or already completed")
)

type Onboarding struct {
	Required      bool       `json:"required"`
	Persona       string     `json:"persona"`
	PersonaOther  string     `json:"persona_other"`
	Goal          string     `json:"goal"`
	CityID        string     `json:"city_id"`
	Profession    string     `json:"profession"`
	Referral      string     `json:"referral"`
	ReferralOther string     `json:"referral_other"`
	Step          int        `json:"step"`
	Version       int64      `json:"version"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type OnboardingInput struct {
	Persona       string `json:"persona"`
	PersonaOther  string `json:"persona_other"`
	Goal          string `json:"goal"`
	CityID        string `json:"city_id"`
	Profession    string `json:"profession"`
	Referral      string `json:"referral"`
	ReferralOther string `json:"referral_other"`
	Step          int    `json:"step"`
	Version       int64  `json:"version"`
	Complete      bool   `json:"complete"`
}

type OnboardingService interface {
	GetOnboarding(context.Context, string) (Onboarding, error)
	SaveOnboarding(context.Context, string, OnboardingInput) (Onboarding, error)
}

const onboardingColumns = `persona, persona_other, goal, city_id, profession, referral, referral_other, step, version, completed_at`

type onboardingQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func scanOnboarding(row pgx.Row) (Onboarding, error) {
	state := Onboarding{Step: 1}
	err := row.Scan(&state.Persona, &state.PersonaOther, &state.Goal, &state.CityID, &state.Profession, &state.Referral, &state.ReferralOther, &state.Step, &state.Version, &state.CompletedAt)
	state.Required = err == nil && state.CompletedAt == nil
	return state, err
}

func (s *AuthService) GetOnboarding(ctx context.Context, userID string) (Onboarding, error) {
	state, err := scanOnboarding(s.pool.QueryRow(ctx, `SELECT `+onboardingColumns+` FROM jobhub.account_onboarding WHERE user_id = $1`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Onboarding{Step: 1}, nil
	}
	if isUndefinedTable(err) {
		return Onboarding{Step: 1}, nil
	}
	return state, err
}

func (s *AuthService) SaveOnboarding(ctx context.Context, userID string, input OnboardingInput) (Onboarding, error) {
	input, err := validateOnboarding(input)
	if err != nil {
		return Onboarding{}, err
	}
	// Compare-and-swap prevents stale tabs/devices from overwriting a newer draft or completion.
	state, err := scanOnboarding(s.pool.QueryRow(ctx, `UPDATE jobhub.account_onboarding
		SET persona=$2, persona_other=$3, goal=$4, city_id=$5, profession=$6,
		    referral=$7, referral_other=$8, step=$9, version=version+1, updated_at=now(),
		    completed_at=CASE WHEN $11 THEN now() ELSE NULL END
		WHERE user_id=$1 AND version=$10 AND completed_at IS NULL
		RETURNING `+onboardingColumns, userID, input.Persona, input.PersonaOther, input.Goal, input.CityID, input.Profession, input.Referral, input.ReferralOther, input.Step, input.Version, input.Complete))
	if errors.Is(err, pgx.ErrNoRows) {
		return Onboarding{}, ErrOnboardingConflict
	}
	return state, err
}

func onboardingTableExists(ctx context.Context, q onboardingQuerier) (bool, error) {
	var exists bool
	if err := q.QueryRow(ctx, `SELECT to_regclass('jobhub.account_onboarding') IS NOT NULL`).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func validateOnboarding(input OnboardingInput) (OnboardingInput, error) {
	input.PersonaOther = strings.TrimSpace(input.PersonaOther)
	input.ReferralOther = strings.TrimSpace(input.ReferralOther)
	input.Profession = strings.TrimSpace(input.Profession)
	invalid := func(field string) (OnboardingInput, error) {
		return OnboardingInput{}, fmt.Errorf("%w: %s", ErrOnboardingValidation, field)
	}
	if input.Step < 1 || input.Step > 5 || input.Version < 0 || (input.Complete && input.Step != 5) {
		return invalid("step")
	}
	if !oneOf(input.Persona, "", "school_student", "university_student", "graduate", "unemployed", "other") {
		return invalid("persona")
	}
	if !oneOf(input.Goal, "", "job", "internship", "browsing") {
		return invalid("goal")
	}
	if !oneOf(input.Referral, "", "telegram", "instagram", "google", "friend", "university", "other", "skip") {
		return invalid("referral")
	}
	if input.CityID != "" && !locations.Valid(input.CityID) {
		return invalid("city_id")
	}
	for _, value := range []string{input.PersonaOther, input.ReferralOther, input.Profession} {
		if !utf8.ValidString(value) || utf8.RuneCountInString(value) > 200 || strings.ContainsRune(value, 0) {
			return invalid("text")
		}
	}
	if input.Persona != "other" {
		input.PersonaOther = ""
	}
	if input.Referral != "other" {
		input.ReferralOther = ""
	}
	if input.Step > 1 && (input.Persona == "" || input.Persona == "other" && input.PersonaOther == "") {
		return invalid("persona")
	}
	if input.Step > 2 && input.Goal == "" {
		return invalid("goal")
	}
	if input.Step > 3 && input.CityID == "" {
		return invalid("city_id")
	}
	if input.Step > 4 && input.Profession == "" {
		return invalid("profession")
	}
	if input.Complete && input.Referral == "other" && input.ReferralOther == "" {
		return invalid("referral_other")
	}
	if input.Complete && input.Referral == "" {
		input.Referral = "skip"
	}
	return input, nil
}

func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
