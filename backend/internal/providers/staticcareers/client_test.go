package staticcareers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func htmlResponse(r *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/html"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

func TestCollectUsesCanonicalURLHashAndExternalFields(t *testing.T) {
	client, err := New(Config{Source: "halyk:careers", Company: "Halyk Bank", ListingURL: "https://example.test/careers", DetailPrefix: "/jobs/"}, 3, 10, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/careers" {
			return htmlResponse(r, `<a href="/jobs/123">job</a>`), nil
		}
		return htmlResponse(r, `<html><h1>Engineer</h1><main>Казахстан, Алматы, KZ. Description.</main></html>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Requests != 2 || !result.Complete || len(result.Items) != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	item := result.Items[0]
	if !strings.HasPrefix(item.ExternalID, "url_") || item.SourceURL != "https://example.test/jobs/123" || item.CompanyNameRaw != "Halyk Bank" || item.LocationRaw != "Казахстан, Алматы, KZ" || item.Title != "Engineer" {
		t.Fatalf("item=%+v", item)
	}
}

func TestOfficialCareerLocationPatterns(t *testing.T) {
	for _, tc := range []struct{ mode, html, location, salary, title string }{
		{"technodom", `<html><head><title>Old title</title></head><body><h1>Оператор</h1><main><h2 class="col-cities-col-item-city">Алматы</h2><p>Текст</p></main></body></html>`, "Алматы", "", "Оператор"},
		{"technodom", `<h1>Кассир</h1><main><h2 class="col-cities-col-item-city">Актау</h2><h2 class="col-cities-col-item-city">Астана</h2></main>`, "Қазақстан — бірнеше қала", "", "Кассир"},
		{"kolesa", `<html><head><title>Карьера</title></head><body><h1>Разработчик</h1><div class="_salary_abc">1 000 000 ₸</div><ul class="_tags_xyz"><li><span></span>Алматы</li><li>Опыт</li></ul><main>Работа</main></body></html>`, "Алматы", "1 000 000 ₸", "Разработчик"},
	} {
		item, ok := normalize(Config{Source: "fixture", LocationMode: tc.mode}, "https://example.test/job/1", []byte(tc.html))
		if !ok || item.LocationRaw != tc.location || item.SalaryRaw != tc.salary || item.Title != tc.title {
			t.Errorf("%s: unexpected job %+v (ok=%v)", tc.mode, item, ok)
		}
	}
}

func TestTechnodomDescriptionExcludesScriptsAndApplicationForm(t *testing.T) {
	body := `<html><head><script>trackUser()</script></head><body><h1>Кассир</h1><div class="row"><h2 class="col-description">Работа с клиентами.</h2></div><div class="row"><h2 class="col-title">Требования</h2><div class="tip">Внимательность.</div></div><div class="row"><h2 class="col-cities-col-item-city">Алматы</h2></div><form>Телефон кандидата</form></body></html>`
	item, ok := normalize(Config{Source: "fixture", LocationMode: "technodom"}, "https://example.test/job/2", []byte(body))
	if !ok || !strings.Contains(item.Description, "## Требования\n- Внимательность.") || strings.Contains(item.Description, "trackUser") || strings.Contains(item.Description, "Телефон кандидата") {
		t.Fatalf("unsafe or unstructured description: %+v", item)
	}
}

func TestCollectStopsAtBudgetAndDoesNotClaimCompleteSnapshot(t *testing.T) {
	client, err := New(Config{Source: "technodom:careers", ListingURL: "https://example.test/careers", DetailPrefix: "/jobs/"}, 2, 10, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/careers" {
			return htmlResponse(r, `<a href="/jobs/1">one</a><a href="/jobs/2">two</a>`), nil
		}
		return htmlResponse(r, `<h1>Job</h1>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Requests != 2 || result.Complete || result.Skipped != 1 || len(result.Items) != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestBankCenterCreditCollectsOnlyNumericVacancyDetails(t *testing.T) {
	client, err := New(Config{Source: "kzcompany:bcc", Company: "Bank CenterCredit", ListingURL: "https://www.bcc.kz/en/career/vacancies/", DetailPrefix: "/en/career/", LocationMode: "bcc", ForcePartial: true}, 5, 20, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/en/career/vacancies/" {
			return htmlResponse(r, `<a href="/en/career/41">job</a><a href="/en/career/vacancies/">listing</a><a href="/en/career/story">story</a>`), nil
		}
		return htmlResponse(r, `<html><body><div><h1>Специалист по рискам</h1><div class="text-sm mb-2">Almaty</div><div class="text-neutral-700"><h3>Обязанности</h3><p>Проверять операции.</p><ul><li>Работать с отчётами.</li></ul></div></div><form>Личные данные</form></body></html>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || len(result.Items) != 1 || result.Items[0].LocationRaw != "Almaty" || result.Items[0].Title != "Специалист по рискам" || result.Items[0].SourceURL != "https://www.bcc.kz/en/career/41/" || !strings.Contains(result.Items[0].Description, "## Обязанности\nПроверять операции.\n- Работать с отчётами.") || strings.Contains(result.Items[0].Description, "Личные данные") {
		t.Fatalf("unexpected BCC result: %+v err=%v", result, err)
	}
}

func TestSulpakOfficePreservesCityAndOriginalApplicationURL(t *testing.T) {
	client, err := New(Config{Source: "kzcompany:sulpak", Company: "Sulpak", ListingURL: "https://job.sulpak.kz/WorkInCompany/Office", DetailPrefix: "/WorkInCompany/ViewVacancy/", LocationMode: "sulpak", ForcePartial: true}, 5, 20, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/WorkInCompany/Office" {
			return htmlResponse(r, `<a href="/WorkInCompany/ViewVacancy/33301">job</a>`), nil
		}
		return htmlResponse(r, `<main><div id="vacancy"><div><strong>Оператор Call-Центра, Алматы</strong></div><div class="vacancy-info"><p>Обучение с нуля.</p><ul><li>Гибкий график.</li></ul></div><form>Телефон кандидата</form></div></main>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || len(result.Items) != 1 {
		t.Fatalf("unexpected Sulpak result: %+v err=%v", result, err)
	}
	item := result.Items[0]
	if item.Title != "Оператор Call-Центра" || item.LocationRaw != "Алматы" || item.SourceURL != "https://job.sulpak.kz/WorkInCompany/ViewVacancy/33301" || strings.Contains(item.Description, "Телефон кандидата") || !strings.Contains(item.Description, "- Гибкий график.") {
		t.Fatalf("unexpected Sulpak vacancy: %+v", item)
	}
}

func TestMagnumOfficeExtractsOnlyVacancyContent(t *testing.T) {
	listing := "https://hr.magnum.kz/office/"
	detail := "https://hr.magnum.kz/vacancy-office/manager-42/"
	client, err := New(Config{Source: "kzcompany:magnum", Company: "Magnum", ListingURL: listing, DetailPrefix: "/vacancy-office/", LocationMode: "magnum", ForcePartial: true}, 2, 6, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() == listing {
			return htmlResponse(r, `<a href="/vacancy-office/manager-42/">Менеджер</a>`), nil
		}
		if r.URL.String() != detail {
			t.Fatalf("unexpected detail URL: %s", r.URL)
		}
		return htmlResponse(r, `<h1 class="vd-post">Менеджер</h1><p class="vd-right-city">Алматы</p>
            <div class="vd-terms-section"><h2>Условия</h2><ul><li>5/2</li></ul></div>
            <div class="vd-unorderedList-section"><h2>Задачи</h2><ul><li>Работать с заказами</li></ul></div>
            <form>Телефон кандидата<script>tracking()</script></form>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || len(result.Items) != 1 {
		t.Fatalf("unexpected Magnum collection: %+v, %v", result, err)
	}
	job := result.Items[0]
	if job.Title != "Менеджер" || job.LocationRaw != "Алматы" || job.SourceURL != detail ||
		!strings.Contains(job.Description, "## Задачи\n- Работать с заказами") ||
		strings.Contains(job.Description, "Телефон кандидата") || strings.Contains(job.Description, "tracking") {
		t.Fatalf("Magnum vacancy content is incorrect: %+v", job)
	}
}

func TestQICHostedPortalPreservesOriginalApplicationURL(t *testing.T) {
	listing := "https://jobs.baiterek.gov.kz/kompanii/?ELEMENT_ID=757"
	detail := "https://jobs.baiterek.gov.kz/vakansii/107749/"
	client, err := New(Config{Source: "kzcompany:qic", Company: "Qazaqstan Investment Corporation", ListingURL: listing, DetailPrefix: "/vakansii/", LocationMode: "qic", ForcePartial: true}, 2, 6, &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() == listing {
			return htmlResponse(r, `<a href="/vakansii/">Вакансии</a><a href="/vakansii/107749/">Главный менеджер</a>`), nil
		}
		if r.URL.String() != detail {
			t.Fatalf("unexpected detail URL: %s", r.URL)
		}
		return htmlResponse(r, `<h1 class="page_title">Главный менеджер</h1><div class="detail_top"><div class="city">Астана</div></div>
			<div class="row detail_text_wrap"><div class="detail_text p-40"><h3>Обязанности</h3><ul><li>Вести отчётность.</li></ul></div></div>
			<form>Личные данные<script>track()</script></form>`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || len(result.Items) != 1 {
		t.Fatalf("unexpected QIC collection: %+v, %v", result, err)
	}
	job := result.Items[0]
	if job.Title != "Главный менеджер" || job.LocationRaw != "Астана" || job.SourceURL != detail ||
		!strings.Contains(job.Description, "## Обязанности\n- Вести отчётность.") ||
		strings.Contains(job.Description, "Личные данные") || strings.Contains(job.Description, "track") {
		t.Fatalf("QIC vacancy content is incorrect: %+v", job)
	}
}
