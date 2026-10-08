package publicapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func jsonResponse(body string, status int) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestArbeitnowPaginationNormalizationAndPartialSnapshot(t *testing.T) {
	requests := 0
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.URL.Query().Get("page") == "2" {
			return jsonResponse(`{"data":[{"slug":"second","title":"Remote support","company_name":"Example","url":"https://www.arbeitnow.com/jobs/second","location":"Germany only","remote":true,"description":"&lt;p&gt;Germany only&lt;/p&gt;"}],"links":{"next":null}}`, 200), nil
		}
		return jsonResponse(`{"data":[{"slug":"first","title":"Engineer","company_name":"Example","url":"https://www.arbeitnow.com/jobs/first","location":"Berlin","job_types":["Part Time"],"description":"<script>alert(1)</script><p>Build products</p>"}],"links":{"next":"https://www.arbeitnow.com/api/job-board-api?page=2"}}`, 200), nil
	})}
	p, err := NewArbeitnow(2, client)
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Collect(context.Background())
	if err != nil || !r.Complete || r.PagesFetched != 2 || requests != 2 || len(r.Items) != 2 {
		t.Fatalf("unexpected pagination result: %+v %v", r, err)
	}
	if strings.Contains(r.Items[0].Description, "alert") || r.Items[0].EmploymentType != "part_time" || r.Items[1].WorkMode != "remote" || !strings.Contains(r.Items[1].Description, "Germany only") || r.Items[1].ExternalPublishedAt != nil {
		t.Fatalf("normalization lost safety or eligibility: %+v", r.Items)
	}
	p.maxPages = 1
	r, err = p.Collect(context.Background())
	if err != nil || r.Complete || len(r.Items) != 1 {
		t.Fatalf("page cap must be partial: %+v %v", r, err)
	}
}

func TestArbeitnowDescriptionKeepsFormatting(t *testing.T) {
	job, ok := normalizeArbeitnow(arbeitnowJob{Slug: "formatted", Title: "Engineer", URL: "https://www.arbeitnow.com/jobs/formatted", Description: "<h2>Requirements</h2><ul><li>EU work permit</li><li>German B2</li></ul>"})
	if !ok || !strings.Contains(job.Description, "## Requirements") || !strings.Contains(job.Description, "- EU work permit") {
		t.Fatalf("description structure lost: %+v", job)
	}
}

func TestLeverPaginationAndMalformedRecord(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("skip") == "2" {
			return jsonResponse(`[{"id":"3","text":"Intern","applyUrl":"https://jobs.lever.co/example/3/apply","categories":{"location":"Paris","commitment":"Intern"}}]`, 200), nil
		}
		return jsonResponse(`[{"id":"1","text":"Engineer","applyUrl":"https://jobs.lever.co/example/1/apply","description":"<p>EU applicants only</p>","lists":[{"text":"Requirements","content":"<li>Must live in EU</li>"}],"workplaceType":"remote","categories":{"location":"EU","commitment":"Full-time"}},{"id":"2","text":"Broken"}]`, 200), nil
	})}
	p, err := NewLever("example", "Example Ltd", 3, client)
	if err != nil {
		t.Fatal(err)
	}
	p.pageSize = 2
	r, err := p.Collect(context.Background())
	if err != nil || r.Complete || r.Malformed != 1 || r.PagesFetched != 2 || len(r.Items) != 2 {
		t.Fatalf("malformed page must not be complete: %+v %v", r, err)
	}
	if r.Items[0].CompanyNameRaw != "Example Ltd" || r.Items[0].WorkMode != "remote" || !strings.Contains(r.Items[0].Description, "Must live in EU") || r.Items[0].EmploymentType != "full_time" || r.Items[1].EmploymentType != "internship" || r.Items[0].ExternalPublishedAt != nil {
		t.Fatalf("incorrect Lever normalization: %+v", r.Items[0])
	}
}

func TestRetryRateLimitAndPartialFailure(t *testing.T) {
	requests := 0
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return jsonResponse(`{}`, 429), nil
		}
		if requests == 2 {
			return jsonResponse(`{"data":[],"links":{"next":"https://www.arbeitnow.com/api/job-board-api?page=2"}}`, 200), nil
		}
		return jsonResponse(`{}`, 500), nil
	})}
	p, _ := NewArbeitnow(3, client)
	r, err := p.Collect(context.Background())
	if err == nil || r.Complete || r.Requests != 5 || r.PagesFetched != 1 {
		t.Fatalf("partial failure must remain incomplete: %+v %v", r, err)
	}
}
