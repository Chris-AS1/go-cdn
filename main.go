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

	// Image list to be used on endpoints
	available_files, err := pg_client.GetFileList()
	sugar.Infow("available files", available_files)
	if err != nil {
		sugar.Panicf("Error retrieving current files: %s", err)
	}

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

	if err = server.SpawnGin(gin_state, available_files); err != nil {
		sugar.Panicf("Gin returned an error: %s", err)
	}
}

/* func main() {
	utils.LoadEnv()
	fileMap = database.BuildFileMap()

	log.Printf("Redis connection: %s", database.ConnectRedis())
	log.Print("Starting Server")

	// go refreshClock()

	r := mux.NewRouter().StrictSlash(true)

	// Disabled
	r.HandleFunc("/", RootHandler)

	// Serving Image Path
	b, err := strconv.ParseBool(utils.EnvSettings.DeliveringSubPathEnable)
	if err != nil {
		log.Panic(err)
	}

	if b {
		log.Printf("Serving Path: /%s/{id}/", utils.EnvSettings.DeliveringSubPath)

		url, url_id := fmt.Sprintf("/%s", utils.EnvSettings.DeliveringSubPath),
			fmt.Sprintf("/%s/{id}", utils.EnvSettings.DeliveringSubPath)

		r.HandleFunc(url, GetImageHandler).Methods("GET")
		r.HandleFunc(url_id, GetImageHandler).Methods("GET")

		// Check if insertion endpoint is enabled
		add, err := strconv.ParseBool(utils.EnvSettings.EnableInsertion)
		if add {
			r.HandleFunc(url, PostImageHandler).Methods("POST")
		}

		if err != nil {
			log.Panic(err)
		}

		// Check if deletion endpoint is enabled
		del, err := strconv.ParseBool(utils.EnvSettings.EnableDeletion)

		if del {
			r.HandleFunc(url_id, DeleteImageHandler).Methods("DELETE")
		}

		if err != nil {
			log.Panic(err)
		}
	} else {
		log.Print("Serving Path: /{id}/")
		r.HandleFunc("/", GetImageHandler).Methods("GET")
		r.HandleFunc("/{id:[0-9]+}", GetImageHandler).Methods("GET")
	}

	// Serve List Path
	r.HandleFunc("/list/", GetListHandler)

	// Use Router
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", utils.EnvSettings.DeliveringPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Serving Port: %s", utils.EnvSettings.DeliveringPort)
	log.Fatal(srv.ListenAndServe())
} */
