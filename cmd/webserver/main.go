package main

import (
	"log/slog"
	"os"

	"SubServices/internal/config"
	"SubServices/internal/storage"
)

func main() {
	slogInit()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local.yaml"
	}

	cfg, err := config.MustLoadConfig(configPath)
	if err != nil {
		slog.Error("Error loading config", slog.Any("error", err))
	}

	slog.Info("Config loaded successfully", slog.String("config_path", configPath), slog.Any("config", cfg))

	pool, err := storage.NewPool(cfg.StoragePath)
	if err != nil {
		slog.Error("StoragePath is incorrect", slog.String("storage_path", cfg.StoragePath), slog.Any("storage", err))
		os.Exit(1)
	}

	err = storage.RunMigrations(cfg.StoragePath)
	if err != nil {
		slog.Error("Failed to run migrations", slog.Any("migrations", err))
		os.Exit(1)
	}
	_ = pool
}

func slogInit() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}
