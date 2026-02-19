package tests

import (
	"testing"

	goconfy "github.com/keksclan/goConfy"
)

type StrictConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func TestStrictModeRejectsUnknown(t *testing.T) {
	input := []byte(`
name: myapp
port: 8080
unknown_field: oops
`)
	_, err := goconfy.Load[StrictConfig](
		goconfy.WithBytes(input),
		goconfy.WithStrictYAML(true),
	)
	if err == nil {
		t.Fatal("expected error for unknown field in strict mode")
	}
}

func TestStrictModeDefault(t *testing.T) {
	// Strict is the default, so unknown fields should be rejected without explicit WithStrictYAML.
	input := []byte(`
name: myapp
port: 8080
extra: bad
`)
	_, err := goconfy.Load[StrictConfig](goconfy.WithBytes(input))
	if err == nil {
		t.Fatal("expected error for unknown field (strict is default)")
	}
}

func TestRelaxedModeAllowsUnknown(t *testing.T) {
	input := []byte(`
name: myapp
port: 8080
unknown_field: ok
`)
	cfg, err := goconfy.Load[StrictConfig](
		goconfy.WithBytes(input),
		goconfy.WithStrictYAML(false),
	)
	if err != nil {
		t.Fatalf("unexpected error in relaxed mode: %v", err)
	}
	if cfg.Name != "myapp" {
		t.Errorf("expected name=myapp, got %q", cfg.Name)
	}
}
