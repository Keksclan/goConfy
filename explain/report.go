package explain

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Source indicates where a configuration value originated from.
type Source string

const (
	SourceBase    Source = "base"
	SourceProfile Source = "profile"
	SourceDotenv  Source = "dotenv"
	SourceEnv     Source = "env"
	SourceDefault Source = "default"
)

// Entry represents a single configuration path and its audit trail.
type Entry struct {
	Path          string   `json:"path"`
	Source        Source   `json:"source"`
	OverriddenBy  []Source `json:"overridden_by,omitempty"`
	IsSecret      bool     `json:"is_secret"`
	ValueRedacted string   `json:"value_redacted"`
	Notes         string   `json:"notes,omitempty"`
}

// Report contains all entries collected during the load process.
type Report struct {
	Entries []Entry `json:"entries"`
}

// Reporter is a function that receives a completed report.
type Reporter func(Report)

// NewReport creates an empty report.
func NewReport() *Report {
	return &Report{
		Entries: make([]Entry, 0),
	}
}

// AddEntry adds or updates an entry in the report.
func (r *Report) AddEntry(path string, source Source, value string, isSecret bool, notes string) {
	if r == nil {
		return
	}

	redacted := value
	if isSecret {
		redacted = "[REDACTED]"
	}

	for i := range r.Entries {
		if r.Entries[i].Path == path {
			// Update existing entry
			if !slices.Contains(r.Entries[i].OverriddenBy, source) {
				r.Entries[i].OverriddenBy = append(r.Entries[i].OverriddenBy, source)
			}
			r.Entries[i].ValueRedacted = redacted
			// We keep the original source as the first one, but track overrides
			if notes != "" {
				if r.Entries[i].Notes != "" {
					r.Entries[i].Notes += "; " + notes
				} else {
					r.Entries[i].Notes = notes
				}
			}
			return
		}
	}

	r.Entries = append(r.Entries, Entry{
		Path:          path,
		Source:        source,
		IsSecret:      isSecret,
		ValueRedacted: redacted,
		Notes:         notes,
	})
}

// WriteText writes the report in a human-readable table format.
func (r Report) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%-40s %-10s %-20s %-20s %s\n", "PATH", "SOURCE", "VALUE (REDACTED)", "OVERRIDES", "NOTES")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 120))
	for _, e := range r.Entries {
		overrides := ""
		if len(e.OverriddenBy) > 0 {
			sources := make([]string, len(e.OverriddenBy))
			for i, s := range e.OverriddenBy {
				sources[i] = string(s)
			}
			overrides = strings.Join(sources, ", ")
		}
		fmt.Fprintf(w, "%-40s %-10s %-20s %-20s %s\n", e.Path, e.Source, e.ValueRedacted, overrides, e.Notes)
	}
	return nil
}

// WriteJSON writes the report in JSON format.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
