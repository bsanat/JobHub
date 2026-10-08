package telegrampublisher

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"jobhub-ai/backend/internal/config"
)

func TestPostgresStoreClaimsOnceAndRetriesConfirmedFailure(t *testing.T) {
	databaseURL := os.Getenv("TELEGRAM_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TELEGRAM_TEST_DATABASE_URL is not set")
	}
	if err := config.ValidateTelegramDevelopmentDatabase("development", databaseURL); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	fixtureIDs := createPublisherFixtures(t, ctx, pool)
	approvedExternalIDs, baselineJobID := createExternalPublisherFixtures(t, ctx, pool)
	channelID := -time.Now().UnixNano()
	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM jobhub.telegram_publications WHERE channel_id = $1", channelID)
	}()
	store := NewPostgresStore(pool, false)
	testChannelID := channelID - 1
	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM jobhub.telegram_publications WHERE channel_id = $1", testChannelID)
	}()
	testStore := NewPostgresStore(pool, true)
	testClaims, err := testStore.ClaimBatch(ctx, testChannelID, 50, 3)
	if err != nil {
		t.Fatal(err)
	}
	var testUnapproved, testDemo bool
	testApprovedCopies := 0
	testPublicProviders := map[string]bool{}
	for _, job := range testClaims {
		testUnapproved = testUnapproved || job.Title == "Unapproved External Engineer"
		testDemo = testDemo || job.Title == "Demo External Engineer"
		if job.Title == "Approved External Engineer" {
			testApprovedCopies++
		}
		if job.Title == "Arbeitnow External Engineer" || job.Title == "Greenhouse External Engineer" || job.Title == "Lever External Engineer" {
			testPublicProviders[job.Title] = true
		}
	}
	if !testUnapproved || testDemo || testApprovedCopies != 2 || len(testPublicProviders) != 3 {
		t.Fatalf("staging test mode eligibility is wrong: unapproved=%t demo=%t approved_copies=%d", testUnapproved, testDemo, testApprovedCopies)
	}
	var preview bytes.Buffer
	previewStats, err := New(Config{ChannelID: channelID, PublicBaseURL: "http://localhost:5173", DryRun: true, BatchSize: 50}, store, nil, &preview, nil).ProcessOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if previewStats.DryRun < 3 || !bytes.Contains(preview.Bytes(), []byte("Backend Engineer")) || !bytes.Contains(preview.Bytes(), []byte("Data Engineer")) || !bytes.Contains(preview.Bytes(), []byte("Approved External Engineer")) {
		t.Fatalf("dry run omitted eligible fixtures: stats=%#v", previewStats)
	}
	if bytes.Contains(preview.Bytes(), []byte("Unapproved External Engineer")) || bytes.Contains(preview.Bytes(), []byte("Demo External Engineer")) {
		t.Fatal("dry run included unapproved or demo external vacancy")
	}
	demoRows, err := pool.Query(ctx, `SELECT j.id::text FROM jobhub.jobs j
		JOIN jobhub.users creator ON creator.id = j.created_by_user_id
		WHERE creator.normalized_email = 'demo-employer@jobhub.local'`)
	if err != nil {
		t.Fatal(err)
	}
	for demoRows.Next() {
		var demoID string
		if err := demoRows.Scan(&demoID); err != nil {
			demoRows.Close()
			t.Fatal(err)
		}
		if bytes.Contains(preview.Bytes(), []byte("/jobs/"+demoID)) {
			demoRows.Close()
			t.Fatal("dry run included a demo vacancy")
		}
	}
	if err := demoRows.Err(); err != nil {
		demoRows.Close()
		t.Fatal(err)
	}
	demoRows.Close()
	var dryRunWrites int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM jobhub.telegram_publications WHERE channel_id = $1", channelID).Scan(&dryRunWrites); err != nil || dryRunWrites != 0 {
		t.Fatalf("dry run created publication rows: count=%d error=%v", dryRunWrites, err)
	}
	claimed, err := store.ClaimBatch(ctx, channelID, 50, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) < 6 {
		t.Fatalf("expected native and approved external/public import fixtures, got %d", len(claimed))
	}
	claimedIDs := make(map[string]bool, len(claimed))
	for _, vacancy := range claimed {
		claimedIDs[vacancy.ID] = true
		if vacancy.Source != "jobhub" && !approvedExternalIDs[vacancy.ID] {
			t.Fatalf("unapproved or demo external vacancy was claimed: %s", vacancy.Source)
		}
	}
	for _, id := range fixtureIDs {
		if !claimedIDs[id] {
			t.Fatalf("native fixture %s was not claimed", id)
		}
	}
	for id := range approvedExternalIDs {
		if !claimedIDs[id] {
			t.Fatalf("approved external fixture %s was not claimed", id)
		}
	}
	if claimedIDs[baselineJobID] {
		t.Fatal("initial public API import was claimed for Telegram")
	}
	var demoClaimed int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM jobhub.telegram_publications tp
		JOIN jobhub.jobs j ON j.id = tp.job_id
		JOIN jobhub.users creator ON creator.id = j.created_by_user_id
		WHERE tp.channel_id = $1 AND creator.normalized_email = 'demo-employer@jobhub.local'`, channelID).Scan(&demoClaimed); err != nil {
		t.Fatal(err)
	}
	if demoClaimed != 0 {
		t.Fatalf("demo vacancies were claimed: %d", demoClaimed)
	}
	if _, err := pool.Exec(ctx, "UPDATE jobhub.telegram_publications SET updated_at = now() - interval '1 hour' WHERE channel_id = $1", channelID); err != nil {
		t.Fatal(err)
	}
	again, err := store.ClaimBatch(ctx, channelID, 50, 3)
	if err != nil || len(again) != 0 {
		t.Fatalf("pending vacancies must not be claimed again: %d, %v", len(again), err)
	}
	if err := store.MarkSent(ctx, claimed[0].ID, channelID, 123); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFailed(ctx, claimed[1].ID, channelID, "TELEGRAM_RATE_LIMITED", false); err != nil {
		t.Fatal(err)
	}
	var attempts int
	if err := pool.QueryRow(ctx, "SELECT attempts FROM jobhub.telegram_publications WHERE channel_id = $1 AND job_id = $2::uuid", channelID, claimed[1].ID).Scan(&attempts); err != nil || attempts != 0 {
		t.Fatalf("rate limit consumed a permanent retry: attempts=%d error=%v", attempts, err)
	}
	beforeDelay, err := store.ClaimBatch(ctx, channelID, 50, 3)
	if err != nil || len(beforeDelay) != 0 {
		t.Fatalf("confirmed failure retried without delay: %d, %v", len(beforeDelay), err)
	}
	if _, err := pool.Exec(ctx, "UPDATE jobhub.telegram_publications SET updated_at = now() - interval '2 minutes' WHERE channel_id = $1 AND job_id = $2::uuid", channelID, claimed[1].ID); err != nil {
		t.Fatal(err)
	}
	retry, err := store.ClaimBatch(ctx, channelID, 50, 3)
	if err != nil || len(retry) != 1 || retry[0].ID != claimed[1].ID {
		t.Fatalf("expected only confirmed failure to retry: %#v, %v", retry, err)
	}
	if err := store.MarkSent(ctx, retry[0].ID, channelID, 124); err != nil {
		t.Fatal(err)
	}
	remaining, err := store.ListEligible(ctx, channelID, 50, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Fatalf("already claimed or sent vacancies appeared again: %d", len(remaining))
	}
}

func createPublisherFixtures(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()
	name := fmt.Sprintf("Publisher test %d", time.Now().UnixNano())
	email := fmt.Sprintf("publisher-test-%d@jobhub.local", time.Now().UnixNano())
	var ownerID, companyID string
	jobIDs := make([]string, 0, 2)
	t.Cleanup(func() {
		cleanup := context.Background()
		for _, id := range jobIDs {
			_, _ = pool.Exec(cleanup, "DELETE FROM jobhub.jobs WHERE id = $1::uuid", id)
		}
		if companyID != "" {
			_, _ = pool.Exec(cleanup, "DELETE FROM jobhub.companies WHERE id = $1::uuid", companyID)
		}
		if ownerID != "" {
			_, _ = pool.Exec(cleanup, "DELETE FROM jobhub.users WHERE id = $1::uuid", ownerID)
		}
	})
	if err := pool.QueryRow(ctx, `INSERT INTO jobhub.users (full_name,email,normalized_email,password_hash,role,status)
		VALUES ($1,$2,$2,'test-only','employer','active') RETURNING id::text`, name, email).Scan(&ownerID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO jobhub.companies (name) VALUES ($1) RETURNING id::text", name).Scan(&companyID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO jobhub.company_memberships (company_id,user_id,role) VALUES ($1::uuid,$2::uuid,'owner')", companyID, ownerID); err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"Backend Engineer", "Data Engineer"} {
		var jobID string
		if err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs
			(source,company_id,created_by_user_id,company_name_raw,title,location_raw,description,description_kind,
			application_method,source_status,publication_status,moderation_status,first_seen_at,last_seen_at,last_synced_at,fresh_until,published_at,work_mode)
			VALUES ('jobhub',$1::uuid,$2::uuid,$3,$4,'Алматы','Test vacancy','full','internal','active','published','approved',
			now(),now(),now(),'infinity'::timestamptz,now(),'on_site') RETURNING id::text`, companyID, ownerID, name, title).Scan(&jobID); err != nil {
			t.Fatal(err)
		}
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs
}

func createExternalPublisherFixtures(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (map[string]bool, string) {
	t.Helper()
	marker := fmt.Sprintf("publisher-external-%d", time.Now().UnixNano())
	type fixture struct {
		source   string
		provider string
		approved bool
		title    string
	}
	fixtures := []fixture{
		{marker + "-approved", "test", true, "Approved External Engineer"},
		{marker + "-unapproved", "test", false, "Unapproved External Engineer"},
		{marker + "-demo", "demo", true, "Demo External Engineer"},
		{marker + "-arbeitnow", "arbeitnow", true, "Arbeitnow External Engineer"},
		{marker + "-greenhouse", "greenhouse", true, "Greenhouse External Engineer"},
		{marker + "-lever", "lever", true, "Lever External Engineer"},
		{marker + "-kzcompany", "kzcompany", true, "Kazakhstan Company Engineer"},
	}
	var sourceNames, jobIDs []string
	t.Cleanup(func() {
		for _, id := range jobIDs {
			_, _ = pool.Exec(context.Background(), "DELETE FROM jobhub.jobs WHERE id = $1::uuid", id)
		}
		for _, source := range sourceNames {
			_, _ = pool.Exec(context.Background(), "DELETE FROM jobhub.job_sources WHERE source = $1", source)
		}
	})
	approvedIDs := map[string]bool{}
	var baselineJobID string
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, `INSERT INTO jobhub.job_sources (source, provider, display_name, enabled, production_permissions_confirmed, created_at, updated_at)
			VALUES ($1, $2, $3, true, $4, now()-interval '2 hours',
			CASE WHEN $2 IN ('arbeitnow','greenhouse','lever') THEN now()-interval '1 hour' ELSE now()-interval '2 hours' END)`, fixture.source, fixture.provider, fixture.title, fixture.approved); err != nil {
			t.Fatal(err)
		}
		sourceNames = append(sourceNames, fixture.source)
		var jobID string
		if err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs
			(source, external_id, source_url, company_name_raw, title, description, application_method,
			source_status, moderation_status, first_seen_at, last_seen_at, last_synced_at, fresh_until, external_published_at)
			VALUES ($1, $2, $4, 'External Test', $3, 'External test vacancy', 'external',
			'active', 'approved', now(), now(), now(), now() + interval '1 day', now()) RETURNING id::text`,
			fixture.source, fixture.source+"-id", fixture.title, "https://example.test/jobs/"+fixture.source).Scan(&jobID); err != nil {
			t.Fatal(err)
		}
		jobIDs = append(jobIDs, jobID)
		if fixture.approved && fixture.provider != "demo" {
			approvedIDs[jobID] = true
		}
		if fixture.provider == "arbeitnow" {
			if err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs
				(source, external_id, source_url, company_name_raw, title, description, application_method,
				source_status, moderation_status, first_seen_at, last_seen_at, last_synced_at, fresh_until, external_published_at)
				VALUES ($1, $2, 'https://example.test/jobs/baseline', 'External Test', 'Initial Arbeitnow Import', 'External test vacancy', 'external',
				'active', 'approved', now()-interval '90 minutes', now(), now(), now()+interval '1 day', now()) RETURNING id::text`,
				fixture.source, fixture.source+"-baseline").Scan(&baselineJobID); err != nil {
				t.Fatal(err)
			}
			jobIDs = append(jobIDs, baselineJobID)
		}
	}
	var duplicateID string
	if err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs
		(source, external_id, source_url, company_name_raw, title, description, application_method,
		source_status, moderation_status, first_seen_at, last_seen_at, last_synced_at, fresh_until, external_published_at)
		VALUES ($1, $2, $3, 'External Test', 'Approved External Engineer', 'External test vacancy', 'external',
		'active', 'approved', now(), now(), now(), now() + interval '1 day', now() + interval '1 hour') RETURNING id::text`,
		fixtures[0].source, fixtures[0].source+"-duplicate", "https://example.test/jobs/"+fixtures[0].source).Scan(&duplicateID); err != nil {
		t.Fatal(err)
	}
	jobIDs = append(jobIDs, duplicateID)
	var distinctURLID string
	if err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs
		(source, external_id, source_url, company_name_raw, title, description, application_method,
		source_status, moderation_status, first_seen_at, last_seen_at, last_synced_at, fresh_until, external_published_at)
		VALUES ($1, $2, 'https://example.test/jobs/distinct-requisition', 'External Test', 'Approved External Engineer', 'External test vacancy', 'external',
		'active', 'approved', now(), now(), now(), now() + interval '1 day', now() + interval '1 hour') RETURNING id::text`,
		fixtures[0].source, fixtures[0].source+"-distinct").Scan(&distinctURLID); err != nil {
		t.Fatal(err)
	}
	jobIDs = append(jobIDs, distinctURLID)
	approvedIDs[distinctURLID] = true
	return approvedIDs, baselineJobID
}
