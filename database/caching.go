package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-cdn/config"
	"log"
	"os"

	"github.com/go-redis/redis/v9"
)

type RedisClient struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	rc := &RedisClient{
		ctx: context.Background(),
	}
	_, err := rc.connect(cfg)
	return rc, err
}

func (rc *RedisClient) connect(cfg *config.Config) (bool, error) {
	log.Print("Connecting to Redis")
	rc.rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.RedisAddress, cfg.Redis.RedisPort),
		Password: cfg.Redis.RedisPassword,
		DB:       cfg.Redis.RedisDB,
	})

	result, err := rc.rdb.Ping(rc.ctx).Result()
	return result == "ping: PONG", err
}

// Hashmap with the current available files, <hash: string>:<filename: string>
func BuildFileMap() map[string]string {
	files, err := os.ReadDir("")
	// TODO Fix
	// files, err := os.ReadDir(dataFolder)
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
func (rc *RedisClient) RecordAccess(file_id string) int64 {
	result, err := rc.rdb.ZIncrBy(rc.ctx, "zset1", 1, file_id).Result()
	if err != nil {
		log.Panic(err)
	}

	return int64(result)
}

func (rc *RedisClient) GetFromCache(file_id string) (bool, []byte) {
	result, err := rc.rdb.Get(rc.ctx, file_id).Result()

	// "" If empty or nil (not error)
	if len(result) == 0 || err == redis.Nil {
		log.Printf("[CACHE] Not found, adding it now [%s]", file_id)
		// buff, err := os.ReadFile(getImagePath(fileMap[file_id]))

		if err != nil {
			log.Print(err)
			return false, nil
		}

		// _, err = rdb_bytes.Set(ctx, file_id, string(buff), 0).Result()

		if err != nil {
			log.Print(err)
		}

		return false, nil
	} else {
		return true, []byte(result)
	}
}

// Every X amount, check that DB 0 (HitN: Filenames) and DB 1 (Filenames: Bytes) are in sync
func (rc *RedisClient) RefreshCache() bool {
	// Get latest 3 scores
	// TODO Check if they're max
	result, err := rc.rdb.ZRangeWithScores(rc.ctx, "zset1", -3, -1).Result()
	if err != nil {
		log.Fatalf("Redis Cache Error %#v", err)
		return false
	}

	// For most hitted images
	for _, z := range result {
		_ = z
		// Hashed ID
		// image_id := z.Member.(string)

		// Check that's not in Cache
		// result2, err := r.ZScan(ctx, "zset1", 0, "", 1).Result()

		// buff, err := os.ReadFile(getImagePath(fileMap[filename]))

	}

	log.Printf("Refreshed Redis File Cache: %#v", result)
	return false
}
