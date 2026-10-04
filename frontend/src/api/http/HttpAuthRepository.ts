import { apiClient } from "../client";
import type {
  AuthSession,
  LoginRequest,
  RegisterRequest,
  OnboardingAnswers,
  OnboardingState,
} from "../models/auth";
import type { AuthRepository } from "../repositories/AuthRepository";
import { mapAuthSession, mapOnboarding } from "./authMappers";
import type { AuthSessionDto, OnboardingDto } from "./dto/authDto";

export class HttpAuthRepository implements AuthRepository {
  async register(input: RegisterRequest): Promise<AuthSession> {
    const { data } = await apiClient.post<AuthSessionDto>("/auth/register", {
      full_name: input.fullName,
      email: input.email,
      password: input.password,
      role: input.role,
      company_name: input.companyName,
    });
    return mapAuthSession(data);
  }

  async login(input: LoginRequest): Promise<AuthSession> {
    const { data } = await apiClient.post<AuthSessionDto>("/auth/login", input);
    return mapAuthSession(data);
  }

  async loginWithGoogleToken(accessToken: string): Promise<AuthSession> {
    const { data } = await apiClient.post<AuthSessionDto>("/auth/oauth/google", {
      access_token: accessToken,
    }, {
      timeout: 15_000,
    });
    return mapAuthSession(data);
  }

  async me(): Promise<AuthSession> {
    const { data } = await apiClient.get<AuthSessionDto>("/auth/me");
    return mapAuthSession(data);
  }

  async logout(csrfToken: string): Promise<void> {
    await apiClient.post("/auth/logout", undefined, {
      headers: { "X-CSRF-Token": csrfToken },
    });
  }

  async onboarding(): Promise<OnboardingState> {
    const { data } = await apiClient.get<OnboardingDto>("/auth/onboarding");
    return mapOnboarding(data);
  }

  async saveOnboarding(
    input: OnboardingAnswers,
    version: number,
    complete: boolean,
    csrfToken: string,
  ): Promise<OnboardingState> {
    const { data } = await apiClient.put<OnboardingDto>(
      "/auth/onboarding",
      {
        persona: input.persona,
        persona_other: input.personaOther,
        goal: input.goal,
        city_id: input.cityId,
        profession: input.profession,
        referral: input.referral,
        referral_other: input.referralOther,
        step: input.step,
        version,
        complete,
      },
      { headers: { "X-CSRF-Token": csrfToken } },
    );
    return mapOnboarding(data);
  }
}
