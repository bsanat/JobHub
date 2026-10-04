import type {
  UserRole,
  UserStatus,
  OnboardingPersona,
  OnboardingGoal,
  OnboardingReferral,
} from "../../models/auth";

export interface AuthSessionDto {
  user: {
    id: string;
    full_name: string;
    email: string;
    role: UserRole;
    status: UserStatus;
  };
  company?: {
    id: string;
    name: string;
  };
  csrf_token: string;
  expires_at: string;
  onboarding_required: boolean;
}

export interface OnboardingDto {
  required: boolean;
  persona: OnboardingPersona;
  persona_other: string;
  goal: OnboardingGoal;
  city_id: string;
  profession: string;
  referral: OnboardingReferral;
  referral_other: string;
  step: number;
  version: number;
  completed_at: string | null;
}
