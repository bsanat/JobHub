import { useId, useState, type CSSProperties } from "react";
import { useTranslation } from "react-i18next";
import type { JobSearchParams } from "../../api/models/job";
import { Button } from "../ui/Button";

type Currency = NonNullable<JobSearchParams["currency"]>;
const limits: Record<Currency, number> = { KZT: 3_000_000, USD: 20_000, EUR: 20_000 };

function sliderCeiling(currency: Currency, search: JobSearchParams) {
  const bound = search.currency === currency ? Math.max(search.salaryMin ?? 0, search.salaryMax ?? 0) : 0;
  return Math.max(limits[currency], (Math.floor(bound / limits[currency]) + 1) * limits[currency]);
}

export function SalaryFilter({ search, onApply }: {
  search: JobSearchParams;
  onApply: (patch: Partial<JobSearchParams>) => void;
}) {
  const { t, i18n } = useTranslation();
  const id = useId();
  const [currency, setCurrency] = useState<Currency>(search.currency ?? "KZT");
  const ceiling = sliderCeiling(currency, search);
  const [minimum, setMinimum] = useState(search.salaryMin ?? 0);
  const [maximum, setMaximum] = useState(search.salaryMax ?? ceiling);
  const step = currency === "KZT" ? 10_000 : 100;
  const format = new Intl.NumberFormat(i18n.language);
  const rangeText = `${format.format(minimum)} – ${format.format(maximum)}${maximum === ceiling ? "+" : ""} ${currency}`;

  return (
    <form className="salary-filter" onSubmit={(event) => {
      event.preventDefault();
      const salaryMin = minimum > 0 ? minimum : undefined;
      const salaryMax = maximum < ceiling ? maximum : undefined;
      onApply({ salaryMin, salaryMax, currency: salaryMin != null || salaryMax != null ? currency : undefined });
    }}>
      <div className="salary-filter__heading">
        <strong id={`${id}-label`}>{t("jobs.salaryRange")}</strong>
        <select aria-label={t("profile.fields.currency")} value={currency} onChange={(event) => {
          const next = event.target.value as Currency;
          setCurrency(next);
          setMinimum(0);
          setMaximum(sliderCeiling(next, search));
        }}>
          <option value="KZT">KZT</option><option value="USD">USD</option><option value="EUR">EUR</option>
        </select>
      </div>
      <div className="salary-filter__slider" style={{ "--range-start": `${minimum / ceiling * 100}%`, "--range-end": `${maximum / ceiling * 100}%` } as CSSProperties}>
        <div className="salary-filter__track" aria-hidden="true"><span /></div>
        <input className="salary-filter__minimum" style={{ zIndex: minimum === ceiling ? 1 : undefined }} type="range" aria-label={t("employer.fields.salaryMin")} aria-valuetext={`${format.format(minimum)} ${currency}`} min={0} max={ceiling} step={step} value={minimum} onChange={(event) => setMinimum(Math.min(Number(event.target.value), maximum))} />
        <input className="salary-filter__maximum" type="range" aria-label={t("employer.fields.salaryMax")} aria-valuetext={`${format.format(maximum)}${maximum === ceiling ? "+" : ""} ${currency}`} min={0} max={ceiling} step={step} value={maximum} onChange={(event) => setMaximum(Math.max(Number(event.target.value), minimum))} />
      </div>
      <div className="salary-filter__footer">
        <output aria-labelledby={`${id}-label`} title={rangeText}>
          <span>{format.format(minimum)} –</span>
          <span>{format.format(maximum)}{maximum === ceiling ? "+" : ""} {currency}</span>
        </output>
        <Button type="submit" size="sm">{t("jobs.applyRange")}</Button>
      </div>
    </form>
  );
}
