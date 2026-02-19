package main

import (
	"fmt"
	"log"

	"github.com/keksclan/goConfy/gen/registry"
	"github.com/keksclan/goConfy/gen/yamltemplate"
	"github.com/keksclan/goConfy/types"
)

// Config is a sample application configuration.
type Config struct {
	Host    string         `yaml:"host" default:"0.0.0.0" desc:"Bind address" env:"APP_HOST" example:"localhost"`
	Port    int            `yaml:"port" default:"8080" desc:"HTTP port" env:"APP_PORT"`
	Debug   bool           `yaml:"debug" default:"false" desc:"Enable debug mode"`
	Timeout types.Duration `yaml:"timeout" default:"30s" desc:"Request timeout" env:"APP_TIMEOUT"`
	DB      DBConfig       `yaml:"db"`
}

// DBConfig holds database settings.
type DBConfig struct {
	DSN      string `yaml:"dsn" default:"postgres://localhost:5432/mydb" desc:"Database connection string" env:"DB_DSN" required:"true"`
	Password string `yaml:"password" secret:"true" env:"DB_PASSWORD" desc:"Database password"`
	MaxConn  int    `yaml:"max_conn" default:"25" desc:"Maximum open connections" env:"DB_MAX_CONN"`
}

// configProvider implements registry.Provider for our Config type.
type configProvider struct{}

func (configProvider) ID() string { return "myservice" }
func (configProvider) New() any   { return &Config{} }

func init() {
	registry.Register(configProvider{})
}

func main() {
	// Demonstrate generating a YAML template programmatically.
	p, ok := registry.Get("myservice")
	if !ok {
		log.Fatal("provider not found")
	}

	yaml, err := yamltemplate.Generate(p.New(), yamltemplate.Options{Profile: "dev"})
	if err != nil {
		log.Fatalf("generate: %v", err)
	}
	fmt.Println(yaml)
}
