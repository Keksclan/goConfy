package main

import (
	"fmt"
	"log"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/types"
)

// Config demonstrates typed decoding with dotenv + YAML macros.
type Config struct {
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

func main() {
	cfg, err := goconfy.Load[Config](
		goconfy.WithFile("examples/dotenv/config.yml"),
		goconfy.WithDotEnvFile("examples/dotenv/.env"),
		goconfy.WithStrictYAML(true),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("App Name:         %s\n", cfg.App.Name)
	fmt.Printf("Version Display:  %v\n", cfg.App.VersionDisplay)
	fmt.Printf("GRPC Port:        %d\n", cfg.Server.GRPC.Port)
	fmt.Printf("GRPC Timeout:     %s\n", cfg.Server.GRPC.Timeout)

	// Redacted JSON dump (secrets would be hidden if tagged).
	out, err := goconfy.DumpRedactedJSON(cfg)
	if err != nil {
		log.Fatalf("failed to dump redacted JSON: %v", err)
	}
	fmt.Println(out)
}
