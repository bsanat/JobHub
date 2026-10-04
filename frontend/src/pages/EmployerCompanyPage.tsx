import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ExternalLink, Pencil, Save, X } from "lucide-react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { repositories } from "../api/repositories";
import type { Company, CompanyProfileInput } from "../api/models/company";
import { useAuth } from "../app/authContext";
import { EmployerWorkspaceLayout } from "../components/employer/EmployerWorkspaceLayout";
import { CompanyIdentity } from "../components/companies/CompanyIdentity";
import { Button } from "../components/ui/Button";
import { ErrorState, JobListSkeleton } from "../components/ui/Feedback";
import { Input, Textarea } from "../components/ui/FormControls";
import { useToast } from "../components/ui/useToast";
import { safeProfileUrl } from "../lib/profileContact";

export function EmployerCompanyPage() {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = useState(false);
  const editButton = useRef<HTMLButtonElement>(null);
  const company = useQuery({
    queryKey: ["employer-company"],
    queryFn: () => repositories.companies.getEmployerCompany(),
  });
  const finishEditing = () => {
    setIsEditing(false);
    requestAnimationFrame(() => editButton.current?.focus());
  };
  return (
    <EmployerWorkspaceLayout>
      <section className="employer-company-page">
        <header className="workbench-page-header">
          <div>
            <h1>{t("companyEdit.title")}</h1>
          </div>
          {company.data && !isEditing ? (
            <Button
              ref={editButton}
              variant="secondary"
              leadingIcon={<Pencil size={17} />}
              onClick={() => setIsEditing(true)}
            >
              {t("companyEdit.edit")}
            </Button>
          ) : null}
        </header>
        {company.isPending ? (
          <JobListSkeleton label={t("companyEdit.loading")} />
        ) : null}
        {company.isError ? (
          <ErrorState
            title={t("companyEdit.errorTitle")}
            description={t("companyEdit.errorDescription")}
            actionLabel={t("common.retry")}
            onAction={() => void company.refetch()}
          />
        ) : null}
        {company.data ? (
          isEditing ? (
            <CompanyEditForm
              key={company.data.id}
              company={company.data}
              onFinish={finishEditing}
            />
          ) : (
            <CompanyView company={company.data} />
          )
        ) : null}
      </section>
    </EmployerWorkspaceLayout>
  );
}

function CompanyView({ company }: { company: Company }) {
  const { t } = useTranslation();
  const website = safeProfileUrl(company.websiteUrl);
  return (
    <div className="company-profile-view">
      <div className="company-profile-view__identity">
        <CompanyIdentity
          company={company}
          verifiedLabel={t("companies.verified")}
        />
        <Link className="text-link" to={`/companies/${company.id}`}>
          {t("companyEdit.publicProfile")}
          <ExternalLink size={16} aria-hidden="true" />
        </Link>
      </div>
      <dl className="account-settings__details">
        <div>
          <dt>{t("companyEdit.industry")}</dt>
          <dd>{company.industry || t("preferences.notSpecified")}</dd>
        </div>
        <div>
          <dt>{t("companyEdit.city")}</dt>
          <dd>{company.city || t("preferences.notSpecified")}</dd>
        </div>
        <div>
          <dt>{t("companyEdit.website")}</dt>
          <dd>
            {website ? (
              <a href={website} target="_blank" rel="noopener noreferrer">
                {website}
                <ExternalLink size={15} aria-hidden="true" />
              </a>
            ) : (
              company.websiteUrl || t("preferences.notSpecified")
            )}
          </dd>
        </div>
      </dl>
      <section
        className="company-profile-view__about"
        aria-labelledby="company-about-title"
      >
        <h2 id="company-about-title">{t("companyEdit.about")}</h2>
        <p>{company.description || t("companies.noDescription")}</p>
      </section>
    </div>
  );
}

function CompanyEditForm({
  company,
  onFinish,
}: {
  company: Company;
  onFinish: () => void;
}) {
  const { t } = useTranslation();
  const auth = useAuth();
  const toast = useToast();
  const queryClient = useQueryClient();
  const [initial] = useState<CompanyProfileInput>(() => ({
    name: company.name,
    description: company.description ?? "",
    websiteUrl: company.websiteUrl ?? "",
    logoUrl: company.logoUrl ?? "",
    industry: company.industry ?? "",
    city: company.city ?? "",
  }));
  const [form, setForm] = useState(initial);
  const isDirty = (Object.keys(initial) as (keyof CompanyProfileInput)[]).some(
    (key) => initial[key] !== form[key],
  );
  useEffect(() => {
    if (!isDirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [isDirty]);
  const save = useMutation({
    mutationFn: () =>
      repositories.companies.updateEmployerCompany(
        form,
        auth.session!.csrfToken,
      ),
    onSuccess: (saved) => {
      queryClient.setQueryData(["employer-company"], saved);
      queryClient.setQueryData(["company", saved.id], saved);
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: ["employer-company"] }),
        queryClient.invalidateQueries({ queryKey: ["auth", "me"] }),
        queryClient.invalidateQueries({ queryKey: ["companies"] }),
        queryClient.invalidateQueries({ queryKey: ["followed-companies"] }),
      ]);
      toast.showToast({ title: t("companyEdit.saved") });
      onFinish();
    },
    onError: () =>
      toast.showToast({ title: t("companyEdit.saveError"), tone: "error" }),
  });
  const set = (field: keyof CompanyProfileInput) => (value: string) =>
    setForm((current) => ({ ...current, [field]: value }));
  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (!save.isPending) save.mutate();
  };
  const cancel = () => {
    if (!isDirty || window.confirm(t("companyEdit.discardConfirm"))) onFinish();
  };
  return (
    <form className="company-edit-form" onSubmit={submit}>
      <fieldset className="company-edit-form__fields" disabled={save.isPending}>
        <Input
          label={t("companyEdit.name")}
          autoFocus
          value={form.name}
          minLength={2}
          maxLength={160}
          required
          onChange={(e) => set("name")(e.target.value)}
        />
        <div className="company-edit-form__row">
          <Input
            label={t("companyEdit.industry")}
            value={form.industry}
            maxLength={120}
            onChange={(e) => set("industry")(e.target.value)}
          />
          <Input
            label={t("companyEdit.city")}
            value={form.city}
            maxLength={120}
            onChange={(e) => set("city")(e.target.value)}
          />
        </div>
        <Textarea
          label={t("companyEdit.about")}
          hint={t("companyEdit.aboutHint")}
          value={form.description}
          maxLength={5000}
          rows={8}
          onChange={(e) => set("description")(e.target.value)}
        />
        <Input
          label={t("companyEdit.website")}
          type="url"
          value={form.websiteUrl}
          maxLength={500}
          placeholder="https://"
          onChange={(e) => set("websiteUrl")(e.target.value)}
        />
        <Input
          label={t("companyEdit.logo")}
          type="url"
          value={form.logoUrl}
          maxLength={500}
          placeholder="https://"
          hint={t("companyEdit.logoHint")}
          onChange={(e) => set("logoUrl")(e.target.value)}
        />
      </fieldset>
      {save.isError ? (
        <p className="auth-error" role="alert">
          {t("companyEdit.saveError")}
        </p>
      ) : null}
      <div className="company-edit-form__actions">
        <span>{isDirty ? t("profile.unsaved") : t("profile.upToDate")}</span>
        <div>
          <Button
            type="button"
            variant="quiet"
            leadingIcon={<X size={17} />}
            disabled={save.isPending}
            onClick={cancel}
          >
            {t("preferences.cancel")}
          </Button>
          <Button
            type="submit"
            leadingIcon={<Save size={18} />}
            isLoading={save.isPending}
          >
            {t("companyEdit.save")}
          </Button>
        </div>
      </div>
    </form>
  );
}
