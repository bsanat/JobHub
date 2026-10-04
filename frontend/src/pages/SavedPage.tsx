import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router-dom";
import { CandidateJobsLayout } from "../components/candidate/CandidateJobsLayout";
import { SavedCompaniesPanel } from "../components/candidate/SavedCompaniesPanel";
import { SavedJobsPanel } from "../components/candidate/SavedJobsPanel";

export function SavedPage() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const showCompanies = params.get("tab") === "companies";

  return (
    <CandidateJobsLayout
      tabTitle={t("savedPage.title")}
      action={
        <Link
          className="ui-button ui-button--secondary ui-button--md"
          to={showCompanies ? "/companies" : "/jobs"}
        >
          {t(showCompanies ? "savedCompanies.browse" : "saved.browse")}
          <ChevronRight size={17} aria-hidden="true" />
        </Link>
      }
    >
      <nav className="saved-section-nav" aria-label={t("savedPage.tabsLabel")}>
        <Link to="/saved" aria-current={!showCompanies ? "page" : undefined}>
          {t("savedPage.jobs")}
        </Link>
        <Link
          to="/saved?tab=companies"
          aria-current={showCompanies ? "page" : undefined}
        >
          {t("savedPage.companies")}
        </Link>
      </nav>
      <section hidden={showCompanies} aria-label={t("saved.title")}>
        <SavedJobsPanel />
      </section>
      <section hidden={!showCompanies} aria-label={t("savedCompanies.title")}>
        <SavedCompaniesPanel />
      </section>
    </CandidateJobsLayout>
  );
}
