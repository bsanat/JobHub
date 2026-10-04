import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Search, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { displayJobLocation } from "../lib/jobLocation";
import { Link, useSearchParams } from "react-router-dom";
import type { PublicationStatus } from "../api/models/job";
import { repositories } from "../api/repositories";
import { useAuth } from "../app/authContext";
import { EmployerJobTable } from "../components/employer/EmployerJobTable";
import { jobUpdatedTime, publicationStatuses } from "../lib/employerVacancies";
import { EmployerWorkspaceLayout } from "../components/employer/EmployerWorkspaceLayout";
import { IconButton } from "../components/ui/Button";
import {
  EmptyState,
  ErrorState,
  JobListSkeleton,
} from "../components/ui/Feedback";
import { useToast } from "../components/ui/useToast";

export function EmployerVacanciesPage() {
  const { t, i18n } = useTranslation();
  const { session } = useAuth();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [params, setParams] = useSearchParams();
  const query = params.get("q") || "";
  const rawStatus = params.get("status");
  const status =
    publicationStatuses.find((value) => value === rawStatus) || "all";
  const sort = params.get("sort") === "oldest" ? "oldest" : "newest";
  const jobs = useQuery({
    queryKey: ["employer", "jobs", "list", session!.user.id],
    queryFn: () => repositories.employerJobs.list(),
  });
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["employer", "jobs"] }),
      queryClient.invalidateQueries({ queryKey: ["jobs"] }),
      queryClient.invalidateQueries({ queryKey: ["employer-company"] }),
      queryClient.invalidateQueries({ queryKey: ["companies"] }),
    ]);
  const transition = useMutation({
    mutationFn: ({ id, status }: { id: string; status: PublicationStatus }) =>
      repositories.employerJobs.transition(id, status, session!.csrfToken),
    onSuccess: async () => {
      await refresh();
      toast.showToast({ title: t("employer.statusUpdated") });
    },
    onError: () =>
      toast.showToast({ title: t("employer.statusError"), tone: "error" }),
  });
  const remove = useMutation({
    mutationFn: (id: string) =>
      repositories.employerJobs.delete(id, session!.csrfToken),
    onSuccess: async () => {
      await refresh();
      toast.showToast({ title: t("employer.vacancyDeleted") });
    },
    onError: () =>
      toast.showToast({ title: t("employer.deleteError"), tone: "error" }),
  });
  const updateFilter = (key: string, value: string) =>
    setParams(
      (current) => {
        const next = new URLSearchParams(current);
        if (value && value !== "all" && !(key === "sort" && value === "newest"))
          next.set(key, value);
        else next.delete(key);
        return next;
      },
      { replace: true },
    );
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const filtered = (jobs.data || [])
    .filter(
      (job) =>
        (status === "all" || job.publicationStatus === status) &&
        (!normalizedQuery ||
          `${job.title} ${displayJobLocation(job, i18n.language)} ${job.location} ${job.category || ""}`
            .toLocaleLowerCase()
            .includes(normalizedQuery)),
    )
    .sort(
      (a, b) =>
        (jobUpdatedTime(b) - jobUpdatedTime(a)) * (sort === "oldest" ? -1 : 1),
    );
  const busy = transition.isPending || remove.isPending;

  return (
    <EmployerWorkspaceLayout>
      <section aria-labelledby="employer-vacancies-title">
        <header className="workbench-page-header">
          <div>
            <h1 id="employer-vacancies-title">{t("employer.myVacancies")}</h1>
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
        <nav
          className="vacancy-filters"
          aria-label={t("employer.statusFilter")}
        >
          {(["all", ...publicationStatuses] as const).map((value) => (
            <button
              key={value}
              type="button"
              aria-pressed={status === value}
              onClick={() => updateFilter("status", value)}
            >
              <span>
                {t(
                  value === "all"
                    ? "employer.all"
                    : `employer.metrics.${value}`,
                )}
              </span>
              <small>
                {jobs.isSuccess
                  ? value === "all"
                    ? jobs.data.length
                    : jobs.data.filter((job) => job.publicationStatus === value)
                        .length
                  : "—"}
              </small>
            </button>
          ))}
        </nav>
        <div className="vacancy-toolbar">
          <label className="vacancy-search">
            <span className="visually-hidden">{t("employer.searchLabel")}</span>
            <Search size={18} aria-hidden="true" />
            <input
              type="search"
              value={query}
              placeholder={t("employer.searchPlaceholder")}
              onChange={(event) => updateFilter("q", event.target.value)}
            />
          </label>
          <label className="vacancy-sort">
            <span className="visually-hidden">{t("jobs.sortLabel")}</span>
            <select
              value={sort}
              onChange={(event) => updateFilter("sort", event.target.value)}
            >
              <option value="newest">{t("jobs.newest")}</option>
              <option value="oldest">{t("jobs.oldest")}</option>
            </select>
          </label>
          {query || status !== "all" ? (
            <IconButton
              label={t("employer.resetFilters")}
              icon={<X size={18} />}
              onClick={() => setParams({}, { replace: true })}
            />
          ) : null}
        </div>
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
              <Link className="text-link" to="/employer/vacancies/new">
                {t("employer.createVacancy")}
              </Link>
            }
          />
        ) : null}
        {jobs.isSuccess && jobs.data.length > 0 && !filtered.length ? (
          <EmptyState
            title={t("employer.filterEmptyTitle")}
            description={t("employer.filterEmptyDescription")}
            actionLabel={t("employer.resetFilters")}
            onAction={() => setParams({}, { replace: true })}
          />
        ) : null}
        {jobs.isSuccess && filtered.length ? (
          <>
            <p className="vacancy-result-count" role="status">
              {t("employer.shownVacancies", {
                count: filtered.length,
                total: jobs.data.length,
              })}
            </p>
            <EmployerJobTable
              jobs={filtered}
              management
              busy={busy}
              pendingId={
                transition.isPending
                  ? transition.variables?.id
                  : remove.isPending
                    ? remove.variables
                    : undefined
              }
              onTransition={(id, nextStatus) => {
                if (
                  !busy &&
                  (nextStatus !== "closed" ||
                    window.confirm(t("employer.closeConfirm")))
                )
                  transition.mutate({ id, status: nextStatus });
              }}
              onDelete={(id) => {
                if (!busy && window.confirm(t("employer.deleteConfirm")))
                  remove.mutate(id);
              }}
            />
          </>
        ) : null}
      </section>
    </EmployerWorkspaceLayout>
  );
}
