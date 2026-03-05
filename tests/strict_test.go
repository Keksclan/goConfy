package tests

import (
	"errors"
	"strings"
	"testing"

	goconfy "github.com/keksclan/goConfy"
)

type StrictConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func TestStrictModeRejectsUnknown(t *testing.T) {
	input := []byte(`name: myapp
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

	var fe *goconfy.FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("expected FieldError, got %T: %v", err, err)
	}

	if fe.Field != "unknown_field" {
		t.Errorf("expected field 'unknown_field', got %q", fe.Field)
	}
	if fe.Line != 3 {
		t.Errorf("expected line 3, got %d", fe.Line)
	}
	if fe.Layer != "base" {
		t.Errorf("expected layer 'base', got %q", fe.Layer)
	}
	if !strings.Contains(fe.Error(), "[base]") {
		t.Errorf("expected error message to contain layer [base], got %q", fe.Error())
	}
	if !strings.Contains(fe.Error(), "line 3") {
		t.Errorf("expected error message to contain line 3, got %q", fe.Error())
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

func TestErrorLayer(t *testing.T) {
	input := []byte(`
server:
  host: "{ENV:MISSING_VAR}"
`)
	_, err := goconfy.Load[StrictConfig](
		goconfy.WithBytes(input),
	)
	if err == nil {
		t.Fatal("expected error for missing env var")
	}

	var fe *goconfy.FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("expected FieldError, got %T: %v", err, err)
	}

	if fe.Layer != "base" {
		t.Errorf("expected layer 'base', got %q", fe.Layer)
	}
	if fe.Path != "server.host" {
		t.Errorf("expected path 'server.host', got %q", fe.Path)
	}
}
