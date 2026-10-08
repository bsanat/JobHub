package greenhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	htmlparser "golang.org/x/net/html"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/safety"
)

const endpoint = "https://boards-api.greenhouse.io/v1/boards/"
const maxBytes = 20 << 20

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)
var idPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

// Budget is shared across all boards and runs in one POC invocation.
// Attempts and raw records (including malformed/duplicate jobs) consume it.
type Budget struct {
	mu            sync.Mutex
	requests      *safety.RequestBudget
	jobs, maxJobs int
}

func NewBudget(requests, jobs int) (*Budget, error) {
	if requests < 1 || requests > 12 || jobs < 1 || jobs > 200 {
		return nil, providers.Failure("INVALID_BUDGET")
	}
	return newBudget(requests, jobs)
}

// NewImportBudget allows larger known boards in the one-shot public importer.
func NewImportBudget(requests, jobs int) (*Budget, error) {
	if requests < 1 || requests > 12 || jobs < 1 || jobs > 5000 {
		return nil, providers.Failure("INVALID_BUDGET")
	}
	return newBudget(requests, jobs)
}

func newBudget(requests, jobs int) (*Budget, error) {
	requestBudget, _ := safety.NewRequestBudget(requests)
	return &Budget{requests: requestBudget, maxJobs: jobs}, nil
}
func (b *Budget) request() bool {
	return b.requests.Acquire() == nil
}
func (b *Budget) records(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n > b.maxJobs-b.jobs {
		return false
	}
	b.jobs += n
	return true
}

type Client struct {
	board  string
	budget *Budget
	http   *http.Client
}

func NewClient(board string, budget *Budget, client *http.Client) (*Client, error) {
	if !tokenPattern.MatchString(board) || budget == nil {
		return nil, providers.Failure("INVALID_SOURCE_CONFIG")
	}
	if client == nil {
		client = &http.Client{}
	}
	copyClient := *client
	copyClient.Timeout = 20 * time.Second
	// No redirects, implicit retries or caller-supplied URL override. Tests inject
	// a transport, never a configurable network destination in the CLI.
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{board: board, budget: budget, http: &copyClient}, nil
}
func (c *Client) Source() string { return "greenhouse:" + c.board }
func (c *Client) Collect(ctx context.Context) (result providers.Result, err error) {
	defer func() {
		m := c.budget.requests.Metrics()
		result.MaxRequests = m.MaxRequests
		result.RequestsUsed = m.RequestsUsed
		result.RemainingRequests = m.Remaining
	}()
	res, err := c.request(ctx, &result)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		category := "HTTP_FAILED"
		switch {
		case res.StatusCode == 401 || res.StatusCode == 403:
			category = "ACCESS_DENIED"
		case res.StatusCode == 429:
			category = "RATE_LIMITED"
		case res.StatusCode == 404:
			category = "BOARD_NOT_FOUND"
		case res.StatusCode >= 300 && res.StatusCode < 400:
			category = "UNEXPECTED_REDIRECT"
		}
		return result, providers.Failure(category)
	}
	contentType, _, contentTypeErr := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if contentTypeErr != nil || contentType != "application/json" {
		return result, providers.Failure("INVALID_CONTENT_TYPE")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBytes+1))
	if err != nil {
		return result, providers.Failure("RESPONSE_READ_FAILED")
	}
	if len(body) > maxBytes {
		return result, providers.Failure("RESPONSE_TOO_LARGE")
	}
	result.ListRequests = 1
	result.PagesFetched = 1
	var response struct {
		Jobs json.RawMessage `json:"jobs"`
		Meta struct {
			Total *int `json:"total"`
		} `json:"meta"`
	}
	if json.Unmarshal(body, &response) != nil || len(response.Jobs) == 0 || bytes.Equal(bytes.TrimSpace(response.Jobs), []byte("null")) {
		return result, providers.Failure("INVALID_RESPONSE")
	}
	var rows []json.RawMessage
	if json.Unmarshal(response.Jobs, &rows) != nil {
		return result, providers.Failure("INVALID_RESPONSE")
	}
	result.Fetched = len(rows)
	if !c.budget.records(len(rows)) {
		return result, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	result.Complete = response.Meta.Total != nil && *response.Meta.Total == len(rows)
	for _, raw := range rows {
		var v vacancy
		if json.Unmarshal(raw, &v) != nil {
			result.Malformed++
			result.Complete = false
			continue
		}
		// Greenhouse identifies prospect posts with null internal_job_id. They are
		// not vacancies. Missing internal_job_id is not asserted to be a prospect.
		if bytes.Equal(bytes.TrimSpace(v.InternalID), []byte("null")) {
			result.Skipped++
			continue
		}
		item, err := normalize(c.Source(), v)
		if err != nil {
			result.Malformed++
			result.Complete = false
			continue
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (c *Client) request(ctx context.Context, result *providers.Result) (*http.Response, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if !c.budget.request() {
			return nil, providers.Failure("REQUEST_BUDGET_EXHAUSTED")
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+c.board+"/jobs?content=true", nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "JobHub-ATS-POC/1.0")
		result.Requests++
		res, err := c.http.Do(req)
		if err != nil {
			if attempt < 2 && c.budget.requests.Metrics().Remaining > 0 && retryPause(ctx, time.Duration(attempt+1)*200*time.Millisecond) {
				continue
			}
			return nil, providers.Failure("TRANSPORT_FAILED")
		}
		if res.StatusCode == 429 && res.Header.Get("Retry-After") != "" {
			seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After"))
			if parseErr != nil || seconds < 0 || seconds > 3 {
				return res, nil
			}
		}
		if (res.StatusCode == 429 || res.StatusCode >= 500) && attempt < 2 && c.budget.requests.Metrics().Remaining > 0 {
			wait := time.Duration(attempt+1) * 200 * time.Millisecond
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds >= 0 && seconds <= 3 {
				wait = time.Duration(seconds) * time.Second
			}
			res.Body.Close()
			if retryPause(ctx, wait) {
				continue
			}
			return nil, providers.Failure("TRANSPORT_FAILED")
		}
		return res, nil
	}
	return nil, providers.Failure("TRANSPORT_FAILED")
}

func retryPause(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

type vacancy struct {
	ID         json.RawMessage `json:"id"`
	InternalID json.RawMessage `json:"internal_job_id"`
	Title      string          `json:"title"`
	URL        string          `json:"absolute_url"`
	Content    string          `json:"content"`
	Company    string          `json:"company_name"`
	Location   struct {
		Name string `json:"name"`
	} `json:"location"`
	Updated string `json:"updated_at"`
}

func normalize(source string, v vacancy) (jobs.ImportedJob, error) {
	id := string(v.ID)
	if len(id) > 0 && id[0] == '"' {
		if json.Unmarshal(v.ID, &id) != nil {
			return jobs.ImportedJob{}, providers.Failure("INVALID_ID")
		}
	}
	// Never parse through float64 or substitute internal_job_id/requisition_id.
	if !idPattern.MatchString(id) {
		return jobs.ImportedJob{}, providers.Failure("INVALID_ID")
	}
	title := plainText(v.Title)
	parsed, err := url.Parse(v.URL)
	if title == "" || err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return jobs.ImportedJob{}, providers.Failure("INVALID_JOB")
	}
	var updated *time.Time
	if v.Updated != "" {
		value, err := time.Parse(time.RFC3339Nano, v.Updated)
		if err != nil {
			return jobs.ImportedJob{}, providers.Failure("INVALID_TIMESTAMP")
		}
		value = value.UTC()
		updated = &value
	}
	return jobs.ImportedJob{Source: source, ExternalID: id, SourceURL: v.URL, Title: title,
		CompanyNameRaw: plainText(v.Company), LocationRaw: plainText(v.Location.Name),
		Description: plainText(v.Content), DescriptionKind: "full", ExternalUpdatedAt: updated, ExternalUpdatedRaw: v.Updated}, nil
}

// Convert the API's entity-encoded HTML to plain text. No HTML is requested or
// rendered, and executable/script/style content is discarded.
func plainText(value string) string { return PlainText(value) }

// PlainText strips executable markup and returns text suitable for storage and display.
func PlainText(value string) string {
	for i := 0; i < 2; i++ {
		value = html.UnescapeString(value)
	}
	node, err := htmlparser.Parse(strings.NewReader(value))
	if err != nil {
		return ""
	}
	var text strings.Builder
	var walk func(*htmlparser.Node)
	walk = func(n *htmlparser.Node) {
		if n.Type == htmlparser.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "noscript") {
			return
		}
		if n.Type == htmlparser.TextNode {
			text.WriteString(n.Data)
			text.WriteByte(' ')
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(strings.Fields(text.String()), " ")
}
