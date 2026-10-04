import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download, FileText, RefreshCw, UserRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useParams } from "react-router-dom";
import type { ApplicationStatus } from "../api/models/candidate";
import { repositories } from "../api/repositories";
import { useAuth } from "../app/authContext";
import { EmployerWorkspaceLayout } from "../components/employer/EmployerWorkspaceLayout";
import { Button } from "../components/ui/Button";
import {
  EmptyState,
  ErrorState,
  JobListSkeleton,
} from "../components/ui/Feedback";
import { useToast } from "../components/ui/useToast";

const actions: Exclude<ApplicationStatus, "sent" | "withdrawn">[] = [
  "viewed",
  "in_review",
  "contacted",
  "interview",
  "offer",
  "rejected",
];
export function EmployerApplicantsPage() {
  const { jobId = "" } = useParams();
  const { t, i18n } = useTranslation();
  const { session } = useAuth();
  const client = useQueryClient();
  const toast = useToast();
  const apps = useQuery({
    queryKey: ["employer", "applications", session?.user.id, jobId],
    queryFn: () => repositories.candidate.listEmployerApplications(jobId),
    enabled: session?.user.role === "employer",
    refetchInterval: 30_000,
  });
  const update = useMutation({
    mutationFn: ({
      id,
      status,
    }: {
      id: string;
      status: Exclude<ApplicationStatus, "sent" | "withdrawn">;
    }) =>
      repositories.candidate.updateApplicationStatus(
        id,
        status,
        session!.csrfToken,
      ),
    onSuccess: async () => {
      await client.invalidateQueries({
        queryKey: ["employer", "applications"],
      });
      toast.showToast({ title: t("employerApplicants.updated") });
    },
    onError: () =>
      toast.showToast({
        tone: "error",
        title: t("employerApplicants.updateError"),
      }),
  });
  const download = async (id: string, name: string) => {
    try {
      const blob = await repositories.candidate.downloadApplicationResume(id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = name;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      toast.showToast({
        tone: "error",
        title: t("employerApplicants.resumeError"),
      });
    }
  };
  return (
    <EmployerWorkspaceLayout>
      <section className="employer-applicants-page">
        <header className="workbench-page-header">
          <div>
            <h1>
              {t(jobId ? "employerApplicants.title" : "employer.applications")}
            </h1>
          </div>
          <div className="employer-applicants-page__actions">
            <Link
              className="text-link"
              to={jobId ? "/employer/applications" : "/employer/vacancies"}
            >
              {t(jobId ? "employerApplicants.all" : "employerApplicants.back")}
            </Link>
            <Button
              variant="quiet"
              aria-label={t("employerApplicants.refresh")}
              title={t("employerApplicants.refresh")}
              disabled={apps.isFetching}
              onClick={() => void apps.refetch()}
            >
              <RefreshCw size={18} aria-hidden="true" />
            </Button>
          </div>
        </header>
        {apps.isPending ? (
          <JobListSkeleton label={t("employerApplicants.loading")} />
        ) : null}
        {apps.isError ? (
          <ErrorState
            title={t("employerApplicants.loadErrorTitle")}
            description={t("employerApplicants.loadErrorDescription")}
            actionLabel={t("common.retry")}
            onAction={() => void apps.refetch()}
          />
        ) : null}
        {apps.data?.length === 0 ? (
          <EmptyState
            title={t("employerApplicants.emptyTitle")}
            description={t(
              jobId
                ? "employerApplicants.emptyDescription"
                : "employerApplicants.emptyAllDescription",
            )}
          />
        ) : null}
        {apps.data?.length ? (
          <div className="applicant-list">
            {apps.data.map((app) => (
              <article className="applicant-row" key={app.id}>
                <div className="applicant-row__identity">
                  <span>
                    <UserRound size={20} />
                  </span>
                  <div>
                    <h2>{app.candidateName}</h2>
                    <p>{app.candidatePosition || t("common.notProvided")}</p>
                    <Link
                      className="text-link"
                      to={`/employer/vacancies/${app.jobId}/applicants`}
                    >
                      {app.jobTitle || t("employerApplicants.vacancy")}
                    </Link>
                    <small>{formatDate(app.createdAt, i18n.language)}</small>
                  </div>
                </div>
                <div className="applicant-row__message">
                  <strong>{t("employerApplicants.message")}</strong>
                  <p>{app.message || t("employerApplicants.noMessage")}</p>
                  <Button
                    size="sm"
                    variant="secondary"
                    leadingIcon={<Download size={15} />}
                    onClick={() =>
                      void download(app.id, app.resume.originalFilename)
                    }
                  >
                    <FileText size={15} />
                    {app.resume.originalFilename}
                  </Button>
                </div>
                <div className="applicant-row__status">
                  <span className={`application-status status-${app.status}`}>
                    {t(`applications.statuses.${app.status}`)}
                  </span>
                  <label>
                    <span>{t("employerApplicants.changeStatus")}</span>
                    <select
                      value={app.status}
                      disabled={
                        app.status === "withdrawn" ||
                        app.status === "rejected" ||
                        update.isPending
                      }
                      onChange={(event) =>
                        update.mutate({
                          id: app.id,
                          status: event.target.value as Exclude<
                            ApplicationStatus,
                            "sent" | "withdrawn"
                          >,
                        })
                      }
                    >
                      <option value={app.status}>
                        {t(`applications.statuses.${app.status}`)}
                      </option>
                      {nextStatuses(app.status).map((status) => (
                        <option value={status} key={status}>
                          {t(`applications.statuses.${status}`)}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </section>
    </EmployerWorkspaceLayout>
  );
}
function formatDate(value: string, language: string) {
  return new Intl.DateTimeFormat(language, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function nextStatuses(current: ApplicationStatus) {
  if (current === "rejected" || current === "withdrawn") return [];
  const rank: Partial<Record<ApplicationStatus, number>> = {
    sent: 0,
    viewed: 1,
    in_review: 2,
    contacted: 3,
    interview: 4,
    offer: 5,
  };
  return actions.filter(
    (status) =>
      status === "rejected" ||
      (rank[status] ?? -1) > (rank[current] ?? Number.MAX_SAFE_INTEGER),
  );
}
