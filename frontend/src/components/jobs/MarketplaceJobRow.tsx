import { Building2, ExternalLink, House, MapPin } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useLocation } from "react-router-dom";
import type { JobSummary } from "../../api/models/job";
import { JobSaveButton } from "./JobSaveButton";
import { formatSalary } from "../../lib/formatSalary";
import { displayJobLocation } from "../../lib/jobLocation";
import { companyPresentationFor } from "../../lib/companyPresentation";
import { CompanyLogo } from "./CompanyLogo";

export function MarketplaceJobRow({ job }: { job: JobSummary }) {
  const { t, i18n } = useTranslation();
  const location = useLocation();
  const external = job.source.type === "external";
  const company = companyPresentationFor(job);
  const salary = job.salaryRaw || (job.salary ? formatSalary(job.salary, i18n.language, t(`employer.salaryPeriods.${job.salary.period}`)) : "");
  const detailTarget = { pathname: `/jobs/${job.id}` };
  const detailState = { from: `${location.pathname}${location.search}` };
  const tags = [...new Set(job.tags.filter(Boolean))];

  return (
    <article className="job-row job-row--offer">
      <div className="job-offer__identity">
        <CompanyLogo job={job} />
        <Link className="job-row__title-link" to={detailTarget} state={detailState}><h3>{job.title}</h3></Link>
      </div>
      {salary ? <strong className="job-offer__salary">{salary}</strong> : null}
      <div className="job-offer__tags">
        {tags.slice(0, 4).map((tag) => <span key={tag}>{tag}</span>)}
        {tags.length > 4 ? <span className="job-offer__extra" title={tags.slice(4).join(", ")}>+{tags.length - 4}</span> : null}
        {external ? <span className="job-offer__source"><ExternalLink size={12} aria-hidden="true" />{t("jobs.externalSource", { source: job.source.name })}</span> : null}
      </div>
      <div className="job-offer__meta">
        {job.workMode ? <span className="job-offer__mode"><House size={15} aria-hidden="true" />{t(`employer.workModes.${job.workMode}`)}</span> : null}
        {job.company.id ? <Link className="job-offer__company" to={`/companies/${job.company.id}`} aria-label={`${t("companies.openProfile")}: ${company.name || t("common.notProvided")}`}><Building2 size={15} aria-hidden="true" />{company.name || t("common.notProvided")}</Link> : <span><Building2 size={15} aria-hidden="true" />{company.name || t("common.notProvided")}</span>}
        <span><MapPin size={15} aria-hidden="true" />{displayJobLocation(job, i18n.language) || t("common.notProvided")}</span>
        <JobSaveButton job={job} compact icon="heart" />
      </div>
    </article>
  );
}
