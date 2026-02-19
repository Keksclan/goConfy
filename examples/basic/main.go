package main

import (
	"fmt"
	"log"

	goconfy "github.com/keksclan/goConfy"
	"github.com/keksclan/goConfy/types"
)

// Config is the application configuration.
type Config struct {
	Host    string         `yaml:"host"`
	Port    int            `yaml:"port"`
	DB      DBConfig       `yaml:"db"`
	Timeout types.Duration `yaml:"timeout"`
}

// DBConfig holds database settings.
type DBConfig struct {
	DSN      string `yaml:"dsn"`
	Password string `yaml:"password" secret:"true"`
}

func main() {
	cfg, err := goconfy.Load[Config](
		goconfy.WithFile("examples/basic/config.yml"),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Timeout: %s\n", cfg.Timeout)

	// Redacted output hides secrets
	redacted := goconfy.Redacted(cfg)
	fmt.Printf("Redacted: %v\n", redacted)

	json, err := goconfy.DumpRedactedJSON(cfg)
	if err != nil {
		log.Fatalf("failed to dump redacted JSON: %v", err)
	}
	fmt.Println(json)
}
