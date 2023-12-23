package main

import (
	"context"
	"encoding/json"
	"errors"
	"go-cdn/internal/config"
	"go-cdn/internal/database/controller"
	"go-cdn/internal/database/repository/postgres"
	"go-cdn/internal/database/repository/redis"
	"go-cdn/internal/discovery/controller"
	"go-cdn/internal/discovery/repository"
	"go-cdn/internal/logger"
	"go-cdn/internal/server"
	"go-cdn/internal/tracing"

	"github.com/gin-gonic/gin"
)

func main() {
	// Yaml Configurations
	cfg, err := config.New()

	// Logger (File with rotation + Console)
	sugar := logger.NewLogger(cfg)
	defer sugar.Sync()

	// Print loaded configs after logger initialization
	if err != nil {
		sugar.Panicw("config load", "err", err)
	}
	dbg, _ := json.Marshal(cfg)
	sugar.Infow("config load", "config", string(dbg), "err", err)

	// Handle Service Discovery Connection/Registration
	dc_builder, err := discovery.NewControllerBuilder().FromConfigs(cfg)
	if err != nil {
		sugar.Panicw("dc builder", "err", err)
	}
	dc := dc_builder.Build()
	if err = dc.RegisterService(); err != nil {
		if errors.Is(err, repository.ErrServiceDisabled) {
			sugar.Errorw("dc registration", "err", err)
		} else {
			sugar.Panicw("dc registration", "err", err)
		}
	}
	defer func() {
		err := dc.DeregisterService()
		if err != nil {
			if errors.Is(err, repository.ErrServiceDisabled) {
				sugar.Errorw("dc deregistration", "err", err)
			} else {
				sugar.Panicw("dc deregistration", "err", err)
			}
		}
	}()

	// Jaeger/OTEL
	if cfg.Telemetry.TelemetryEnable {
		trace_ctx := context.Background()
		shutdown, err := tracing.InstallExportPipeline(trace_ctx, dc, cfg)
		if err != nil {
			sugar.Panicw("jaeger/otel setup", "err", err)
		}
		defer func() {
			if err := shutdown(trace_ctx); err != nil {
				sugar.Panicw("jaeger/otel close", "err", err)
			}
		}()

		// Main span trace
		_, span := tracing.Tracer.Start(trace_ctx, "main")
		defer span.End()
	}

	// DB Repo
	pg_repo, err := postgres.NewPostgresRepository(dc, cfg)
	if err != nil {
		sugar.Panicw("database repo creation", "err", err)
	}
	db := database.NewController(pg_repo)
	defer db.Close()

	// Cache Repo
	var cache *database.Controller
	if cfg.Cache.RedisEnable {
		rd_repo, err := redis.NewRedisRepository(dc, cfg)
		if err != nil {
			sugar.Panicw("redis repo creation", "err", err)
		}
		cache = database.NewController(rd_repo)
	}

	// Gin Setup
	ginServer := server.New(cfg, db, cache, sugar)
	ginServer.Spawn(
		server.WithMode(gin.ReleaseMode),
	)
}
