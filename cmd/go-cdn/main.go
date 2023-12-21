package main

import (
	"context"
	"encoding/json"
	"go-cdn/internal/config"
	"go-cdn/internal/consul"
	"go-cdn/internal/database"
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

	// Handle Consul Connection/Registration
	var csl_client *consul.ConsulClient
	if cfg.Consul.ConsulEnable {
		csl_client, err = consul.NewConsulClient(cfg)
		if err != nil {
			sugar.Panicw("consul connection", "err", err)
		}

		if err = csl_client.RegisterService(cfg); err != nil {
			sugar.Panicw("consul service registration", "err", err)
		}
		defer func() {
			err := csl_client.DeregisterService(cfg)
			if err != nil {
				sugar.Panicw("consul servie deregistration", "err", err)
			}
		}()
	}

	// Jaeger/OTEL
	if cfg.Telemetry.TelemetryEnable {
		trace_ctx := context.Background()
		shutdown, err := tracing.InstallExportPipeline(trace_ctx, csl_client, cfg)
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

	// Postgres Repo/Ctrl
	pg_repo, err := database.NewPostgresRepository(csl_client, cfg)
	if err != nil {
		sugar.Panicw("database repo creation", "err", err)
	}
	db := database.NewController(pg_repo)
	defer db.Close()

	// Redis Connection
	var cache *database.Controller
	if cfg.Redis.RedisEnable {
		rd_repo, err := database.NewRedisRepository(csl_client, cfg)
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
