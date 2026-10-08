// Command reconcile-kz-legacy-duplicates retires duplicate records created when
// a Kazakhstan company already had a working official-careers importer.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"jobhub-ai/backend/internal/database"
)

var legacySources = map[string]string{
	"airastana": "airastana:careers",
	"kcell":     "kcell:careers",
}

type report struct {
	Source         string `json:"source"`
	ExistingSource string `json:"existing_source"`
	ActiveNew      int    `json:"active_new"`
	ExactMatches   int    `json:"exact_matches"`
	UniqueNew      int    `json:"unique_new"`
	Retired        int    `json:"retired"`
	DryRun         bool   `json:"dry_run"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("reconcile-kz-legacy-duplicates", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	id := flags.String("source", "", "airastana or kcell")
	dry := flags.Bool("dry-run", false, "read-only exact-URL comparison")
	once := flags.Bool("once", false, "retire only fully duplicated records")
	if flags.Parse(args) != nil || flags.NArg() != 0 || *dry == *once {
		return errors.New("usage: reconcile-kz-legacy-duplicates --source airastana|kcell (--dry-run|--once)")
	}
	oldSource, ok := legacySources[*id]
	if !ok {
		return errors.New("unsupported legacy source")
	}
	if *once && os.Getenv("KZ_IMPORT_WRITE_ENABLED") != "true" {
		return errors.New("KZ_IMPORT_WRITE_ENABLED=true is required")
	}
	if *once && os.Getenv("KZ_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED") != "true" {
		return errors.New("production permission review is required")
	}
	key := "KZ_IMPORT_READ_DATABASE_URL"
	if *once {
		key = "DATABASE_URL"
	}
	raw, err := checkedDatabaseURL(key)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Connect(ctx, raw, "exec")
	if err != nil {
		return errors.New("database connection failed")
	}
	defer pool.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.New("reconciliation transaction failed")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if *dry {
		if _, err := tx.Exec(ctx, "SET TRANSACTION READ ONLY"); err != nil {
			return errors.New("read-only transaction failed")
		}
	}
	newSource := "kzcompany:" + *id
	if *once {
		for _, source := range []string{oldSource, newSource} {
			if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, source); err != nil {
				return errors.New("source reconciliation lock failed")
			}
		}
	}
	r := report{Source: newSource, ExistingSource: oldSource, DryRun: *dry}
	// The exact original URL must already be public in the established importer.
	// Legacy imports can use source_status=unknown while remaining public.
	err = tx.QueryRow(ctx, `SELECT count(*)::int,
count(*) FILTER (WHERE EXISTS (SELECT 1 FROM jobhub.jobs old
 WHERE old.source=$2 AND jobhub.job_is_public(old) AND old.source_url=added.source_url))::int
FROM jobhub.jobs added WHERE added.source=$1 AND added.source_status='active'`, newSource, oldSource).Scan(&r.ActiveNew, &r.ExactMatches)
	if err != nil {
		return errors.New("duplicate comparison failed")
	}
	r.UniqueNew = r.ActiveNew - r.ExactMatches
	if *once {
		var oldEnabled bool
		err = tx.QueryRow(ctx, `SELECT enabled FROM jobhub.job_sources WHERE source=$1`, oldSource).Scan(&oldEnabled)
		if err != nil || !oldEnabled {
			return errors.New("established source is not publicly enabled")
		}
		tag, err := tx.Exec(ctx, `UPDATE jobhub.jobs added SET source_status='removed', external_archived_at=now(), updated_at=now()
WHERE added.source=$1 AND added.source_status='active' AND EXISTS (
 SELECT 1 FROM jobhub.jobs old WHERE old.source=$2 AND jobhub.job_is_public(old) AND old.source_url=added.source_url
)`, newSource, oldSource)
		if err != nil || tag.RowsAffected() != int64(r.ExactMatches) {
			return errors.New("duplicate retirement failed")
		}
		r.Retired = r.ExactMatches
		if err := tx.Commit(ctx); err != nil {
			return errors.New("duplicate retirement commit failed")
		}
	}
	return json.NewEncoder(os.Stdout).Encode(r)
}

func checkedDatabaseURL(key string) (string, error) {
	if os.Getenv("APP_ENV") != "production" || os.Getenv("KZ_IMPORT_TARGET") != "production" {
		return "", errors.New("reconciliation requires inspected production target")
	}
	raw := strings.TrimSpace(os.Getenv(key))
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.User == nil || u.Hostname() == "" ||
		u.Hostname() != os.Getenv("KZ_IMPORT_DATABASE_HOST") || u.User.Username() != os.Getenv("KZ_IMPORT_DATABASE_USER") ||
		os.Getenv("KZ_IMPORT_DATABASE_HOST") == "" || os.Getenv("KZ_IMPORT_DATABASE_USER") == "" {
		return "", errors.New("database identity does not match inspected production target")
	}
	return raw, nil
}
