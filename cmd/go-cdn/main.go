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
)

func main() {
	// Yaml Configurations
	cfg, err := config.NewConfig()
	// Logger (File with rotation + Console)
	sugar := logger.NewLogger(&cfg)
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
		csl_client, err = consul.NewConsulClient(&cfg)
		if err != nil {
			sugar.Panicw("consul connection", "err", err)
		}

		if err = csl_client.RegisterService(&cfg); err != nil {
			sugar.Panicw("consul service registration", "err", err)
		}
		defer func() {
			err := csl_client.DeregisterService(&cfg)
			if err != nil {
				sugar.Panicw("consul servie deregistration", "err", err)
			}
		}()
	}

	// Jaeger/OTEL
	if cfg.Telemetry.TelemetryEnable {
		trace_ctx := context.Background()
		shutdown, err := tracing.InstallExportPipeline(trace_ctx, csl_client, &cfg)
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
	db_repo, err := database.NewPostgresRepo(csl_client, &cfg)
	if err != nil {
		sugar.Panicw("database repo creation", "err", err)
	}
	db_ctrl := database.NewController(db_repo)
	defer db_ctrl.Close()

	// Redis Connection
	var rd_client *database.RedisClient
	if cfg.Redis.RedisEnable {
		rd_client, err = database.NewRedisClient(csl_client, &cfg)
		if err != nil {
			sugar.Panicw("redis connection", "err", err)
		}
	}

	// Gin Setup
	// gin.SetMode(gin.ReleaseMode) // Release Mode
	ginServer := &server.GinServer{
		Config:      &cfg,
		RedisClient: rd_client,
		PgClient:    db_repo,
		Sugar:       sugar,
	}

	ginServer.Spawn()
}
