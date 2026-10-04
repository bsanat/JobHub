package candidates

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSupabaseFileStoreLifecycle(t *testing.T) {
	key := strings.Repeat("a", 64) + ".pdf"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "Bearer secret" || r.Header.Get("apikey") != "secret" {
			t.Errorf("missing server-side authorization")
		}
		switch r.Method {
		case http.MethodPost:
			if r.URL.Path != "/storage/v1/object/resumes/"+key || r.Header.Get("Content-Type") != "application/pdf" || r.Header.Get("x-upsert") != "false" {
				t.Errorf("invalid upload request: %s %s", r.Method, r.URL.Path)
			}
			contents, _ := io.ReadAll(r.Body)
			if string(contents) != "%PDF-1.7" {
				t.Errorf("upload body = %q", contents)
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			if r.URL.Path != "/storage/v1/object/authenticated/resumes/"+key {
				t.Errorf("invalid download path: %s", r.URL.Path)
			}
			_, _ = io.WriteString(w, "%PDF-1.7")
		case http.MethodDelete:
			if r.URL.Path != "/storage/v1/object/resumes" {
				t.Errorf("invalid delete path: %s", r.URL.Path)
			}
			var body struct {
				Prefixes []string `json:"prefixes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Prefixes) != 1 || body.Prefixes[0] != key {
				t.Errorf("invalid delete body: %+v, %v", body, err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	store, err := NewSupabaseFileStore(server.URL, "secret", "resumes", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), key, strings.NewReader("%PDF-1.7")); err != nil {
		t.Fatal(err)
	}
	file, err := store.Open(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil || string(contents) != "%PDF-1.7" {
		t.Fatalf("download = %q, %v", contents, err)
	}
	if err := store.Delete(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want 3", requests)
	}
}

func TestSupabaseFileStoreRejectsInvalidKeyAndConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("invalid key reached storage")
	}))
	defer server.Close()
	store, err := NewSupabaseFileStore(server.URL, "secret", "resumes", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), "../unsafe.pdf", strings.NewReader("bad")); err == nil {
		t.Fatal("invalid upload key accepted")
	}
	if _, err := store.Open(context.Background(), "../unsafe.pdf"); err == nil {
		t.Fatal("invalid download key accepted")
	}
	if err := store.Delete(context.Background(), "../unsafe.pdf"); err == nil {
		t.Fatal("invalid delete key accepted")
	}
	for _, tc := range []struct{ url, key, bucket string }{
		{"http://example.com", "secret", "resumes"},
		{"https://example.com/path", "secret", "resumes"},
		{"https://example.com", "", "resumes"},
		{"https://example.com", "secret", "../resumes"},
	} {
		if _, err := NewSupabaseFileStore(tc.url, tc.key, tc.bucket, nil); err == nil {
			t.Fatalf("invalid configuration accepted: %q", tc.url)
		}
	}
}

func TestSupabaseFileStoreDoesNotReturnProviderBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "sensitive provider details")
	}))
	defer server.Close()
	store, err := NewSupabaseFileStore(server.URL, "secret", "resumes", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(context.Background(), strings.Repeat("b", 64)+".pdf", strings.NewReader("%PDF"))
	if err == nil || strings.Contains(err.Error(), "sensitive provider details") {
		t.Fatalf("unsafe error: %v", err)
	}
}
