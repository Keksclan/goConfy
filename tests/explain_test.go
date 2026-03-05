package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/explain"
)

type ExplainConfig struct {
	App struct {
		Name  string `yaml:"name"`
		Port  int    `yaml:"port"`
		Token string `yaml:"token" secret:"true"`
	} `yaml:"app"`
	DB struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
	} `yaml:"db"`
}

func TestExplainReporting(t *testing.T) {
	yamlData := `
app:
  name: "MyApp"
  port: 8080
  token: "{ENV:APP_TOKEN:default-token}"
db:
  host: "localhost"
profiles:
  dev:
    db:
      host: "dev-host"
`
	os.Setenv("APP_TOKEN", "super-secret")
	defer os.Unsetenv("APP_TOKEN")

	var capturedReport explain.Report
	reporter := func(r explain.Report) {
		capturedReport = r
	}

	_, err := goconfy.Load[ExplainConfig](
		goconfy.WithBytes([]byte(yamlData)),
		goconfy.WithExplainReporter(reporter),
		goconfy.WithEnableProfiles(true),
		goconfy.WithProfile("dev"),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify report entries
	entries := make(map[string]explain.Entry)
	for _, e := range capturedReport.Entries {
		entries[e.Path] = e
	}

	// 1. Base value
	if e, ok := entries["app.name"]; !ok || e.Source != explain.SourceBase || e.ValueRedacted != "MyApp" {
		t.Errorf("app.name report incorrect: %+v", e)
	}

	// 2. Secret redaction + Env expansion
	if e, ok := entries["app.token"]; !ok {
		t.Error("app.token missing from report")
	} else {
		if !e.IsSecret {
			t.Error("app.token should be marked as secret")
		}
		if e.ValueRedacted != "[REDACTED]" {
			t.Errorf("app.token value should be redacted, got: %s", e.ValueRedacted)
		}
		if !strings.Contains(e.Notes, "expanded {ENV:APP_TOKEN}") {
			t.Errorf("app.token notes should mention env expansion, got: %s", e.Notes)
		}
	}

	// 3. Profile override
	if e, ok := entries["db.host"]; !ok {
		t.Error("db.host missing from report")
	} else {
		if e.ValueRedacted != "dev-host" {
			t.Errorf("db.host value should be overridden by profile, got: %s", e.ValueRedacted)
		}
		if e.Source != explain.SourceProfile {
			t.Errorf("db.host source should be profile, got: %s", e.Source)
		}
		foundBase := false
		for _, s := range e.OverriddenBy {
			if s == explain.SourceBase {
				foundBase = true
				break
			}
		}
		if !foundBase {
			t.Errorf("db.host should show base override, got: %v", e.OverriddenBy)
		}
	}
}

func TestExplainDisabled_NoAllocations(t *testing.T) {
	yamlData := `app: { name: "test" }`

	// Just verify it works without reporter
	_, err := goconfy.Load[ExplainConfig](
		goconfy.WithBytes([]byte(yamlData)),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
}

func TestExplainSecretInSlice(t *testing.T) {
	type Config struct {
		Tokens []string `yaml:"tokens" secret:"true"`
	}
	yamlData := `
tokens:
  - "secret-token-1"
  - "secret-token-2"
`
	var capturedReport explain.Report
	reporter := func(r explain.Report) {
		capturedReport = r
	}

	_, err := goconfy.Load[Config](
		goconfy.WithBytes([]byte(yamlData)),
		goconfy.WithExplainReporter(reporter),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	foundTokens := 0
	for _, e := range capturedReport.Entries {
		if strings.HasPrefix(e.Path, "tokens[") {
			foundTokens++
			if !e.IsSecret {
				t.Errorf("Path %s should be marked as secret", e.Path)
			}
			if e.ValueRedacted != "[REDACTED]" {
				t.Errorf("Path %s value should be redacted, got: %s", e.Path, e.ValueRedacted)
			}
		}
	}
	if foundTokens != 2 {
		t.Errorf("Expected 2 token entries in report, found %d", foundTokens)
	}
}

func TestExplainSecretInNestedSlice(t *testing.T) {
	type Server struct {
		Password string `yaml:"password" secret:"true"`
	}
	type Config struct {
		Servers []Server `yaml:"servers"`
	}
	yamlData := `
servers:
  - password: "pass1"
  - password: "pass2"
`
	var capturedReport explain.Report
	reporter := func(r explain.Report) {
		capturedReport = r
	}

	_, err := goconfy.Load[Config](
		goconfy.WithBytes([]byte(yamlData)),
		goconfy.WithExplainReporter(reporter),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	foundPasswords := 0
	for _, e := range capturedReport.Entries {
		if strings.HasPrefix(e.Path, "servers[") && strings.HasSuffix(e.Path, ".password") {
			foundPasswords++
			if !e.IsSecret {
				t.Errorf("Path %s should be marked as secret", e.Path)
			}
			if e.ValueRedacted != "[REDACTED]" {
				t.Errorf("Path %s value should be redacted, got: %s", e.Path, e.ValueRedacted)
			}
		}
	}
	if foundPasswords != 2 {
		t.Errorf("Expected 2 password entries in report, found %d", foundPasswords)
	}
}
