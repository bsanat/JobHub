import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../app/authContext";
import { JobHubLoader } from "../components/ui/JobHubLoader";
import { SiteShell } from "../layouts/SiteShell";
import { onboardingDestination, onboardingPath } from "../lib/onboarding";
import { safeReturnTo } from "../lib/returnTo";
import { supabase } from "../lib/supabase";

export function AuthCallbackPage() {
  const { t } = useTranslation();
  const auth = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [error, setError] = useState("");
  const handledRef = useRef(false);

  useEffect(() => {
    let active = true;
    async function completeGoogleSignIn() {
      if (handledRef.current) return;
      handledRef.current = true;
      if (!supabase) {
        setError(t("auth.errors.googleNotConfigured"));
        return;
      }
      const { data, error: sessionError } = await supabase.auth.getSession();
      if (sessionError || !data.session?.access_token) {
        if (active) setError(t("auth.errors.google"));
        return;
      }
      try {
        const session = await auth.loginWithGoogleToken(data.session.access_token);
        await supabase.auth.signOut();
        const destination = onboardingDestination(safeReturnTo(params.get("returnTo")), session.user.role);
        if (active) navigate(session.onboardingRequired ? onboardingPath(destination) : destination, { replace: true });
      } catch {
        if (active) setError(t("auth.errors.google"));
      }
    }
    void completeGoogleSignIn();
    return () => {
      active = false;
    };
  }, [auth, navigate, params, t]);

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
