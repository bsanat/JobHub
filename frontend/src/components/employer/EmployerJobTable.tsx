import {
  ArrowRight,
  ExternalLink,
  MoreHorizontal,
  Pause,
  Pencil,
  Play,
  Trash2,
  Users,
  XCircle,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import type { JobSummary, PublicationStatus } from "../../api/models/job";
import { displayJobLocation } from "../../lib/jobLocation";
import { formatSalary } from "../../lib/formatSalary";
import { jobUpdatedTime } from "../../lib/employerVacancies";

interface EmployerJobTableProps {
  jobs: JobSummary[];
  management?: boolean;
  busy?: boolean;
  pendingId?: string;
  onTransition?: (id: string, status: PublicationStatus) => void;
  onDelete?: (id: string) => void;
}

export function EmployerJobTable({
  jobs,
  management = false,
  busy,
  pendingId,
  onTransition,
  onDelete,
}: EmployerJobTableProps) {
  const { t, i18n } = useTranslation();
  return (
    <div className="employer-table-scroll">
      <table
        className={`employer-table${management ? " employer-table--management" : ""}`}
      >
        <caption className="visually-hidden">
          {t(management ? "employer.myVacancies" : "employer.recentVacancies")}
        </caption>
        <thead>
          <tr>
            <th scope="col">{t("employer.table.vacancy")}</th>
            <th scope="col">{t("employer.table.status")}</th>
            <th scope="col">
              {t(
                management ? "employer.table.salary" : "employer.table.updated",
              )}
            </th>
            <th scope="col">
              <span className="visually-hidden">
                {t("employer.table.actions")}
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((job) => {
            const destination =
              job.publicationStatus === "closed"
                ? `/employer/vacancies/${job.id}/applicants`
                : `/employer/vacancies/${job.id}/edit`;
            return (
              <tr key={job.id} aria-busy={pendingId === job.id || undefined}>
                <td>
                  <Link className="employer-table__title" to={destination}>
                    {job.title}
                  </Link>
                  <p>
                    {displayJobLocation(job, i18n.language)}
                    {job.workMode
                      ? ` · ${t(`employer.workModes.${job.workMode}`)}`
                      : ""}
                  </p>
                  {management ? (
                    <small>
                      {t("employer.updatedAt", {
                        date: formatDate(jobUpdatedTime(job), i18n.language),
                      })}
                    </small>
                  ) : null}
                </td>
                <td>
                  <span
                    className={`publication-label publication-label--${job.publicationStatus || "unknown"}`}
                  >
                    <span aria-hidden="true" />
                    {job.publicationStatus
                      ? t(`employer.statuses.${job.publicationStatus}`)
                      : t("common.notProvided")}
                  </span>
                  {job.moderationStatus &&
                  job.moderationStatus !== "approved" ? (
                    <small className="employer-table__moderation">
                      {t(`employer.moderation.${job.moderationStatus}`)}
                    </small>
                  ) : null}
                </td>
                <td className="employer-table__secondary">
                  {management
                    ? job.salary
                      ? formatSalary(
                          job.salary,
                          i18n.language,
                          t(`employer.salaryPeriods.${job.salary.period}`),
                        )
                      : t("common.notProvided")
                    : formatDate(jobUpdatedTime(job), i18n.language)}
                </td>
                <td>
                  {management ? (
                    <div className="employer-table__actions">
                      <Link
                        className="ui-button ui-button--quiet ui-button--sm"
                        to={`/employer/vacancies/${job.id}/applicants`}
                      >
                        <Users size={16} aria-hidden="true" />
                        <span>{t("employer.applicants")}</span>
                      </Link>
                      <details
                        className="vacancy-menu"
                        onBlur={(event) => {
                          if (
                            !event.currentTarget.contains(
                              event.relatedTarget as Node | null,
                            )
                          )
                            event.currentTarget.removeAttribute("open");
                        }}
                        onKeyDown={(event) => {
                          if (event.key === "Escape") {
                            event.currentTarget.removeAttribute("open");
                            event.currentTarget
                              .querySelector("summary")
                              ?.focus();
                          }
                        }}
                      >
                        <summary
                          aria-label={t("employer.vacancyActions", {
                            title: job.title,
                          })}
                          title={t("employer.table.actions")}
                        >
                          <MoreHorizontal size={19} />
                        </summary>
                        <div
                          className="vacancy-menu__content"
                          onClick={(event) =>
                            event.currentTarget
                              .closest("details")
                              ?.removeAttribute("open")
                          }
                        >
                          {job.publicationStatus !== "closed" ? (
                            <Link to={destination}>
                              <Pencil size={16} />
                              {t("employer.edit")}
                            </Link>
                          ) : null}
                          {job.publicationStatus === "published" ? (
                            <Link to={`/jobs/${job.id}`}>
                              <ExternalLink size={16} />
                              {t("employer.view")}
                            </Link>
                          ) : null}
                          {job.publicationStatus === "draft" ||
                          job.publicationStatus === "paused" ? (
                            <button
                              type="button"
                              disabled={busy}
                              onClick={() =>
                                onTransition?.(job.id, "published")
                              }
                            >
                              <Play size={16} />
                              {t("employer.publish")}
                            </button>
                          ) : null}
                          {job.publicationStatus === "published" ? (
                            <button
                              type="button"
                              disabled={busy}
                              onClick={() => onTransition?.(job.id, "paused")}
                            >
                              <Pause size={16} />
                              {t("employer.pause")}
                            </button>
                          ) : null}
                          {job.publicationStatus !== "closed" ? (
                            <button
                              type="button"
                              className="is-destructive"
                              disabled={busy}
                              onClick={() => onTransition?.(job.id, "closed")}
                            >
                              <XCircle size={16} />
                              {t("employer.close")}
                            </button>
                          ) : null}
                          {job.publicationStatus === "draft" ||
                          job.publicationStatus === "closed" ? (
                            <button
                              type="button"
                              className="is-destructive"
                              disabled={busy}
                              onClick={() => onDelete?.(job.id)}
                            >
                              <Trash2 size={16} />
                              {t("employer.delete")}
                            </button>
                          ) : null}
                        </div>
                      </details>
                    </div>
                  ) : (
                    <Link
                      className="employer-table__open"
                      to={destination}
                      aria-label={t("employer.openVacancy", {
                        title: job.title,
                      })}
                      title={t("employer.view")}
                    >
                      <ArrowRight size={18} />
                    </Link>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function formatDate(value: number, language: string) {
  return value
    ? new Intl.DateTimeFormat(language, {
        day: "numeric",
        month: "short",
        year: "numeric",
      }).format(value)
    : "—";
}
