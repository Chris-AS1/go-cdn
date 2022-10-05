package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go-cdn/utils"
	"log"
	"os"

	"github.com/go-redis/redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func ConnectRedis() string {
	log.Print("Connecting to Redis...")
	rdb = redis.NewClient(&redis.Options{
		Addr:     utils.EnvSettings.RedisURL,
		Password: "", // No Password
		DB:       0,  // use default DB
	})

	result, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	return result
}

// Hashmap with the current available files, <hash: string>:<filename: string>
func BuildFileMap() map[string]string {
	files, err := os.ReadDir(dataFolder)
	var ret = make(map[string]string)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		sum := sha256.Sum256([]byte(file.Name()))
		sum2 := hex.EncodeToString(sum[:])
		ret[sum2[:6]] = file.Name()
	}

	return ret
}

// Records image access on Redis - Most used cache
func recordAccess(file_id string) int64 {
	result, err := rdb.Incr(ctx, file_id).Result()
	if err != nil {
		log.Panic(err)
	}

	return result
}
