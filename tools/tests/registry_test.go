package tests

import (
	"slices"
	"testing"

	"github.com/keksclan/goConfy/tools/generator/registry"
)

type testProvider struct {
	id string
}

func (p testProvider) ID() string { return p.id }
func (p testProvider) New() any   { return &SampleConfig{} }

func TestRegistryRegisterAndGet(t *testing.T) {
	registry.Register(testProvider{id: "test-svc"})

	p, ok := registry.Get("test-svc")
	if !ok {
		t.Fatal("expected provider to be found")
	}
	if p.ID() != "test-svc" {
		t.Fatalf("expected id %q, got %q", "test-svc", p.ID())
	}

	cfg := p.New()
	if cfg == nil {
		t.Fatal("expected non-nil config instance")
	}
	if _, ok := cfg.(*SampleConfig); !ok {
		t.Fatalf("expected *SampleConfig, got %T", cfg)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	_, ok := registry.Get("nonexistent-id-12345")
	if ok {
		t.Fatal("expected provider NOT to be found")
	}
}

func TestRegistryList(t *testing.T) {
	registry.Register(testProvider{id: "list-test"})
	ids := registry.List()
	if !slices.Contains(ids, "list-test") {
		t.Fatalf("expected ids to contain %q, got %v", "list-test", ids)
	}
}
