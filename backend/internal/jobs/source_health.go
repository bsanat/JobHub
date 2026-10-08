package jobs

import (
	"context"
	"errors"
	"strings"
	"time"
)

type SourceHealth struct {
	Source          string     `json:"source"`
	LastAttempt     *time.Time `json:"last_attempt"`
	LastSuccess     *time.Time `json:"last_success"`
	RequestCount    int        `json:"request_count"`
	JobsSeen        int        `json:"jobs_seen"`
	FailureCategory *string    `json:"failure_category"`
	Status          string     `json:"status"`
}

// LegacyPublicURLs returns only publicly visible originals from the two
// established career importers. New Kazakhstan imports use them to avoid
// briefly republishing the same vacancy under another source ID.
func (s *PostgresStore) LegacyPublicURLs(ctx context.Context, source string) (map[string]bool, error) {
	if source != "airastana:careers" && source != "kcell:careers" {
		return nil, errors.New("unsupported legacy career source")
	}
	rows, err := s.pool.Query(ctx, `SELECT j.source_url FROM jobhub.jobs j
JOIN jobhub.job_sources src ON src.source=j.source
WHERE j.source=$1 AND src.enabled AND src.production_permissions_confirmed
AND jobhub.job_is_public(j) AND j.source_url IS NOT NULL`, source)
	if err != nil {
		return nil, errors.New("legacy career URL read failed")
	}
	defer rows.Close()
	urls := make(map[string]bool)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, errors.New("legacy career URL read failed")
		}
		urls[url] = true
	}
	if rows.Err() != nil {
		return nil, errors.New("legacy career URL read failed")
	}
	return urls, nil
}

// SourceHealth is a read-only projection of real run records, not dry-run output.
// No cached registry columns or new schema are required for the bounded POC.
func (s *PostgresStore) SourceHealth(ctx context.Context, source string) (SourceHealth, error) {
	h := SourceHealth{Source: source}
	err := s.pool.QueryRow(ctx, `
 SELECT r.started_at,
   (SELECT max(finished_at) FROM jobhub.ingestion_runs WHERE source=$1 AND status='succeeded'),
   COALESCE(r.request_count,0),COALESCE(r.fetched_count,0),r.error_code,
   CASE WHEN NOT s.enabled THEN 'disabled'
        WHEN s.production_permissions_confirmed THEN 'permission_review_required'
        WHEN r.status='failed' THEN 'failing'
        WHEN r.status='succeeded' THEN 'healthy_development_only'
        WHEN r.status='running' THEN 'running'
        ELSE 'never_synced' END
 FROM jobhub.job_sources s
 LEFT JOIN LATERAL (SELECT * FROM jobhub.ingestion_runs WHERE source=s.source ORDER BY started_at DESC,id DESC LIMIT 1) r ON true
 WHERE s.source=$1`, source).Scan(&h.LastAttempt, &h.LastSuccess, &h.RequestCount, &h.JobsSeen, &h.FailureCategory, &h.Status)
	if err != nil {
		return h, errors.New("source health read failed")
	}
	return h, nil
}

// RegisterDevelopmentSource only inserts an unapproved local source. Conflicts
// are checked, not overwritten; it cannot enable an existing disabled source.
// The caller must first enforce the isolated-development database boundary.
func (s *PostgresStore) RegisterDevelopmentSource(ctx context.Context, source, provider, displayName string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO jobhub.job_sources(source,provider,display_name,enabled,production_permissions_confirmed) VALUES($1,$2,$3,true,false) ON CONFLICT(source) DO NOTHING`, source, provider, displayName)
	if err != nil {
		return errors.New("development source registration failed")
	}
	var allowed bool
	err = s.pool.QueryRow(ctx, `SELECT provider=$2 AND enabled AND NOT production_permissions_confirmed FROM jobhub.job_sources WHERE source=$1`, source, provider).Scan(&allowed)
	if err != nil || !allowed {
		return errors.New("source is disabled, incompatible, or approved for production; POC refused")
	}
	return nil
}

// RegisterPublicImportSource records a disabled, unapproved source. Enabling
// display and confirming redistribution remain separate operator decisions.
func (s *PostgresStore) RegisterPublicImportSource(ctx context.Context, source, provider, displayName string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO jobhub.job_sources(source,provider,display_name,enabled,production_permissions_confirmed,attribution_text) VALUES($1,$2,$3,false,false,$4) ON CONFLICT(source) DO NOTHING`, source, provider, displayName, "Original source: "+displayName)
	if err != nil {
		return errors.New("public source registration failed")
	}
	var compatible bool
	err = s.pool.QueryRow(ctx, `SELECT provider=$2 FROM jobhub.job_sources WHERE source=$1`, source, provider).Scan(&compatible)
	if err != nil || !compatible {
		return errors.New("public source provider mismatch")
	}
	return nil
}

// RegisterKZCompanySource keeps company-career imports hidden until a separate
// approved activation, while setting the existing Kazakhstan market filter.
func (s *PostgresStore) RegisterKZCompanySource(ctx context.Context, source, displayName string) error {
	if !strings.HasPrefix(source, "kzcompany:") || strings.TrimPrefix(source, "kzcompany:") == "" || strings.TrimSpace(displayName) == "" {
		return errors.New("invalid Kazakhstan company source")
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO jobhub.job_sources(source,provider,market,display_name,enabled,production_permissions_confirmed,attribution_text)
VALUES($1,'kzcompany','KZ',$2,false,false,$3)
ON CONFLICT(source) DO UPDATE SET market='KZ'
WHERE jobhub.job_sources.provider='kzcompany'`, source, displayName, "Official company careers: "+displayName)
	if err != nil {
		return errors.New("Kazakhstan source registration failed")
	}
	var compatible bool
	err = s.pool.QueryRow(ctx, `SELECT provider='kzcompany' AND market='KZ' FROM jobhub.job_sources WHERE source=$1`, source).Scan(&compatible)
	if err != nil || !compatible {
		return errors.New("Kazakhstan source provider mismatch")
	}
	return nil
}

// ActivateKZCompanyAfterImport is called only after an explicitly approved
// --once run. It never activates a source that imported zero real jobs.
func (s *PostgresStore) ActivateKZCompanyAfterImport(ctx context.Context, source string) error {
	if !strings.HasPrefix(source, "kzcompany:") || strings.TrimPrefix(source, "kzcompany:") == "" {
		return errors.New("invalid Kazakhstan company source")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE jobhub.job_sources s
SET enabled=true, production_permissions_confirmed=true, updated_at=now()
WHERE s.source=$1 AND s.provider='kzcompany' AND s.market='KZ'
  AND NOT s.enabled AND NOT s.production_permissions_confirmed
  AND EXISTS (SELECT 1 FROM jobhub.ingestion_runs r WHERE r.source=s.source AND r.status='succeeded' AND r.normalized_count>0)
  AND EXISTS (SELECT 1 FROM jobhub.jobs j WHERE j.source=s.source AND j.source_status='active' AND j.moderation_status='approved')`, source)
	if err != nil {
		return errors.New("Kazakhstan source activation failed")
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	var active bool
	err = s.pool.QueryRow(ctx, `SELECT enabled AND production_permissions_confirmed AND market='KZ'
FROM jobhub.job_sources WHERE source=$1 AND provider='kzcompany'`, source).Scan(&active)
	if err != nil || !active {
		return errors.New("Kazakhstan source activation requires a successful import")
	}
	return nil
}

// ActivateArbeitnowAfterInitialSync opens a successfully imported source once.
// updated_at records the Telegram baseline; the publisher only selects jobs
// first seen after this moment. Repeated imports must not move that baseline.
func (s *PostgresStore) ActivateArbeitnowAfterInitialSync(ctx context.Context) error {
	const source = "arbeitnow:global"
	tag, err := s.pool.Exec(ctx, `UPDATE jobhub.job_sources s
SET enabled=true, production_permissions_confirmed=true, updated_at=now()
WHERE s.source=$1 AND s.provider='arbeitnow' AND NOT s.enabled
  AND EXISTS (SELECT 1 FROM jobhub.ingestion_runs r
              WHERE r.source=s.source AND r.status='succeeded' AND r.normalized_count>0)
  AND EXISTS (SELECT 1 FROM jobhub.jobs j
              WHERE j.source=s.source AND j.source_status='active' AND j.moderation_status='approved')`, source)
	if err != nil {
		return errors.New("Arbeitnow activation failed")
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	var active bool
	err = s.pool.QueryRow(ctx, `SELECT enabled AND production_permissions_confirmed AND updated_at>created_at
		FROM jobhub.job_sources WHERE source=$1 AND provider='arbeitnow'`, source).Scan(&active)
	if err != nil || !active {
		return errors.New("Arbeitnow activation requires a successful import and an unused publication baseline")
	}
	return nil
}
