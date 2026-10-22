package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JathinShyam/NexusLink/internal/analytics"
	"github.com/JathinShyam/NexusLink/internal/config"
	"github.com/JathinShyam/NexusLink/internal/logging"
)

func main() {
	cfg, err := config.LoadConsumer()
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
	defer func() { _ = logCloser.Close() }()

	store, err := analytics.NewClickHouseStore(
		cfg.ClickHouseURL, cfg.ClickHouseDatabase, cfg.ClickHouseUser, cfg.ClickHousePassword,
	)
	if err != nil {
		log.Error("clickhouse store init failed", "error", err)
		os.Exit(1)
	}

	consumer, err := analytics.NewKafkaConsumer(
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID,
		store, cfg.BatchSize, cfg.FlushInterval, log,
	)
	if err != nil {
		log.Error("kafka consumer init failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = consumer.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	if err := store.Ping(pingCtx); err != nil {
		log.Warn("clickhouse unreachable at startup; inserts will retry", "error", err)
	}
	cancelPing()

	health := startHealthServer(cfg.HealthPort, log, store, consumer)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = health.Shutdown(shutdownCtx)
	}()

	log.Info("analytics consumer started",
		"brokers", cfg.KafkaBrokers,
		"topic", cfg.KafkaTopic,
		"group", cfg.KafkaGroupID,
		"clickhouse", cfg.ClickHouseURL,
		"batch_size", cfg.BatchSize,
		"flush_interval", cfg.FlushInterval,
	)

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("consumer stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("analytics consumer stopped")
}

func startHealthServer(port int, log *slog.Logger, store analytics.Store, consumer *analytics.KafkaConsumer) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := store.Ping(readyCtx); err != nil {
			http.Error(w, "clickhouse: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		if err := consumer.Ping(readyCtx); err != nil {
			http.Error(w, "kafka: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("consumer health server error", "error", err)
		}
	}()
	return srv
}
