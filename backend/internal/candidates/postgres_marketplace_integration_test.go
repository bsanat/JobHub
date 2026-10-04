package candidates

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"jobhub-ai/backend/internal/database"
)

func TestEmployerApplicationsIntegration(t *testing.T) {
	raw := os.Getenv("TEST_DATABASE_URL")
	u, err := url.Parse(raw)
	if raw == "" || err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") || !strings.HasSuffix(u.Path, "_test") {
		t.Skip("requires an isolated local TEST_DATABASE_URL ending in _test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Connect(ctx, raw, "cache_statement")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := NewPostgresStore(pool)
	var users, companies, jobs []string
	defer func() {
		for _, query := range []string{
			`DELETE FROM jobhub.application_status_events WHERE application_id IN (SELECT id FROM jobhub.applications WHERE candidate_id=ANY($1::uuid[]))`,
			`DELETE FROM jobhub.applications WHERE candidate_id=ANY($1::uuid[])`,
			`DELETE FROM jobhub.jobs WHERE created_by_user_id=ANY($1::uuid[])`,
			`DELETE FROM jobhub.users WHERE id=ANY($1::uuid[])`,
		} {
			if _, err := pool.Exec(context.Background(), query, users); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
		if _, err := pool.Exec(context.Background(), `DELETE FROM jobhub.companies WHERE id=ANY($1::uuid[])`, companies); err != nil {
			t.Errorf("cleanup companies: %v", err)
		}
	}()
	for i, role := range []string{"employer", "employer", "job_seeker"} {
		var id string
		email := fmt.Sprintf("applications-%d-%d@example.com", time.Now().UnixNano(), i)
		err := pool.QueryRow(ctx, `INSERT INTO jobhub.users(full_name,email,normalized_email,password_hash,role) VALUES('Applications Fixture',$1,$1,'unused',$2) RETURNING id::text`, email, role).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		users = append(users, id)
	}
	for _, owner := range users[:2] {
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO jobhub.companies(name) VALUES($1) RETURNING id::text`, "Fixture "+owner).Scan(&id); err != nil {
			t.Fatal(err)
		}
		companies = append(companies, id)
		if _, err := pool.Exec(ctx, `INSERT INTO jobhub.company_memberships(company_id,user_id,role) VALUES($1,$2,'owner')`, id, owner); err != nil {
			t.Fatal(err)
		}
	}
	resume, _, err := store.ReplaceResume(ctx, users[2], Resume{storageKey: users[2] + ".pdf", OriginalFilename: "fixture.pdf", ContentType: "application/pdf", SizeBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	var applications []Application
	for i, owner := range []int{0, 0, 1} {
		var id string
		err := pool.QueryRow(ctx, `INSERT INTO jobhub.jobs(source,company_id,created_by_user_id,title,application_method,publication_status,first_seen_at,last_seen_at,last_synced_at,fresh_until) VALUES('jobhub',$1,$2,$3,'internal','published',now(),now(),now(),now()+interval '1 day') RETURNING id::text`, companies[owner], users[owner], fmt.Sprintf("Fixture vacancy %d", i)).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, id)
		app, err := store.CreateApplication(ctx, users[2], id, resume.ID, "Fixture application")
		if err != nil {
			t.Fatal(err)
		}
		applications = append(applications, app)
	}
	if _, err := pool.Exec(ctx, `UPDATE jobhub.jobs SET publication_status='closed' WHERE id=$1`, jobs[0]); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []struct {
		owner, job string
		count      int
	}{
		{users[0], "", 2}, {users[1], "", 1}, {users[2], "", 0}, {users[0], jobs[0], 1},
	} {
		items, err := store.ListEmployerApplications(ctx, scope.owner, scope.job)
		if err != nil || len(items) != scope.count {
			t.Fatalf("employer list: count=%d want=%d err=%v", len(items), scope.count, err)
		}
		for _, item := range items {
			if item.JobTitle == "" || item.CandidateName == "" || item.Resume.ID != resume.ID {
				t.Fatalf("missing applicant details: %+v", item)
			}
		}
	}
	if _, err := store.ListEmployerApplications(ctx, users[0], jobs[2]); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign vacancy list: %v", err)
	}
	if _, err := store.GetEmployerApplicationResume(ctx, users[1], applications[0].ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign resume: %v", err)
	}
	if _, err := store.UpdateEmployerApplication(ctx, users[1], applications[0].ID, "viewed"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign status update: %v", err)
	}
	if _, err := store.UpdateEmployerApplication(ctx, users[0], applications[0].ID, "in_review"); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListEmployerApplications(ctx, users[0], jobs[0])
	if err != nil || len(items) != 1 || items[0].Status != "in_review" || items[0].JobAvailable {
		t.Fatalf("closed vacancy application state: %+v %v", items, err)
	}
}
