export type UserRole = "job_seeker" | "employer" | "admin";
export type UserStatus = "active" | "suspended";

export interface AuthUser {
  id: string;
  fullName: string;
  email: string;
  role: UserRole;
  status: UserStatus;
}

export interface AuthCompany {
  id: string;
  name: string;
}

export interface AuthSession {
  user: AuthUser;
  company?: AuthCompany;
  csrfToken: string;
  expiresAt: string;
  onboardingRequired: boolean;
}

export type OnboardingPersona =
  | ""
  | "school_student"
  | "university_student"
  | "graduate"
  | "unemployed"
  | "other";
export type OnboardingGoal = "" | "job" | "internship" | "browsing";
export type OnboardingReferral =
  | ""
  | "telegram"
  | "instagram"
  | "google"
  | "friend"
  | "university"
  | "other"
  | "skip";

export interface OnboardingAnswers {
  persona: OnboardingPersona;
  personaOther: string;
  goal: OnboardingGoal;
  cityId: string;
  profession: string;
  referral: OnboardingReferral;
  referralOther: string;
  step: number;
}

export interface OnboardingState extends OnboardingAnswers {
  required: boolean;
  version: number;
  completedAt: string | null;
}

export interface RegisterRequest {
  fullName: string;
  email: string;
  password: string;
  role: "job_seeker" | "employer";
  companyName?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}
