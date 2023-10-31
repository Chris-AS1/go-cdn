package main

import (
	"context"
	"encoding/json"
	"go-cdn/config"
	"go-cdn/consul"
	"go-cdn/database"
	"go-cdn/server"
	"go-cdn/tracing"

	"go.uber.org/zap"
)

func main() {
	// Logger
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()
	sugar := logger.Sugar()

	// Yaml Configurations
	cfg, err := config.NewConfig()
	dbg, _ := json.Marshal(cfg)
	sugar.Info("Loaded following configs:", string(dbg))

	// Jaeger/OTEL
	trace_ctx := context.Background()
	shutdown, err := tracing.InstallExportPipeline(trace_ctx, &cfg)
	if err != nil {
		sugar.Panic(err)
	}
	defer func() {
		if err := shutdown(trace_ctx); err != nil {
			sugar.Panic(err)
		}
	}()
	// Main span trace
	_, span := tracing.Tracer.Start(trace_ctx, "main")
	defer span.End()

	// Handle Consul Connection/Registration
	if err != nil {
		sugar.Panic("Error reading config file, %s", err)
	}

	csl_client, err := consul.NewConsulClient(&cfg)
	if err != nil {
		sugar.Panicf("Couldn't get Consul Client, connection failed: %s", err)
	}

	if err = csl_client.RegisterService(&cfg); err != nil {
		sugar.Panicf("Couldn't register Consul Service: %s", err)
	}
	defer csl_client.DeregisterService(&cfg)

	// Postgres Connection
	pg_client, err := database.NewPostgresClient(csl_client, &cfg)
	if err != nil {
		sugar.Panicf("Couldn't connect to Postgres: %s", err)
	}
	defer pg_client.CloseConnection()
	if err = pg_client.MigrateDB(); err != nil {
		sugar.Panicf("Couldn't apply migrations to Postgres: %s", err)
	}

	// Redis Connection
	var rd_client *database.RedisClient
	if cfg.Redis.RedisEnable {
		rd_client, err = database.NewRedisClient(csl_client, &cfg)
		if err != nil {
			sugar.Panicf("Couldn't connect to Redis: %s", err)
		}
	}

	// Gin Setup
	// gin.SetMode(gin.ReleaseMode) // Release Mode

	gin_state := &server.GinState{
		Config:      &cfg,
		RedisClient: rd_client,
		PgClient:    pg_client,
		Sugar:       sugar,
	}

	if err = server.SpawnGin(gin_state); err != nil {
		sugar.Panicf("Gin returned an error: %s", err)
	}
}
