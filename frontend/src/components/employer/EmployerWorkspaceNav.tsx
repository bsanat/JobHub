import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";
import { employerDestinations } from "./employerNavigation";

export function EmployerWorkspaceNav() {
  const { t } = useTranslation();
  return (
    <nav className="workbench-nav" aria-label={t("employer.navLabel")}>
      {employerDestinations.map(([path, key, Icon]) => (
        <NavLink key={path} to={path} end={path === "/employer"}>
          <Icon size={18} aria-hidden="true" />
          <span>{t(`employer.${key}`)}</span>
        </NavLink>
      ))}
    </nav>
  );
}
