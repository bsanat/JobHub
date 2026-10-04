import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronRight, MapPin, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { repositories } from "../../api/repositories";
import { useAuth } from "../../app/authContext";
import { CandidateListToolbar } from "./CandidateListToolbar";
import { CompanyIdentity } from "../companies/CompanyIdentity";
import { IconButton } from "../ui/Button";
import { EmptyState, ErrorState, JobListSkeleton } from "../ui/Feedback";
import { useToast } from "../ui/useToast";

export function SavedCompaniesPanel() {
  const { t } = useTranslation();
  const { session } = useAuth();
  const client = useQueryClient();
  const toast = useToast();
  const [query, setQuery] = useState("");
  const companies = useQuery({
    queryKey: ["followed-companies"],
    queryFn: () => repositories.companies.listFollowing(),
  });
  const unfollow = useMutation({
    mutationFn: (companyId: string) =>
      repositories.companies.unfollow(companyId, session!.csrfToken),
    onSuccess: async (_, companyId) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ["followed-companies"] }),
        client.invalidateQueries({ queryKey: ["company-follow", companyId] }),
        client.invalidateQueries({ queryKey: ["company", companyId] }),
      ]);
      toast.showToast({ title: t("savedCompanies.removed") });
    },
    onError: () =>
      toast.showToast({ tone: "error", title: t("savedCompanies.error") }),
  });
  const search = query.trim().toLocaleLowerCase();
  const visible = companies.data?.filter((company) =>
    [company.name, company.industry, company.city]
      .join(" ")
      .toLocaleLowerCase()
      .includes(search),
  );
  return (
    <>
      {companies.isPending ? (
        <JobListSkeleton label={t("savedCompanies.loading")} />
      ) : null}
      {companies.isError ? (
        <ErrorState
          title={t("savedCompanies.loadErrorTitle")}
          description={t("savedCompanies.loadErrorDescription")}
          actionLabel={t("common.retry")}
          onAction={() => void companies.refetch()}
        />
      ) : null}
      {companies.data?.length === 0 ? (
        <EmptyState
          title={t("savedCompanies.emptyTitle")}
          description={t("savedCompanies.emptyDescription")}
        />
      ) : null}
      {companies.data?.length ? (
        <>
          <CandidateListToolbar
            query={query}
            onQueryChange={setQuery}
            searchLabel={t("candidateLists.searchCompanies")}
            count={visible?.length ?? 0}
          />
          <div className="saved-company-list">
            {visible?.map((company) => (
              <article key={company.id} className="saved-company-row">
                <div>
                  <Link
                    className="saved-company-row__name"
                    to={`/companies/${company.id}`}
                  >
                    <CompanyIdentity
                      company={company}
                      verifiedLabel={t("companies.verified")}
                    />
                  </Link>
                  <div className="saved-company-row__info">
                    <span>
                      {company.industry || t("savedCompanies.industryMissing")}
                    </span>
                    {company.city ? (
                      <span>
                        <MapPin size={14} aria-hidden="true" />
                        {company.city}
                      </span>
                    ) : null}
                  </div>
                </div>
                <p className="saved-company-row__description">
                  {company.description || t("companies.noDescription")}
                </p>
                <div className="saved-company-row__actions">
                  <Link
                    className="ui-icon-button ui-icon-button--quiet"
                    aria-label={t("companies.openProfile")}
                    title={t("companies.openProfile")}
                    to={`/companies/${company.id}`}
                  >
                    <ChevronRight size={19} />
                  </Link>
                  <IconButton
                    variant="quiet"
                    label={t("savedCompanies.unfollow")}
                    icon={
                      unfollow.isPending &&
                      unfollow.variables === company.id ? (
                        <span className="ui-spinner" />
                      ) : (
                        <Trash2 size={17} />
                      )
                    }
                    disabled={unfollow.isPending}
                    onClick={() => unfollow.mutate(company.id)}
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
