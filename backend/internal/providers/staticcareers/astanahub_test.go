package staticcareers

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const astanaHubFixtureListing = `<div class="vacancy-item" onclick="window.location.href='https://astanahub.com/ru/vacancy/6348'"></div>
<div class="vacancy-item" onclick="window.location.href='https://astanahub.com/ru/vacancy/6348'"></div>
<div class="vacancy-item" onclick="window.location.href='https://evil.example/ru/vacancy/9'"></div>`

func astanaHubFixtureDetail(employer, status, city string) string {
	return `<html><h1>Бизнес аналитик</h1><div class="pills"><span>` + status + `</span><span>` + city + `</span><span>Полная занятость</span></div>` +
		`<div class="tiny-content-v2"><h3>Обязанности</h3><ul><li>Анализировать требования</li></ul><script>bad()</script></div>` +
		`<a href="/account/v2/company/10658/profile/activity/"><span>` + employer + `</span></a>` +
		`<span x-text="displayDuration('2026-10-08T17:39:44+05:00')"></span></html>`
}

func TestAstanaHubOnlyAcceptsOfficialOpenEmployer(t *testing.T) {
	links := astanaHubListings([]byte(astanaHubFixtureListing))
	if len(links) != 1 || links[0] != "https://astanahub.com/ru/vacancy/6348" {
		t.Fatalf("unexpected detail links: %v", links)
	}
	cfg := Config{Source: "kzcompany:astana_hub", Company: "Astana Hub"}
	item, ok := astanaHubVacancy(cfg, links[0], []byte(astanaHubFixtureDetail(astanaHubEmployerName, "Опубликовано", "г.Астана")))
	if !ok || item.ExternalID != "6348" || item.SourceURL != links[0] || item.LocationRaw != "г.Астана" || item.ExternalPublishedAt == nil || !strings.Contains(item.Description, "## Обязанности") || strings.Contains(item.Description, "bad()") {
		t.Fatalf("unexpected normalized vacancy: %+v (ok=%v)", item, ok)
	}
	for _, tc := range []struct{ employer, status, city string }{
		{"Another marketplace employer", "Опубликовано", "г.Астана"},
		{astanaHubEmployerName, "Закрыто", "г.Астана"},
		{astanaHubEmployerName, "Опубликовано", "Berlin"},
	} {
		if _, ok := astanaHubVacancy(cfg, links[0], []byte(astanaHubFixtureDetail(tc.employer, tc.status, tc.city))); ok {
			t.Fatalf("accepted invalid marketplace vacancy: %+v", tc)
		}
	}
}

func TestAstanaHubPartialListingNeverDeactivates(t *testing.T) {
	client, err := New(Config{Source: "kzcompany:astana_hub", Company: "Astana Hub", ListingURL: "https://astanahub.com/ru/vacancy/?company_id=10658&opened=True", DetailPrefix: "/ru/vacancy/", LocationMode: "astanahub", ForcePartial: true}, 2, 10,
		&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/ru/vacancy/" {
				return htmlResponse(r, astanaHubFixtureListing), nil
			}
			return htmlResponse(r, astanaHubFixtureDetail(astanaHubEmployerName, "Опубликовано", "г.Астана")), nil
		})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Collect(context.Background())
	if err != nil || result.Complete || result.Fetched != 1 || len(result.Items) != 1 {
		t.Fatalf("paginated source must remain partial: %+v (%v)", result, err)
	}
}
