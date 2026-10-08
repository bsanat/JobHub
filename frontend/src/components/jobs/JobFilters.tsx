import { useTranslation } from "react-i18next";
import type { EmploymentType, ExperienceLevel, JobSearchParams, WorkMode } from "../../api/models/job";
import { CitySelect } from "../location/CitySelect";
import { Button } from "../ui/Button";
import { Checkbox, RadioGroup } from "../ui/FormControls";
import { SalaryFilter } from "./SalaryFilter";

const modes: WorkMode[] = ["on_site", "hybrid", "remote"];
const experiences: ExperienceLevel[] = ["no_experience", "junior", "middle", "senior", "lead"];
const employments: EmploymentType[] = ["full_time", "part_time", "contract", "temporary", "internship"];

export function JobFilters({ search, onChange, onReset }: { search: JobSearchParams; onChange: (next: JobSearchParams) => void; onReset: () => void }) {
  const { t } = useTranslation();
  const update = (patch: Partial<JobSearchParams>) => onChange({ ...search, ...patch, page: 1 });
  const toggleMode = (mode: WorkMode, checked: boolean) => update({ workModes: checked ? [...(search.workModes ?? []), mode] : (search.workModes ?? []).filter((item) => item !== mode) });
  return (
    <div className="job-filters">
      <RadioGroup name="region" label={t("jobs.regionLabel")} value={search.region ?? ""} onChange={(value) => update({ region: (value || undefined) as JobSearchParams["region"], city: undefined })} options={[{ value: "", label: t("jobs.anyOption") }, { value: "kazakhstan", label: t("jobs.regionKazakhstan") }, { value: "international", label: t("jobs.regionInternational") }]} />
      {search.region === "international" ? <p className="filter-region-note">{t("jobs.internationalNote")}</p> : <CitySelect label={t("search.locationLabel")} anyLabel={t("jobs.anyLocation")} value={search.city} onChange={(city) => update({ city: city || undefined })} />}
      <SalaryFilter key={`${search.salaryMin ?? ""}:${search.salaryMax ?? ""}:${search.currency ?? ""}`} search={search} onApply={update} />
      <fieldset className="filter-fieldset"><legend>{t("jobs.workMode")}</legend>{modes.map((mode) => <Checkbox key={mode} label={t(`employer.workModes.${mode}`)} checked={search.workModes?.includes(mode) ?? false} onChange={(event) => toggleMode(mode, event.target.checked)} />)}</fieldset>
      <RadioGroup name="employment" label={t("employer.fields.employmentType")} value={search.employment ?? ""} onChange={(value) => update({ employment: (value || undefined) as EmploymentType | undefined })} options={[{ value: "", label: t("jobs.anyOption") }, ...employments.map((value) => ({ value, label: t(`employer.employmentTypes.${value}`) }))]} />
      <RadioGroup name="experience" label={t("employer.fields.experienceLevel")} value={search.experience ?? ""} onChange={(value) => update({ experience: (value || undefined) as ExperienceLevel | undefined })} options={[{ value: "", label: t("jobs.anyOption") }, ...experiences.map((value) => ({ value, label: t(`employer.experienceLevels.${value}`) }))]} />
      <RadioGroup name="date" label={t("jobs.datePosted")} value={search.datePosted ?? ""} onChange={(value) => update({ datePosted: (value || undefined) as JobSearchParams["datePosted"] })} options={[{ value: "", label: t("jobs.anyTime") }, ...["24h", "3d", "7d", "30d"].map((value) => ({ value, label: t(`jobs.dateOptions.${value}`) }))]} />
      <Button type="button" variant="secondary" onClick={onReset}>{t("jobs.reset")}</Button>
    </div>
  );
}
