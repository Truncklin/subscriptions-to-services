package main

import (
	"log/slog"
	"os"

	"github.com/spf13/pflag"

	"SubServices/internal/config"
)

func main() {
	slogInit()

	configPath := pafseFlags()

	cfg, err := config.MustLoadConfig(configPath)
	if err != nil {
		slog.Error("Error loading config", slog.Any("error", err))
	}

	slog.Info("Config loaded successfully", slog.Any("config", cfg))
}

func slogInit() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

func pafseFlags() string {

	var configPath string

	pflag.StringVar(&configPath, "config", "configs/local.yaml", "Path to config file")
	pflag.Parse()

	return configPath
}
