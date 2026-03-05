package tests

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keksclan/goConfy/tools/internal/tools/generator/commands"
	"github.com/keksclan/goConfy/tools/generator/registry"
)

var updateGolden = flag.Bool("update", false, "update golden files")

type goldenTestConfig struct {
	App      string   `yaml:"app" desc:"The application name" default:"myapp"`
	Port     int      `yaml:"port" desc:"The port to listen on" env:"PORT" default:"8080"`
	Debug    bool     `yaml:"debug" desc:"Enable debug mode" default:"false"`
	APIKey   string   `yaml:"api_key" desc:"Secret API key" secret:"true" env:"API_KEY"`
	Servers  []string `yaml:"servers" desc:"List of servers" example:"s1,s2"`
	Database struct {
		Host string `yaml:"host" default:"localhost"`
		User string `yaml:"user" default:"admin"`
	} `yaml:"database"`
}

type goldenTestProvider struct{}

func (p *goldenTestProvider) ID() string { return "test-config" }
func (p *goldenTestProvider) New() any   { return &goldenTestConfig{} }

func init() {
	registry.Register(&goldenTestProvider{})
}

func TestGolden(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("init", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "init-config.yml")
		args := []string{"-id", "test-config", "-out", outPath, "-profile", "dev", "-dotenv"}

		if err := commands.RunInit(args); err != nil {
			t.Fatalf("RunInit failed: %v", err)
		}

		compareGolden(t, outPath, "init-config.yml")
		compareGolden(t, filepath.Join(tempDir, ".env"), "init.env")
	})

	t.Run("fmt", func(t *testing.T) {
		inPath := filepath.Join(tempDir, "fmt-in.yml")
		content := []byte("app:  \"myapp\"\nport: 8080\ndatabase:\n  host: localhost\n")
		if err := os.WriteFile(inPath, content, 0644); err != nil {
			t.Fatalf("failed to write input: %v", err)
		}

		outPath := filepath.Join(tempDir, "fmt-out.yml")
		args := []string{"-in", inPath, "-out", outPath}
		if err := commands.RunFmt(args); err != nil {
			t.Fatalf("RunFmt failed: %v", err)
		}

		compareGolden(t, outPath, "fmt-out.yml")
	})

	t.Run("dump", func(t *testing.T) {
		inPath := filepath.Join(tempDir, "dump-in.yml")
		content := []byte("app: myapp\nport: 9000\napi_key: secret-value\n")
		if err := os.WriteFile(inPath, content, 0644); err != nil {
			t.Fatalf("failed to write input: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		args := []string{"-id", "test-config", "-in", inPath}
		err := commands.RunDump(args)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("RunDump failed: %v", err)
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(r); err != nil {
			t.Fatalf("failed to read from pipe: %v", err)
		}

		dumpOutPath := filepath.Join(tempDir, "dump-out.json")
		if err := os.WriteFile(dumpOutPath, buf.Bytes(), 0644); err != nil {
			t.Fatalf("failed to write dump output: %v", err)
		}

		compareGolden(t, dumpOutPath, "dump-out.json")
	})
}

func compareGolden(t *testing.T, actualPath, goldenName string) {
	t.Helper()
	actual, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("failed to read actual output: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden", goldenName)

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
			t.Fatalf("failed to create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file %s does not exist. run with -update to create it.", goldenPath)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	// OS-independent comparison: normalize line endings
	actualStr := strings.ReplaceAll(string(actual), "\r\n", "\n")
	expectedStr := strings.ReplaceAll(string(expected), "\r\n", "\n")

	if actualStr != expectedStr {
		t.Errorf("output does not match golden file %s\nACTUAL:\n%s\nEXPECTED:\n%s", goldenName, actualStr, expectedStr)

		// Optional: write actual to a file for easier debugging
		diffPath := filepath.Join("testdata", "golden", goldenName+".actual")
		_ = os.WriteFile(diffPath, []byte(actualStr), 0644)
		t.Errorf("actual output written to %s for comparison", diffPath)
	}
}
