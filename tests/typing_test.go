package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/types"
)

type TypedConfig struct {
	App struct {
		Name           string `yaml:"name"`
		VersionDisplay bool   `yaml:"version_display"`
	} `yaml:"app"`
	Server struct {
		GRPC struct {
			Port    int            `yaml:"port"`
			Timeout types.Duration `yaml:"timeout"`
		} `yaml:"grpc"`
	} `yaml:"server"`
}

func TestTypedDecodingFromMacroDefaults(t *testing.T) {
	yamlData := []byte(`
app:
  name: "{ENV:APP_NAME:todo-service}"
  version_display: "{ENV:APP_VERSION_DISPLAY:true}"
server:
  grpc:
    port: "{ENV:GRPC_PORT:50051}"
    timeout: "{ENV:GRPC_TIMEOUT:3s}"
`)
	noLookup := func(string) (string, bool) { return "", false }

	cfg, err := goconfy.Load[TypedConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(noLookup),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.Name != "todo-service" {
		t.Errorf("expected name=%q, got %q", "todo-service", cfg.App.Name)
	}
	if cfg.App.VersionDisplay != true {
		t.Errorf("expected version_display=true, got %v", cfg.App.VersionDisplay)
	}
	if cfg.Server.GRPC.Port != 50051 {
		t.Errorf("expected port=50051, got %d", cfg.Server.GRPC.Port)
	}
	if time.Duration(cfg.Server.GRPC.Timeout) != 3*time.Second {
		t.Errorf("expected timeout=3s, got %s", cfg.Server.GRPC.Timeout)
	}
}

func TestTypedDecodingFromDotenv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	envContent := "APP_NAME=\"todo-service-dev\"\nGRPC_PORT=60051\nGRPC_TIMEOUT=\"5s\"\nAPP_VERSION_DISPLAY=false\n"
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	yamlData := []byte(`
app:
  name: "{ENV:APP_NAME:todo-service}"
  version_display: "{ENV:APP_VERSION_DISPLAY:true}"
server:
  grpc:
    port: "{ENV:GRPC_PORT:50051}"
    timeout: "{ENV:GRPC_TIMEOUT:3s}"
`)
	noLookup := func(string) (string, bool) { return "", false }

	cfg, err := goconfy.Load[TypedConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithDotEnvFile(envFile),
		goconfy.WithEnvLookup(noLookup),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.Name != "todo-service-dev" {
		t.Errorf("expected name=%q, got %q", "todo-service-dev", cfg.App.Name)
	}
	if cfg.App.VersionDisplay != false {
		t.Errorf("expected version_display=false, got %v", cfg.App.VersionDisplay)
	}
	if cfg.Server.GRPC.Port != 60051 {
		t.Errorf("expected port=60051, got %d", cfg.Server.GRPC.Port)
	}
	if time.Duration(cfg.Server.GRPC.Timeout) != 5*time.Second {
		t.Errorf("expected timeout=5s, got %s", cfg.Server.GRPC.Timeout)
	}
}

func TestTypedDecodingEmptyDefault(t *testing.T) {
	yamlData := []byte(`
app:
  name: "{ENV:MISSING_VAR:}"
  version_display: true
server:
  grpc:
    port: 8080
    timeout: 1s
`)
	noLookup := func(string) (string, bool) { return "", false }

	cfg, err := goconfy.Load[TypedConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(noLookup),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.Name != "" {
		t.Errorf("expected empty name, got %q", cfg.App.Name)
	}
}

func TestTypedDecodingIntFromEnv(t *testing.T) {
	yamlData := []byte(`
app:
  name: test
  version_display: true
server:
  grpc:
    port: "{ENV:GRPC_PORT:50051}"
    timeout: 1s
`)
	lookup := func(key string) (string, bool) {
		if key == "GRPC_PORT" {
			return "9090", true
		}
		return "", false
	}

	cfg, err := goconfy.Load[TypedConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(lookup),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.GRPC.Port != 9090 {
		t.Errorf("expected port=9090, got %d", cfg.Server.GRPC.Port)
	}
}

func TestTypedDecodingBoolFromEnv(t *testing.T) {
	yamlData := []byte(`
app:
  name: test
  version_display: "{ENV:APP_VERSION_DISPLAY:true}"
server:
  grpc:
    port: 8080
    timeout: 1s
`)
	lookup := func(key string) (string, bool) {
		if key == "APP_VERSION_DISPLAY" {
			return "false", true
		}
		return "", false
	}

	cfg, err := goconfy.Load[TypedConfig](
		goconfy.WithBytes(yamlData),
		goconfy.WithEnvLookup(lookup),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.App.VersionDisplay != false {
		t.Errorf("expected version_display=false, got %v", cfg.App.VersionDisplay)
	}
}
