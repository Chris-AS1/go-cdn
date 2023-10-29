package main

import (
	"encoding/json"
	"go-cdn/config"
	"go-cdn/consul"
	"go-cdn/database"
	"go-cdn/server"

	"go.uber.org/zap"
)

/* // Root Handle - Version Number
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "API v1")
}

// Lists files on a directory
func GetListHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Authenticate

	w.WriteHeader(http.StatusOK)
	for k, v := range fileMap {
		io.WriteString(w, k+" "+v+"\n")
	}
}

// Builds the correct path given the filename
func getImagePath(filename string) string {
	return fmt.Sprintf("%s/%s", dataFolder, filename)
}
*/

func main() {
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.NewConfig()
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
	defer csl_client.DeregisterService(&cfg)

	// Handle Postgres Connection
	pg_client, err := database.NewPostgresClient(csl_client, &cfg)
	if err != nil {
		sugar.Panicf("Couldn't connect to Postgres: %s", err)
	}
	defer pg_client.CloseConnection()

	// Handle Redis Connection
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
