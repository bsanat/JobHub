import { useTranslation } from "react-i18next";

export function LanguageSelect() {
  const { t, i18n } = useTranslation();
  const language = i18n.resolvedLanguage === "ru" ? "ru" : "kk";
  return (
    <label className="language-select compact-select">
      <span className="visually-hidden">{t("common.language")}</span>
      <span className="compact-select__value" aria-hidden="true">{language === "ru" ? "🇷🇺 Рус" : "🇰🇿 Қаз"}</span>
      <select value={language} onChange={(event) => void i18n.changeLanguage(event.target.value)}>
        <option value="kk">🇰🇿 Қазақша</option>
        <option value="ru">🇷🇺 Русский</option>
      </select>
    </label>
  );
}
