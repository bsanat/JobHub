import type { AuthSession, OnboardingState } from "../models/auth";
import type { AuthSessionDto, OnboardingDto } from "./dto/authDto";

export function mapAuthSession(dto: AuthSessionDto): AuthSession {
  return {
    user: {
      id: dto.user.id,
      fullName: dto.user.full_name,
      email: dto.user.email,
      role: dto.user.role,
      status: dto.user.status,
    },
    company: dto.company,
    csrfToken: dto.csrf_token,
    expiresAt: dto.expires_at,
    onboardingRequired: dto.onboarding_required === true,
  };
}

export function mapOnboarding(dto: OnboardingDto): OnboardingState {
  return {
    required: dto.required,
    persona: dto.persona,
    personaOther: dto.persona_other,
    goal: dto.goal,
    cityId: dto.city_id,
    profession: dto.profession,
    referral: dto.referral,
    referralOther: dto.referral_other,
    step: dto.step,
    version: dto.version,
    completedAt: dto.completed_at,
  };
}
