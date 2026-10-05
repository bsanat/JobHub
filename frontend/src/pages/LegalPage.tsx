import { ArrowLeft } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { SiteShell } from "../layouts/SiteShell";
import "./legal.css";

type LegalDocument = "privacy" | "terms";

export function LegalPage({ document }: { document: LegalDocument }) {
  const { t } = useTranslation();
  const root = `legal.${document}`;
  const sections = document === "privacy"
    ? ["operator", "collect", "use", "providers", "sharing", "rights", "children", "changes", "contact"]
    : ["service", "accounts", "seeker", "employer", "content", "external", "telegram", "availability", "liability", "contact"];

  return (
    <SiteShell>
      <article className="legal-page page-container">
        <Link className="legal-page__back" to="/">
          <ArrowLeft size={17} aria-hidden="true" />
          {t("legal.back")}
        </Link>
        <header className="legal-page__header">
          <h1>{t(`${root}.title`)}</h1>
          <p>{t(`${root}.updated`)}</p>
          <div className="legal-page__intro">{t(`${root}.intro`)}</div>
        </header>
        {sections.map((section) => {
          const content = t(`${root}.${section}`, { returnObjects: true });
          return (
            <section className="legal-page__section" key={section}>
              <h2>{t(`${root}.${section}Title`)}</h2>
              {Array.isArray(content) ? (
                <ul>{content.map((item) => <li key={item}>{item}</li>)}</ul>
              ) : (
                <p>{typeof content === "string" ? content : ""}</p>
              )}
            </section>
          );
        })}
      </article>
    </SiteShell>
  );
}
