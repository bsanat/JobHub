import { zodResolver } from "@hookform/resolvers/zod";
import { BriefcaseBusiness, Eye, EyeOff, UserRound } from "lucide-react";
import { useState, type FormEvent } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../app/authContext";
import type { UserRole } from "../api/models/auth";
import { GoogleMark } from "../components/auth/GoogleMark";
import { Button, IconButton } from "../components/ui/Button";
import { Input } from "../components/ui/FormControls";
import { SiteShell } from "../layouts/SiteShell";
import { safeReturnTo } from "../lib/returnTo";
import { googleAuthConfigured, googleRedirectTo, neonAuth, neonAuthConfigured } from "../lib/neonAuth";
import { authErrorCode, emailCodeSendErrorKey } from "./authUtils";
import { onboardingDestination, onboardingPath } from "../lib/onboarding";

const schema = z.object({
  fullName: z.string().trim().min(1),
  email: z.email(),
  password: z.string().min(10),
  confirmPassword: z.string().min(1),
  companyName: z.string().optional(),
}).refine((values) => values.password === values.confirmPassword, { path: ["confirmPassword"] });

type RegisterValues = z.infer<typeof schema>;
const verificationCodePattern = /^\d{6}$/;

export function RegisterPage() {
  const { t } = useTranslation();
  const auth = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [role, setRole] = useState<Extract<UserRole, "job_seeker" | "employer">>("job_seeker");
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState("");
  const [pendingRegistration, setPendingRegistration] = useState<RegisterValues | null>(null);
  const [verificationCode, setVerificationCode] = useState("");
  const [isVerifying, setIsVerifying] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [verificationNotice, setVerificationNotice] = useState("");
  const { register, handleSubmit, formState } = useForm<RegisterValues>({
    resolver: zodResolver(schema),
    defaultValues: { fullName: "", email: "", password: "", confirmPassword: "", companyName: "" },
  });

  const onSubmit = handleSubmit(async (values) => {
    setServerError("");
    if (role === "employer" && !values.companyName?.trim()) {
      setServerError(t("auth.validation.companyRequired"));
      return;
    }
    if (!neonAuthConfigured || !neonAuth) {
      setServerError(t("auth.errors.emailVerificationUnavailable"));
      return;
    }
    try {
      const { error } = await neonAuth.emailOtp.sendVerificationOtp({
        email: values.email,
        type: "sign-in",
      });
      if (error) {
        setServerError(t(emailCodeSendErrorKey(error)));
        return;
      }
      setPendingRegistration(values);
      setVerificationCode("");
      setVerificationNotice("");
    } catch {
      setServerError(t("auth.errors.codeSend"));
    }
  });

  const finishRegistration = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setServerError("");
    setVerificationNotice("");
    if (!pendingRegistration || !neonAuth || !verificationCodePattern.test(verificationCode)) {
      setServerError(t("auth.validation.verificationCode"));
      return;
    }
    setIsVerifying(true);
    try {
      const { error } = await neonAuth.signIn.emailOtp({
        email: pendingRegistration.email,
        otp: verificationCode,
      });
      if (error) {
        setServerError(t("auth.errors.codeInvalid"));
        return;
      }
      const { data: token, error: tokenError } = await neonAuth.token();
      if (tokenError || !token?.token) throw new Error("Email verification token unavailable");
      const session = await auth.register({
        fullName: pendingRegistration.fullName,
        email: pendingRegistration.email,
        password: pendingRegistration.password,
        role,
        companyName: pendingRegistration.companyName,
        emailVerificationToken: token.token,
      });
      const destination = onboardingDestination(role === "employer" ? "/employer" : params.get("returnTo"), role);
      navigate(session.onboardingRequired ? onboardingPath(destination) : destination, { replace: true });
    } catch (error) {
      const code = authErrorCode(error);
      if (code === "EMAIL_ALREADY_EXISTS") setServerError(t("auth.errors.emailExists"));
      else if (code === "EMAIL_VERIFICATION_INVALID") setServerError(t("auth.errors.codeInvalid"));
      else if (code === "EMAIL_VERIFICATION_UNAVAILABLE") setServerError(t("auth.errors.emailVerificationUnavailable"));
      else setServerError(t("auth.errors.generic"));
    } finally {
      setIsVerifying(false);
      await neonAuth.signOut();
    }
  };

  const resendCode = async () => {
    if (!pendingRegistration || !neonAuth) return;
    setServerError("");
    setVerificationNotice("");
    setIsResending(true);
    try {
      const { error } = await neonAuth.emailOtp.sendVerificationOtp({
        email: pendingRegistration.email,
        type: "sign-in",
      });
      if (error) setServerError(t(emailCodeSendErrorKey(error)));
      else setVerificationNotice(t("auth.codeResent"));
    } catch {
      setServerError(t("auth.errors.codeSend"));
    } finally {
      setIsResending(false);
    }
  };

  const startGoogleRegistration = async () => {
    setServerError("");
    if (!googleAuthConfigured || !neonAuth) {
      setServerError(t("auth.errors.googleNotConfigured"));
      return;
    }
    const returnTo = safeReturnTo(role === "employer" ? "/employer" : params.get("returnTo"));
    try {
      const { error } = await neonAuth.signIn.social({
        provider: "google",
        callbackURL: googleRedirectTo(returnTo),
      });
      if (error) setServerError(t("auth.errors.google"));
    } catch {
      setServerError(t("auth.errors.google"));
    }
  };

  return (
    <SiteShell>
      <main className="auth-page page-container">
        <section className="auth-panel auth-panel--wide" aria-labelledby="register-title">
          <p className="auth-eyebrow">{t("auth.registerEyebrow")}</p>
          <h1 id="register-title">{t("auth.registerTitle")}</h1>
          {pendingRegistration ? (
            <form className="auth-form" onSubmit={finishRegistration} noValidate>
              <p className="auth-verification-copy">{t("auth.verificationDescription", { email: pendingRegistration.email })}</p>
              <Input
                label={t("auth.verificationCode")}
                value={verificationCode}
                onChange={(event) => setVerificationCode(event.target.value.replace(/\D/g, "").slice(0, 6))}
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                error={serverError === t("auth.validation.verificationCode") ? serverError : undefined}
              />
              {serverError && serverError !== t("auth.validation.verificationCode") ? <p className="auth-error" role="alert">{serverError}</p> : null}
              {verificationNotice ? <p className="auth-success" role="status">{verificationNotice}</p> : null}
              <Button type="submit" isLoading={isVerifying}>{t("auth.verifyAndRegister")}</Button>
              <div className="auth-verification-actions">
                <Button type="button" variant="secondary" isLoading={isResending} onClick={resendCode}>{t("auth.resendCode")}</Button>
                <Button type="button" variant="quiet" onClick={() => { setPendingRegistration(null); setVerificationCode(""); setServerError(""); setVerificationNotice(""); }}>{t("auth.changeEmail")}</Button>
              </div>
            </form>
          ) : <>
          <div className="role-picker" role="radiogroup" aria-label={t("auth.chooseRole")}>
            <button type="button" className={role === "job_seeker" ? "active" : ""} onClick={() => setRole("job_seeker")}><UserRound size={20} /><span>{t("auth.roles.job_seeker")}</span></button>
            <button type="button" className={role === "employer" ? "active" : ""} onClick={() => setRole("employer")}><BriefcaseBusiness size={20} /><span>{t("auth.roles.employer")}</span></button>
          </div>
          <button type="button" className="auth-google-button" onClick={startGoogleRegistration}>
            <GoogleMark />
            {t("auth.googleRegister")}
          </button>
          <p className="auth-oauth-note">{t("auth.googleRegisterNote")}</p>
          <div className="auth-divider"><span>{t("auth.orEmail")}</span></div>
          <form className="auth-form" onSubmit={onSubmit} noValidate>
            <Input label={t("auth.fullName")} autoComplete="name" error={formState.errors.fullName ? t("auth.validation.fullName") : undefined} {...register("fullName")} />
            <Input label={role === "employer" ? t("auth.workEmail") : t("auth.email")} type="email" autoComplete="email" error={formState.errors.email ? t("auth.validation.email") : undefined} {...register("email")} />
            {role === "employer" ? <Input label={t("auth.companyName")} autoComplete="organization" error={serverError === t("auth.validation.companyRequired") ? serverError : undefined} {...register("companyName")} /> : null}
            <div className="password-field">
              <Input label={t("auth.password")} type={showPassword ? "text" : "password"} autoComplete="new-password" error={formState.errors.password ? t("auth.validation.passwordPolicy") : undefined} {...register("password")} />
              <IconButton type="button" label={showPassword ? t("auth.hidePassword") : t("auth.showPassword")} icon={showPassword ? <EyeOff size={18} /> : <Eye size={18} />} variant="quiet" onClick={() => setShowPassword((value) => !value)} />
            </div>
            <Input label={t("auth.confirmPassword")} type={showPassword ? "text" : "password"} autoComplete="new-password" error={formState.errors.confirmPassword ? t("auth.validation.passwordMatch") : undefined} {...register("confirmPassword")} />
            {serverError && serverError !== t("auth.validation.companyRequired") ? <p className="auth-error" role="alert">{serverError}</p> : null}
            <Button type="submit" isLoading={formState.isSubmitting}>{t("auth.registerSubmit")}</Button>
          </form>
          <p className="auth-switch">{t("auth.hasAccount")} <Link to={params.has("returnTo") ? `/login?returnTo=${encodeURIComponent(safeReturnTo(params.get("returnTo")))}` : "/login"}>{t("auth.loginLink")}</Link></p>
          </>}
        </section>
      </main>
    </SiteShell>
  );
}
