package tests

import (
	"strings"
	"testing"

	"github.com/keksclan/goConfy/tools/generator/yamltemplate"
)

// SampleConfig is a test config struct with various tags.
type SampleConfig struct {
	Host    string         `yaml:"host" default:"localhost" desc:"Application host" env:"APP_HOST" example:"myapp.example.com"`
	Port    int            `yaml:"port" default:"8080" desc:"Application port" env:"APP_PORT"`
	Debug   bool           `yaml:"debug" default:"false" desc:"Enable debug mode"`
	DB      SampleDBConfig `yaml:"db"`
	Tags    []string       `yaml:"tags" example:"api,web" desc:"Service tags"`
	EnvTags []string       `yaml:"env_tags" env:"APP_TAGS" sep:"," desc:"Tags from env"`
}

type SampleDBConfig struct {
	DSN      string `yaml:"dsn" default:"postgres://localhost/mydb" desc:"Database connection string" env:"DB_DSN" required:"true"`
	Password string `yaml:"password" secret:"true" env:"DB_PASSWORD" desc:"Database password" default:"supersecret"`
	MaxConn  int    `yaml:"max_conn" default:"10" desc:"Max connections"`
}

func TestGenerateYAMLTemplate(t *testing.T) {
	out, err := yamltemplate.Generate(&SampleConfig{}, yamltemplate.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should contain YAML keys.
	assertContains(t, out, "host:")
	assertContains(t, out, "port:")
	assertContains(t, out, "db:")
	assertContains(t, out, "dsn:")
	assertContains(t, out, "password:")

	// Should contain comments.
	assertContains(t, out, "# Application host")
	assertContains(t, out, "# env: APP_HOST")
	assertContains(t, out, "# required: true")
	assertContains(t, out, "# secret: true")
	assertContains(t, out, "# example: myapp.example.com")

	// Should contain macro strings for env fields.
	assertContains(t, out, `"{ENV:APP_HOST:localhost}"`)
	assertContains(t, out, `"{ENV:APP_PORT:8080}"`)
	assertContains(t, out, `"{ENV:DB_DSN:postgres://localhost/mydb}"`)

	// Secret with env should have empty default in macro.
	assertContains(t, out, `"{ENV:DB_PASSWORD:}"`)

	// Secret default value must NOT appear in output.
	assertNotContains(t, out, "supersecret")

	// Non-env bool field should have plain default.
	assertContains(t, out, "debug: false")
}

func TestGenerateYAMLTemplateSlices(t *testing.T) {
	out, err := yamltemplate.Generate(&SampleConfig{}, yamltemplate.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Tags (no env) should use example items as list.
	assertContains(t, out, "tags:")
	assertContains(t, out, "- api")
	assertContains(t, out, "- web")

	// EnvTags (with env) should use macro format.
	assertContains(t, out, "env_tags:")
	assertContains(t, out, `"{ENV:APP_TAGS:}"`)
}

func TestGenerateYAMLTemplateWithProfile(t *testing.T) {
	out, err := yamltemplate.Generate(&SampleConfig{}, yamltemplate.Options{Profile: "dev"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	assertContains(t, out, "profiles:")
	assertContains(t, out, "dev:")
	assertContains(t, out, "staging:")
	assertContains(t, out, "prod:")
}

func TestGenerateDotEnv(t *testing.T) {
	out, err := yamltemplate.GenerateDotEnv(&SampleConfig{})
	if err != nil {
		t.Fatalf("GenerateDotEnv failed: %v", err)
	}

	// Should contain env vars.
	assertContains(t, out, "APP_HOST=localhost")
	assertContains(t, out, "APP_PORT=8080")
	assertContains(t, out, "DB_DSN=postgres://localhost/mydb")

	// Secret should have empty value.
	assertContains(t, out, "DB_PASSWORD=\n")

	// Secret default must NOT appear.
	assertNotContains(t, out, "supersecret")

	// Should contain description comments.
	assertContains(t, out, "# Application host")
	assertContains(t, out, "# Database password")
}

func TestGenerateNilValue(t *testing.T) {
	_, err := yamltemplate.Generate(nil, yamltemplate.Options{})
	if err == nil {
		t.Fatal("expected error for nil value")
	}
}

func TestGenerateNonStruct(t *testing.T) {
	s := "not a struct"
	_, err := yamltemplate.Generate(&s, yamltemplate.Options{})
	if err == nil {
		t.Fatal("expected error for non-struct value")
	}
}

func TestGenerateStableOutput(t *testing.T) {
	// Generate twice and ensure output is identical (stable).
	out1, err := yamltemplate.Generate(&SampleConfig{}, yamltemplate.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	out2, err := yamltemplate.Generate(&SampleConfig{}, yamltemplate.Options{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if out1 != out2 {
		t.Fatalf("output is not stable:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", needle, haystack)
	}
}
