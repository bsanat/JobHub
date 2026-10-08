package staticcareers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/locations"
	"jobhub-ai/backend/internal/providers"
)

// The bank's official careers page embeds the same vacancy data shown to
// visitors in __NEXT_DATA__. Do not infer publication dates from these IDs.
type alatauVacancy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Area        struct {
		Name string `json:"name"`
	} `json:"area"`
	Employment struct {
		Name string `json:"name"`
	} `json:"employment"`
	Salary *struct {
		From     *float64 `json:"from"`
		To       *float64 `json:"to"`
		Currency string   `json:"currency"`
	} `json:"salary"`
}

var alatauID = regexp.MustCompile(`^[0-9]+$`)

// Explicit names observed on the official bank listing. Unknown or foreign
// locations are skipped until checked; remote never implies KZ eligibility.
var alatauKazakhstanCities = map[string]bool{
	"аксукент": true, "актау": true, "актобе": true, "алматы": true,
	"алтай (зыряновск)": true, "астана": true, "астраханка": true,
	"атырау": true, "бесколь": true, "жанаозен": true, "караганда": true,
	"каскелен": true, "кокшетау": true, "конаев": true, "костанай": true,
	"павлодар": true, "петропавловск": true, "степногорск": true,
	"тараз": true, "темиртау": true, "уральск": true,
	"усть-каменогорск": true, "шу": true, "шымкент": true,
	"щучинск": true, "экибастуз": true,
}

func (c *Client) collectAlatau(body []byte, out providers.Result) (providers.Result, error) {
	out.PagesFetched = 1
	var vacancies []alatauVacancy
	if err := extractAlatauVacancies(body, &vacancies); err != nil {
		return out, providers.Failure("MALFORMED_LISTING")
	}
	out.Fetched = len(vacancies)
	if len(vacancies) > c.maxJobs {
		return out, providers.Failure("JOB_BUDGET_EXHAUSTED")
	}
	seen := make(map[string]bool, len(vacancies))
	for _, vacancy := range vacancies {
		city := strings.TrimSpace(vacancy.Area.Name)
		if !alatauID.MatchString(vacancy.ID) || seen[vacancy.ID] ||
			strings.TrimSpace(vacancy.Name) == "" || !alatauKazakhstanCities[strings.ToLower(city)] {
			out.Skipped++
			continue
		}
		seen[vacancy.ID] = true
		description, err := html.Parse(strings.NewReader(vacancy.Description))
		if err != nil {
			out.Malformed++
			continue
		}
		formatted := formattedContent(description)
		if formatted == "" {
			out.Malformed++
			continue
		}
		// The official bank listing is Kazakhstan-specific. Keep its city and
		// add the documented country only for cities not in JobHub's city map.
		if locations.Normalize(city) == "" {
			city += ", Казахстан"
		}
		out.Items = append(out.Items, jobs.ImportedJob{
			Source: c.cfg.Source, ExternalID: vacancy.ID,
			SourceURL:      "https://alataucitybank.kz/career/vacancy/" + vacancy.ID,
			CompanyNameRaw: c.cfg.Company, Title: strings.TrimSpace(vacancy.Name),
			LocationRaw: city, Description: formatted, DescriptionKind: "full",
			EmploymentTypeRaw: strings.TrimSpace(vacancy.Employment.Name),
			SalaryRaw:         alatauSalary(vacancy.Salary),
		})
	}
	// The embedded result has no pagination/completeness contract. New jobs are
	// upserted, but disappearing rows are never deactivated from this snapshot.
	out.Complete = false
	return out, nil
}

func extractAlatauVacancies(body []byte, into *[]alatauVacancy) error {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	var payload string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "script" {
			for _, attr := range node.Attr {
				if attr.Key == "id" && attr.Val == "__NEXT_DATA__" && node.FirstChild != nil {
					payload = node.FirstChild.Data
					return
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if payload == "" {
		return fmt.Errorf("embedded vacancy data missing")
	}
	var page struct {
		Props struct {
			PageProps struct {
				InitialState struct {
					API struct {
						Queries map[string]struct {
							Data struct {
								AllVacancies json.RawMessage `json:"allVacancies"`
							} `json:"data"`
						} `json:"queries"`
					} `json:"api"`
				} `json:"initialState"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal([]byte(payload), &page); err != nil {
		return err
	}
	const unfilteredQuery = `GetAllVacanciesWithFilter({"searchText":"","selectedCategory":"","selectedCity":"","selectedSchedules":""})`
	for key, query := range page.Props.PageProps.InitialState.API.Queries {
		if key != unfilteredQuery || len(query.Data.AllVacancies) == 0 {
			continue
		}
		return json.Unmarshal(query.Data.AllVacancies, into)
	}
	return fmt.Errorf("embedded vacancy collection missing")
}

func alatauSalary(salary *struct {
	From     *float64 `json:"from"`
	To       *float64 `json:"to"`
	Currency string   `json:"currency"`
}) string {
	if salary == nil || (salary.From == nil && salary.To == nil) || strings.TrimSpace(salary.Currency) == "" {
		return ""
	}
	if salary.From != nil && salary.To != nil {
		return fmt.Sprintf("%.0f–%.0f %s", *salary.From, *salary.To, salary.Currency)
	}
	if salary.From != nil {
		return fmt.Sprintf("от %.0f %s", *salary.From, salary.Currency)
	}
	return fmt.Sprintf("до %.0f %s", *salary.To, salary.Currency)
}
