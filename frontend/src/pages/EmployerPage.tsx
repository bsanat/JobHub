import { useQuery } from "@tanstack/react-query";
import { ArrowRight, Building2, Plus } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { repositories } from "../api/repositories";
import { useAuth } from "../app/authContext";
import { EmployerJobTable } from "../components/employer/EmployerJobTable";
import { jobUpdatedTime, publicationStatuses } from "../lib/employerVacancies";
import { EmployerWorkspaceLayout } from "../components/employer/EmployerWorkspaceLayout";
import {
  EmptyState,
  ErrorState,
  JobListSkeleton,
} from "../components/ui/Feedback";

export function EmployerPage() {
  const { t } = useTranslation();
  const { session } = useAuth();
  const jobs = useQuery({
    queryKey: ["employer", "jobs", "list", session!.user.id],
    queryFn: () => repositories.employerJobs.list(),
  });
  const recentJobs = jobs.data
    ? [...jobs.data]
        .sort((a, b) => jobUpdatedTime(b) - jobUpdatedTime(a))
        .slice(0, 5)
    : [];
  return (
    <EmployerWorkspaceLayout>
      <section
        className="employer-overview"
        aria-labelledby="employer-overview-title"
      >
        <header className="workbench-page-header">
          <div>
            <h1 id="employer-overview-title">{t("employer.overview")}</h1>
            <p>{session!.company?.name}</p>
          </div>
          <Link
            className="ui-button ui-button--primary ui-button--md"
            to="/employer/vacancies/new"
          >
            <Plus size={18} aria-hidden="true" />
            <span>{t("employer.createVacancy")}</span>
          </Link>
        </header>
        <div
          className="employer-metrics"
          aria-label={t("employer.vacancySummary")}
        >
          {publicationStatuses.map((status) => (
            <Link
              key={status}
              to={`/employer/vacancies?status=${status}`}
              className={`employer-metric employer-metric--${status}`}
            >
              <span>{t(`employer.metrics.${status}`)}</span>
              <strong>
                {jobs.isSuccess
                  ? jobs.data.filter((job) => job.publicationStatus === status)
                      .length
                  : "—"}
              </strong>
            </Link>
          ))}
        </div>
        <section aria-labelledby="recent-vacancies-title">
          <header className="workbench-section-header">
            <h2 id="recent-vacancies-title">{t("employer.recentVacancies")}</h2>
            <Link className="text-link" to="/employer/vacancies">
              {t("employer.allVacancies")}
              <ArrowRight size={17} aria-hidden="true" />
            </Link>
          </header>
          {jobs.isPending ? (
            <JobListSkeleton label={t("employer.loadingVacancies")} />
          ) : null}
          {jobs.isError ? (
            <ErrorState
              title={t("employer.loadErrorTitle")}
              description={t("employer.loadErrorDescription")}
              actionLabel={t("common.retry")}
              onAction={() => void jobs.refetch()}
            />
          ) : null}
          {jobs.isSuccess && !jobs.data.length ? (
            <EmptyState
              title={t("employer.emptyTitle")}
              description={t("employer.emptyDescription")}
              secondary={
                <Link
                  className="ui-button ui-button--primary ui-button--md"
                  to="/employer/vacancies/new"
                >
                  <Plus size={17} />
                  {t("employer.createVacancy")}
                </Link>
              }
            />
          ) : null}
          {jobs.isSuccess && recentJobs.length ? (
            <EmployerJobTable jobs={recentJobs} />
          ) : null}
        </section>
        <div className="employer-company-strip">
          <span className="workbench__avatar" aria-hidden="true">
            <Building2 size={22} />
          </span>
          <div>
            <strong>{session!.company?.name || t("common.notProvided")}</strong>
            <span>{t("companyEdit.title")}</span>
          </div>
          <Link className="text-link" to="/employer/company">
            {t("employer.companySettings")}
            <ArrowRight size={17} aria-hidden="true" />
          </Link>
        </div>
      </section>
    </EmployerWorkspaceLayout>
  );
}
