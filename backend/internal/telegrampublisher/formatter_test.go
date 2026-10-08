package telegrampublisher

import (
	"strings"
	"testing"
)

func TestFormatVacancy(t *testing.T) {
	msg := FormatVacancy(Vacancy{
		ID: "11111111-1111-4111-8111-111111111111", Title: "Backend Engineer", CompanyName: "Kaspi.kz",
		Location: "Almaty", WorkMode: "hybrid", ExperienceLevel: "middle", SalaryMin: "700000.00", SalaryMax: "1200000.00", SalaryCurrency: "kzt", SalaryPeriod: "month", SalaryVisible: true,
	}, "https://jobhub.kz")
	for _, want := range []string{"<b>Backend Engineer</b>", "Kaspi.kz", "700 000–1 200 000 ₸ /ай", "Аралас / Гибрид", "Middle", "https://jobhub.kz/jobs/11111111-1111-4111-8111-111111111111"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing %q:\n%s", want, msg)
		}
	}
}

func TestFormatVacancyEscapesEmployerText(t *testing.T) {
	msg := FormatVacancy(Vacancy{ID: "id", Title: "Go <script>", CompanyName: "A & B", SalaryRaw: "<private>", SalaryVisible: true}, "https://jobhub.kz")
	for _, want := range []string{"Go &lt;script&gt;", "A &amp; B", "&lt;private&gt;"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing escaped %q: %s", want, msg)
		}
	}
}

func TestFormatVacancyHidesInvisibleSalary(t *testing.T) {
	msg := FormatVacancy(Vacancy{ID: "id", Title: "Backend Engineer", SalaryRaw: "1 000 000 ₸", SalaryVisible: false}, "https://jobhub.kz")
	if strings.Contains(msg, "1 000 000") {
		t.Fatalf("hidden salary leaked:\n%s", msg)
	}
}

func TestFormatVacancyWithoutPublicSiteLink(t *testing.T) {
	msg := FormatVacancy(Vacancy{ID: "job-id", Title: "Backend Engineer", Description: "Build <secure> APIs https://example.test/apply and www.example.test/jobs"}, "")
	if strings.Contains(msg, "/jobs/") || strings.Contains(msg, "localhost") || strings.Contains(msg, "example.test") || !strings.Contains(msg, "Build &lt;secure&gt; APIs") || !strings.Contains(msg, "Сипаттама / Описание") {
		t.Fatalf("unexpected linkless post: %s", msg)
	}
}

func TestFormatExternalVacancyCreditsSourceWithoutLink(t *testing.T) {
	msg := FormatVacancy(Vacancy{Source: "approved:external", SourceName: "Test & Partner", Attribution: "With permission <2026>", Title: "Engineer"}, "")
	for _, want := range []string{"Дереккөз / Источник: Test &amp; Partner", "With permission &lt;2026&gt;"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing %q: %s", want, msg)
		}
	}
	if strings.Contains(msg, "http") {
		t.Fatalf("linkless external message contains a URL: %s", msg)
	}
}

func TestFormatPublicImportLinksOriginalAndPreservesRemoteCaution(t *testing.T) {
	msg := FormatVacancy(Vacancy{ID: "job-id", Source: "arbeitnow:global", SourceName: "Arbeitnow", OriginalURL: "https://www.arbeitnow.com/jobs/role?a=1&b=2", Title: "Remote Engineer", WorkMode: "remote"}, "https://jobhub.kz")
	for _, want := range []string{"Дереккөз / Источник: Arbeitnow", `href="https://www.arbeitnow.com/jobs/role?a=1&amp;b=2"`, `href="https://www.arbeitnow.com"`, "Проверьте страну и условия отклика"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("public import message missing %q: %s", want, msg)
		}
	}
	if strings.Contains(FormatVacancy(Vacancy{Source: "lever:board", OriginalURL: "javascript:alert(1)", Title: "Engineer"}, "https://jobhub.kz"), "javascript:") {
		t.Fatal("unsafe original URL was linked")
	}
}
