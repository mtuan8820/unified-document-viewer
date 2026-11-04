package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mtuan8820/unified-document-viewer/api"
	"github.com/mtuan8820/unified-document-viewer/internal/adapters"
	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
	"github.com/mtuan8820/unified-document-viewer/internal/cache"
	"github.com/mtuan8820/unified-document-viewer/internal/config"
	"github.com/mtuan8820/unified-document-viewer/internal/httpapi"
	"github.com/mtuan8820/unified-document-viewer/internal/snapshot"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// init redis client
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	// The client connects lazily, so ping once to surface a bad address early.
	// Not fatal: on cache errors the aggregator falls back to upstream.
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis unavailable at startup, requests will skip the cache", "addr", cfg.RedisAddr, "error", err)
	}

	redisCache := cache.New(redisClient, 5*time.Minute)

	// init postgres pool. pgxpool.New only parses the URL; it connects lazily.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Same idea as redis: warn, don't exit. Snapshot errors are logged by the
	// aggregator and the source falls back to timeout/error status.
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := pool.Ping(pingCtx); err != nil {
		logger.Warn("postgres unavailable at startup, snapshot fallback will not work", "error", err)
	}
	cancel()

	snapshotStore := snapshot.New(pool)

	// One client shared by both adapters (safe for concurrent use).
	// No Timeout here: the aggregator's per-source context sets the deadline.
	client := &http.Client{}

	agg := &aggregator.Aggregator{
		Logger:         logger,
		Cache:          redisCache,
		Snapshot:       snapshotStore,
		SnapshotMaxAge: cfg.SnapshotMaxAge,
		Timeout:        cfg.Upstream.Timeout,
		Retries:        cfg.Upstream.Retries,
		RetryDelay:     cfg.Upstream.RetryDelay,
		Sources: []aggregator.Source{
			adapters.NewSalesSystemAdapter(cfg.Upstream.SalesURL, client),
			adapters.NewServiceSystemAdapter(cfg.Upstream.ServiceURL, client),
		},
	}

	r := gin.Default()
	api.RegisterHandlers(r, httpapi.NewHandler(agg))
	if err := r.Run(cfg.HTTPAddr); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
