package telegrampublisher

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"jobhub-ai/backend/internal/providers/telegram"
)

type fakeStore struct {
	items     []Vacancy
	listed    bool
	claimed   bool
	sent      []string
	failed    []string
	failCodes []string
	counted   []bool
}

func (f *fakeStore) ClaimBatch(context.Context, int64, int, int) ([]Vacancy, error) {
	f.claimed = true
	if len(f.items) == 0 {
		return nil, nil
	}
	item := f.items[0]
	f.items = f.items[1:]
	return []Vacancy{item}, nil
}
func (f *fakeStore) ListEligible(context.Context, int64, int, int) ([]Vacancy, error) {
	f.listed = true
	return f.items, nil
}
func (f *fakeStore) MarkSent(_ context.Context, jobID string, _ int64, _ int64) error {
	f.sent = append(f.sent, jobID)
	return nil
}
func (f *fakeStore) MarkFailed(_ context.Context, jobID string, _ int64, code string, countAttempt bool) error {
	f.failed = append(f.failed, jobID)
	f.failCodes = append(f.failCodes, code)
	f.counted = append(f.counted, countAttempt)
	return nil
}

type fakeSender struct {
	failFirst      bool
	rateLimitFirst bool
	unknownFirst   bool
	calls          int
}

func (f *fakeSender) SendHTMLMessageAndReturnID(context.Context, int64, string) (int64, error) {
	f.calls++
	if f.unknownFirst && f.calls == 1 {
		return 0, errors.New("timeout after request")
	}
	if f.failFirst && f.calls == 1 {
		return 0, &telegram.APIError{Category: "TELEGRAM_INVALID_MESSAGE"}
	}
	if f.rateLimitFirst && f.calls == 1 {
		return 0, &telegram.APIError{Category: "TELEGRAM_RATE_LIMITED", RetryAfter: 3 * time.Second}
	}
	return int64(100 + f.calls), nil
}

func TestPublisherKeepsUncertainDeliveryPending(t *testing.T) {
	store := &fakeStore{items: []Vacancy{{ID: "job-1", Title: "Backend"}}}
	stats, err := New(Config{ChannelID: -1001, PublicBaseURL: "https://jobhub.kz"}, store, &fakeSender{unknownFirst: true}, nil, nil).ProcessOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Unknown != 1 || len(store.failed) != 0 || len(store.sent) != 0 {
		t.Fatalf("uncertain delivery must remain pending: store=%#v stats=%#v", store, stats)
	}
}

func TestPublisherDryRunDoesNotSendOrMark(t *testing.T) {
	store := &fakeStore{items: []Vacancy{{ID: "job-1", Title: "Backend Engineer"}}}
	sender := &fakeSender{}
	var out bytes.Buffer
	stats, err := New(Config{ChannelID: -1001, PublicBaseURL: "https://jobhub.kz", DryRun: true}, store, sender, &out, nil).ProcessOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !store.listed || store.claimed || sender.calls != 0 || len(store.sent) != 0 || stats.DryRun != 1 || !bytes.Contains(out.Bytes(), []byte("Backend Engineer")) {
		t.Fatalf("unexpected dry-run state: store=%#v sender=%#v stats=%#v output=%q", store, sender, stats, out.String())
	}
}

func TestPublisherMarksFailuresAndContinues(t *testing.T) {
	store := &fakeStore{items: []Vacancy{{ID: "job-1", Title: "Backend"}, {ID: "job-2", Title: "Frontend"}}}
	stats, err := New(Config{ChannelID: -1001, PublicBaseURL: "https://jobhub.kz"}, store, &fakeSender{failFirst: true}, nil, nil).ProcessOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sent != 1 || stats.Failed != 1 || len(store.failed) != 1 || store.failed[0] != "job-1" || store.failCodes[0] != "TELEGRAM_INVALID_MESSAGE" || len(store.sent) != 1 || store.sent[0] != "job-2" {
		t.Fatalf("unexpected retry tracking: store=%#v stats=%#v", store, stats)
	}
}

func TestPublisherStopsBatchOnRateLimit(t *testing.T) {
	store := &fakeStore{items: []Vacancy{{ID: "job-1", Title: "Backend"}, {ID: "job-2", Title: "Frontend"}}}
	sender := &fakeSender{rateLimitFirst: true}
	stats, err := New(Config{ChannelID: -1001, PublicBaseURL: "https://jobhub.kz"}, store, sender, nil, nil).ProcessOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Failed != 1 || stats.RetryAfter != 3*time.Second || sender.calls != 1 || len(store.items) != 1 || len(store.counted) != 1 || store.counted[0] {
		t.Fatalf("rate limit should leave the next vacancy unclaimed: stats=%#v sender=%#v store=%#v", stats, sender, store)
	}
}

func TestEligibleForPublication(t *testing.T) {
	base := EligibilityCandidate{Source: "jobhub", SourceApproved: true, PublicationStatus: "published", ModerationStatus: "approved", PublicVisible: true, MaxAttempts: 3}
	if !EligibleForPublication(base) {
		t.Fatal("expected native approved public vacancy to be eligible")
	}
	external := base
	external.Source = "approved:external"
	external.PublicationStatus = ""
	if !EligibleForPublication(external) {
		t.Fatal("expected permitted public external vacancy to be eligible")
	}
	testExternal := external
	testExternal.SourceApproved = false
	testExternal.TestSources = true
	if !EligibleForPublication(testExternal) {
		t.Fatal("expected staging test source to be eligible with explicit opt-in")
	}
	for name, candidate := range map[string]EligibilityCandidate{
		"source permission missing": {Source: "jooble:kz", ModerationStatus: "approved", PublicVisible: true, MaxAttempts: 3},
		"demo source":               {Source: "demo:external", SourceApproved: true, DemoSource: true, ModerationStatus: "approved", PublicVisible: true, MaxAttempts: 3},
		"draft":                     {Source: "jobhub", SourceApproved: true, PublicationStatus: "draft", ModerationStatus: "approved", PublicVisible: true, MaxAttempts: 3},
		"rejected":                  {Source: "jobhub", SourceApproved: true, PublicationStatus: "published", ModerationStatus: "rejected", PublicVisible: true, MaxAttempts: 3},
		"private":                   {Source: "jobhub", SourceApproved: true, PublicationStatus: "published", ModerationStatus: "approved", PublicVisible: false, MaxAttempts: 3},
		"already sent":              {Source: "jobhub", SourceApproved: true, PublicationStatus: "published", ModerationStatus: "approved", PublicVisible: true, AlreadySent: true, MaxAttempts: 3},
		"retry exhausted":           {Source: "jobhub", SourceApproved: true, PublicationStatus: "published", ModerationStatus: "approved", PublicVisible: true, Attempts: 3, MaxAttempts: 3},
	} {
		t.Run(name, func(t *testing.T) {
			if EligibleForPublication(candidate) {
				t.Fatal("candidate should not be eligible")
			}
		})
	}
}
