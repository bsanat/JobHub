package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestOnboardingValidation(t *testing.T) {
	valid := OnboardingInput{Persona: "university_student", Goal: "internship", CityID: "almaty", Profession: "  Разработка  ", Step: 5, Complete: true}
	for _, tc := range []struct {
		name      string
		change    func(*OnboardingInput)
		wantError bool
	}{
		{"skip optional referral", func(*OnboardingInput) {}, false},
		{"custom answers", func(v *OnboardingInput) {
			v.Persona = "other"
			v.PersonaOther = "  Мансап ауыстыру  "
			v.Referral = "other"
			v.ReferralOther = "Конференция"
		}, false},
		{"first empty draft", func(v *OnboardingInput) { *v = OnboardingInput{Step: 1} }, false},
		{"partial other draft", func(v *OnboardingInput) { *v = OnboardingInput{Persona: "other", Step: 1} }, false},
		{"invalid persona", func(v *OnboardingInput) { v.Persona = "employer" }, true},
		{"invalid goal", func(v *OnboardingInput) { v.Goal = "anything" }, true},
		{"invalid city", func(v *OnboardingInput) { v.CityID = "Atlantis" }, true},
		{"invalid referral", func(v *OnboardingInput) { v.Referral = "spam" }, true},
		{"missing other", func(v *OnboardingInput) { v.Persona = "other" }, true},
		{"missing profession", func(v *OnboardingInput) { v.Profession = " " }, true},
		{"missing city", func(v *OnboardingInput) { v.CityID = "" }, true},
		{"missing goal", func(v *OnboardingInput) { v.Goal = "" }, true},
		{"missing referral other", func(v *OnboardingInput) { v.Referral = "other" }, true},
		{"200 unicode characters", func(v *OnboardingInput) { v.Profession = strings.Repeat("қ", 200) }, false},
		{"long text", func(v *OnboardingInput) { v.Profession = strings.Repeat("қ", 201) }, true},
		{"null byte", func(v *OnboardingInput) { v.Profession = "x\x00y" }, true},
		{"negative version", func(v *OnboardingInput) { v.Version = -1 }, true},
		{"step zero", func(v *OnboardingInput) { v.Step = 0 }, true},
		{"early completion", func(v *OnboardingInput) { v.Step = 2 }, true},
		{"cannot advance missing persona", func(v *OnboardingInput) { *v = OnboardingInput{Step: 2} }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			got, err := validateOnboarding(input)
			if tc.wantError != errors.Is(err, ErrOnboardingValidation) {
				t.Fatalf("validation error=%v, wantError=%v", err, tc.wantError)
			}
			if err == nil && got.Complete && input.Referral == "" && got.Referral != "skip" {
				t.Fatal("optional answer must become skip")
			}
		})
	}
}

func TestOnboardingNormalizesUnusedCustomAnswers(t *testing.T) {
	got, err := validateOnboarding(OnboardingInput{Persona: "graduate", PersonaOther: "old answer", Referral: "telegram", ReferralOther: "old source", Step: 1})
	if err != nil || got.PersonaOther != "" || got.ReferralOther != "" {
		t.Fatalf("unused custom answers were retained: %#v %v", got, err)
	}
}
