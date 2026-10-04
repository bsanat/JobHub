package candidates

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUnavailableFileStoreRejectsOperations(t *testing.T) {
	store := UnavailableFileStore{}
	ctx := context.Background()
	if err := store.Put(ctx, "file.pdf", strings.NewReader("data")); !errors.Is(err, ErrFileStorageUnavailable) {
		t.Fatalf("Put error = %v", err)
	}
	if file, err := store.Open(ctx, "file.pdf"); file != nil || !errors.Is(err, ErrFileStorageUnavailable) {
		t.Fatalf("Open = %v, %v", file, err)
	}
	if err := store.Delete(ctx, "file.pdf"); !errors.Is(err, ErrFileStorageUnavailable) {
		t.Fatalf("Delete error = %v", err)
	}
}
