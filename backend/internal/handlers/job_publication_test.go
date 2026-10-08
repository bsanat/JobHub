package handlers

import (
	"testing"
	"time"

	"jobhub-ai/backend/internal/jobs"
)

func TestImportedJobDoesNotInventPublicationDateFromFirstSeen(t *testing.T) {
	seen := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	job := jobs.Job{Source: "arbeitnow:global", FirstSeenAt: seen}
	if publicationDate(job) != nil {
		t.Fatal("first seen is not a publication date")
	}
	job.ExternalPublishedAt = &seen
	if got := publicationDate(job); got == nil || !got.Equal(seen) {
		t.Fatal("explicit external publication date missing")
	}
}
