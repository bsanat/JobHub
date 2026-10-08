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
	"regexp"
	"strings"
	"time"

	"jobhub-ai/backend/internal/database"
	"jobhub-ai/backend/internal/ingestion"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/providers/greenhouse"
	"jobhub-ai/backend/internal/providers/publicapi"
)

type board struct {
	ID      string `json:"id"`
	Company string `json:"company"`
}
type sourceConfig struct {
	Enabled    bool    `json:"enabled"`
	Arbeitnow  bool    `json:"arbeitnow"`
	Greenhouse []board `json:"greenhouse"`
	Lever      []board `json:"lever"`
	MaxPages   int     `json:"max_pages"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Getenv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer, getenv func(string) string) error {
	flags := flag.NewFlagSet("import-public-jobs", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("config", "", "server-side JSON config")
	dry := flags.Bool("dry-run", false, "fetch and normalize without database writes")
	once := flags.Bool("once", false, "run one import and exit")
	if flags.Parse(args) != nil || flags.NArg() != 0 || *path == "" || *dry == *once {
		return errors.New("usage: import-public-jobs --config <file> (--dry-run | --once)")
	}
	config, err := loadConfig(*path)
	if err != nil {
		return err
	}
	if !config.Enabled {
		return errors.New("public imports are disabled in config")
	}
	if config.MaxPages == 0 {
		config.MaxPages = 10
	}
	if config.MaxPages < 1 || config.MaxPages > 50 {
		return errors.New("invalid max_pages")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	var store *jobs.PostgresStore
	if *once {
		if getenv("PUBLIC_IMPORT_WRITE_ENABLED") != "true" {
			return errors.New("PUBLIC_IMPORT_WRITE_ENABLED=true is required for --once")
		}
		env := getenv("APP_ENV")
		if (env != "development" && env != "test" && env != "staging" && env != "production") || env != getenv("PUBLIC_IMPORT_TARGET") {
			return errors.New("APP_ENV and PUBLIC_IMPORT_TARGET must explicitly match")
		}
		if (env == "staging" || env == "production") && getenv("PUBLIC_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED") != "true" {
			return errors.New("staging/production public imports require explicit permission review")
		}
		raw := getenv("DATABASE_URL")
		dbURL, err := url.Parse(raw)
		if err != nil || dbURL.Hostname() == "" || (dbURL.Scheme != "postgres" && dbURL.Scheme != "postgresql") {
			return errors.New("DATABASE_URL must be a PostgreSQL URL")
		}
		if dbURL.Hostname() != getenv("PUBLIC_IMPORT_DATABASE_HOST") {
			return errors.New("PUBLIC_IMPORT_DATABASE_HOST must match the inspected database hostname")
		}
		mode := getenv("PGX_QUERY_EXEC_MODE")
		if mode == "" {
			mode = "cache_statement"
		}
		pool, err := database.Connect(ctx, raw, mode)
		if err != nil {
			return errors.New("database connection failed")
		}
		defer pool.Close()
		store = jobs.NewPostgresStore(pool, env == "production")
	}
	encoder := json.NewEncoder(out)
	for _, configured := range configuredSources(config) {
		var p providers.VacancyProvider
		switch configured.kind {
		case "arbeitnow":
			p, err = publicapi.NewArbeitnow(config.MaxPages, nil)
		case "greenhouse":
			budget, budgetErr := greenhouse.NewImportBudget(3, 5000)
			if budgetErr != nil {
				return budgetErr
			}
			p, err = greenhouse.NewClient(configured.id, budget, &http.Client{})
		case "lever":
			p, err = publicapi.NewLever(configured.id, configured.company, config.MaxPages, nil)
		}
		if err != nil {
			return err
		}
		if configured.kind == "greenhouse" {
			p = companyFallback{p, configured.company}
		}
		if store != nil {
			if err := store.RegisterPublicImportSource(ctx, p.Source(), configured.kind, configured.display); err != nil {
				return err
			}
		}
		var target jobs.Store
		if store != nil {
			target = store
		}
		report, err := ingestion.NewService(target, nil, logger, 0).Run(ctx, p, ingestion.RunOptions{DryRun: *dry, Sample: 5, NoAutomaticExpiry: true})
		if encodeErr := encoder.Encode(report); encodeErr != nil {
			return errors.New("report output failed")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type configuredSource struct{ kind, id, company, display string }

func configuredSources(c sourceConfig) []configuredSource {
	items := []configuredSource{}
	if c.Arbeitnow {
		items = append(items, configuredSource{kind: "arbeitnow", display: "Arbeitnow"})
	}
	for _, b := range c.Greenhouse {
		items = append(items, configuredSource{kind: "greenhouse", id: b.ID, company: b.Company, display: "Greenhouse"})
	}
	for _, b := range c.Lever {
		items = append(items, configuredSource{kind: "lever", id: b.ID, company: b.Company, display: "Lever"})
	}
	return items
}

func loadConfig(path string) (sourceConfig, error) {
	var c sourceConfig
	f, err := os.Open(path)
	if err != nil {
		return c, errors.New("cannot read public import config")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil || len(raw) > 65536 {
		return c, errors.New("invalid public import config size")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&c) != nil || decoder.Decode(new(any)) != io.EOF {
		return c, errors.New("invalid public import JSON")
	}
	if len(configuredSources(c)) == 0 || len(c.Greenhouse)+len(c.Lever) > 20 {
		return c, errors.New("invalid source selection")
	}
	validID := regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)
	seen := map[string]bool{}
	for _, selection := range configuredSources(c) {
		if selection.kind == "arbeitnow" {
			continue
		}
		key := selection.kind + ":" + selection.id
		if !validID.MatchString(selection.id) || strings.TrimSpace(selection.company) == "" || seen[key] {
			return c, errors.New("invalid or duplicate company board")
		}
		seen[key] = true
	}
	return c, nil
}

type companyFallback struct {
	providers.VacancyProvider
	company string
}

func (c companyFallback) Collect(ctx context.Context) (providers.Result, error) {
	result, err := c.VacancyProvider.Collect(ctx)
	for i := range result.Items {
		if result.Items[i].CompanyNameRaw == "" {
			result.Items[i].CompanyNameRaw = c.company
		}
	}
	return result, err
}
