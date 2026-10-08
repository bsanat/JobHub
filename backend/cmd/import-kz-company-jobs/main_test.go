package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jobhub-ai/backend/internal/ingestion"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/kzcompanies"
)

func TestFixtureDryRunDoesNotNeedDatabaseOrEnablement(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "sources.json")
	fixturePath := filepath.Join(dir, "vacancies.json")
	cfg := kzcompanies.Config{Sources: []kzcompanies.Source{{
		ID: "halyk", Company: "Halyk Bank", OfficialWebsite: "https://halykbank.kz/",
		CareersURL: "https://halykbank.kz/about/career/vacancies", Adapter: "halyk",
		SourceType: "static_html", Status: "supported", Country: "Kazakhstan",
		RobotsNote: "Reviewed fixture only", TermsNote: "No live access",
	}}}
	writeJSON := func(path string, value any) {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON(configPath, cfg)
	writeJSON(fixturePath, []jobs.ImportedJob{{
		ExternalID: "123", SourceURL: "https://halykbank.kz/about/career/vacancies/123",
		Title: "Инженер", LocationRaw: "Астана", Description: "Тест", DescriptionKind: "full",
	}})
	var out, logs bytes.Buffer
	getenv := func(string) string { t.Fatal("fixture dry-run accessed an environment/database setting"); return "" }
	err := run([]string{"--config", configPath, "--source", "halyk", "--dry-run", "--fixture", fixturePath}, &out, &logs, getenv)
	if err != nil {
		t.Fatal(err)
	}
	var report ingestion.Report
	if json.Unmarshal(out.Bytes(), &report) != nil || !report.DryRun || report.Stats.NormalizedCount != 1 ||
		report.Stats.InsertedCount != 0 || report.Stats.UpdatedCount != 0 || report.Stats.DeactivatedCount != 0 ||
		len(report.Samples) != 1 || report.Samples[0].ApplicationURL != "https://halykbank.kz/about/career/vacancies/123" {
		t.Fatalf("unexpected fixture report: %s", out.String())
	}
	if !strings.Contains(logs.String(), "source started") || !strings.Contains(logs.String(), "source completed") {
		t.Fatalf("missing lifecycle logs: %s", logs.String())
	}
}

func TestOnceRejectsDisabledSourceBeforeDatabaseAccess(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"--config", filepath.Join("..", "..", "config", "kz-company-sources.json"), "--source", "halyk", "--once"}, &out, &out, func(string) string {
		t.Fatal("disabled import queried environment")
		return ""
	})
	if err == nil || !strings.Contains(err.Error(), "no enabled") {
		t.Fatalf("expected disabled-source failure, got %v", err)
	}
}

func TestProbeRejectsRobotsBlockedSourceWithoutNetwork(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"--config", filepath.Join("..", "..", "config", "kz-company-sources.json"), "--source", "halyk", "--dry-run", "--probe"}, &out, &out, func(string) string {
		t.Fatal("blocked probe queried environment")
		return ""
	})
	if err == nil || !strings.Contains(err.Error(), "no enabled") {
		t.Fatalf("expected blocked probe failure, got %v", err)
	}
}

func TestPreviewSnapshotOnlyDeactivatesAfterCompleteSync(t *testing.T) {
	existing := map[string]string{"same": "active", "missing": "active", "closed": "removed"}
	partial := previewSnapshot([]string{"same", "new"}, existing, false)
	if partial.WouldInsert != 1 || partial.WouldUpdate != 1 || partial.WouldDeactivate != 0 {
		t.Fatalf("unexpected partial preview: %+v", partial)
	}
	complete := previewSnapshot([]string{"same", "new"}, existing, true)
	if complete.WouldDeactivate != 1 {
		t.Fatalf("unexpected complete preview: %+v", complete)
	}
}

func TestDatabaseIdentityRequiresInspectedUserAndHost(t *testing.T) {
	values := map[string]string{
		"APP_ENV": "production", "KZ_IMPORT_TARGET": "production", "KZ_IMPORT_DATABASE_HOST": "db.example.test",
		"KZ_IMPORT_DATABASE_USER": "expected", "KZ_IMPORT_READ_DATABASE_URL": "postgresql://other:secret@db.example.test:5432/postgres",
	}
	_, err := checkedDatabaseURL(func(key string) string { return values[key] }, "KZ_IMPORT_READ_DATABASE_URL")
	if err == nil {
		t.Fatal("different database user must be rejected")
	}
}
