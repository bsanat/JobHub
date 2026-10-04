package candidates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var storageBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type SupabaseFileStore struct {
	baseURL string
	bucket  string
	key     string
	client  *http.Client
}

func NewSupabaseFileStore(projectURL, secretKey, bucket string, client *http.Client) (*SupabaseFileStore, error) {
	parsed, err := url.Parse(projectURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" || (parsed.Scheme != "https" && !(parsed.Scheme == "http" && parsed.Hostname() == "127.0.0.1")) {
		return nil, fmt.Errorf("SUPABASE_URL must be a project HTTPS origin")
	}
	if !storageBucketPattern.MatchString(bucket) {
		return nil, fmt.Errorf("SUPABASE_RESUME_BUCKET is invalid")
	}
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("SUPABASE_SECRET_KEY is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &SupabaseFileStore{baseURL: projectURL, bucket: bucket, key: secretKey, client: client}, nil
}

func (s *SupabaseFileStore) objectURL(key string, authenticated bool) (string, error) {
	if !storageKeyPattern.MatchString(key) {
		return "", fmt.Errorf("invalid storage key")
	}
	path := "/storage/v1/object/"
	if authenticated {
		path += "authenticated/"
	}
	return s.baseURL + path + s.bucket + "/" + key, nil
}

func (s *SupabaseFileStore) request(ctx context.Context, method, target string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.key)
	req.Header.Set("Authorization", "Bearer "+s.key)
	return req, nil
}

func (s *SupabaseFileStore) Put(ctx context.Context, key string, source io.Reader) error {
	target, err := s.objectURL(key, false)
	if err != nil {
		return err
	}
	req, err := s.request(ctx, http.MethodPost, target, source)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/pdf")
	req.Header.Set("x-upsert", "false")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("upload resume: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload resume: storage returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *SupabaseFileStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := s.objectURL(key, true)
	if err != nil {
		return nil, err
	}
	req, err := s.request(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download resume: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download resume: storage returned HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (s *SupabaseFileStore) Delete(ctx context.Context, key string) error {
	if !storageKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid storage key")
	}
	data, err := json.Marshal(struct {
		Prefixes []string `json:"prefixes"`
	}{Prefixes: []string{key}})
	if err != nil {
		return err
	}
	req, err := s.request(ctx, http.MethodDelete, s.baseURL+"/storage/v1/object/"+s.bucket, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete resume: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("delete resume: storage returned HTTP %d", resp.StatusCode)
	}
	return nil
}
