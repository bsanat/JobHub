package staticcareers

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/locations"
	"jobhub-ai/backend/internal/providers"
)

// QSamruk company pages include subsidiary vacancies after the direct employer
// list. Only the first accordion is considered, and each detail must identify
// the configured employer before it may become a JobHub vacancy.
var qsamrukDetailPath = regexp.MustCompile(`^/vacancy/([0-9]+)$`)

type qsamrukListing struct {
	url, salary string
}

func (c *Client) collectQSamruk(ctx context.Context, body []byte, out providers.Result) (providers.Result, error) {
	list := qsamrukListings(body, c.cfg.ListingURL)
	out.PagesFetched = 1
	out.Fetched = len(list)
	if len(list) > c.maxJobs {
		return out, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	for _, listing := range list {
		page, err := c.fetch(ctx, listing.url, &out, true)
		if err != nil {
			out.Skipped++
			continue
		}
		item, ok := qsamrukVacancy(c.cfg, listing, page)
		if !ok {
			out.Malformed++
			continue
		}
		out.Items = append(out.Items, item)
	}
	// The employer page may omit older records; a missing listing is never
	// evidence that an existing vacancy was withdrawn.
	out.Complete = false
	return out, nil
}

func qsamrukListings(body []byte, listingURL string) []qsamrukListing {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil
	}
	base, err := url.Parse(listingURL)
	if err != nil {
		return nil
	}
	var accordion *html.Node
	var first func(*html.Node)
	first = func(n *html.Node) {
		if accordion != nil {
			return
		}
		if n.Type == html.ElementNode && hasClassToken(n, "accordeon") && hasClassToken(n, "allopen") {
			accordion = n
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			first(ch)
		}
	}
	first(doc)
	if accordion == nil {
		return nil
	}
	seen := map[string]bool{}
	var result []qsamrukListing
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" && hasClassToken(n, "vacancy-item") {
			href := nodeAttr(n, "href")
			candidate, err := url.Parse(href)
			if err == nil {
				resolved := base.ResolveReference(candidate)
				if resolved.Scheme == "https" && resolved.Host == base.Host && qsamrukDetailPath.MatchString(resolved.Path) && resolved.RawQuery == "" && !seen[resolved.Path] {
					seen[resolved.Path] = true
					var salary string
					var findSalary func(*html.Node)
					findSalary = func(ch *html.Node) {
						if ch.Type == html.ElementNode && hasClassToken(ch, "salary") {
							salary = nodeText(ch)
							return
						}
						for child := ch.FirstChild; child != nil && salary == ""; child = child.NextSibling {
							findSalary(child)
						}
					}
					findSalary(n)
					result = append(result, qsamrukListing{url: resolved.String(), salary: salary})
				}
			}
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(accordion)
	return result
}

func qsamrukVacancy(cfg Config, listing qsamrukListing, body []byte) (jobs.ImportedJob, bool) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return jobs.ImportedJob{}, false
	}
	var title, employerPath, city, status, description, salary, employment string
	var walk func(*html.Node, bool, bool)
	walk = func(n *html.Node, header, content bool) {
		if n.Type == html.ElementNode {
			if hasClassToken(n, "vacancy_header") {
				header = true
			}
			if hasClassToken(n, "vacancy_content") {
				content = true
			}
			if header {
				switch {
				case n.Data == "h1" && title == "":
					title = nodeText(n)
				case n.Data == "a" && hasClassToken(n, "comp"):
					employerPath = nodeAttr(n, "href")
				case n.Data == "a" && hasClassToken(n, "city"):
					city = nodeText(n)
				case hasClassToken(n, "status"):
					status = nodeText(n)
				case hasClassToken(n, "salary") && salary == "":
					salary = nodeText(n)
				case hasClassToken(n, "extra_info"):
					for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
						if ch.Type == html.ElementNode && ch.Data == "span" {
							employment = nodeText(ch)
						}
					}
				}
			}
			if content && hasClassToken(n, "descr") && description == "" {
				description = formattedContent(n)
				return
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch, header, content)
		}
	}
	walk(doc, false, false)
	employerPath = strings.TrimPrefix(strings.TrimPrefix(employerPath, "/ru"), "/kz")
	if employerPath != cfg.EmployerPath || (status != "Вакансия открыта" && status != "Бос орын ашық") ||
		locations.Normalize(city) == "" || title == "" || description == "" ||
		strings.Contains(title, "Все права защищены") || strings.Contains(title, "Барлық құқықтар қорғалған") {
		return jobs.ImportedJob{}, false
	}
	u, err := url.Parse(listing.url)
	if err != nil || !qsamrukDetailPath.MatchString(u.Path) {
		return jobs.ImportedJob{}, false
	}
	if salary == "" {
		salary = listing.salary
	}
	return jobs.ImportedJob{
		Source: cfg.Source, ExternalID: strings.TrimPrefix(u.Path, "/vacancy/"), SourceURL: listing.url,
		CompanyNameRaw: cfg.Company, Title: title, LocationRaw: city, Description: description,
		DescriptionKind: "full", EmploymentTypeRaw: employment, SalaryRaw: salary,
	}, true
}

func nodeAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}
