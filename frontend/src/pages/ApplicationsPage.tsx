import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BriefcaseBusiness, CalendarDays, ChevronRight, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { repositories } from "../api/repositories";
import { useAuth } from "../app/authContext";
import { Button } from "../components/ui/Button";
import {
  EmptyState,
  ErrorState,
  JobListSkeleton,
} from "../components/ui/Feedback";
import { useToast } from "../components/ui/useToast";
import { CandidateJobsLayout } from "../components/candidate/CandidateJobsLayout";
import { CandidateListToolbar } from "../components/candidate/CandidateListToolbar";
import type { ApplicationStatus } from "../api/models/candidate";

export function ApplicationsPage() {
  const { t, i18n } = useTranslation();
  const { session } = useAuth();
  const client = useQueryClient();
  const toast = useToast();
  const [status, setStatus] = useState<"all" | ApplicationStatus>("all");
  const [query, setQuery] = useState("");
  const apps = useQuery({
    queryKey: ["applications"],
    queryFn: () => repositories.candidate.listApplications(),
  });
  const withdraw = useMutation({
    mutationFn: (id: string) =>
      repositories.candidate.withdraw(id, session!.csrfToken),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ["applications"] });
      toast.showToast({ title: t("applications.withdrawn") });
    },
    onError: () =>
      toast.showToast({
        tone: "error",
        title: t("applications.withdrawError"),
      }),
  });
  const search = query.trim().toLocaleLowerCase();
  const visibleApplications = apps.data?.filter(
    (application) =>
      (status === "all" || application.status === status) &&
      [
        application.jobTitle,
        application.companyName,
        !application.jobAvailable ? t("applications.unavailableJob") : "",
      ]
        .join(" ")
        .toLocaleLowerCase()
        .includes(search),
  );
  return (
    <CandidateJobsLayout
      tabTitle={t("applications.title")}
      action={
        <Link
          className="ui-button ui-button--secondary ui-button--md"
          to="/jobs"
        >
          {t("applications.browse")}
          <ChevronRight size={17} />
        </Link>
      }
    >
      {apps.data?.length ? (
        <CandidateListToolbar
          query={query}
          onQueryChange={setQuery}
          searchLabel={t("candidateLists.searchApplications")}
          count={visibleApplications?.length ?? 0}
        >
          <label className="application-filter">
            <span className="visually-hidden">
              {t("applications.filterLabel")}
            </span>
            <select
              value={status}
              onChange={(event) =>
                setStatus(event.target.value as "all" | ApplicationStatus)
              }
            >
              <option value="all">{t("applications.filterAll")}</option>
              {applicationStatuses.map((value) => (
                <option key={value} value={value}>
                  {t(`applications.statuses.${value}`)}
                </option>
              ))}
            </select>
          </label>
        </CandidateListToolbar>
      ) : null}
      {apps.isPending ? (
        <JobListSkeleton label={t("applications.loading")} />
      ) : null}
      {apps.isError ? (
        <ErrorState
          title={t("applications.loadErrorTitle")}
          description={t("applications.loadErrorDescription")}
          actionLabel={t("common.retry")}
          onAction={() => void apps.refetch()}
        />
      ) : null}
      {apps.data?.length === 0 ? (
        <EmptyState
          illustration="/illustrations/empty-applications.png"
          title={t("applications.emptyTitle")}
          description={t("applications.emptyDescription")}
        />
      ) : null}
      {visibleApplications?.length ? (
        <div className="application-list">
          {visibleApplications.map((app) => (
            <article className="application-row" key={app.id}>
              <div className="application-row__icon">
                <BriefcaseBusiness size={19} aria-hidden="true" />
              </div>
              <div className="application-row__main">
                <div>
                  <h2>
                    {app.jobAvailable ? (
                      <Link to={`/jobs/${app.jobId}`}>{app.jobTitle}</Link>
                    ) : (
                      t("applications.unavailableJob")
                    )}
                  </h2>
                  <span className={`application-status status-${app.status}`}>
                    {t(`applications.statuses.${app.status}`)}
                  </span>
                </div>
                {app.jobAvailable ? (
                  <p>{app.companyName}</p>
                ) : (
                  <p>{t("applications.unavailableHint")}</p>
                )}
                <div className="application-row__meta">
                  <span>
                    <CalendarDays size={15} aria-hidden="true" />
                    {t("applications.submittedAt", {
                      date: formatDate(app.createdAt, i18n.language),
                    })}
                  </span>
                  <span>
                    {t("applications.updatedAt", {
                      date: formatDate(app.latestStatusAt, i18n.language),
                    })}
                  </span>
                </div>
                {app.message ? (
                  <p className="application-message">{app.message}</p>
                ) : null}
              </div>
              <div className="application-row__actions">
                {app.jobAvailable ? (
                  <Link
                    className="ui-icon-button ui-icon-button--quiet"
                    aria-label={t("applications.viewJob")}
                    title={t("applications.viewJob")}
                    to={`/jobs/${app.jobId}`}
                  >
                    <ChevronRight size={19} />
                  </Link>
                ) : null}
                {app.status !== "rejected" && app.status !== "withdrawn" ? (
                  <Button
                    size="sm"
                    variant="quiet"
                    leadingIcon={<X size={16} />}
                    isLoading={
                      withdraw.isPending && withdraw.variables === app.id
                    }
                    disabled={withdraw.isPending}
                    onClick={() =>
                      window.confirm(t("applications.withdrawConfirm")) &&
                      withdraw.mutate(app.id)
                    }
                  >
                    {t("applications.withdraw")}
                  </Button>
                ) : null}
              </div>
            </article>
          ))}
        </div>
      ) : null}
      {apps.data?.length && visibleApplications?.length === 0 ? (
        <p className="section-empty application-filter-empty">
          {query
            ? t("candidateLists.noResults")
            : t("applications.filterEmpty")}
        </p>
      ) : null}
    </CandidateJobsLayout>
  );
}
function formatDate(value: string, language: string) {
  return new Intl.DateTimeFormat(language, { dateStyle: "medium" }).format(
    new Date(value),
  );
}
const applicationStatuses: ApplicationStatus[] = [
  "sent",
  "viewed",
  "in_review",
  "contacted",
  "interview",
  "offer",
  "rejected",
  "withdrawn",
];
