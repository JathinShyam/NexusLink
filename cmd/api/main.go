package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JathinShyam/NexusLink/internal/config"
	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/JathinShyam/NexusLink/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logging.New(cfg.LogLevel, cfg.Env)
	log.Info("starting nexuslink api",
		"env", cfg.Env,
		"port", cfg.Port,
	)

	srv := server.New(log, cfg.Addr())

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
}
