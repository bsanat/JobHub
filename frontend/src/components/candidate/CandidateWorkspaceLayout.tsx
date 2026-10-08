import type { PropsWithChildren } from "react";
import { Bookmark, FileText, Send, Settings2, UserRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { NavLink, useLocation, useNavigate } from "react-router-dom";
import { UserAvatar } from "./UserAvatar";
import { SiteShell } from "../../layouts/SiteShell";
import { useAuth } from "../../app/authContext";

const destinations = [
  ["/profile", "workspace.nav.profile", UserRound],
  ["/resume", "workspace.nav.resume", FileText],
  ["/saved", "savedPage.title", Bookmark],
  ["/applications", "applications.title", Send],
  ["/account", "employer.account", Settings2],
] as const;

function primaryPath(pathname: string) {
  if (pathname === "/resume" || pathname === "/preferences") return "/resume";
  return pathname;
}

export function CandidateWorkspaceLayout({ children }: PropsWithChildren) {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const { session } = useAuth();
  const activePath = primaryPath(location.pathname);

  return (
    <SiteShell variant="workspace">
      <div className="workbench workbench--candidate page-container">
        <aside className="workbench__sidebar">
          <div className="workbench__identity">
            <UserAvatar fullName={session?.user.fullName ?? ""} />
            <div>
              <strong>{session?.user.fullName}</strong>
              <span>{t("auth.roles.job_seeker")}</span>
            </div>
          </div>
          <nav
            className="workbench-nav"
            aria-label={t("workspace.navigationLabel")}
          >
            {destinations.map(([path, key, Icon]) => (
              <NavLink
                key={path}
                to={path}
                end
                aria-current={activePath === path ? "page" : undefined}
              >
                <Icon size={18} aria-hidden="true" />
                <span>{t(key)}</span>
              </NavLink>
            ))}
          </nav>
        </aside>
        <label className="workbench__mobile-nav">
          <span>{t("workspace.navigationLabel")}</span>
          <select
            value={activePath}
            onChange={(event) => navigate(event.target.value)}
          >
            {destinations.map(([path, key]) => (
              <option key={path} value={path}>
                {t(key)}
              </option>
            ))}
          </select>
        </label>
        <div className="workbench__content">{children}</div>
      </div>
    </SiteShell>
  );
}
