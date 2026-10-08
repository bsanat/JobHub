package kzcompanies

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Source struct {
	ID                   string `json:"id"`
	Company              string `json:"company"`
	OfficialWebsite      string `json:"official_website"`
	CareersURL           string `json:"careers_url"`
	SourceType           string `json:"source_type"`
	Adapter              string `json:"adapter"`
	APIEndpoint          string `json:"api_endpoint,omitempty"`
	DetailPrefix         string `json:"detail_prefix,omitempty"`
	Country              string `json:"country"`
	Status               string `json:"status"`
	Enabled              bool   `json:"enabled"`
	PermissionsConfirmed bool   `json:"permissions_confirmed"`
	RobotsReviewed       bool   `json:"robots_reviewed"`
	RobotsNote           string `json:"robots_note"`
	TermsNote            string `json:"terms_note"`
}

type Config struct {
	Enabled bool     `json:"enabled"`
	Sources []Source `json:"sources"`
}

var sourceID = regexp.MustCompile(`^[a-z][a-z0-9_]{1,49}$`)

func Load(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, errors.New("cannot read Kazakhstan source config")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 128*1024+1))
	if err != nil || len(raw) > 128*1024 {
		return cfg, errors.New("invalid Kazakhstan source config size")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&cfg) != nil || dec.Decode(new(any)) != io.EOF {
		return cfg, errors.New("invalid Kazakhstan source config JSON")
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if len(c.Sources) == 0 || len(c.Sources) > 50 {
		return errors.New("invalid Kazakhstan source count")
	}
	seen := make(map[string]bool, len(c.Sources))
	for _, s := range c.Sources {
		if !sourceID.MatchString(s.ID) || seen[s.ID] || strings.TrimSpace(s.Company) == "" || s.Country != "Kazakhstan" ||
			(s.Status != "supported" && s.Status != "needs_manual_review" && s.Status != "blocked") ||
			strings.TrimSpace(s.RobotsNote) == "" || strings.TrimSpace(s.TermsNote) == "" {
			return errors.New("invalid or duplicate Kazakhstan source")
		}
		seen[s.ID] = true
		for _, raw := range []string{s.OfficialWebsite, s.CareersURL, s.APIEndpoint} {
			if raw == "" {
				continue
			}
			u, err := url.Parse(raw)
			if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
				return errors.New("invalid Kazakhstan source URL")
			}
		}
		if s.DetailPrefix != "" && !strings.HasPrefix(s.DetailPrefix, "/") {
			return errors.New("invalid detail prefix")
		}
		switch s.Adapter {
		case "":
		case "airastana", "kcell", "halyk", "technodom", "kolesa", "bcc", "sulpak":
			if (s.Status != "supported" && s.Status != "blocked") || s.CareersURL == "" {
				return errors.New("unsupported Kazakhstan adapter configuration")
			}
		default:
			return errors.New("unknown Kazakhstan source adapter")
		}
		if s.Enabled && (!c.Enabled || s.Status != "supported" || s.Adapter == "" || !s.PermissionsConfirmed || !s.RobotsReviewed) {
			return errors.New("enabled source lacks explicit permission and robots review")
		}
	}
	return nil
}

func (c Config) Find(id string) (Source, bool) {
	for _, s := range c.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}
