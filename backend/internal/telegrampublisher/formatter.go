package telegrampublisher

import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

var messageURLPattern = regexp.MustCompile(`(?i)\b(?:https?://|www\.)[^\s<>]+`)

func FormatVacancy(v Vacancy, publicBaseURL string) string {
	lines := []string{
		"💼 <b>" + html.EscapeString(shortText(v.Title, 160)) + "</b>",
		"",
	}
	if value := shortText(v.CompanyName, 100); value != "" {
		lines = append(lines, "🏢 Компания: "+html.EscapeString(value))
	}
	if value := shortText(v.Location, 100); value != "" {
		lines = append(lines, "📍 Қала / Город: "+html.EscapeString(value))
	}
	if value := salaryLine(v); value != "" {
		lines = append(lines, "💰 Жалақы / Зарплата: <b>"+html.EscapeString(shortText(value, 120))+"</b>")
	}
	if value := labelValue(v.WorkMode, map[string]string{"remote": "Қашықтан / Удалённо", "hybrid": "Аралас / Гибрид", "on_site": "Кеңседе / В офисе"}); value != "" {
		lines = append(lines, "🧩 Формат: "+html.EscapeString(shortText(value, 80)))
	}
	if value := labelValue(v.ExperienceLevel, map[string]string{"no_experience": "Тәжірибесіз / Без опыта", "junior": "Junior", "middle": "Middle", "senior": "Senior", "lead": "Lead"}); value != "" {
		lines = append(lines, "📌 Тәжірибе / Опыт: "+html.EscapeString(shortText(value, 80)))
	}
	if value := labelValue(v.EmploymentType, map[string]string{"full_time": "Толық жұмыс / Полная занятость", "part_time": "Жарты күн / Частичная занятость", "contract": "Келісімшарт / Контракт", "temporary": "Уақытша / Временная работа", "internship": "Тағылымдама / Стажировка"}); value != "" {
		lines = append(lines, "🕘 Жұмыспен қамту / Занятость: "+html.EscapeString(shortText(value, 80)))
	}
	description := v.Description
	if strings.TrimSpace(publicBaseURL) == "" {
		description = messageURLPattern.ReplaceAllString(description, "")
	}
	if value := shortText(description, 260); value != "" {
		lines = append(lines, "", "📝 Сипаттама / Описание:", html.EscapeString(value))
	}
	if v.Source != "jobhub" {
		if value := shortText(v.SourceName, 100); value != "" {
			lines = append(lines, "", "Дереккөз / Источник: "+html.EscapeString(value))
		}
		attribution := v.Attribution
		if strings.TrimSpace(publicBaseURL) == "" {
			attribution = messageURLPattern.ReplaceAllString(attribution, "")
		}
		if value := shortText(attribution, 180); value != "" {
			lines = append(lines, html.EscapeString(value))
		}
	}
	if strings.TrimSpace(publicBaseURL) != "" {
		lines = append(lines, "", "🔗 <a href=\""+html.EscapeString(vacancyURL(publicBaseURL, v.ID))+"\">Вакансияны ашу / Открыть вакансию</a>")
	}
	return strings.Join(lines, "\n")
}

func shortText(raw string, limit int) string {
	value := []rune(strings.Join(strings.Fields(raw), " "))
	if len(value) <= limit {
		return string(value)
	}
	return string(value[:limit-1]) + "…"
}

func salaryLine(v Vacancy) string {
	if !v.SalaryVisible {
		return ""
	}
	if raw := strings.TrimSpace(v.SalaryRaw); raw != "" {
		return raw
	}
	min, max := normalizeMoney(v.SalaryMin), normalizeMoney(v.SalaryMax)
	switch {
	case min != "" && max != "":
		return min + "–" + max + salarySuffix(v)
	case min != "":
		return min + salarySuffix(v) + " бастап / от"
	case max != "":
		return max + salarySuffix(v) + " дейін / до"
	default:
		return ""
	}
}

func salarySuffix(v Vacancy) string {
	currency := strings.ToUpper(strings.TrimSpace(v.SalaryCurrency))
	if currency == "KZT" {
		currency = "₸"
	}
	period := strings.TrimSpace(v.SalaryPeriod)
	if period == "month" {
		period = "/ай"
	}
	if currency == "" && period == "" {
		return ""
	}
	if currency == "" {
		return " " + period
	}
	if period == "" {
		return " " + currency
	}
	return " " + currency + " " + period
}

func normalizeMoney(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if idx := strings.IndexByte(raw, '.'); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.ReplaceAll(raw, " ", "")
	if raw == "" {
		return ""
	}
	chunks := []string{}
	for len(raw) > 3 {
		chunks = append([]string{raw[len(raw)-3:]}, chunks...)
		raw = raw[:len(raw)-3]
	}
	chunks = append([]string{raw}, chunks...)
	return strings.Join(chunks, " ")
}

func labelValue(raw string, labels map[string]string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if value, ok := labels[raw]; ok {
		return value
	}
	return raw
}

func vacancyURL(publicBaseURL, id string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	return base + "/jobs/" + url.PathEscape(id)
}
