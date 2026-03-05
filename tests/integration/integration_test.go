package integration

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/explain"
)

type FullConfig struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	DB struct {
		Password string        `yaml:"password" secret:"true"`
		URL      string        `yaml:"url"`
		Timeout  time.Duration `yaml:"timeout"`
	} `yaml:"db"`
	SecretKey  string `yaml:"secret_key" secret:"true"`
	Debug      bool   `yaml:"debug"`
	Normalized bool   `yaml:"-"`
}

func (c *FullConfig) Normalize() {
	c.Normalized = true
	if c.Server.Host == "" {
		c.Server.Host = "default-host"
	}
}

func (c *FullConfig) Validate() error {
	var errs []error
	if c.Server.Port <= 0 {
		errs = append(errs, errors.New("port must be positive"))
	}
	if c.DB.Password == "" {
		errs = append(errs, errors.New("db password is required"))
	}
	if len(errs) > 0 {
		return &goconfy.MultiError{Errors: errs}
	}
	return nil
}

func TestFullPipeline(t *testing.T) {
	// Setup env for test
	t.Setenv("SECRET_KEY", "env-secret")

	// 1. Load with profiles, dotenv, and macros
	cfg, err := goconfy.Load[FullConfig](
		goconfy.WithFile("testdata/base.yml"),
		goconfy.WithDotEnvFile("testdata/.env"),
		goconfy.WithProfile("dev"),
		goconfy.WithEnableProfiles(true),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// 2. Verify Merge and Profiles (dev.yml overrides port to 9090)
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090 from dev profile, got %d", cfg.Server.Port)
	}

	// 3. Verify ENV Macros from .env (DB_PASSWORD from .env)
	if cfg.DB.Password != "secret-pass" {
		t.Errorf("expected db password 'secret-pass' from .env, got %q", cfg.DB.Password)
	}

	// 4. Verify OS ENV Precedence (SECRET_KEY from OS env should win if implemented that way,
	// but usually .env is loaded first. Let's see how goConfy handles it.)
	// Actually, the security model says "OS env precedence".
	// Let's check if SECRET_KEY is "env-secret" (OS) or "real-secret" (.env).
	if cfg.SecretKey != "env-secret" {
		t.Errorf("expected secret_key 'env-secret' (OS env precedence), got %q", cfg.SecretKey)
	}

	// 5. Verify FILE Macros
	expectedURL := "postgres://localhost:5432/mydb"
	if strings.TrimSpace(cfg.DB.URL) != expectedURL {
		t.Errorf("expected db url %q, got %q", expectedURL, cfg.DB.URL)
	}

	// 6. Verify Normalize and Validate
	if !cfg.Normalized {
		t.Error("Normalize() was not called")
	}

	// 7. Verify Redaction
	redacted, err := goconfy.DumpRedactedJSON(cfg)
	if err != nil {
		t.Fatalf("failed to dump redacted: %v", err)
	}
	redactedStr := string(redacted)
	if strings.Contains(redactedStr, "secret-pass") {
		t.Error("DB password leaked in redacted dump")
	}
	if strings.Contains(redactedStr, "env-secret") {
		t.Error("Secret key leaked in redacted dump")
	}
	if !strings.Contains(redactedStr, "[REDACTED]") {
		t.Error("Redacted dump does not contain [REDACTED] placeholder")
	}

	// 8. Verify Duration
	if cfg.DB.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.DB.Timeout)
	}
}

func TestNegativeScenarios(t *testing.T) {
	t.Run("MissingRequiredEnvMacro", func(t *testing.T) {
		// Clear env
		os.Unsetenv("DB_PASSWORD")

		yaml := `
server:
  port: 8080
db:
  password: "{ENV:DB_PASSWORD}"
`
		_, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err == nil {
			t.Fatal("expected error for missing required env macro")
		}
		if !strings.Contains(err.Error(), "DB_PASSWORD") {
			t.Errorf("expected error message to mention DB_PASSWORD, got %v", err)
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		yaml := `
server:
  port: invalid
`
		_, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err == nil {
			t.Fatal("expected error for invalid yaml type")
		}
	})

	t.Run("UnknownKeyStrict", func(t *testing.T) {
		yaml := `
unknown_key: value
`
		_, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err == nil {
			t.Fatal("expected error for unknown key in strict mode")
		}
	})

	t.Run("InvalidFileMacro", func(t *testing.T) {
		yaml := `
db:
  url: "{FILE:non-existent.txt:}"
`
		_, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err == nil {
			t.Fatal("expected error for non-existent file macro without default")
		}
	})

	t.Run("DotenvNoMutation", func(t *testing.T) {
		// Key only in .env, not in OS env
		key := "ONLY_IN_DOTENV"
		os.Unsetenv(key)

		yaml := `
server:
  port: 8080
  host: "{ENV:ONLY_IN_DOTENV}"
db:
  password: pass
  url: url
`
		// Create a temp .env file
		dotenv := "ONLY_IN_DOTENV=secret-val"
		err := os.WriteFile("testdata/temp.env", []byte(dotenv), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("testdata/temp.env")

		_, err = goconfy.Load[FullConfig](
			goconfy.WithBytes([]byte(yaml)),
			goconfy.WithDotEnvFile("testdata/temp.env"),
		)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if val, ok := os.LookupEnv(key); ok {
			t.Errorf("%s was mutated into OS environment: %s", key, val)
		}
	})

	t.Run("MacroInterpolationRestriction", func(t *testing.T) {
		t.Setenv("FOO", "bar")
		yaml := `
server:
  port: 8080
  host: "prefix-{ENV:FOO}-suffix"
db:
  password: pass
`
		cfg, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Host != "prefix-{ENV:FOO}-suffix" {
			t.Errorf("expected no partial interpolation, got %q", cfg.Server.Host)
		}
	})

	t.Run("MultiErrorAggregation", func(t *testing.T) {
		yaml := `
server:
  port: -1
db:
  password: ""
`
		_, err := goconfy.Load[FullConfig](goconfy.WithBytes([]byte(yaml)))
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "port must be positive") || !strings.Contains(err.Error(), "db password is required") {
			t.Errorf("expected aggregated errors, got: %v", err)
		}
	})

	t.Run("ExplainRedaction", func(t *testing.T) {
		t.Setenv("SECRET_KEY", "very-secret")

		var capturedReport explain.Report
		_, err := goconfy.Load[FullConfig](
			goconfy.WithBytes([]byte(`
server:
  port: 8080
db:
  password: "pass"
secret_key: "{ENV:SECRET_KEY}"
`)),
			goconfy.WithExplainReporter(func(r explain.Report) {
				capturedReport = r
			}),
		)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		// Verify report entries
		for _, entry := range capturedReport.Entries {
			if entry.Path == "db.password" || entry.Path == "secret_key" {
				if !entry.IsSecret {
					t.Errorf("path %q should be marked as secret in report", entry.Path)
				}
				if entry.ValueRedacted != "[REDACTED]" {
					t.Errorf("path %q value should be redacted in report, got %q", entry.Path, entry.ValueRedacted)
				}
			}
		}
	})
}
