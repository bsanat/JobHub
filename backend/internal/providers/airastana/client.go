package airastana

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/safety"
)

const (
	Source           = "airastana:careers"
	baseURL          = "https://job.airastana.com"
	maxResponseBytes = 5 << 20
)

var vacancyIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type Client struct {
	http        *http.Client
	budget      *safety.RequestBudget
	maxJobs     int
	detailLimit int
}

type vacancy struct {
	ID          string `json:"ID"`
	Job         string `json:"Job"`
	Description string `json:"Description"`
	CityName    string `json:"CityName"`
	CountryName string `json:"CountryName"`
	Category    string `json:"CategoryName"`
	StartDate   string `json:"StartDate"`
	FinalDate   string `json:"FinalDate"`
}

func NewClient(maxRequests, maxJobs int, client *http.Client) (*Client, error) {
	return NewClientWithDetailLimit(maxRequests, maxJobs, 0, client)
}

func NewClientWithDetailLimit(maxRequests, maxJobs, detailLimit int, client *http.Client) (*Client, error) {
	if maxRequests < 1 || maxRequests > 10 || maxJobs < 1 || maxJobs > 200 || detailLimit < 0 || detailLimit > maxRequests-1 {
		return nil, providers.Failure("INVALID_BUDGET")
	}
	if client == nil {
		client = &http.Client{}
	}
	configured := *client
	configured.Timeout = 20 * time.Second
	configured.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	budget, _ := safety.NewRequestBudget(maxRequests)
	return &Client{http: &configured, budget: budget, maxJobs: maxJobs, detailLimit: detailLimit}, nil
}

func (*Client) Source() string { return Source }

func (c *Client) Collect(ctx context.Context) (result providers.Result, err error) {
	defer func() {
		metrics := c.budget.Metrics()
		result.MaxRequests = metrics.MaxRequests
		result.RequestsUsed = metrics.RequestsUsed
		result.RemainingRequests = metrics.Remaining
	}()
	if err := c.budget.Acquire(); err != nil {
		return result, providers.Failure("REQUEST_BUDGET_EXHAUSTED")
	}
	result.Requests++
	result.ListRequests++
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/vacancies/get", strings.NewReader("{}"))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "JobHub-Career-POC/1.0")
	response, err := c.http.Do(req)
	if err != nil {
		return result, providers.Failure("TRANSPORT_FAILED")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return result, providers.Failure("HTTP_FAILED")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return result, providers.Failure("INVALID_RESPONSE")
	}
	var rawItems []json.RawMessage
	if err := json.Unmarshal(body, &rawItems); err != nil || rawItems == nil {
		return result, providers.Failure("INVALID_RESPONSE")
	}
	result.PagesFetched = 1
	result.Fetched = len(rawItems)
	if result.Fetched > c.maxJobs {
		return result, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	for _, raw := range rawItems {
		var v vacancy
		if err := json.Unmarshal(raw, &v); err != nil {
			result.Malformed++
			continue
		}
		item, ok := normalize(v)
		if !ok {
			result.Malformed++
			continue
		}
		if item.Description == "" && result.DetailRequests < c.detailLimit {
			if detailed, ok := c.detail(ctx, v.ID, &result); ok {
				item.Description = detailed.Description
				item.DescriptionKind = "full"
			} else {
				result.DetailUnavailable++
			}
		}
		result.Items = append(result.Items, item)
	}
	// The upstream list response has no reliable total or pagination marker.
	return result, nil
}

func (c *Client) detail(ctx context.Context, id string, result *providers.Result) (vacancy, bool) {
	if c.budget.Acquire() != nil {
		return vacancy{}, false
	}
	result.Requests++
	result.DetailRequests++
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/vacancies/"+id, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "JobHub-Career-POC/1.0")
	response, err := c.http.Do(req)
	if err != nil {
		return vacancy{}, false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return vacancy{}, false
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return vacancy{}, false
	}
	var item vacancy
	if json.Unmarshal(body, &item) != nil || item.ID != id || strings.TrimSpace(item.Description) == "" {
		return vacancy{}, false
	}
	return item, true
}

func normalize(v vacancy) (jobs.ImportedJob, bool) {
	if !vacancyIDPattern.MatchString(v.ID) || strings.TrimSpace(v.Job) == "" {
		return jobs.ImportedJob{}, false
	}
	published, err := parseDate(v.StartDate)
	if err != nil {
		return jobs.ImportedJob{}, false
	}
	expires, err := parseDate(v.FinalDate)
	if err != nil {
		return jobs.ImportedJob{}, false
	}
	location := strings.Trim(strings.Join([]string{strings.TrimSpace(v.CityName), strings.TrimSpace(v.CountryName)}, ", "), ", ")
	description := strings.TrimSpace(v.Description)
	kind := "snippet"
	if description != "" {
		kind = "full"
	}
	return jobs.ImportedJob{
		Source: Source, ExternalID: v.ID, SourceURL: baseURL + "/vacancies/detail/" + v.ID,
		CompanyNameRaw: "Air Astana", Title: strings.TrimSpace(v.Job), LocationRaw: location,
		Description: description, DescriptionKind: kind, Category: strings.TrimSpace(v.Category),
		ExternalPublishedAt: published, ExternalExpiresAt: expires,
	}, true
}

func parseDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
		if value, err := time.Parse(layout, raw); err == nil {
			return &value, nil
		}
	}
	return nil, providers.Failure("INVALID_DATE")
}
