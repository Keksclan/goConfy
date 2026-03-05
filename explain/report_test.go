package explain

import (
	"testing"
)

func TestAddEntry_Override(t *testing.T) {
	r := NewReport()

	// Initial entry from base
	r.AddEntry("app.name", SourceBase, "initial", false, "from base")

	// Override with env
	r.AddEntry("app.name", SourceEnv, "env-name", false, "from env")

	// Override with profile
	r.AddEntry("app.name", SourceProfile, "profile-name", false, "from profile")

	var entry *Entry
	for i := range r.Entries {
		if r.Entries[i].Path == "app.name" {
			entry = &r.Entries[i]
			break
		}
	}

	if entry == nil {
		t.Fatal("entry not found")
	}

	if entry.ValueRedacted != "profile-name" {
		t.Errorf("expected ValueRedacted to be 'profile-name', got '%s'", entry.ValueRedacted)
	}

	// This is where the bug is expected
	if entry.Source != SourceProfile {
		t.Errorf("expected Source to be '%s', got '%s'", SourceProfile, entry.Source)
	}

	expectedOverrides := []Source{SourceBase, SourceEnv}
	if len(entry.OverriddenBy) != len(expectedOverrides) {
		t.Errorf("expected %d overrides, got %d: %v", len(expectedOverrides), len(entry.OverriddenBy), entry.OverriddenBy)
	}

	for i, s := range expectedOverrides {
		if entry.OverriddenBy[i] != s {
			t.Errorf("expected override %d to be '%s', got '%s'", i, s, entry.OverriddenBy[i])
		}
	}
}
