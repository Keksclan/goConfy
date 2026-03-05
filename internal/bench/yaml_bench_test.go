package bench

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keksclan/goConfy/internal/yamlparse"
	"gopkg.in/yaml.v3"
)

func BenchmarkParseSmallV3(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "basic", "config.yml"))
	if err != nil {
		b.Fatalf("failed to read small config file: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := yamlparse.ParseBytes(data)
		if err != nil {
			b.Fatalf("yaml parse error: %v", err)
		}
	}
}

func BenchmarkParseMediumV3(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("..", "..", "config.example.yml"))
	if err != nil {
		b.Fatalf("failed to read medium config file: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := yamlparse.ParseBytes(data)
		if err != nil {
			b.Fatalf("yaml parse error: %v", err)
		}
	}
}

func BenchmarkParseLargeV3(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&buf, "item_%d:\n  key1: value1\n  key2: value2\n  nested:\n    field: {ENV:ITEM_%d_NESTED:default}\n", i, i)
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := yamlparse.ParseBytes(data)
		if err != nil {
			b.Fatalf("yaml parse error: %v", err)
		}
	}
}

func BenchmarkStrictDecodingV3(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("..", "..", "config.example.yml"))
	if err != nil {
		b.Fatalf("failed to read example config file for strict decoding: %v", err)
	}
	type Config struct {
		App struct {
			Name     string `yaml:"name"`
			Env      string `yaml:"env"`
			Port     string `yaml:"port"`
			LogLevel string `yaml:"log_level"`
		} `yaml:"app"`
		DB struct {
			Host     string `yaml:"host"`
			Port     string `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Name     string `yaml:"name"`
			MaxConns int    `yaml:"max_conns"`
			Timeout  string `yaml:"timeout"`
		} `yaml:"db"`
		Redis struct {
			URL string `yaml:"url"`
			TTL string `yaml:"ttl"`
		} `yaml:"redis"`
		Features struct {
			EnableTracing bool `yaml:"enable_tracing"`
			EnableMetrics bool `yaml:"enable_metrics"`
		} `yaml:"features"`
		Profiles map[string]any `yaml:"profiles"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		_ = dec.Decode(&cfg)
	}
}
