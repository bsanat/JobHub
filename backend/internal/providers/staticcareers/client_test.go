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
