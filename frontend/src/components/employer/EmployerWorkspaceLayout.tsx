import { Building2, ExternalLink } from "lucide-react";
import type { PropsWithChildren } from "react";
import { useTranslation } from "react-i18next";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../app/authContext";
import { SiteShell } from "../../layouts/SiteShell";
import { EmployerWorkspaceNav } from "./EmployerWorkspaceNav";
import { employerDestinations } from "./employerNavigation";

export function EmployerWorkspaceLayout({ children }: PropsWithChildren) {
  const { t } = useTranslation();
  const { session } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const company = session?.company;
  const activePath = location.pathname.startsWith("/employer/vacancies")
    ? "/employer/vacancies"
    : location.pathname;

  return (
    <SiteShell variant="workspace">
      <div className="workbench page-container">
        <aside className="workbench__sidebar">
          <div className="workbench__identity">
            <span className="workbench__avatar" aria-hidden="true">
              <Building2 size={22} />
            </span>
            <div>
              <strong>{company?.name || t("common.notProvided")}</strong>
              <span>{t("auth.roles.employer")}</span>
            </div>
          </div>
          <EmployerWorkspaceNav />
          {company ? (
            <Link
              className="workbench__public-link"
              to={`/companies/${company.id}`}
            >
              <ExternalLink size={16} aria-hidden="true" />
              {t("employer.publicCompany")}
            </Link>
          ) : null}
        </aside>
        <label className="workbench__mobile-nav">
          <span>{t("employer.navLabel")}</span>
          <select
            value={activePath}
            onChange={(event) => navigate(event.target.value)}
          >
            {employerDestinations.map(([path, key]) => (
              <option key={path} value={path}>
                {t(`employer.${key}`)}
              </option>
            ))}
          </select>
        </label>
        <div className="workbench__content">{children}</div>
      </div>
    </SiteShell>
  );
}
