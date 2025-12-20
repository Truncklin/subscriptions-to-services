package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string           `yaml:"env" envDefault:"localhost"`
	StoragePath string           `yaml:"storage_path" env-required:"true"`
	HttpServer  HttpServerConfig `yaml:"http_server"`
}

type HttpServerConfig struct {
	Host        string        `yaml:"host" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoadConfig(configPath string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Printf("Failed to read config file: %v", err)
		return nil, err
	}
	return &cfg, nil
}
