package ingestion

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/providers/jooble"
)

type fakeProvider struct {
	vacancies []jooble.Vacancy
	err       error
	requests  int
}

type snapshotProvider struct {
	result providers.Result
	err    error
}

func (snapshotProvider) Source() string                                      { return "arbeitnow:global" }
func (p snapshotProvider) Collect(context.Context) (providers.Result, error) { return p.result, p.err }

func TestOnlyCompleteSuccessfulSnapshotArchivesMissingJobs(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, nil, nil, 0)
	now := time.Date(2026, 10, 9, 1, 0, 0, 0, time.UTC)
	job := func(id string) jobs.ImportedJob {
		return jobs.ImportedJob{Source: "arbeitnow:global", ExternalID: id, Title: id, SourceURL: "https://www.arbeitnow.com/jobs/" + id, DescriptionKind: "full"}
	}
	service.now = func() time.Time { return now }
	if _, err := service.Run(context.Background(), snapshotProvider{result: providers.Result{Items: []jobs.ImportedJob{job("a"), job("b")}, Complete: true}}, RunOptions{}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	if _, err := service.Run(context.Background(), snapshotProvider{result: providers.Result{Items: []jobs.ImportedJob{job("a")}, Complete: false}}, RunOptions{}); err != nil {
		t.Fatal(err)
	}
	if store.archived["arbeitnow:global/b"] {
		t.Fatal("partial snapshot archived an unseen job")
	}
	now = now.Add(time.Hour)
	if _, err := service.Run(context.Background(), snapshotProvider{err: providers.Failure("RATE_LIMITED")}, RunOptions{}); err == nil {
		t.Fatal("expected sync failure")
	}
	if store.archived["arbeitnow:global/b"] {
		t.Fatal("failed sync archived a job")
	}
	now = now.Add(time.Hour)
	if _, err := service.Run(context.Background(), snapshotProvider{result: providers.Result{Items: []jobs.ImportedJob{job("a")}, Complete: true}}, RunOptions{}); err != nil {
		t.Fatal(err)
	}
	if !store.archived["arbeitnow:global/b"] || store.archived["arbeitnow:global/a"] {
		t.Fatal("complete snapshot did not reconcile accurately")
	}
}

func (p *fakeProvider) Search(context.Context, jooble.Search) ([]jooble.Vacancy, error) {
	p.requests++
	return p.vacancies, p.err
}
func (p *fakeProvider) RequestCount() int { return p.requests }

type storedJob struct {
	localID    string
	job        jobs.ImportedJob
	firstSeen  time.Time
	lastSeen   time.Time
	lastSynced time.Time
}

type memoryStore struct {
	items    map[string]storedJob
	runs     map[string]string
	archived map[string]bool
	nextRun  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{items: map[string]storedJob{}, runs: map[string]string{}, archived: map[string]bool{}}
}
func (s *memoryStore) BeginIngestionRun(_ context.Context, _ string, _ int) (string, error) {
	s.nextRun++
	id := string(rune('0' + s.nextRun))
	s.runs[id] = "running"
	return id, nil
}
func (s *memoryStore) FailIngestionRun(_ context.Context, id string, _ jobs.ImportStats, _, _ string) error {
	s.runs[id] = "failed"
	return nil
}
func (s *memoryStore) CompleteIngestionRun(_ context.Context, id, source string, imported []jobs.ImportedJob, observedAt, _ time.Time, stats jobs.ImportStats, complete bool) (jobs.ImportStats, error) {
	for _, item := range imported {
		key := source + "/" + item.ExternalID
		current, exists := s.items[key]
		if !exists {
			current = storedJob{localID: "local-1", firstSeen: observedAt}
			stats.InsertedCount++
		} else {
			stats.UpdatedCount++
		}
		current.job = item
		current.lastSeen = observedAt
		current.lastSynced = observedAt
		s.items[key] = current
		s.archived[key] = false
	}
	if complete {
		for key, current := range s.items {
			if strings.HasPrefix(key, source+"/") && current.lastSynced.Before(observedAt) {
				s.archived[key] = true
			}
		}
	}
	s.runs[id] = "succeeded"
	return stats, nil
}
func (s *memoryStore) Search(context.Context, jobs.SearchParams) (jobs.SearchResult, error) {
	return jobs.SearchResult{}, nil
}
func (s *memoryStore) Get(context.Context, string) (jobs.Job, error) {
	return jobs.Job{}, jobs.ErrNotFound
}

func TestImportIsIdempotentUpdatesInPlaceAndFailureKeepsFreshness(t *testing.T) {
	store := newMemoryStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	base := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	search := []jooble.Search{{Keywords: "Go", Location: "Казахстан", Page: 1, ResultsPerPage: 10}}

	firstProvider := &fakeProvider{vacancies: []jooble.Vacancy{{ID: "42", Title: "Go Developer", Link: "https://kz.jooble.org/jdp/42"}}}
	first := NewService(store, firstProvider, logger, 72*time.Hour)
	first.now = func() time.Time { return base }
	stats, err := first.RunJooble(context.Background(), search)
	if err != nil || stats.InsertedCount != 1 {
		t.Fatalf("first import failed: %#v %v", stats, err)
	}

	secondProvider := &fakeProvider{vacancies: []jooble.Vacancy{{ID: "42", Title: "Senior Go Developer", Link: "https://kz.jooble.org/jdp/42"}}}
	second := NewService(store, secondProvider, logger, 72*time.Hour)
	second.now = func() time.Time { return base.Add(time.Hour) }
	stats, err = second.RunJooble(context.Background(), search)
	if err != nil || stats.UpdatedCount != 1 || stats.InsertedCount != 0 || len(store.items) != 1 {
		t.Fatalf("repeat import was not idempotent: %#v %v", stats, err)
	}
	beforeFailure := store.items["jooble:kz/42"]
	if beforeFailure.localID != "local-1" || beforeFailure.job.Title != "Senior Go Developer" || !beforeFailure.firstSeen.Equal(base) || !beforeFailure.lastSynced.Equal(base.Add(time.Hour)) {
		t.Fatalf("update did not preserve identity/freshness: %#v", beforeFailure)
	}

	failedProvider := &fakeProvider{err: errors.New("provider unavailable")}
	failed := NewService(store, failedProvider, logger, 72*time.Hour)
	failed.now = func() time.Time { return base.Add(2 * time.Hour) }
	if _, err := failed.RunJooble(context.Background(), search); err == nil {
		t.Fatal("expected failed import")
	}
	afterFailure := store.items["jooble:kz/42"]
	if afterFailure.job.Title != beforeFailure.job.Title || !afterFailure.lastSeen.Equal(beforeFailure.lastSeen) || !afterFailure.lastSynced.Equal(beforeFailure.lastSynced) {
		t.Fatal("failed import corrupted stored job freshness")
	}
}
