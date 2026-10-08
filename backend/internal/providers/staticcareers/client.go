// Package staticcareers provides a bounded, configuration-driven adapter for
// public career pages whose vacancy details are available in ordinary HTML.
package staticcareers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/locations"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/safety"
)

type Config struct {
	Source, Company, ListingURL, DetailPrefix, LocationMode string
	ForcePartial                                            bool
}
type Client struct {
	cfg     Config
	http    *http.Client
	budget  *safety.RequestBudget
	maxJobs int
}

func New(cfg Config, maxRequests, maxJobs int, client *http.Client) (*Client, error) {
	if strings.TrimSpace(cfg.Source) == "" || strings.TrimSpace(cfg.ListingURL) == "" || maxRequests < 1 || maxRequests > 60 || maxJobs < 1 || maxJobs > 200 || (cfg.LocationMode != "" && cfg.LocationMode != "technodom" && cfg.LocationMode != "kolesa" && cfg.LocationMode != "bcc" && cfg.LocationMode != "sulpak" && cfg.LocationMode != "magnum" && cfg.LocationMode != "qic") {
		return nil, providers.Failure("INVALID_BUDGET")
	}
	if client == nil {
		client = &http.Client{}
	}
	c := *client
	c.Timeout = 20 * time.Second
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	b, _ := safety.NewRequestBudget(maxRequests)
	return &Client{cfg: cfg, http: &c, budget: b, maxJobs: maxJobs}, nil
}
func (c *Client) Source() string { return c.cfg.Source }
func (c *Client) Collect(ctx context.Context) (out providers.Result, err error) {
	defer func() {
		m := c.budget.Metrics()
		out.MaxRequests = m.MaxRequests
		out.RequestsUsed = m.RequestsUsed
		out.RemainingRequests = m.Remaining
	}()
	body, err := c.fetch(ctx, c.cfg.ListingURL, &out, false)
	if err != nil {
		return out, err
	}
	links := detailLinks(body, c.cfg.ListingURL, c.cfg.DetailPrefix, c.cfg.LocationMode)
	out.Fetched = len(links)
	if out.Fetched > c.maxJobs {
		return out, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	out.PagesFetched = 1
	for _, link := range links {
		d, e := c.fetch(ctx, link, &out, true)
		if e != nil {
			out.Skipped++
			continue
		}
		item, ok := normalize(c.cfg, link, d)
		if !ok {
			out.Malformed++
			continue
		}
		out.Items = append(out.Items, item)
	}
	out.Complete = !c.cfg.ForcePartial && out.Skipped == 0 && out.Malformed == 0 && len(out.Items) == len(links)
	return out, nil
}
func (c *Client) fetch(ctx context.Context, raw string, out *providers.Result, detail bool) ([]byte, error) {
	if c.budget.Acquire() != nil {
		return nil, providers.Failure("REQUEST_BUDGET_EXHAUSTED")
	}
	out.Requests++
	if detail {
		out.DetailRequests++
	} else {
		out.ListRequests++
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	req.Header.Set("User-Agent", "JobHub-Career-POC/1.0")
	res, e := c.http.Do(req)
	if e != nil {
		return nil, providers.Failure("TRANSPORT_FAILED")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, providers.Failure("HTTP_FAILED")
	}
	b, e := io.ReadAll(io.LimitReader(res.Body, 3<<20))
	if e != nil {
		return nil, providers.Failure("RESPONSE_READ_FAILED")
	}
	return b, nil
}
func detailLinks(body []byte, listing, prefix, mode string) []string {
	base, _ := url.Parse(listing)
	seen := map[string]bool{}
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					u, e := url.Parse(strings.TrimSpace(a.Val))
					if e == nil {
						u = base.ResolveReference(u)
						if mode == "bcc" && bccDetailPath.MatchString(u.Path) && !strings.HasSuffix(u.Path, "/") {
							u.Path += "/"
						}
						raw := u.String()
						if u.Host == base.Host && strings.HasPrefix(u.Path, prefix) && raw != listing && !seen[raw] &&
							(mode != "bcc" || bccDetailPath.MatchString(u.Path)) &&
							(mode != "qic" || qicDetailPath.MatchString(u.Path)) {
							seen[raw] = true
							out = append(out, raw)
						}
					}
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	doc, e := html.Parse(strings.NewReader(string(body)))
	if e == nil {
		walk(doc)
	}
	return out
}

var bccDetailPath = regexp.MustCompile(`^/en/career/[0-9]+/?$`)
var qicDetailPath = regexp.MustCompile(`^/vakansii/[0-9]+/?$`)

var locationPattern = regexp.MustCompile(`(?i)(Казахстан,\s*[А-ЯA-ZЁӘІҢҒҚӨҰҮҺа-яёәіңғқөұүһ ,/-]+(?:,\s*KZ)?)`)

func normalize(cfg Config, raw string, body []byte) (jobs.ImportedJob, bool) {
	doc, e := html.Parse(strings.NewReader(string(body)))
	if e != nil {
		return jobs.ImportedJob{}, false
	}
	var title, pageTitle, loc, desc, salary string
	var text func(*html.Node)
	text = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1":
				if title == "" {
					title = nodeText(n)
				}
			case "title":
				pageTitle = nodeText(n)
			case "main", "article":
				if desc == "" {
					desc = nodeText(n)
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			text(ch)
		}
	}
	text(doc)
	if title == "" {
		title = pageTitle
	}
	if title == "" {
		title = "Vacancy"
	}
	if desc == "" {
		desc = nodeText(doc)
	}
	if cfg.LocationMode == "technodom" {
		if structured := technodomDescription(doc); structured != "" {
			desc = structured
		}
	}
	if cfg.LocationMode == "bcc" {
		if city, body := bccVacancyContent(doc); city != "" {
			loc, desc = city, body
		}
	}
	if cfg.LocationMode == "sulpak" {
		if name, city, body := sulpakVacancyContent(doc); name != "" {
			title, loc, desc = name, city, body
		}
	}
	if cfg.LocationMode == "magnum" {
		loc, desc, salary = magnumVacancyContent(doc)
	}
	if cfg.LocationMode == "qic" {
		loc, desc = qicVacancyContent(doc)
	}
	if match := locationPattern.FindString(desc); match != "" {
		loc = strings.TrimSpace(match)
	}
	if specific, rawSalary := specificLocation(doc, cfg.LocationMode); specific != "" {
		loc = specific
		salary = rawSalary
	}
	title = strings.TrimSpace(title)
	if title == "" || title == "Vacancy" {
		return jobs.ImportedJob{}, false
	}
	sum := sha256.Sum256([]byte(raw))
	id := "url_" + hex.EncodeToString(sum[:8])
	return jobs.ImportedJob{Source: cfg.Source, ExternalID: id, SourceURL: raw, CompanyNameRaw: cfg.Company, Title: title, LocationRaw: strings.TrimSpace(loc), Description: strings.TrimSpace(desc), DescriptionKind: "full", SalaryRaw: salary}, true
}

func bccVacancyContent(doc *html.Node) (string, string) {
	var city, description string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
				if sibling.Type != html.ElementNode {
					continue
				}
				if city == "" && classContains(sibling, "text-sm") {
					candidate := nodeText(sibling)
					if locations.Normalize(candidate) != "" {
						city = candidate
					}
				}
				if city != "" && classContains(sibling, "text-neutral-700") {
					description = formattedContent(sibling)
					return
				}
			}
		}
		for child := n.FirstChild; child != nil && description == ""; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return city, description
}

func sulpakVacancyContent(doc *html.Node) (string, string, string) {
	var name, city, description string
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inVacancy bool) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "vacancy" {
					inVacancy = true
				}
			}
			if inVacancy && name == "" && n.Data == "strong" {
				full := nodeText(n)
				if comma := strings.LastIndex(full, ","); comma > 0 {
					candidate := strings.TrimSpace(full[comma+1:])
					if locations.Normalize(candidate) != "" {
						name, city = strings.TrimSpace(full[:comma]), candidate
					}
				}
			}
			if inVacancy && description == "" && classContains(n, "vacancy-info") {
				description = formattedContent(n)
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, inVacancy)
		}
	}
	walk(doc, false)
	return name, city, description
}

func magnumVacancyContent(doc *html.Node) (string, string, string) {
	var city, salary string
	var sections []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if city == "" && classContains(n, "vd-right-city") {
				city = nodeText(n)
			}
			if salary == "" && classContains(n, "vd-salary") {
				salary = nodeText(n)
			}
			if classContains(n, "vd-terms-section") || classContains(n, "vd-unorderedList-section") {
				if content := formattedContent(n); content != "" {
					sections = append(sections, content)
				}
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return strings.TrimSpace(city), strings.Join(sections, "\n"), strings.TrimSpace(salary)
}

func qicVacancyContent(doc *html.Node) (string, string) {
	var city, description string
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inHeader bool) {
		if n.Type == html.ElementNode {
			if classContains(n, "detail_top") {
				inHeader = true
			}
			if inHeader && city == "" && classContains(n, "city") {
				candidate := nodeText(n)
				if locations.Normalize(candidate) != "" {
					city = candidate
				}
			}
			if description == "" && classContains(n, "detail_text") {
				description = formattedContent(n)
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, inHeader)
		}
	}
	walk(doc, false)
	return strings.TrimSpace(city), strings.TrimSpace(description)
}

func formattedContent(root *html.Node) string {
	var lines []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var prefix string
			switch n.Data {
			case "h2", "h3", "h4":
				prefix = "## "
			case "li":
				prefix = "- "
			case "p":
				prefix = ""
			default:
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					walk(child)
				}
				return
			}
			if value := nodeText(n); value != "" {
				lines = append(lines, prefix+value)
			}
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return strings.Join(lines, "\n")
}

func specificLocation(doc *html.Node, mode string) (string, string) {
	if mode == "" {
		return "", ""
	}
	var cities []string
	var salary string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch mode {
			case "technodom":
				if n.Data == "h2" && classContains(n, "col-cities-col-item-city") {
					city := strings.TrimSpace(nodeText(n))
					if city != "" {
						cities = append(cities, city)
					}
				}
			case "kolesa":
				if n.Data == "ul" && classContains(n, "_tags_") && len(cities) == 0 {
					for child := n.FirstChild; child != nil; child = child.NextSibling {
						if child.Type == html.ElementNode && child.Data == "li" {
							city := strings.TrimSpace(nodeText(child))
							if locations.Normalize(city) != "" {
								cities = append(cities, city)
								break
							}
						}
					}
				}
				if salary == "" && classContains(n, "_salary_") {
					salary = strings.TrimSpace(nodeText(n))
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if len(cities) == 1 {
		return cities[0], salary
	}
	if len(cities) > 1 {
		return "Қазақстан — бірнеше қала", salary
	}
	return "", salary
}

func classContains(n *html.Node, fragment string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, fragment) {
			return true
		}
	}
	return false
}

func technodomDescription(doc *html.Node) string {
	var intro string
	var sections []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if intro == "" && classContains(n, "col-description") {
				intro = nodeText(n)
			}
			if n.Data == "div" && hasClassToken(n, "row") {
				var heading string
				var tips []string
				var read func(*html.Node)
				read = func(child *html.Node) {
					if child.Type == html.ElementNode {
						if heading == "" && child.Data == "h2" && classContains(child, "col-title") {
							heading = nodeText(child)
						}
						if hasClassToken(child, "tip") {
							if value := nodeText(child); value != "" {
								tips = append(tips, "- "+value)
							}
							return
						}
					}
					for c := child.FirstChild; c != nil; c = c.NextSibling {
						read(c)
					}
				}
				read(n)
				lower := strings.ToLower(heading)
				if len(tips) > 0 && (strings.Contains(lower, "требован") || strings.Contains(lower, "обязан") || strings.Contains(lower, "предлага") || strings.Contains(lower, "талап") || strings.Contains(lower, "міндет") || strings.Contains(lower, "ұсынамыз")) {
					sections = append(sections, "## "+heading+"\n"+strings.Join(tips, "\n"))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if intro == "" && len(sections) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(append([]string{intro}, sections...), "\n\n"))
}

func hasClassToken(n *html.Node, wanted string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			for _, token := range strings.Fields(attr.Val) {
				if token == wanted {
					return true
				}
			}
		}
	}
	return false
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.ElementNode {
			switch x.Data {
			case "script", "style", "noscript", "svg", "form", "nav", "footer", "iframe":
				return
			}
		}
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
			b.WriteByte(' ')
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
