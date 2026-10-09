import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../app/authContext";
import { JobHubLoader } from "../components/ui/JobHubLoader";
import { SiteShell } from "../layouts/SiteShell";
import { onboardingDestination, onboardingPath } from "../lib/onboarding";
import { safeReturnTo } from "../lib/returnTo";
import { neonAuth } from "../lib/neonAuth";

export function AuthCallbackPage() {
  const { t } = useTranslation();
  const auth = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [error, setError] = useState("");
  const handledRef = useRef(false);
  const returnTo = safeReturnTo(params.get("returnTo"));

  useEffect(() => {
    if (!auth.session || !handledRef.current) return;
    const destination = onboardingDestination(returnTo, auth.session.user.role);
    navigate(auth.session.onboardingRequired ? onboardingPath(destination) : destination, { replace: true });
  }, [auth.session, navigate, returnTo]);

  useEffect(() => {
    let active = true;
    async function completeGoogleSignIn() {
      if (handledRef.current) return;
      handledRef.current = true;
      if (!neonAuth) {
        setError(t("auth.errors.googleNotConfigured"));
        return;
      }
      try {
        const { data, error: tokenError } = await neonAuth.token();
        if (tokenError || !data?.token) {
          if (active) setError(t("auth.errors.google"));
          return;
        }
        const session = await auth.loginWithGoogleToken(data.token);
        await neonAuth.signOut();
        const destination = onboardingDestination(returnTo, session.user.role);
        navigate(session.onboardingRequired ? onboardingPath(destination) : destination, { replace: true });
      } catch {
        if (active) setError(t("auth.errors.google"));
      }
    }
    void completeGoogleSignIn();
    return () => {
      active = false;
    };
  }, [auth, navigate, params, returnTo, t]);

  return (
    <SiteShell>
      <main className="auth-page page-container">
        <section className="auth-panel" aria-labelledby="auth-callback-title">
          <p className="auth-eyebrow">{t("auth.googleCallbackEyebrow")}</p>
          <h1 id="auth-callback-title">{t("auth.googleCallbackTitle")}</h1>
          {error ? <p className="auth-error" role="alert">{error}</p> : <JobHubLoader />}
        </section>
      </main>
    </SiteShell>
  );
}
