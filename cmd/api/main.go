package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JathinShyam/NexusLink/internal/auth"
	"github.com/JathinShyam/NexusLink/internal/cache"
	"github.com/JathinShyam/NexusLink/internal/config"
	"github.com/JathinShyam/NexusLink/internal/idgen"
	"github.com/JathinShyam/NexusLink/internal/links"
	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/JathinShyam/NexusLink/internal/metrics"
	"github.com/JathinShyam/NexusLink/internal/ratelimit"
	"github.com/JathinShyam/NexusLink/internal/redirect"
	"github.com/JathinShyam/NexusLink/internal/server"
	"github.com/JathinShyam/NexusLink/internal/shorten"
	"github.com/JathinShyam/NexusLink/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Error("invalid redis url", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOpts)
	defer func() { _ = redisClient.Close() }()

	urlCache := cacheForRedis(ctx, log, redisClient)

	repo := storage.NewPostgresRepository(pool)
	generator, err := idgen.NewSnowflakeGenerator(cfg.MachineID)
	if err != nil {
		log.Error("snowflake generator init failed", "error", err)
		os.Exit(1)
	}

	m := metrics.New()
	resolver := redirect.NewResolver(repo, urlCache, m)
	shortenService := shorten.NewService(repo, generator, cfg.BaseURL, urlCache)
	linksService := links.NewService(repo, urlCache)
	jwtService := auth.NewJWT(cfg.JWTSecret, cfg.JWTTTL)

	shortenLimiter := ratelimit.NewFixedWindowLimiter(
		redisClient, "rl:shorten", cfg.ShortenRateLimit, cfg.ShortenRateWindow,
	)
	redirectLimiter := ratelimit.NewTokenBucketLimiter(
		redisClient, "rl:redirect", cfg.RedirectBucketCapacity, cfg.RedirectRefillPerSec,
	)

	shortenHandler := ratelimit.Middleware(shortenLimiter, "shorten", m)(
		auth.OptionalMiddleware(jwtService)(
			http.HandlerFunc(shorten.NewHandler(shortenService).Create),
		),
	)
	redirectHandler := ratelimit.Middleware(redirectLimiter, "redirect", m)(
		http.HandlerFunc(redirect.NewHandler(resolver).Serve),
	)
	deleteHandler := auth.Middleware(jwtService)(
		http.HandlerFunc(links.NewHandler(linksService).Delete),
	)

	srv := server.New(log, cfg.Addr(), server.Options{
		Shorten:  shortenHandler,
		Redirect: redirectHandler,
		Delete:   deleteHandler,
		Metrics:  m,
		Ready:    readyCheck(pool, redisClient),
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

func cacheForRedis(ctx context.Context, log *slog.Logger, client *redis.Client) cache.URLCache {
	if err := client.Ping(ctx).Err(); err != nil {
		log.Warn("redis unavailable, using noop cache", "error", err)
		return cache.NoopCache{}
	}
	log.Info("redis cache connected")
	return cache.NewRedisCache(client)
}

func readyCheck(pool *pgxpool.Pool, redisClient *redis.Client) server.ReadyChecker {
	return func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			return fmt.Errorf("database: %w", err)
		}
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis: %w", err)
		}
		return nil
	}
}
