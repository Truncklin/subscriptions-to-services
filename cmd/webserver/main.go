package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SubServices/internal/config"
	"SubServices/internal/http/handlers"
	"SubServices/internal/http/router"
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
		os.Exit(1)
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

	// Инициализация HTTP
	h := handlers.NewHandler(pool)
	r := router.InitRouter(h)

	srv := &http.Server{
		Addr:         cfg.HttpServer.Host,
		Handler:      r,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting HTTP server", slog.String("host", cfg.HttpServer.Host))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	} else {
		slog.Info("Server stopped gracefully")
	}

	pool.Close()
}

func slogInit() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}
