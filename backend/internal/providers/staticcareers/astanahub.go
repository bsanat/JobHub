package staticcareers

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/locations"
	"jobhub-ai/backend/internal/providers"
)

var astanaHubDetailURL = regexp.MustCompile(`https://astanahub\.com/ru/vacancy/([0-9]+)`)
var astanaHubPublication = regexp.MustCompile(`displayDuration\('([^']+)'\)`)

const astanaHubEmployerProfile = "/account/v2/company/10658/profile/activity/"
const astanaHubEmployerName = "Autonomous cluster fund «Astana Hub»"

func (c *Client) collectAstanaHub(ctx context.Context, body []byte, out providers.Result) (providers.Result, error) {
	links := astanaHubListings(body)
	out.PagesFetched = 1
	out.Fetched = len(links)
	if len(links) > c.maxJobs {
		return out, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	for _, link := range links {
		page, err := c.fetch(ctx, link, &out, true)
		if err != nil {
			out.Skipped++
			continue
		}
		item, ok := astanaHubVacancy(c.cfg, link, page)
		if !ok {
			out.Malformed++
			continue
		}
		out.Items = append(out.Items, item)
	}
	// The HTML listing is paginated. It supplies fresh jobs for upsert but a
	// job absent from the first page may still be live on later pages.
	out.Complete = false
	return out, nil
}

func astanaHubListings(body []byte) []string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var links []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && hasClassToken(n, "vacancy-item") {
			match := astanaHubDetailURL.FindString(nodeAttr(n, "onclick"))
			if match != "" && !seen[match] {
				seen[match] = true
				links = append(links, match)
			}
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)
	return links
}

func astanaHubVacancy(cfg Config, raw string, body []byte) (jobs.ImportedJob, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host != "astanahub.com" || u.Scheme != "https" || !regexp.MustCompile(`^/ru/vacancy/[0-9]+$`).MatchString(u.Path) {
		return jobs.ImportedJob{}, false
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return jobs.ImportedJob{}, false
	}
	var title, city, employment, description string
	var officialEmployer, published bool
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inPills bool) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" && title == "" {
				title = nodeText(n)
			}
			if n.Data == "a" && nodeAttr(n, "href") == astanaHubEmployerProfile && strings.Contains(nodeText(n), astanaHubEmployerName) {
				officialEmployer = true
			}
			if hasClassToken(n, "pills") {
				inPills = true
			}
			if inPills && n.Data == "span" {
				value := nodeText(n)
				if city == "" && locations.Normalize(value) != "" {
					city = value
				}
				if value == "Полная занятость" || value == "Частичная занятость" || value == "Стажировка" {
					employment = value
				}
				if value == "Опубликовано" {
					published = true
				}
			}
			if description == "" && hasClassToken(n, "tiny-content-v2") {
				description = formattedContent(n)
				return
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch, inPills)
		}
	}
	walk(doc, false)
	if !officialEmployer || !published || title == "" || description == "" || locations.Normalize(city) == "" {
		return jobs.ImportedJob{}, false
	}
	var publication *time.Time
	if match := astanaHubPublication.FindSubmatch(body); len(match) == 2 {
		if parsed, err := time.Parse(time.RFC3339Nano, string(match[1])); err == nil {
			publication = &parsed
		}
	}
	return jobs.ImportedJob{
		Source: cfg.Source, ExternalID: strings.TrimPrefix(u.Path, "/ru/vacancy/"), SourceURL: raw,
		CompanyNameRaw: cfg.Company, Title: title, LocationRaw: city, Description: description,
		DescriptionKind: "full", EmploymentTypeRaw: employment, ExternalPublishedAt: publication,
	}, true
}
