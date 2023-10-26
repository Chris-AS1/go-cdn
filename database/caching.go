package database

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
	ctx       = context.Background()
	rdb       *redis.Client
	rdb_bytes *redis.Client
)

func ConnectRedis() string {
	log.Print("Connecting to Redis...")
	rdb = redis.NewClient(&redis.Options{
		Addr:     utils.EnvSettings.RedisURL,
		Password: "", // No Password
		DB:       0,  // use default DB
	})

	rdb_bytes = redis.NewClient(&redis.Options{
		Addr:     utils.EnvSettings.RedisURL,
		Password: "", // No Password
		DB:       1,  // use DB 1
	})

	result, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	result, err = rdb_bytes.Ping(ctx).Result()
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
	result, err := rdb.ZIncrBy(ctx, "zset1", 1, file_id).Result()
	if err != nil {
		log.Panic(err)
	}

	return int64(result)
}

func getFromCache(file_id string) (bool, []byte) {
	result, err := rdb_bytes.Get(ctx, file_id).Result()

	// "" If empty or nil (not error)
	if len(result) == 0 || err == redis.Nil {
		log.Printf("[CACHE] Not found, adding it now [%s]", file_id)
		buff, err := os.ReadFile(getImagePath(fileMap[file_id]))

		if err != nil {
			log.Print(err)
			return false, nil
		}

		_, err = rdb_bytes.Set(ctx, file_id, string(buff), 0).Result()

		if err != nil {
			log.Print(err)
		}

		return false, nil
	} else {
		return true, []byte(result)
	}
}

// Every X amount, check that DB 0 (HitN: Filenames) and DB 1 (Filenames: Bytes) are in sync
func refreshCache() bool {
	// Get latest 3 scores
	// TODO Check if they're max
	result, err := rdb.ZRangeWithScores(ctx, "zset1", -3, -1).Result()
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
