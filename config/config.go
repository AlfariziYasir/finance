package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME"`
	AppVersion  string `env:"APP_VERSION"`
	DSN         string `env:"DSN"`
	LogLevel    string `env:"LOG_LEVEL"`
	AppHost     string `env:"APP_HOST"`
	HttpPort    int    `env:"HTTP_PORT"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("read env error: %w", err)
	}

	return &cfg, nil
}
