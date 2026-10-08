package staticcareers

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const qsamrukFixtureList = `<html><div class="accordeon allopen">
<a class="vacancy-item" href="/vacancy/123"><div class="salary">от 400 000 KZT</div></a>
<a class="vacancy-item" href="/vacancy/123">same job in another category</a>
<a class="vacancy-item" href="/vacancy/124">another direct job</a>
<a class="vacancy-item" href="https://evil.example/vacancy/999">foreign host</a>
</div><div class="accordeon allopen"><a class="vacancy-item" href="/vacancy/125">subsidiary</a></div></html>`

func qsamrukFixtureDetail(employer, status, city string) string {
	return `<html><section class="vacancy"><div class="vacancy_header"><h1>Инженер</h1>` +
		`<a class="comp" href="` + employer + `">АО «KEGOC»</a><a class="city">` + city + `</a>` +
		`<div class="extra_info"><span>Без опыта</span><span>Полная занятость</span></div>` +
		`<div class="vacancy_info"><div class="status">` + status + `</div></div></div>` +
		`<div class="vacancy_content"><div class="descr"><h1>Обязанности</h1>` +
		`<ul><li>Работать с оборудованием</li></ul><script>bad()</script></div></div></section></html>`
}

func TestQSamrukListingsDeduplicatesAndExcludesSubsidiaries(t *testing.T) {
	got := qsamrukListings([]byte(qsamrukFixtureList), "https://qsamruk.kz/company/32")
	if len(got) != 2 || got[0].url != "https://qsamruk.kz/vacancy/123" || got[0].salary != "от 400 000 KZT" || got[1].url != "https://qsamruk.kz/vacancy/124" {
		t.Fatalf("unexpected direct vacancies: %+v", got)
	}
}

func TestQSamrukNormalizesAndGuardsEmployerStatusLocation(t *testing.T) {
	cfg := Config{Source: "kzcompany:kegoc", Company: "KEGOC", EmployerPath: "/company/ao-kegoc"}
	listing := qsamrukListing{url: "https://qsamruk.kz/vacancy/123", salary: "от 400 000 KZT"}
	item, ok := qsamrukVacancy(cfg, listing, []byte(qsamrukFixtureDetail("/company/ao-kegoc", "Вакансия открыта", "г.Астана")))
	if !ok || item.ExternalID != "123" || item.SourceURL != listing.url || item.LocationRaw != "г.Астана" || item.SalaryRaw != listing.salary || item.EmploymentTypeRaw != "Полная занятость" || !strings.Contains(item.Description, "## Обязанности") || strings.Contains(item.Description, "bad()") {
		t.Fatalf("unexpected normalized vacancy: %+v (ok=%v)", item, ok)
	}
	for _, tc := range []struct{ employer, status, city string }{
		{"/company/other", "Вакансия открыта", "г.Астана"},
		{"/company/ao-kegoc", "Вакансия закрыта", "г.Астана"},
		{"/company/ao-kegoc", "Вакансия открыта", "Unknownland"},
	} {
		if _, ok := qsamrukVacancy(cfg, listing, []byte(qsamrukFixtureDetail(tc.employer, tc.status, tc.city))); ok {
			t.Fatalf("accepted invalid vacancy: %+v", tc)
		}
	}
}

func TestQSamrukPartialFetchNeverReportsComplete(t *testing.T) {
	client, err := New(Config{Source: "kzcompany:kegoc", Company: "KEGOC", ListingURL: "https://qsamruk.kz/company/32", DetailPrefix: "/vacancy/", EmployerPath: "/company/ao-kegoc", LocationMode: "qsamruk", ForcePartial: true}, 4, 10,
		&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/company/32":
				return htmlResponse(r, qsamrukFixtureList), nil
			case "/vacancy/123":
				return htmlResponse(r, qsamrukFixtureDetail("/company/ao-kegoc", "Вакансия открыта", "г.Астана")), nil
			default:
				return &http.Response{StatusCode: 503, Body: http.NoBody, Request: r}, nil
			}
		})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || result.Fetched != 2 || len(result.Items) != 1 || result.Skipped != 1 {
		t.Fatalf("partial source must preserve existing jobs: %+v (%v)", result, err)
	}
}
