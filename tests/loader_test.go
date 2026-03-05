package tests

import (
	"errors"
	"fmt"
	"testing"

	goconfy "github.com/keksclan/goConfy"
)

type SimpleConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func TestLoadBasic(t *testing.T) {
	yaml := []byte(`
host: localhost
port: 9090
`)
	cfg, err := goconfy.Load[SimpleConfig](goconfy.WithBytes(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected host=localhost, got %q", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port=9090, got %d", cfg.Port)
	}
}

func TestLoadWithEnvMacro(t *testing.T) {
	yaml := []byte(`
host: "{ENV:TEST_HOST:fallback}"
port: 3000
`)
	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yaml),
		goconfy.WithEnvLookup(func(key string) (string, bool) {
			if key == "TEST_HOST" {
				return "envhost", true
			}
			return "", false
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "envhost" {
		t.Errorf("expected host=envhost, got %q", cfg.Host)
	}
}

func TestLoadWithDefault(t *testing.T) {
	yaml := []byte(`
host: "{ENV:MISSING_HOST:defaulthost}"
port: 3000
`)
	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yaml),
		goconfy.WithEnvLookup(func(string) (string, bool) { return "", false }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "defaulthost" {
		t.Errorf("expected host=defaulthost, got %q", cfg.Host)
	}
}

func TestLoadNoInput(t *testing.T) {
	_, err := goconfy.Load[SimpleConfig]()
	if err == nil {
		t.Fatal("expected error for no input")
	}
}

func TestMustLoadPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from MustLoad")
		}
	}()
	goconfy.MustLoad[SimpleConfig]()
}

type ValidatableConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c *ValidatableConfig) Validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	return nil
}

func TestLoadValidation(t *testing.T) {
	yaml := []byte(`
host: localhost
port: -1
`)
	_, err := goconfy.Load[ValidatableConfig](goconfy.WithBytes(yaml))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

type NormalizableConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c *NormalizableConfig) Normalize() {
	if c.Port == 0 {
		c.Port = 8080
	}
}

func TestLoadNormalize(t *testing.T) {
	yaml := []byte(`
host: localhost
port: 0
`)
	cfg, err := goconfy.Load[NormalizableConfig](goconfy.WithBytes(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port=8080 after normalize, got %d", cfg.Port)
	}
}

func TestFieldError(t *testing.T) {
	tests := []struct {
		name     string
		err      *goconfy.FieldError
		expected string
	}{
		{
			name:     "field only",
			err:      &goconfy.FieldError{Field: "port", Message: "must be positive"},
			expected: `field "port": must be positive`,
		},
		{
			name:     "line only",
			err:      &goconfy.FieldError{Line: 3, Message: "some error"},
			expected: "line 3: some error",
		},
		{
			name:     "line and column",
			err:      &goconfy.FieldError{Line: 3, Column: 5, Message: "some error"},
			expected: "line 3, col 5: some error",
		},
		{
			name:     "path and line",
			err:      &goconfy.FieldError{Path: "db.port", Line: 10, Message: "invalid"},
			expected: `path "db.port": line 10: invalid`,
		},
		{
			name:     "layer, path, line, col",
			err:      &goconfy.FieldError{Layer: "base", Path: "db.port", Line: 10, Column: 2, Message: "invalid"},
			expected: `[base] path "db.port": line 10, col 2: invalid`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestMultiError(t *testing.T) {
	me := &goconfy.MultiError{
		Errors: []error{
			fmt.Errorf("err1"),
			fmt.Errorf("err2"),
		},
	}
	if me.Error() != "err1; err2" {
		t.Errorf("unexpected error string: %q", me.Error())
	}
	unwrapped := me.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("expected 2 unwrapped errors, got %d", len(unwrapped))
	}

	// Test errors.Is works through MultiError
	target := fmt.Errorf("err1")
	_ = target
	if !errors.Is(me, me.Errors[0]) {
		t.Error("errors.Is failed to find wrapped error in MultiError")
	}
}
