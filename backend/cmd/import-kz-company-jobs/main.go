package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jobhub-ai/backend/internal/database"
	"jobhub-ai/backend/internal/ingestion"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/kzcompanies"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/providers/airastana"
	"jobhub-ai/backend/internal/providers/kcell"
	"jobhub-ai/backend/internal/providers/staticcareers"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out, logs io.Writer, getenv func(string) string) error {
	flags := flag.NewFlagSet("import-kz-company-jobs", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("config", "", "server-side allowlist JSON")
	sourceID := flags.String("source", "", "one configured company ID")
	fixture := flags.String("fixture", "", "offline imported-job JSON array (dry-run only)")
	probe := flags.Bool("probe", false, "explicit read-only source fetch without enabling imports (dry-run only)")
	compareDB := flags.Bool("compare-db", false, "read-only comparison with the inspected target database (dry-run only)")
	dry := flags.Bool("dry-run", false, "collect without database writes or Telegram actions")
	once := flags.Bool("once", false, "run one authorized import and exit")
	activate := flags.Bool("activate", false, "after a successful approved --once, enable public display")
	if flags.Parse(args) != nil || flags.NArg() != 0 || *path == "" || *dry == *once || (*fixture != "" && (!*dry || *sourceID == "" || *probe)) || (*compareDB && (!*dry || *probe)) || (*probe && (!*dry || *sourceID == "")) || (*activate && !*once) {
		return errors.New("usage: import-kz-company-jobs --config <file> [--source <id>] (--dry-run | --once) [--fixture <file> | --probe] [--compare-db] [--activate]")
	}
	cfg, err := kzcompanies.Load(*path)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	selected := make([]kzcompanies.Source, 0, len(cfg.Sources))
	for _, candidate := range cfg.Sources {
		if *sourceID != "" && candidate.ID != *sourceID {
			continue
		}
		if *probe && (candidate.Status != "supported" || !candidate.RobotsReviewed || candidate.Adapter == "") {
			logger.Warn("Kazakhstan source probe refused", "source", candidate.ID, "status", candidate.Status, "robots_note", candidate.RobotsNote)
			continue
		}
		if *fixture == "" && !*probe && (!cfg.Enabled || !candidate.Enabled) {
			if *sourceID != "" {
				logger.Warn("Kazakhstan source access not enabled", "source", candidate.ID, "status", candidate.Status, "robots_note", candidate.RobotsNote, "terms_note", candidate.TermsNote)
			}
			continue
		}
		if candidate.Adapter != "" {
			selected = append(selected, candidate)
		}
	}
	if len(selected) == 0 || (*fixture != "" && len(selected) != 1) {
		return errors.New("no enabled supported Kazakhstan source selected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	var store *jobs.PostgresStore
	var comparisonPool *pgxpool.Pool
	if *compareDB {
		raw, identityErr := checkedDatabaseURL(getenv, "KZ_IMPORT_READ_DATABASE_URL")
		if identityErr != nil {
			return identityErr
		}
		comparisonPool, err = database.Connect(ctx, raw, "exec")
		if err != nil {
			return errors.New("read-only database connection failed")
		}
		defer comparisonPool.Close()
	}
	if *once {
		if getenv("KZ_IMPORT_WRITE_ENABLED") != "true" {
			return errors.New("KZ_IMPORT_WRITE_ENABLED=true is required for --once")
		}
		env := getenv("APP_ENV")
		if (env == "staging" || env == "production") && getenv("KZ_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED") != "true" {
			return errors.New("staging/production permission review is required")
		}
		raw, identityErr := checkedDatabaseURL(getenv, "DATABASE_URL")
		if identityErr != nil {
			return identityErr
		}
		mode := getenv("PGX_QUERY_EXEC_MODE")
		if mode == "" {
			mode = "exec"
		}
		pool, connectErr := database.Connect(ctx, raw, mode)
		if connectErr != nil {
			return errors.New("database connection failed")
		}
		defer pool.Close()
		store = jobs.NewPostgresStore(pool, env == "production")
	}
	encoder := json.NewEncoder(out)
	for _, source := range selected {
		var upstream providers.VacancyProvider
		if *fixture != "" {
			items, fixtureErr := readFixture(*fixture)
			if fixtureErr != nil {
				return fixtureErr
			}
			upstream = fixtureProvider{items: items}
		} else {
			upstream, err = providerFor(source)
			if err != nil {
				return err
			}
		}
		provider := kzcompanies.Provider{SourceConfig: source, Upstream: upstream}
		logger.Info("Kazakhstan source started", "source", provider.Source())
		var target jobs.Store
		if store != nil {
			if err := store.RegisterKZCompanySource(ctx, provider.Source(), source.Company); err != nil {
				return err
			}
			target = store
		}
		report, collectErr := ingestion.NewService(target, nil, logger, 0).Run(ctx, provider, ingestion.RunOptions{DryRun: *dry, Sample: 5, NoAutomaticExpiry: true})
		output := struct {
			ingestion.Report
			Preview *preview `json:"preview,omitempty"`
		}{Report: report}
		if *dry {
			output.Preview = &preview{CandidateCount: len(report.ValidatedIDs), ComparisonAvailable: false}
			if comparisonPool != nil {
				output.Preview, err = compareSnapshot(ctx, comparisonPool, provider.Source(), report.ValidatedIDs, report.CompleteSnapshot)
				if err != nil {
					return err
				}
			}
		}
		if err := encoder.Encode(output); err != nil {
			return errors.New("report output failed")
		}
		logger.Info("Kazakhstan source completed", "source", provider.Source(), "fetched", report.Stats.FetchedCount,
			"imported", report.Stats.InsertedCount, "updated", report.Stats.UpdatedCount,
			"deactivated", report.Stats.DeactivatedCount, "skipped", report.Stats.SkippedCount,
			"parsing_errors", report.Malformed, "failure", report.FailureCategory)
		if collectErr != nil {
			return collectErr
		}
		if *activate && store != nil {
			if err := store.ActivateKZCompanyAfterImport(ctx, provider.Source()); err != nil {
				return err
			}
			logger.Info("Kazakhstan source activated for public display", "source", provider.Source())
		}
	}
	return nil
}

func checkedDatabaseURL(getenv func(string) string, key string) (string, error) {
	env := getenv("APP_ENV")
	if (env != "development" && env != "test" && env != "staging" && env != "production") || env != getenv("KZ_IMPORT_TARGET") {
		return "", errors.New("APP_ENV and KZ_IMPORT_TARGET must explicitly match")
	}
	raw := strings.TrimSpace(getenv(key))
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.User == nil ||
		getenv("KZ_IMPORT_DATABASE_HOST") == "" || u.Hostname() != getenv("KZ_IMPORT_DATABASE_HOST") ||
		getenv("KZ_IMPORT_DATABASE_USER") == "" || u.User.Username() != getenv("KZ_IMPORT_DATABASE_USER") {
		return "", errors.New("database identity does not match inspected target")
	}
	return raw, nil
}

type preview struct {
	CandidateCount      int  `json:"candidate_count"`
	ComparisonAvailable bool `json:"comparison_available"`
	WouldInsert         int  `json:"would_insert,omitempty"`
	WouldUpdate         int  `json:"would_update,omitempty"`
	WouldDeactivate     int  `json:"would_deactivate,omitempty"`
}

func compareSnapshot(ctx context.Context, pool *pgxpool.Pool, source string, ids []string, complete bool) (*preview, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, errors.New("read-only comparison failed")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `SELECT external_id, source_status FROM jobhub.jobs WHERE source=$1`, source)
	if err != nil {
		return nil, errors.New("read-only comparison failed")
	}
	existing := map[string]string{}
	for rows.Next() {
		var id, status string
		if rows.Scan(&id, &status) != nil {
			rows.Close()
			return nil, errors.New("read-only comparison failed")
		}
		existing[id] = status
	}
	if rows.Err() != nil {
		rows.Close()
		return nil, errors.New("read-only comparison failed")
	}
	rows.Close()
	return previewSnapshot(ids, existing, complete), nil
}

func previewSnapshot(ids []string, existing map[string]string, complete bool) *preview {
	p := &preview{CandidateCount: len(ids), ComparisonAvailable: true}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		seen[id] = true
		if _, ok := existing[id]; ok {
			p.WouldUpdate++
		} else {
			p.WouldInsert++
		}
	}
	if complete {
		for id, status := range existing {
			if !seen[id] && (status == "active" || status == "unknown") {
				p.WouldDeactivate++
			}
		}
	}
	return p
}

func providerFor(source kzcompanies.Source) (providers.VacancyProvider, error) {
	switch source.Adapter {
	case "airastana":
		return airastana.NewClient(10, 200, nil)
	case "kcell":
		return kcell.NewClient(5, 200, nil)
	case "halyk", "technodom", "kolesa", "bcc", "sulpak":
		budget := 12
		if source.Adapter == "technodom" {
			budget = 60
		} else if source.Adapter == "bcc" {
			budget = 30
		}
		return staticcareers.New(staticcareers.Config{
			Source: "kzcompany:" + source.ID, Company: source.Company,
			ListingURL: source.CareersURL, DetailPrefix: source.DetailPrefix, LocationMode: source.Adapter,
			ForcePartial: source.Adapter == "bcc" || source.Adapter == "sulpak",
		}, budget, 200, &http.Client{})
	default:
		return nil, errors.New("source has no verified adapter")
	}
}

func readFixture(path string) ([]jobs.ImportedJob, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("cannot read fixture")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (5<<20)+1))
	if err != nil || len(raw) > 5<<20 {
		return nil, errors.New("invalid fixture size")
	}
	var items []jobs.ImportedJob
	if json.Unmarshal(raw, &items) != nil || items == nil {
		return nil, errors.New("invalid fixture JSON")
	}
	return items, nil
}

type fixtureProvider struct{ items []jobs.ImportedJob }

func (fixtureProvider) Source() string { return "fixture" }
func (f fixtureProvider) Collect(context.Context) (providers.Result, error) {
	return providers.Result{Items: f.items, Fetched: len(f.items), Complete: true, PagesFetched: 1}, nil
}
