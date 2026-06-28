package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JathinShyam/NexusLink/internal/config"
	"github.com/JathinShyam/NexusLink/internal/idgen"
	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/JathinShyam/NexusLink/internal/redirect"
	"github.com/JathinShyam/NexusLink/internal/server"
	"github.com/JathinShyam/NexusLink/internal/shorten"
	"github.com/JathinShyam/NexusLink/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, logCloser, err := logging.Setup(logging.Options{
		Level:   cfg.LogLevel,
		Env:     cfg.Env,
		LogDir:  cfg.LogDir,
		LogFile: cfg.LogFile,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logging.CloseCloser(logCloser); err != nil {
			log.Error("failed to close log file", "error", err)
		}
	}()

	log.Info("starting nexuslink api",
		"env", cfg.Env,
		"port", cfg.Port,
		"log_dir", cfg.LogDir,
		"log_file", cfg.LogFile,
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	repo := storage.NewPostgresRepository(pool)
	generator := idgen.NewSequenceGenerator(pool)
	shortenService := shorten.NewService(repo, generator, cfg.BaseURL)

	srv := server.New(log, cfg.Addr(), server.Options{
		Shorten:  shorten.NewHandler(shortenService),
		Redirect: redirect.NewHandler(repo),
	})

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
}
