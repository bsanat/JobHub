package telegrampublisher

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool        *pgxpool.Pool
	testSources bool
}

func NewPostgresStore(pool *pgxpool.Pool, testSources bool) *PostgresStore {
	return &PostgresStore{pool: pool, testSources: testSources}
}

func (s *PostgresStore) ClaimBatch(ctx context.Context, channelID int64, limit, maxAttempts int) ([]Vacancy, error) {
	return s.queryVacancies(ctx, claimEligibleSQL, channelID, limit, maxAttempts)
}

func (s *PostgresStore) ListEligible(ctx context.Context, channelID int64, limit, maxAttempts int) ([]Vacancy, error) {
	return s.queryVacancies(ctx, listEligibleSQL, channelID, limit, maxAttempts)
}

func (s *PostgresStore) MarkSent(ctx context.Context, jobID string, channelID, messageID int64) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE jobhub.telegram_publications
SET status = 'sent',
    telegram_message_id = $3,
    error_code = NULL,
    sent_at = now(),
    updated_at = now()
WHERE job_id = $1::uuid AND channel_id = $2`, jobID, channelID, messageID)
	if err != nil {
		return fmt.Errorf("mark telegram publication sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("telegram publication claim not found")
	}
	return nil
}

func (s *PostgresStore) MarkFailed(ctx context.Context, jobID string, channelID int64, errorCode string, countAttempt bool) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE jobhub.telegram_publications
SET status = 'failed',
    error_code = $3,
    attempts = attempts + CASE WHEN $4 THEN 1 ELSE 0 END,
    updated_at = now()
WHERE job_id = $1::uuid AND channel_id = $2`, jobID, channelID, errorCode, countAttempt)
	if err != nil {
		return fmt.Errorf("mark telegram publication failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("telegram publication claim not found")
	}
	return nil
}

func (s *PostgresStore) queryVacancies(ctx context.Context, query string, channelID int64, limit, maxAttempts int) ([]Vacancy, error) {
	rows, err := s.pool.Query(ctx, query, channelID, limit, maxAttempts, s.testSources)
	if err != nil {
		return nil, fmt.Errorf("query telegram publication vacancies: %w", err)
	}
	defer rows.Close()
	items := []Vacancy{}
	for rows.Next() {
		var item Vacancy
		var salaryRaw, salaryMin, salaryMax, salaryCurrency, salaryPeriod sql.NullString
		if err := rows.Scan(&item.ID, &item.Source, &item.SourceName, &item.Attribution, &item.OriginalURL, &item.Title, &item.Description, &item.CompanyName, &item.Location, &item.WorkMode, &item.ExperienceLevel, &item.EmploymentType, &salaryRaw, &salaryMin, &salaryMax, &salaryCurrency, &salaryPeriod, &item.SalaryVisible); err != nil {
			return nil, fmt.Errorf("scan telegram publication vacancy: %w", err)
		}
		item.SalaryRaw = nullString(salaryRaw)
		item.SalaryMin = nullString(salaryMin)
		item.SalaryMax = nullString(salaryMax)
		item.SalaryCurrency = nullString(salaryCurrency)
		item.SalaryPeriod = nullString(salaryPeriod)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate telegram publication vacancies: %w", err)
	}
	return items, nil
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

const vacancySelectColumns = `
    j.id::text,
    j.source,
    s.display_name,
    COALESCE(s.attribution_text, ''),
    COALESCE(j.source_url, ''),
    j.title,
    COALESCE(j.description, '') AS description,
    COALESCE(c.name, j.company_name_raw, '') AS company_name,
    COALESCE(j.location_raw, '') AS location_raw,
    COALESCE(j.work_mode, '') AS work_mode,
    COALESCE(j.experience_level, '') AS experience_level,
    COALESCE(j.employment_type, j.employment_type_raw, '') AS employment_type,
    j.salary_raw,
    j.salary_min::text,
    j.salary_max::text,
    j.salary_currency,
    j.salary_period,
    j.salary_visible`

const canonicalVacancySQL = `
  AND NOT EXISTS (
      SELECT 1
      FROM jobhub.jobs duplicate
      JOIN jobhub.job_sources duplicate_source ON duplicate_source.source = duplicate.source
      WHERE duplicate.id <> j.id
        AND duplicate_source.enabled
        AND (duplicate_source.production_permissions_confirmed OR $4)
        AND duplicate_source.provider <> 'demo'
        AND jobhub.job_is_public(duplicate)
        AND lower(btrim(duplicate.title)) = lower(btrim(j.title))
        AND lower(btrim(COALESCE((SELECT name FROM jobhub.companies WHERE id = duplicate.company_id), duplicate.company_name_raw, '')))
            = lower(btrim(COALESCE((SELECT name FROM jobhub.companies WHERE id = j.company_id), j.company_name_raw, '')))
        AND lower(btrim(COALESCE(duplicate.location_raw, ''))) = lower(btrim(COALESCE(j.location_raw, '')))
        AND COALESCE(duplicate.description, '') = COALESCE(j.description, '')
        AND NOT EXISTS (
            SELECT 1 FROM jobhub.users creator
            WHERE creator.id = duplicate.created_by_user_id
              AND creator.normalized_email = 'demo-employer@jobhub.local'
        )
        AND (COALESCE(duplicate.published_at, duplicate.external_published_at, duplicate.first_seen_at), duplicate.id)
            < (COALESCE(j.published_at, j.external_published_at, j.first_seen_at), j.id)
  )`

const listEligibleSQL = `
SELECT` + vacancySelectColumns + `
FROM jobhub.jobs j
JOIN jobhub.job_sources s ON s.source = j.source
LEFT JOIN jobhub.companies c ON c.id = j.company_id
LEFT JOIN jobhub.telegram_publications tp ON tp.job_id = j.id AND tp.channel_id = $1
WHERE s.enabled
  AND (s.production_permissions_confirmed OR $4)
  AND s.provider NOT IN ('demo', 'kzcompany')
  AND (s.provider NOT IN ('arbeitnow', 'greenhouse', 'lever') OR (s.updated_at > s.created_at AND j.first_seen_at > s.updated_at))
  AND jobhub.job_is_public(j)
  AND NOT EXISTS (
      SELECT 1 FROM jobhub.users creator
      WHERE creator.id = j.created_by_user_id
        AND creator.normalized_email = 'demo-employer@jobhub.local'
  )
  ` + canonicalVacancySQL + `
  AND (
      tp.id IS NULL
      OR (tp.status = 'failed' AND tp.attempts < $3 AND tp.updated_at < now() - interval '1 minute')
  )
ORDER BY COALESCE(j.published_at, j.external_published_at, j.first_seen_at) ASC, j.id ASC
LIMIT $2`

const claimEligibleSQL = `
WITH eligible AS (
    SELECT j.id
    FROM jobhub.jobs j
    JOIN jobhub.job_sources s ON s.source = j.source
    LEFT JOIN jobhub.telegram_publications tp ON tp.job_id = j.id AND tp.channel_id = $1
    WHERE s.enabled
      AND (s.production_permissions_confirmed OR $4)
      AND s.provider NOT IN ('demo', 'kzcompany')
      AND (s.provider NOT IN ('arbeitnow', 'greenhouse', 'lever') OR (s.updated_at > s.created_at AND j.first_seen_at > s.updated_at))
      AND jobhub.job_is_public(j)
      AND NOT EXISTS (
          SELECT 1 FROM jobhub.users creator
          WHERE creator.id = j.created_by_user_id
            AND creator.normalized_email = 'demo-employer@jobhub.local'
      )
      ` + canonicalVacancySQL + `
      AND (
          tp.id IS NULL
          OR (tp.status = 'failed' AND tp.attempts < $3 AND tp.updated_at < now() - interval '1 minute')
      )
    ORDER BY COALESCE(j.published_at, j.external_published_at, j.first_seen_at) ASC, j.id ASC
    LIMIT $2
),
claimed AS (
    INSERT INTO jobhub.telegram_publications (job_id, channel_id, status)
    SELECT id, $1, 'pending'
    FROM eligible
    ON CONFLICT (job_id, channel_id) DO UPDATE
    SET status = 'pending',
        error_code = NULL,
        updated_at = now()
    WHERE jobhub.telegram_publications.status = 'failed'
      AND jobhub.telegram_publications.attempts < $3
      AND jobhub.telegram_publications.updated_at < now() - interval '1 minute'
    RETURNING job_id
)
SELECT` + vacancySelectColumns + `
FROM claimed
JOIN jobhub.jobs j ON j.id = claimed.job_id
JOIN jobhub.job_sources s ON s.source = j.source
LEFT JOIN jobhub.companies c ON c.id = j.company_id
ORDER BY COALESCE(j.published_at, j.external_published_at, j.first_seen_at) ASC, j.id ASC`
