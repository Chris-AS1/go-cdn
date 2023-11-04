package main

import (
	"context"
	"encoding/json"
	"go-cdn/config"
	"go-cdn/consul"
	"go-cdn/database"
	"go-cdn/logger"
	"go-cdn/server"
	"go-cdn/tracing"
)

func main() {
	// Yaml Configurations
	cfg, err := config.NewConfig()

	// Logger (File with rotation + Console)
	sugar := logger.NewLogger(&cfg)
	defer sugar.Sync()

	// Print loaded configs after logger initialization
	dbg, _ := json.Marshal(cfg)
	sugar.Info("Loaded following configs:", string(dbg))

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
	defer func() {
		err := csl_client.DeregisterService(&cfg)
		if err != nil {
			sugar.Panicf("Couldn't de-register Consul Service: %s", err)
		}
	}()

	// Jaeger/OTEL
	trace_ctx := context.Background()
	shutdown, err := tracing.InstallExportPipeline(trace_ctx, csl_client, &cfg)
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
	ginServer := &server.GinServer{
		Config:      &cfg,
		RedisClient: rd_client,
		PgClient:    pg_client,
		Sugar:       sugar,
	}

	ginServer.Spawn()
}
