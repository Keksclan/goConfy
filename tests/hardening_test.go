package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/explain"
)

type SecretConfig struct {
	DBPassword string `yaml:"db_password"`
	ApiToken   string `yaml:"apiToken"`
	UserKey    string `yaml:"user_key"`
	SafeField  string `yaml:"safe_field"`
	Nested     struct {
		SecretValue string `yaml:"secret_value"`
	} `yaml:"nested"`
}

func TestRedactByConvention(t *testing.T) {
	cfg := SecretConfig{
		DBPassword: "password123",
		ApiToken:   "token456",
		UserKey:    "key789",
		SafeField:  "not-a-secret",
	}
	cfg.Nested.SecretValue = "nested-secret"

	// 1. Without convention (default)
	redacted := goconfy.Redacted(cfg)
	m := redacted.(map[string]any)

	if m["db_password"] != "password123" {
		t.Errorf("expected db_password to be plaintext, got %v", m["db_password"])
	}

	// 2. With convention
	redacted = goconfy.Redacted(cfg, goconfy.WithRedactByConventionOption(true))
	m = redacted.(map[string]any)

	secrets := []string{"db_password", "apiToken", "user_key"}
	for _, s := range secrets {
		if m[s] != "[REDACTED]" {
			t.Errorf("expected %s to be redacted, got %v", s, m[s])
		}
	}
	if m["safe_field"] != "not-a-secret" {
		t.Errorf("expected safe_field to be plaintext, got %v", m["safe_field"])
	}

	nested := m["nested"].(map[string]any)
	if nested["secret_value"] != "[REDACTED]" {
		t.Errorf("expected nested.secret_value to be redacted, got %v", nested["secret_value"])
	}
}

func TestFileMacro(t *testing.T) {
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(secretFile, []byte("  file-secret-value  \n"), 0600)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	yaml := []byte(strings.ReplaceAll(`
host: "{FILE:PATH}"
port: "{FILE:MISSING:3000}"
`, "PATH", secretFile))

	cfg, err := goconfy.Load[SimpleConfig](goconfy.WithBytes(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "file-secret-value" {
		t.Errorf("expected host=file-secret-value (trimmed), got %q", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected port=3000 (default), got %d", cfg.Port)
	}
}

func TestExplainWithConventionAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("file-val"), 0600); err != nil {
		t.Fatal(err)
	}

	yaml := []byte(strings.ReplaceAll(`
db_password: "{FILE:PATH}"
`, "PATH", secretFile))

	var report explain.Report
	_, err := goconfy.Load[SecretConfig](
		goconfy.WithBytes(yaml),
		goconfy.WithRedactByConvention(true),
		goconfy.WithExplainReporter(func(r explain.Report) {
			report = r
		}),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	found := false
	for _, e := range report.Entries {
		if e.Path == "db_password" {
			found = true
			if !e.IsSecret {
				t.Error("expected db_password to be marked as secret by convention")
			}
			if e.ValueRedacted != "[REDACTED]" {
				t.Errorf("expected redacted value, got %q", e.ValueRedacted)
			}
			if e.Source != explain.SourceFile {
				t.Errorf("expected source=file, got %q", e.Source)
			}
		}
	}
	if !found {
		t.Fatal("db_password entry not found in report")
	}
}
