import { useTranslation } from "react-i18next";
import type { ReactNode } from "react";
import type { CandidateProfile } from "../../api/models/candidate";
import { cityName } from "../../lib/cities";

interface CandidateIdentityHeaderProps {
  profile: CandidateProfile;
  titleId?: string;
  actions?: ReactNode;
}

export function CandidateIdentityHeader({
  profile,
  titleId,
  actions,
}: CandidateIdentityHeaderProps) {
  const { t, i18n } = useTranslation();
  const city = cityName(profile.cityId, i18n.language) || profile.city;
  const position =
    profile.desiredPosition ||
    profile.currentPosition ||
    t("profile.positionMissing");

  return (
    <header className="candidate-identity-header">
      <div className="profile-avatar" aria-hidden="true">
        {initials(profile.fullName)}
      </div>
      <div className="candidate-identity-header__copy">
        <h1 id={titleId}>{profile.fullName}</h1>
        <p className="candidate-identity-header__position">{position}</p>
        <div className="candidate-identity-header__meta">
          {city ? <span>{city}</span> : null}
          <span
            className={`search-status search-status--${profile.searchStatus}`}
          >
            {t(`profile.searchStatuses.${profile.searchStatus}`)}
          </span>
        </div>
      </div>
      <div className="candidate-identity-header__tools">
        {actions}
        <div
          className="profile-completion"
          aria-label={t("profile.completionLabel", {
            value: profile.completion.percentage,
          })}
        >
          <strong>{profile.completion.percentage}%</strong>
          <span>{t("profile.complete")}</span>
          <div aria-hidden="true">
            <span style={{ width: `${profile.completion.percentage}%` }} />
          </div>
        </div>
      </div>
    </header>
  );
}

function initials(value: string) {
  return value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toLocaleUpperCase())
    .join("");
}
