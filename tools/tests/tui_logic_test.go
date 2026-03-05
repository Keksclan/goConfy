package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keksclan/goConfy/tools/internal/tools/tui/logic"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/state"
)

// TestRedactYAML_DefaultPaths verifies that built-in secret paths are redacted.
func TestRedactYAML_DefaultPaths(t *testing.T) {
	input := `
redis:
  host: localhost
  password: mysecretpass
postgres:
  url: "postgres://user:pass@host/db"
auth:
  opaque:
    client_id: myid
    client_secret: topsecret
`
	out, err := logic.RedactYAML(input, nil)
	if err != nil {
		t.Fatalf("RedactYAML failed: %v", err)
	}

	if strings.Contains(out, "mysecretpass") {
		t.Error("redis.password was not redacted")
	}
	if strings.Contains(out, "topsecret") {
		t.Error("auth.opaque.client_secret was not redacted")
	}
	if strings.Contains(out, "postgres://user:pass@host/db") {
		t.Error("postgres.url was not redacted")
	}
	if !strings.Contains(out, "******") {
		t.Error("expected redaction placeholder '******' in output")
	}
	// Non-secret fields should remain.
	if !strings.Contains(out, "localhost") {
		t.Error("redis.host should not be redacted")
	}
	if !strings.Contains(out, "myid") {
		t.Error("auth.opaque.client_id should not be redacted")
	}
}

// TestRedactYAML_ExtraPaths verifies custom dot-paths are redacted.
func TestRedactYAML_ExtraPaths(t *testing.T) {
	input := `
app:
  api_key: "abc123"
  name: "myapp"
`
	out, err := logic.RedactYAML(input, []string{"app.api_key"})
	if err != nil {
		t.Fatalf("RedactYAML failed: %v", err)
	}
	if strings.Contains(out, "abc123") {
		t.Error("app.api_key was not redacted")
	}
	if !strings.Contains(out, "myapp") {
		t.Error("app.name should not be redacted")
	}
}

// TestRedactYAML_EmptyInput handles empty YAML gracefully.
func TestRedactYAML_EmptyInput(t *testing.T) {
	out, err := logic.RedactYAML("{}", nil)
	if err != nil {
		t.Fatalf("RedactYAML on empty input failed: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output for empty YAML object")
	}
}

// TestContainsSecretPaths checks heuristic detection.
func TestContainsSecretPaths(t *testing.T) {
	yamlWithSecret := `
redis:
  password: secret
`
	if !logic.ContainsSecretPaths(yamlWithSecret, nil) {
		t.Error("expected ContainsSecretPaths to return true for redis.password")
	}

	yamlWithout := `
app:
  name: hello
`
	if logic.ContainsSecretPaths(yamlWithout, nil) {
		t.Error("expected ContainsSecretPaths to return false for non-secret YAML")
	}
}

// TestLoadRawYAML reads a file and returns its contents.
func TestLoadRawYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")
	content := "host: localhost\nport: 8080\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := logic.LoadRawYAML(path)
	if err != nil {
		t.Fatalf("LoadRawYAML failed: %v", err)
	}
	if got != content {
		t.Errorf("expected %q, got %q", content, got)
	}
}

// TestLoadRawYAML_NotFound returns error for missing file.
func TestLoadRawYAML_NotFound(t *testing.T) {
	_, err := logic.LoadRawYAML("/nonexistent/path.yml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// TestExpandedYAML verifies macro expansion works via the logic wrapper.
func TestExpandedYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(cfgPath, []byte("host: \"{ENV:TEST_HOST:fallback}\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := state.DefaultConfig()
	cfg.ConfigPath = cfgPath

	out, err := logic.ExpandedYAML(cfg)
	if err != nil {
		t.Fatalf("ExpandedYAML failed: %v", err)
	}
	// Should use the default value since TEST_HOST is not set.
	if !strings.Contains(out, "fallback") {
		t.Errorf("expected 'fallback' in expanded output, got: %s", out)
	}
}

// TestMergedYAML verifies profile merge works.
func TestMergedYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	yaml := `
host: localhost
port: 8080
profiles:
  prod:
    port: 443
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := state.DefaultConfig()
	cfg.ConfigPath = cfgPath
	cfg.ActiveProfile = "prod"

	out, err := logic.MergedYAML(cfg)
	if err != nil {
		t.Fatalf("MergedYAML failed: %v", err)
	}
	if !strings.Contains(out, "443") {
		t.Errorf("expected merged port 443 in output, got: %s", out)
	}
	if strings.Contains(out, "profiles") {
		t.Errorf("profiles section should be removed after merge, got: %s", out)
	}
}

// TestFormatYAML verifies formatting round-trip.
func TestFormatYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(cfgPath, []byte("host:   localhost\nport:  8080\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := state.DefaultConfig()
	cfg.ConfigPath = cfgPath

	out, err := logic.FormatYAML(cfg)
	if err != nil {
		t.Fatalf("FormatYAML failed: %v", err)
	}
	if !strings.Contains(string(out), "host:") {
		t.Errorf("expected formatted YAML with 'host:', got: %s", string(out))
	}
}

// TestListProfiles parses profile names from YAML.
func TestListProfiles(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	yaml := `
host: localhost
profiles:
  dev:
    port: 9090
  staging:
    port: 8081
  prod:
    port: 443
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	names, hasProfiles, err := logic.ListProfiles(cfgPath)
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if !hasProfiles {
		t.Error("expected hasProfiles to be true")
	}
	if len(names) != 3 {
		t.Errorf("expected 3 profiles, got %d: %v", len(names), names)
	}
}

// TestListProfiles_NoProfiles verifies behavior when no profiles section exists.
func TestListProfiles_NoProfiles(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(cfgPath, []byte("host: localhost\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	names, hasProfiles, err := logic.ListProfiles(cfgPath)
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if hasProfiles {
		t.Error("expected hasProfiles to be false")
	}
	if len(names) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(names))
	}
}

// TestWriteFormatted writes formatted output and verifies.
func TestWriteFormatted(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	cfg := state.DefaultConfig()
	cfg.OutputPath = outPath

	data := []byte("host: localhost\nport: 8080\n")
	written, err := logic.WriteFormatted(cfg, data)
	if err != nil {
		t.Fatalf("WriteFormatted failed: %v", err)
	}
	if written != outPath {
		t.Errorf("expected written path %q, got %q", outPath, written)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(data) {
		t.Errorf("file content mismatch: expected %q, got %q", string(data), string(content))
	}
}
