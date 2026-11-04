package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mtuan8820/unified-document-viewer/api"
	"github.com/mtuan8820/unified-document-viewer/internal/adapters"
	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
	"github.com/mtuan8820/unified-document-viewer/internal/config"
	"github.com/mtuan8820/unified-document-viewer/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// One client shared by both adapters (safe for concurrent use).
	// No Timeout here: the aggregator's per-source context sets the deadline.
	client := &http.Client{}

	agg := &aggregator.Aggregator{
		Logger:         logger,
		Cache:          nil, // TODO: Redis cache
		Snapshot:       nil, // TODO: Postgres snapshot store
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
