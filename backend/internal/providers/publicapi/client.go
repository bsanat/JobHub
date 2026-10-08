package publicapi

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/providers/greenhouse"
)

const arbeitnowURL = "https://www.arbeitnow.com/api/job-board-api"
const leverURL = "https://api.lever.co/v0/postings/"
const maxResponseBytes = 20 << 20

var boardPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

type Client struct {
	source, kind, board, company, base string
	http                               *http.Client
	maxPages, pageSize                 int
}

func NewArbeitnow(maxPages int, client *http.Client) (*Client, error) {
	return newClient("arbeitnow", "", "", arbeitnowURL, maxPages, client)
}

func NewLever(board, company string, maxPages int, client *http.Client) (*Client, error) {
	return newClient("lever", board, company, leverURL, maxPages, client)
}

func newClient(kind, board, company, base string, maxPages int, client *http.Client) (*Client, error) {
	if maxPages < 1 || maxPages > 50 || (kind == "lever" && (!boardPattern.MatchString(board) || strings.TrimSpace(company) == "")) {
		return nil, providers.Failure("INVALID_SOURCE_CONFIG")
	}
	if client == nil {
		client = &http.Client{}
	}
	copyClient := *client
	copyClient.Timeout = 20 * time.Second
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	source := "arbeitnow:global"
	if kind == "lever" {
		source = "lever:" + board
	}
	return &Client{source: source, kind: kind, board: board, company: strings.TrimSpace(company), base: base, http: &copyClient, maxPages: maxPages, pageSize: 100}, nil
}

func (c *Client) Source() string { return c.source }

func (c *Client) Collect(ctx context.Context) (providers.Result, error) {
	if c.kind == "arbeitnow" {
		return c.collectArbeitnow(ctx)
	}
	return c.collectLever(ctx)
}

func (c *Client) fetch(ctx context.Context, rawURL string, result *providers.Result) ([]byte, error) {
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, providers.Failure("INVALID_SOURCE_CONFIG")
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "JobHub-PublicJobs/1.0")
		result.Requests++
		res, err := c.http.Do(req)
		if err != nil {
			if attempt < 2 && ctx.Err() == nil {
				if !pause(ctx, time.Duration(attempt+1)*200*time.Millisecond) {
					return nil, providers.Failure("TRANSPORT_FAILED")
				}
				continue
			}
			return nil, providers.Failure("TRANSPORT_FAILED")
		}
		if res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500 {
			if res.StatusCode == http.StatusTooManyRequests && res.Header.Get("Retry-After") != "" {
				seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After"))
				if parseErr != nil || seconds < 0 || seconds > 3 {
					res.Body.Close()
					return nil, providers.Failure("RATE_LIMITED")
				}
			}
			wait := time.Duration(attempt+1) * 200 * time.Millisecond
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds >= 0 && seconds <= 3 {
				wait = time.Duration(seconds) * time.Second
			}
			res.Body.Close()
			if attempt < 2 && pause(ctx, wait) {
				continue
			}
			if res.StatusCode == http.StatusTooManyRequests {
				return nil, providers.Failure("RATE_LIMITED")
			}
			return nil, providers.Failure("HTTP_FAILED")
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return nil, providers.Failure("HTTP_FAILED")
		}
		contentType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
		if err != nil || contentType != "application/json" {
			res.Body.Close()
			return nil, providers.Failure("INVALID_CONTENT_TYPE")
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
		res.Body.Close()
		if err != nil || len(body) > maxResponseBytes {
			return nil, providers.Failure("INVALID_RESPONSE")
		}
		result.ListRequests++
		result.PagesFetched++
		return body, nil
	}
	return nil, providers.Failure("HTTP_FAILED")
}

func pause(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (c *Client) collectArbeitnow(ctx context.Context) (providers.Result, error) {
	result := providers.Result{}
	for page := 1; page <= c.maxPages; page++ {
		u := c.base
		if page > 1 {
			u += "?page=" + strconv.Itoa(page)
		}
		body, err := c.fetch(ctx, u, &result)
		if err != nil {
			return result, err
		}
		var response struct {
			Data  []json.RawMessage `json:"data"`
			Links *struct {
				Next json.RawMessage `json:"next"`
			} `json:"links"`
		}
		if json.Unmarshal(body, &response) != nil || response.Data == nil || response.Links == nil || len(response.Links.Next) == 0 {
			return result, providers.Failure("INVALID_RESPONSE")
		}
		result.Fetched += len(response.Data)
		for _, raw := range response.Data {
			var row arbeitnowJob
			if json.Unmarshal(raw, &row) != nil {
				result.Malformed++
				continue
			}
			item, ok := normalizeArbeitnow(row)
			if !ok {
				result.Malformed++
				continue
			}
			result.Items = append(result.Items, item)
		}
		if string(response.Links.Next) == "null" {
			result.Complete = result.Malformed == 0
			return result, nil
		}
		var nextURL string
		if json.Unmarshal(response.Links.Next, &nextURL) != nil || nextURL == "" {
			return result, providers.Failure("INVALID_PAGINATION")
		}
		next, err := url.Parse(nextURL)
		base, _ := url.Parse(c.base)
		if err != nil || next.Scheme != "https" || next.Host != base.Host || next.Path != base.Path || next.Query().Get("page") != strconv.Itoa(page+1) {
			return result, providers.Failure("INVALID_PAGINATION")
		}
	}
	return result, nil // A page cap is a partial snapshot; never archive absent jobs.
}

type arbeitnowJob struct {
	Slug, Company, Title, Description, URL, Location string
	Remote                                           bool
	JobTypes                                         []string
}

func (j *arbeitnowJob) UnmarshalJSON(raw []byte) error {
	var v struct {
		Slug        string   `json:"slug"`
		Company     string   `json:"company_name"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		URL         string   `json:"url"`
		Location    string   `json:"location"`
		Remote      bool     `json:"remote"`
		JobTypes    []string `json:"job_types"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	*j = arbeitnowJob{v.Slug, v.Company, v.Title, v.Description, v.URL, v.Location, v.Remote, v.JobTypes}
	return nil
}

func normalizeArbeitnow(v arbeitnowJob) (jobs.ImportedJob, bool) {
	if strings.TrimSpace(v.Slug) == "" || !validURL(v.URL) || strings.TrimSpace(v.Title) == "" {
		return jobs.ImportedJob{}, false
	}
	mode := ""
	if v.Remote {
		mode = "remote"
	}
	description := greenhouse.FormattedText(v.Description)
	kind := "snippet"
	if description != "" {
		kind = "full"
	}
	return jobs.ImportedJob{Source: "arbeitnow:global", ExternalID: v.Slug, SourceURL: v.URL, UpstreamSourceName: "Arbeitnow", CompanyNameRaw: greenhouse.PlainText(v.Company), Title: greenhouse.PlainText(v.Title), LocationRaw: greenhouse.PlainText(v.Location), Description: description, DescriptionKind: kind, EmploymentTypeRaw: strings.Join(v.JobTypes, ", "), EmploymentType: normalizeEmployment(v.JobTypes...), WorkMode: mode}, true
}

func (c *Client) collectLever(ctx context.Context) (providers.Result, error) {
	result := providers.Result{}
	for page := 0; page < c.maxPages; page++ {
		u := c.base + c.board + "?mode=json&skip=" + strconv.Itoa(page*c.pageSize) + "&limit=" + strconv.Itoa(c.pageSize)
		body, err := c.fetch(ctx, u, &result)
		if err != nil {
			return result, err
		}
		var rows []json.RawMessage
		if json.Unmarshal(body, &rows) != nil || rows == nil {
			return result, providers.Failure("INVALID_RESPONSE")
		}
		result.Fetched += len(rows)
		for _, raw := range rows {
			var row leverJob
			if json.Unmarshal(raw, &row) != nil {
				result.Malformed++
				continue
			}
			item, ok := normalizeLever(c.source, c.company, row)
			if !ok {
				result.Malformed++
				continue
			}
			result.Items = append(result.Items, item)
		}
		if len(rows) < c.pageSize {
			result.Complete = result.Malformed == 0
			return result, nil
		}
	}
	return result, nil
}

type leverJob struct {
	ID               string `json:"id"`
	Title            string `json:"text"`
	HostedURL        string `json:"hostedUrl"`
	ApplyURL         string `json:"applyUrl"`
	Description      string `json:"description"`
	DescriptionPlain string `json:"descriptionPlain"`
	AdditionalPlain  string `json:"additionalPlain"`
	WorkplaceType    string `json:"workplaceType"`
	Country          string `json:"country"`
	Lists            []struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	} `json:"lists"`
	Categories struct {
		Location     string   `json:"location"`
		Commitment   string   `json:"commitment"`
		AllLocations []string `json:"allLocations"`
	} `json:"categories"`
}

func normalizeLever(source, company string, v leverJob) (jobs.ImportedJob, bool) {
	link := v.ApplyURL
	if link == "" {
		link = v.HostedURL
	}
	if strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.Title) == "" || !validURL(link) {
		return jobs.ImportedJob{}, false
	}
	mode := ""
	switch strings.ToLower(v.WorkplaceType) {
	case "remote":
		mode = "remote"
	case "hybrid":
		mode = "hybrid"
	case "on-site", "onsite":
		mode = "on_site"
	}
	description := v.DescriptionPlain
	if description == "" {
		description = v.Description
	}
	parts := []string{greenhouse.FormattedText(description)}
	for _, list := range v.Lists {
		parts = append(parts, strings.TrimSpace("## "+greenhouse.PlainText(list.Text)+"\n\n"+greenhouse.FormattedText(list.Content)))
	}
	parts = append(parts, greenhouse.FormattedText(v.AdditionalPlain))
	fullDescription := strings.TrimSpace(strings.Join(parts, "\n\n"))
	kind := "snippet"
	if fullDescription != "" {
		kind = "full"
	}
	locations := append([]string{v.Categories.Location}, v.Categories.AllLocations...)
	seen := map[string]bool{}
	locationParts := []string{}
	for _, raw := range locations {
		value := greenhouse.PlainText(raw)
		if value != "" && !seen[value] {
			seen[value] = true
			locationParts = append(locationParts, value)
		}
	}
	if v.Country != "" && !strings.Contains(strings.Join(locationParts, " "), v.Country) {
		locationParts = append(locationParts, greenhouse.PlainText(v.Country))
	}
	return jobs.ImportedJob{Source: source, ExternalID: v.ID, SourceURL: link, UpstreamSourceName: "Lever", CompanyNameRaw: company, Title: greenhouse.PlainText(v.Title), LocationRaw: strings.Join(locationParts, "; "), Description: fullDescription, DescriptionKind: kind, EmploymentTypeRaw: greenhouse.PlainText(v.Categories.Commitment), EmploymentType: normalizeEmployment(v.Categories.Commitment), WorkMode: mode}, true
}

func normalizeEmployment(values ...string) string {
	chosen := ""
	for _, value := range values {
		var normalized string
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "full time", "full-time", "fulltime":
			normalized = "full_time"
		case "part time", "part-time", "parttime":
			normalized = "part_time"
		case "contract", "contractor":
			normalized = "contract"
		case "temporary", "temp":
			normalized = "temporary"
		case "intern", "internship":
			normalized = "internship"
		}
		if normalized != "" && chosen != "" && chosen != normalized {
			return ""
		}
		if normalized != "" {
			chosen = normalized
		}
	}
	return chosen
}

func validURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}
