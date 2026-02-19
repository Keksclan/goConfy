package tests

import (
	"testing"
	"time"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/types"
)

type DurationConfig struct {
	Timeout types.Duration `yaml:"timeout"`
}

func TestDurationParsing(t *testing.T) {
	input := []byte(`timeout: 5m`)
	cfg, err := goconfy.Load[DurationConfig](goconfy.WithBytes(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Duration(cfg.Timeout) != 5*time.Minute {
		t.Errorf("expected 5m, got %s", cfg.Timeout)
	}
}

func TestDurationParsingSeconds(t *testing.T) {
	input := []byte(`timeout: 30s`)
	cfg, err := goconfy.Load[DurationConfig](goconfy.WithBytes(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Duration(cfg.Timeout) != 30*time.Second {
		t.Errorf("expected 30s, got %s", cfg.Timeout)
	}
}

func TestDurationString(t *testing.T) {
	d := types.Duration(5 * time.Minute)
	if d.String() != "5m0s" {
		t.Errorf("expected 5m0s, got %q", d.String())
	}
}

func TestDurationInvalidFails(t *testing.T) {
	input := []byte(`timeout: notaduration`)
	_, err := goconfy.Load[DurationConfig](goconfy.WithBytes(input))
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestRedactionHidesSecrets(t *testing.T) {
	type SecretConfig struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password" secret:"true"`
	}

	cfg := SecretConfig{Host: "localhost", Password: "s3cret"}
	redacted := goconfy.Redacted(cfg)

	m, ok := redacted.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", redacted)
	}
	if m["password"] != "[REDACTED]" {
		t.Errorf("expected password=[REDACTED], got %v", m["password"])
	}
	if m["host"] != "localhost" {
		t.Errorf("expected host=localhost, got %v", m["host"])
	}
}

func TestRedactionByPath(t *testing.T) {
	type DBConfig struct {
		DSN string `yaml:"dsn"`
	}
	type Config struct {
		DB DBConfig `yaml:"db"`
	}

	cfg := Config{DB: DBConfig{DSN: "postgres://secret"}}
	redacted := goconfy.Redacted(cfg, goconfy.WithRedactPaths([]string{"db.dsn"}))

	m, ok := redacted.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", redacted)
	}
	dbMap, ok := m["db"].(map[string]any)
	if !ok {
		t.Fatalf("expected db to be map[string]any, got %T", m["db"])
	}
	if dbMap["dsn"] != "[REDACTED]" {
		t.Errorf("expected dsn=[REDACTED], got %v", dbMap["dsn"])
	}
}

func TestDumpRedactedJSON(t *testing.T) {
	type Config struct {
		Host     string `yaml:"host" json:"host"`
		Password string `yaml:"password" json:"password" secret:"true"`
	}

	cfg := Config{Host: "localhost", Password: "secret"}
	json, err := goconfy.DumpRedactedJSON(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if json == "" {
		t.Fatal("expected non-empty JSON")
	}
	if !contains(json, "[REDACTED]") {
		t.Errorf("expected JSON to contain [REDACTED], got %s", json)
	}
	if contains(json, "secret") && !contains(json, "[REDACTED]") {
		t.Error("password should be redacted in JSON output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
