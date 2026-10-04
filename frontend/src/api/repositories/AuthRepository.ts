import type {
  AuthSession,
  LoginRequest,
  RegisterRequest,
  OnboardingAnswers,
  OnboardingState,
} from "../models/auth";

export interface AuthRepository {
  register(input: RegisterRequest): Promise<AuthSession>;
  login(input: LoginRequest): Promise<AuthSession>;
  loginWithGoogleToken(accessToken: string): Promise<AuthSession>;
  me(): Promise<AuthSession>;
  logout(csrfToken: string): Promise<void>;
  onboarding(): Promise<OnboardingState>;
  saveOnboarding(
    input: OnboardingAnswers,
    version: number,
    complete: boolean,
    csrfToken: string,
  ): Promise<OnboardingState>;
}
