package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/internal/dotenv"
)

// --- Dotenv Parsing Tests ---

func TestDotenvParseCommentsAndBlankLines(t *testing.T) {
	input := `
# This is a comment
KEY1=value1

# Another comment
KEY2=value2

`
	vars, err := dotenv.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got %q", vars["KEY1"])
	}
	if vars["KEY2"] != "value2" {
		t.Errorf("expected KEY2=value2, got %q", vars["KEY2"])
	}
	if len(vars) != 2 {
		t.Errorf("expected 2 keys, got %d", len(vars))
	}
}

func TestDotenvParseExportPrefix(t *testing.T) {
	input := `export MY_KEY=exported_value
export ANOTHER=hello`
	vars, err := dotenv.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["MY_KEY"] != "exported_value" {
		t.Errorf("expected MY_KEY=exported_value, got %q", vars["MY_KEY"])
	}
	if vars["ANOTHER"] != "hello" {
		t.Errorf("expected ANOTHER=hello, got %q", vars["ANOTHER"])
	}
}

func TestDotenvParseSingleQuoted(t *testing.T) {
	input := `KEY='single quoted value'
ESCAPE='no \n escape here'`
	vars, err := dotenv.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "single quoted value" {
		t.Errorf("expected literal single-quoted value, got %q", vars["KEY"])
	}
	// Single quotes: no escape processing.
	if vars["ESCAPE"] != `no \n escape here` {
		t.Errorf("expected literal backslash-n in single quotes, got %q", vars["ESCAPE"])
	}
}

func TestDotenvParseDoubleQuoted(t *testing.T) {
	input := `KEY="double quoted value"
NEWLINE="line1\nline2"
TAB="col1\tcol2"
ESCAPED_QUOTE="say \"hello\""
ESCAPED_BACKSLASH="path\\to\\file"`
	vars, err := dotenv.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "double quoted value" {
		t.Errorf("expected double-quoted value, got %q", vars["KEY"])
	}
	if vars["NEWLINE"] != "line1\nline2" {
		t.Errorf("expected newline escape, got %q", vars["NEWLINE"])
	}
	if vars["TAB"] != "col1\tcol2" {
		t.Errorf("expected tab escape, got %q", vars["TAB"])
	}
	if vars["ESCAPED_QUOTE"] != `say "hello"` {
		t.Errorf("expected escaped quotes, got %q", vars["ESCAPED_QUOTE"])
	}
	if vars["ESCAPED_BACKSLASH"] != `path\to\file` {
		t.Errorf("expected escaped backslashes, got %q", vars["ESCAPED_BACKSLASH"])
	}
}

// --- Precedence Tests ---

func TestDotenvOSPrecedenceDefault(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_HOST=from_dotenv\n"), 0644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`host: "{ENV:APP_HOST:fallback}"
port: 8080`)

	// Base lookup simulates OS env returning a value.
	osLookup := func(key string) (string, bool) {
		if key == "APP_HOST" {
			return "from_os", true
		}
		return "", false
	}

	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(osLookup),
		goconfy.WithDotEnvFile(envFile),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default: OS wins.
	if cfg.Host != "from_os" {
		t.Errorf("expected host=from_os (OS precedence), got %q", cfg.Host)
	}
}

func TestDotenvPrecedenceReversed(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_HOST=from_dotenv\n"), 0644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`host: "{ENV:APP_HOST:fallback}"
port: 8080`)

	osLookup := func(key string) (string, bool) {
		if key == "APP_HOST" {
			return "from_os", true
		}
		return "", false
	}

	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(osLookup),
		goconfy.WithDotEnvFile(envFile),
		goconfy.WithDotEnvOSPrecedence(false),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Dotenv wins when OS precedence is false.
	if cfg.Host != "from_dotenv" {
		t.Errorf("expected host=from_dotenv (dotenv precedence), got %q", cfg.Host)
	}
}

func TestDotenvFallbackToDotenv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_HOST=from_dotenv\n"), 0644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`host: "{ENV:APP_HOST:fallback}"
port: 8080`)

	// OS lookup returns nothing.
	osLookup := func(key string) (string, bool) {
		return "", false
	}

	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(osLookup),
		goconfy.WithDotEnvFile(envFile),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "from_dotenv" {
		t.Errorf("expected host=from_dotenv (fallback), got %q", cfg.Host)
	}
}

// --- Optional Behavior Tests ---

func TestDotenvMissingFileErrorsByDefault(t *testing.T) {
	yamlData := []byte(`host: localhost
port: 8080`)

	_, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithDotEnvFile("/nonexistent/path/.env"),
	)
	if err == nil {
		t.Fatal("expected error for missing dotenv file")
	}
}

func TestDotenvMissingFileOptional(t *testing.T) {
	yamlData := []byte(`host: "{ENV:SOME_KEY:default_val}"
port: 8080`)

	cfg, err := goconfy.Load[SimpleConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithDotEnvFile("/nonexistent/path/.env"),
		goconfy.WithDotEnvOptional(true),
		goconfy.WithEnvLookup(func(string) (string, bool) { return "", false }),
	)
	if err != nil {
		t.Fatalf("unexpected error with optional dotenv: %v", err)
	}
	if cfg.Host != "default_val" {
		t.Errorf("expected host=default_val, got %q", cfg.Host)
	}
}

// --- Integration Test ---

func TestDotenvIntegrationYAMLMacro(t *testing.T) {
	dir := t.TempDir()

	// Write .env file.
	envFile := filepath.Join(dir, ".env")
	envContent := "APP_NAME=my-awesome-service\nAPP_PORT=9090\n"
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`name: "{ENV:APP_NAME:todo-service}"
port: 8080`)

	type AppConfig struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	}

	// Use a lookup that returns nothing to ensure dotenv is the source.
	cfg, err := goconfy.Load[AppConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithDotEnvFile(envFile),
		goconfy.WithEnvLookup(func(string) (string, bool) { return "", false }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "my-awesome-service" {
		t.Errorf("expected name=my-awesome-service, got %q", cfg.Name)
	}
}
