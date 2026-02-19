package tests

import (
	"testing"

	goconfy "github.com/keksclan/goConfy"
)

type AppConfig struct {
	App struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"app"`
}

func TestProfileMerge(t *testing.T) {
	input := []byte(`
app:
  port: 8080
  host: localhost
profiles:
  prod:
    app:
      port: 80
`)
	cfg, err := goconfy.Load[AppConfig](
		goconfy.WithBytes(input),
		goconfy.WithEnableProfiles(true),
		goconfy.WithProfile("prod"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.Port != 80 {
		t.Errorf("expected port=80 after profile merge, got %d", cfg.App.Port)
	}
	if cfg.App.Host != "localhost" {
		t.Errorf("expected host=localhost (unchanged), got %q", cfg.App.Host)
	}
}

func TestProfileMergeNoProfile(t *testing.T) {
	input := []byte(`
app:
  port: 8080
  host: localhost
profiles:
  prod:
    app:
      port: 80
`)
	cfg, err := goconfy.Load[AppConfig](
		goconfy.WithBytes(input),
		goconfy.WithEnableProfiles(true),
		goconfy.WithProfile(""),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected port=8080 (no profile applied), got %d", cfg.App.Port)
	}
}

func TestProfileNotEnabled(t *testing.T) {
	// When profiles are NOT enabled, the "profiles" key remains and strict decode
	// should fail because it's an unknown field.
	input := []byte(`
app:
  port: 8080
profiles:
  prod:
    app:
      port: 80
`)
	_, err := goconfy.Load[AppConfig](
		goconfy.WithBytes(input),
		goconfy.WithStrictYAML(true),
	)
	if err == nil {
		t.Fatal("expected error: profiles key should cause strict decode failure when profiles not enabled")
	}
}
