package main

import (
	"context"
	"encoding/json"
	"errors"
	"go-cdn/internal/config"
	database "go-cdn/internal/database/controller"
	"go-cdn/internal/database/repository/postgres"
	"go-cdn/internal/database/repository/redis"
	discovery "go-cdn/internal/discovery/controller"
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
		sugar.Errorw("config load", "err", err)
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
		if !errors.Is(err, repository.ErrServiceDisabled) {
			sugar.Panicw("dc registration", "err", err)
		}
	}
	defer func() {
		err := dc.DeregisterService()
		if err != nil {
			if !errors.Is(err, repository.ErrServiceDisabled) {
				sugar.Errorw("dc deregistration", "err", err)
			}
		}
	}()

	// Tracing Setup
	var mctx context.Context
	if cfg.Telemetry.TelemetryEnable {
		tctx := context.Background()
		shutdown, err := tracing.InstallExportPipeline(tctx, dc, cfg)
		if err != nil {
			sugar.Panicw("jaeger/otel setup", "err", err)
		}
		defer func() {
			if err := shutdown(tctx); err != nil {
				sugar.Panicw("jaeger/otel close", "err", err)
			}
		}()

		// Main span trace
		local_mctx, span := tracing.Tracer.Start(tctx, "main")
		defer span.End()
		mctx = local_mctx
	}

	// That's a workaround in the case telemetry is disabled. It will attempt to connect to DefaultCollectorHost.
	if mctx == nil {
		mctx = context.Background()
	}

	// DB Repo
	pg_repo, err := postgres.New(mctx, dc, cfg)
	if err != nil {
		sugar.Panicw("database repo creation", "err", err)
	}
	db := database.New(pg_repo)
	defer db.Close()

	// Cache Repo
	var cache *database.Controller
	if cfg.Cache.RedisEnable {
		rd_repo, err := redis.New(mctx, dc, cfg)
		if err != nil {
			sugar.Panicw("redis repo creation", "err", err)
		}
		cache = database.New(rd_repo)
	}

	// Gin Setup
	ginServer := server.New(cfg, db, cache, sugar)
	ginServer.Spawn(
		server.WithMode(gin.ReleaseMode),
	)
}
