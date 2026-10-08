import { useQuery } from "@tanstack/react-query";
import { ArrowRight, BriefcaseBusiness, GraduationCap, MapPin } from "lucide-react";
import { useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import type { JobSearchParams } from "../api/models/job";
import { repositories } from "../api/repositories";
import { MarketplaceJobRow } from "../components/jobs/MarketplaceJobRow";
import { CitySelect } from "../components/location/CitySelect";
import { ErrorState, JobListSkeleton } from "../components/ui/Feedback";
import { SiteShell } from "../layouts/SiteShell";
import { serializeJobSearchParams } from "../lib/jobSearchParams";
import { usePreferredSearchCity } from "../lib/searchCityPreference";
import "../styles/students.css";

const sections = [
  { key: "internships", filter: { employment: "internship" }, icon: GraduationCap },
  { key: "firstJobs", filter: { employment: "full_time", experience: "no_experience" }, icon: BriefcaseBusiness },
  { key: "partTime", filter: { employment: "part_time" }, icon: BriefcaseBusiness },
] as const;

export function StudentsPage() {
  const { t } = useTranslation();
  const preferredCity = usePreferredSearchCity();
  const [city, setCity] = useState(preferredCity ?? "");

  return (
    <SiteShell>
      <div className="page-container students-page">
        <header className="students-page__intro">
          <div>
            <h1>{t("students.title")}</h1>
            <p>{t("students.description")}</p>
          </div>
          <CitySelect label={t("students.cityLabel")} anyLabel={t("students.allCities")} value={city} onChange={setCity} />
        </header>
        <div className="students-page__sections">
          {sections.map(({ key, filter, icon: Icon }) => (
            <StudentJobsSection key={key} sectionKey={key} filter={filter} city={city} onClearCity={() => setCity("")} icon={<Icon size={22} aria-hidden="true" />} />
          ))}
        </div>
      </div>
    </SiteShell>
  );
}

function StudentJobsSection({ sectionKey, filter, city, icon, onClearCity }: {
  sectionKey: "internships" | "firstJobs" | "partTime";
  filter: Pick<JobSearchParams, "employment" | "experience">;
  city: string;
  icon: ReactNode;
  onClearCity: () => void;
}) {
  const { t } = useTranslation();
  const search: JobSearchParams = { ...filter, city: city || undefined, page: 1, pageSize: 6, sort: "newest" };
  const jobs = useQuery({
    queryKey: ["jobs", "students", sectionKey, city],
    queryFn: () => repositories.jobs.search(search),
    staleTime: 2 * 60_000,
  });
  const allJobsUrl = `/jobs?${serializeJobSearchParams(search).toString()}`;
  const titleId = `students-${sectionKey}`;

  return (
    <section className="students-section" aria-labelledby={titleId}>
      <div className="students-section__heading">
        <span className="students-section__icon">{icon}</span>
        <div>
          <h2 id={titleId}>{t(`students.sections.${sectionKey}.title`)}</h2>
          <p>{t(`students.sections.${sectionKey}.description`)}</p>
        </div>
        {jobs.data && jobs.data.total > 0 ? <Link to={allJobsUrl}>{t("students.viewAll")}<ArrowRight size={17} aria-hidden="true" /></Link> : null}
      </div>
      {jobs.isPending ? <JobListSkeleton label={t("jobs.loading")} /> : null}
      {jobs.isError ? <ErrorState title={t("jobs.errorTitle")} description={t("jobs.errorDescription")} actionLabel={t("common.retry")} onAction={() => void jobs.refetch()} /> : null}
      {jobs.data?.items.length ? (
        <div className="job-list">{jobs.data.items.map((job) => <MarketplaceJobRow key={job.id} job={job} />)}</div>
      ) : null}
      {jobs.data?.items.length === 0 ? (
        <div className="students-section__empty">
          <MapPin size={20} aria-hidden="true" />
          <p>{t("students.empty")}</p>
          {city ? <button type="button" onClick={onClearCity}>{t("students.allCities")}</button> : <Link to="/jobs">{t("students.browseAll")}</Link>}
        </div>
      ) : null}
    </section>
  );
}
