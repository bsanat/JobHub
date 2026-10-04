import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bookmark, ChevronRight, MapPin, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { repositories } from "../../api/repositories";
import { useAuth } from "../../app/authContext";
import { CandidateListToolbar } from "./CandidateListToolbar";
import { IconButton } from "../ui/Button";
import { EmptyState, ErrorState, JobListSkeleton } from "../ui/Feedback";
import { useToast } from "../ui/useToast";

export function SavedJobsPanel() {
  const { t } = useTranslation();
  const { session } = useAuth();
  const client = useQueryClient();
  const toast = useToast();
  const [query, setQuery] = useState("");
  const saved = useQuery({
    queryKey: ["saved-jobs"],
    queryFn: () => repositories.candidate.listSavedJobs(),
  });
  const unsave = useMutation({
    mutationFn: (jobId: string) =>
      repositories.candidate.unsaveJob(jobId, session!.csrfToken),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ["saved-jobs"] });
      toast.showToast({ title: t("saved.removed") });
    },
    onError: () => toast.showToast({ tone: "error", title: t("saved.error") }),
  });
  const search = query.trim().toLocaleLowerCase();
  const visible = saved.data?.filter((item) =>
    [
      item.title,
      item.companyName,
      item.location,
      item.sourceName,
      !item.available ? t("saved.unavailable") : "",
    ]
      .join(" ")
      .toLocaleLowerCase()
      .includes(search),
  );
  return (
    <>
      {saved.isPending ? <JobListSkeleton label={t("saved.loading")} /> : null}
      {saved.isError ? (
        <ErrorState
          title={t("saved.loadErrorTitle")}
          description={t("saved.loadErrorDescription")}
          actionLabel={t("common.retry")}
          onAction={() => void saved.refetch()}
        />
      ) : null}
      {saved.data?.length === 0 ? (
        <EmptyState
          illustration="/illustrations/empty-saved-jobs.png"
          title={t("saved.emptyTitle")}
          description={t("saved.emptyDescription")}
        />
      ) : null}
      {saved.data?.length ? (
        <>
          <CandidateListToolbar
            query={query}
            onQueryChange={setQuery}
            searchLabel={t("candidateLists.searchJobs")}
            count={visible?.length ?? 0}
          />
          <div className="candidate-job-list">
            {visible?.map((item) => (
              <article
                className={`candidate-job-row${item.available ? "" : " is-unavailable"}`}
                key={item.jobId}
              >
                <span className="candidate-row-mark" aria-hidden="true">
                  <Bookmark size={19} />
                </span>
                <div>
                  <h2>
                    {item.available ? (
                      <Link to={`/jobs/${item.jobId}`}>{item.title}</Link>
                    ) : (
                      t("saved.unavailable")
                    )}
                  </h2>
                  <p>
                    {item.available
                      ? item.companyName
                      : t("saved.unavailableHint")}
                  </p>
                  {item.available && item.location ? (
                    <span>
                      <MapPin size={14} aria-hidden="true" />
                      {item.location}
                    </span>
                  ) : null}
                  {item.available && item.sourceName ? (
                    <small>
                      {t("saved.source", { source: item.sourceName })}
                    </small>
                  ) : null}
                </div>
                <div className="candidate-job-row__actions">
                  {item.available ? (
                    <Link
                      className="ui-icon-button ui-icon-button--quiet"
                      aria-label={t("saved.openJob", { title: item.title })}
                      title={t("saved.openJob", { title: item.title })}
                      to={`/jobs/${item.jobId}`}
                    >
                      <ChevronRight size={19} />
                    </Link>
                  ) : null}
                  <IconButton
                    variant="quiet"
                    label={t("saved.unsave")}
                    icon={
                      unsave.isPending && unsave.variables === item.jobId ? (
                        <span className="ui-spinner" />
                      ) : (
                        <Trash2 size={17} />
                      )
                    }
                    disabled={unsave.isPending}
                    onClick={() => unsave.mutate(item.jobId)}
                  />
                </div>
              </article>
            ))}
          </div>
          {!visible?.length ? (
            <p className="section-empty">{t("candidateLists.noResults")}</p>
          ) : null}
        </>
      ) : null}
    </>
  );
}
