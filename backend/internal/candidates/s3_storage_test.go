package candidates

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestS3FileStoreLifecycle(t *testing.T) {
	key := strings.Repeat("a", 64) + ".pdf"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") || r.URL.Path != "/uploads/"+key {
			t.Error("missing SigV4 or wrong path")
		}
		switch r.Method {
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			if string(b) != "%PDF-1.7" || r.Header.Get("Content-Type") != "application/pdf" || r.Header.Get("If-None-Match") != "*" {
				t.Error("invalid upload")
			}
		case http.MethodGet:
			_, _ = io.WriteString(w, "%PDF-1.7")
		case http.MethodDelete:
			w.WriteHeader(204)
		default:
			t.Error("unexpected method")
		}
	}))
	defer server.Close()
	store, err := NewS3FileStore(server.URL, "us-east-2", "test", "test-secret", "uploads", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Put(context.Background(), key, strings.NewReader("%PDF-1.7")); err != nil {
		t.Fatal(err)
	}
	file, err := store.Open(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(file)
	file.Close()
	if string(b) != "%PDF-1.7" {
		t.Fatal("wrong body")
	}
	if err = store.Delete(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"../private.pdf", "bad"} {
		if err = store.Put(context.Background(), k, strings.NewReader("bad")); err == nil {
			t.Fatal("unsafe upload")
		}
		if _, err = store.Open(context.Background(), k); err == nil {
			t.Fatal("unsafe download")
		}
		if err = store.Delete(context.Background(), k); err == nil {
			t.Fatal("unsafe deletion")
		}
	}
	if requests != 3 {
		t.Fatalf("unsafe key reached storage: %d", requests)
	}
}

func TestS3ErrorsDoNotExposePrivateDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = io.WriteString(w, `<Error><Code>AccessDenied</Code><Message>private-secret-object</Message></Error>`)
	}))
	defer server.Close()
	store, _ := NewS3FileStore(server.URL, "us-east-2", "test", "secret", "uploads", server.Client())
	err := store.Put(context.Background(), strings.Repeat("b", 64)+".pdf", strings.NewReader("%PDF"))
	if err == nil || strings.Contains(err.Error(), "private-secret") {
		t.Fatalf("unsafe error: %v", err)
	}
	for _, u := range []string{"http://example.com", "https://secret@example.com", "https://example.com?token=bad"} {
		if _, err := NewS3FileStore(u, "us-east-2", "test", "secret", "uploads", nil); err == nil {
			t.Fatal("unsafe endpoint")
		}
	}
}
