package kzcompanies

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
)

type fakeProvider struct {
	result providers.Result
	err    error
}

func (fakeProvider) Source() string { return "upstream" }
func (f fakeProvider) Collect(context.Context) (providers.Result, error) {
	return f.result, f.err
}

func TestKazakhstanEligibility(t *testing.T) {
	for _, tc := range []struct {
		location string
		want     bool
	}{
		{"Алматы", true},
		{"Astana, Kazakhstan", true},
		{"Қашықтан — Қазақстан", true},
		{"Талдықорған", true},
		{"Remote — Kazakhstan residents only", true},
		{"Remote", false},
		{"Worldwide remote", false},
		{"Berlin, Germany", false},
		{"Almaty, Germany", false},
		{"", false},
	} {
		if got := KazakhstanEligible(tc.location); got != tc.want {
			t.Errorf("location %q: got %v, want %v", tc.location, got, tc.want)
		}
	}
}

func TestRemoteRequiresExplicitKazakhstanEligibility(t *testing.T) {
	if !kazakhstanRemoteEligible("Remote", "Applicants must be based in Kazakhstan.") {
		t.Fatal("explicit Kazakhstan remote eligibility should be accepted")
	}
	if kazakhstanRemoteEligible("Remote worldwide", "Our company has an office in Kazakhstan.") {
		t.Fatal("an office address is not applicant eligibility")
	}
}

func TestProviderKeepsApplicationParametersAndNeverCompletesEmptySnapshot(t *testing.T) {
	cfg := Source{ID: "halyk", Company: "Halyk Bank", CareersURL: "https://halykbank.kz/about/career/vacancies"}
	url := "https://halykbank.kz/about/career/vacancies/42?application=abc&utm_source=listing"
	item := jobs.ImportedJob{ExternalID: "42", SourceURL: url, Title: "Analyst", LocationRaw: "Алматы", DescriptionKind: "full"}
	result, err := (Provider{SourceConfig: cfg, Upstream: fakeProvider{result: providers.Result{Complete: true, Fetched: 1, Items: []jobs.ImportedJob{item}}}}).Collect(context.Background())
	if err != nil || len(result.Items) != 1 || result.Items[0].SourceURL != url {
		t.Fatalf("original application URL changed: %+v, %v", result, err)
	}
	empty, err := (Provider{SourceConfig: cfg, Upstream: fakeProvider{result: providers.Result{Complete: true}}}).Collect(context.Background())
	if err != nil || empty.Complete {
		t.Fatalf("empty snapshot must not reconcile: %+v, %v", empty, err)
	}
}

func TestProviderSkipsOnlyExactLegacyOriginalURLs(t *testing.T) {
	cfg := Source{ID: "kcell", Company: "Kcell", CareersURL: "https://jobs.kcell.kz/"}
	duplicate := jobs.ImportedJob{ExternalID: "1", SourceURL: "https://jobs.kcell.kz/job/1", Title: "Engineer", LocationRaw: "Алматы", DescriptionKind: "full"}
	unique := jobs.ImportedJob{ExternalID: "2", SourceURL: "https://jobs.kcell.kz/job/2", Title: "Analyst", LocationRaw: "Астана", DescriptionKind: "full"}
	result, err := (Provider{SourceConfig: cfg, Upstream: fakeProvider{result: providers.Result{Fetched: 2, Complete: true, Items: []jobs.ImportedJob{duplicate, unique}}}, ExcludedOriginalURLs: map[string]bool{duplicate.SourceURL: true}}).Collect(context.Background())
	if err != nil || !result.Complete || result.Skipped != 1 || len(result.Items) != 1 || result.Items[0].SourceURL != unique.SourceURL {
		t.Fatalf("legacy duplicate filtering changed unique jobs or snapshot safety: %+v, %v", result, err)
	}
}

func TestProviderFiltersDeduplicatesAndSanitizes(t *testing.T) {
	cfg := Source{ID: "halyk", Company: "Halyk Bank", CareersURL: "https://halykbank.kz/about/career/vacancies"}
	sourceURL := "https://halykbank.kz/about/career/vacancies/42"
	makeItem := func(id, rawURL, place, description string) jobs.ImportedJob {
		return jobs.ImportedJob{Source: "upstream", ExternalID: id, SourceURL: rawURL, Title: "Аналитик",
			LocationRaw: place, Description: description, DescriptionKind: "full"}
	}
	upstream := fakeProvider{result: providers.Result{Complete: true, Fetched: 5, Items: []jobs.ImportedJob{
		makeItem("42", sourceURL, "Алматы", "<h2>Міндеттер</h2><script>steal()</script><ul><li>SQL</li></ul>"),
		makeItem("duplicate", sourceURL+"?utm_source=feed", "Алматы", "repeat"),
		makeItem("foreign", "https://halykbank.kz/about/career/vacancies/43", "Berlin", "foreign"),
		makeItem("unknown", "https://halykbank.kz/about/career/vacancies/44", "", "unknown"),
		makeItem("offsite", "https://example.com/job/45", "Алматы", "offsite"),
	}}}
	result, err := (Provider{SourceConfig: cfg, Upstream: upstream}).Collect(context.Background())
	if err != nil || result.Complete || len(result.Items) != 1 || result.Skipped != 4 {
		t.Fatalf("unexpected filtered result: %+v, %v", result, err)
	}
	job := result.Items[0]
	if job.Source != "kzcompany:halyk" || job.CompanyNameRaw != "Halyk Bank" || job.SourceURL != sourceURL ||
		!strings.Contains(job.Description, "## Міндеттер") || strings.Contains(job.Description, "steal") {
		t.Fatalf("source, original URL or sanitized description invalid: %+v", job)
	}
}

func TestConfigIsDisabledAndRequiresReviewedAccess(t *testing.T) {
	path := filepath.Join("..", "..", "config", "kz-company-sources.json")
	cfg, err := Load(path)
	if err != nil || len(cfg.Sources) != 30 || cfg.Enabled {
		t.Fatalf("invalid checked-in allowlist: %v, count=%d", err, len(cfg.Sources))
	}
	for _, source := range cfg.Sources {
		if source.Enabled || source.PermissionsConfirmed {
			t.Fatalf("source unexpectedly enabled: %s", source.ID)
		}
	}
	cfg.Enabled = true
	cfg.Sources[0].Enabled = true
	if cfg.Validate() == nil {
		t.Fatal("unreviewed source must not be enabled")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
