package kzcompanies

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/locations"
	"jobhub-ai/backend/internal/providers"
	"jobhub-ai/backend/internal/providers/greenhouse"
)

var htmlTag = regexp.MustCompile(`(?i)</?[a-z][^>]*>`)

// Provider keeps company imports distinct from existing native and career
// sources. It never converts a location-free or globally remote post into a
// Kazakhstan vacancy.
type Provider struct {
	SourceConfig         Source
	Upstream             providers.VacancyProvider
	ExcludedOriginalURLs map[string]bool
}

func (p Provider) Source() string { return "kzcompany:" + p.SourceConfig.ID }

func (p Provider) Collect(ctx context.Context) (providers.Result, error) {
	result, err := p.Upstream.Collect(ctx)
	if err != nil {
		result.Complete = false
		return result, err
	}
	// An empty response can be a changed page contract or a failed renderer.
	// Never interpret it as permission to remove all existing vacancies.
	if result.Fetched == 0 {
		result.Complete = false
	}
	items := make([]jobs.ImportedJob, 0, len(result.Items))
	seenURLs := make(map[string]bool, len(result.Items))
	career, _ := url.Parse(p.SourceConfig.CareersURL)
	for _, item := range result.Items {
		original, parseErr := url.Parse(item.SourceURL)
		if parseErr != nil || original.Scheme != "https" || original.Hostname() != career.Hostname() || original.User != nil || original.Fragment != "" {
			result.Skipped++
			result.Complete = false
			continue
		}
		// Preserve the original application URL, including parameters a career
		// site may require. Ignore tracking parameters only for deduplication.
		identity := *original
		query := identity.Query()
		for name := range query {
			if strings.HasPrefix(strings.ToLower(name), "utm_") {
				query.Del(name)
			}
		}
		identity.RawQuery = query.Encode()
		key := identity.String()
		if seenURLs[key] {
			result.Skipped++
			continue
		}
		seenURLs[key] = true
		if p.ExcludedOriginalURLs[item.SourceURL] {
			result.Skipped++
			continue
		}
		if !KazakhstanEligible(item.LocationRaw) && !kazakhstanRemoteEligible(item.LocationRaw, item.Description) {
			result.Skipped++
			if strings.TrimSpace(item.LocationRaw) == "" {
				result.Complete = false
			}
			continue
		}
		item.Source = p.Source()
		item.CompanyNameRaw = p.SourceConfig.Company
		if htmlTag.MatchString(item.Description) {
			item.Description = greenhouse.FormattedText(item.Description)
		}
		items = append(items, item)
	}
	result.Items = items
	return result, nil
}

// A generic remote label alone is not eligibility. Require an explicit
// applicant-location statement in the original description.
func kazakhstanRemoteEligible(location, description string) bool {
	place := normalizePlace(location)
	if hasForeignCountry(place) || (!strings.Contains(place, "remote") && !strings.Contains(place, "қашықтан") && !strings.Contains(place, "удаленно") && !strings.Contains(place, "удалённо")) {
		return false
	}
	text := normalizePlace(description)
	for _, phrase := range []string{
		"remote from kazakhstan", "remote in kazakhstan", "based in kazakhstan", "residents of kazakhstan",
		"қазақстаннан қашықтан", "қазақстанда тұратын", "удаленно из казахстана", "удалённо из казахстана", "резидентов казахстана",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func KazakhstanEligible(location string) bool {
	if strings.TrimSpace(location) == "" {
		return false
	}
	normalized := normalizePlace(location)
	// A city name in a cross-border listing must not override an explicit
	// foreign location. Ambiguous multinational posts need manual review.
	if hasForeignCountry(normalized) {
		return false
	}
	for _, country := range []string{"kazakhstan", "қазақстан", "казахстан", "republic of kazakhstan"} {
		if strings.Contains(" "+normalized+" ", " "+country+" ") {
			return true
		}
	}
	for _, city := range locations.Cities {
		for _, name := range []string{city.ID, city.KK, city.RU, city.EN} {
			if strings.Contains(" "+normalized+" ", " "+normalizePlace(name)+" ") {
				return true
			}
		}
	}
	for _, name := range []string{"талгар", "talgar", "қаскелең", "каскелен", "kaskelen", "талдықорған", "талдыкорган", "taldykorgan", "жезқазған", "жезказган", "zhezkazgan", "темиртау", "temirtau", "рудный", "rudny", "байконур", "baikonur", "екібастұз", "экибастуз", "ekibastuz", "көкшетау", "кокшетау", "kokshetau"} {
		if strings.Contains(" "+normalized+" ", " "+normalizePlace(name)+" ") {
			return true
		}
	}
	return false
}

func hasForeignCountry(normalized string) bool {
	for _, country := range []string{"germany", "россия", "russia", "united states", "usa", "poland", "польша", "uzbekistan", "өзбекстан", "узбекистан", "kyrgyzstan", "кыргызстан", "turkey", "түркия", "турция"} {
		if strings.Contains(" "+normalized+" ", " "+country+" ") {
			return true
		}
	}
	return false
}

func normalizePlace(value string) string {
	value = strings.ToLower(value)
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, value)
	return strings.Join(strings.Fields(value), " ")
}
