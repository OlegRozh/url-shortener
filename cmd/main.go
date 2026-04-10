package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/OlegRozh/url-shortener/internal/config"
	"github.com/OlegRozh/url-shortener/internal/http-server/server"
	"github.com/OlegRozh/url-shortener/internal/lib/logger/sl"
	"github.com/lmittmann/tint"

	"github.com/OlegRozh/url-shortener/storage/postgres"
	_ "github.com/lib/pq"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")
	// Создаем контекст (например, без тайм-аута — контекст "без ограничений")
	ctx := context.Background()

	storage, err := postgres.NewPostgresStorage(ctx, &cfg, log)
	if err != nil {
		log.Error("failed to init storage", slog.Any("error", err))
		os.Exit(1)
		return
	}
	defer storage.Close()

	srv := server.New(log, &cfg, storage)
	if err := srv.Start(); err != nil {
		log.Error("server error", sl.Err(err))
		os.Exit(1)
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}))
}
