import { LogOut, UserRound } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "../app/authContext";
import { CandidateWorkspaceLayout } from "../components/candidate/CandidateWorkspaceLayout";
import { EmployerWorkspaceLayout } from "../components/employer/EmployerWorkspaceLayout";
import { Button } from "../components/ui/Button";
import { SiteShell } from "../layouts/SiteShell";

export function AccountPage() {
  const { t } = useTranslation();
  const auth = useAuth();
  const navigate = useNavigate();
  const [logoutPending, setLogoutPending] = useState(false);
  const [logoutError, setLogoutError] = useState("");
  const user = auth.session?.user;

  if (!user) return <Navigate to="/" replace />;

  const logout = async () => {
    if (logoutPending) return;
    setLogoutError("");
    setLogoutPending(true);
    try {
      await auth.logout();
      navigate("/", { replace: true });
    } catch {
      setLogoutError(t("auth.logoutError"));
      setLogoutPending(false);
    }
  };
  const content = (
    <section className="account-settings" aria-labelledby="account-title">
      <header className="workbench-page-header">
        <div>
          <h1 id="account-title">{t("auth.accountSettings")}</h1>
        </div>
      </header>
      <div className="account-settings__identity">
        <span className="account-settings__avatar" aria-hidden="true">
          <UserRound size={28} />
        </span>
        <div>
          <h2>{user.fullName}</h2>
          <span>{t(`auth.roles.${user.role}`)}</span>
        </div>
      </div>
      <dl className="account-settings__details">
        <div>
          <dt>{t("auth.fullName")}</dt>
          <dd>{user.fullName}</dd>
        </div>
        <div>
          <dt>{t("auth.email")}</dt>
          <dd>{user.email}</dd>
        </div>
        <div>
          <dt>{t("auth.role")}</dt>
          <dd>{t(`auth.roles.${user.role}`)}</dd>
        </div>
        <div>
          <dt>{t("auth.accountStatus")}</dt>
          <dd>
            <span className={`account-status account-status--${user.status}`}>
              {t(`auth.statuses.${user.status}`)}
            </span>
          </dd>
        </div>
        {auth.session?.company ? (
          <div>
            <dt>{t("employer.company")}</dt>
            <dd>{auth.session.company.name}</dd>
          </div>
        ) : null}
      </dl>
      <section
        className="account-settings__session"
        aria-labelledby="account-session-title"
      >
        <h2 id="account-session-title">{t("auth.sessionTitle")}</h2>
        <Button
          variant="secondary"
          leadingIcon={<LogOut size={17} />}
          isLoading={logoutPending}
          onClick={() => void logout()}
        >
          {t("auth.logout")}
        </Button>
        {logoutError ? (
          <p className="auth-error account-logout-error" role="alert">
            {logoutError}
          </p>
        ) : null}
      </section>
    </section>
  );
  if (user.role === "employer")
    return <EmployerWorkspaceLayout>{content}</EmployerWorkspaceLayout>;
  if (user.role === "job_seeker")
    return <CandidateWorkspaceLayout>{content}</CandidateWorkspaceLayout>;
  return (
    <SiteShell>
      <div className="account-page page-container">{content}</div>
    </SiteShell>
  );
}
