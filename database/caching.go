package database

import (
	"context"
	"fmt"
	"go-cdn/config"
	"go-cdn/consul"
	"go-cdn/utils"
	"log"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v9"
)

type RedisClient struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisClient(csl *consul.ConsulClient, cfg *config.Config) (*RedisClient, error) {
	rc := &RedisClient{
		ctx: context.Background(),
	}
	err := rc.connect(csl, cfg)
	return rc, err
}

func (pg *RedisClient) GetConnectionString(csl *consul.ConsulClient, cfg *config.Config) (string, error) {
	var err error
	var address string
	var port int
	if cfg.Consul.ConsulEnable {
		// Discovers Redis from Consul
		address, port, err = csl.DiscoverService(cfg.Redis.RedisAddress)
		if err != nil {
			return "", err
		}
	} else {
		cfg_adr := strings.Split(cfg.Redis.RedisAddress, ":")
		if len(cfg_adr) != 2 {
			return "", fmt.Errorf("wrong address format")
		}
		address = cfg_adr[0]
		port, _ = strconv.Atoi(cfg_adr[1])
	}

	connStr := fmt.Sprintf("%s:%d", address, port)
	return connStr, nil

}

func (rc *RedisClient) connect(csl *consul.ConsulClient, cfg *config.Config) error {
	address, err := rc.GetConnectionString(csl, cfg)
	if err != nil {
		return err
	}

	rc.rdb = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: cfg.Redis.RedisPassword,
		DB:       cfg.Redis.RedisDB,
	})

	_, err = rc.rdb.Ping(rc.ctx).Result()
	return err
}

func (rc *RedisClient) GetFromCache(id_hash string) ([]byte, error) {
	result, err := rc.rdb.Get(rc.ctx, id_hash).Bytes()

	// Documentation at https://redis.uptrace.dev/guide/go-redis.html#redis-nil
	switch {
	case err == redis.Nil:
		return nil, utils.ErrorRedisKeyDoesNotExist
	case err != nil:
		return nil, err
	}

	return result, nil
}

func (rc *RedisClient) AddToCache(id_hash string, content []byte) (string, error) {
	result, err := rc.rdb.Set(rc.ctx, id_hash, content, 0).Result()
	return result, err
}

func (rc *RedisClient) RemoveFromCache(id_hash string) (int64, error) {
	result, err := rc.rdb.Del(rc.ctx, id_hash).Result()
	return result, err
}

// Records image access on Redis - Most used cache
func (rc *RedisClient) RecordAccess(file_id string) int64 {
	result, err := rc.rdb.ZIncrBy(rc.ctx, "zset1", 1, file_id).Result()
	if err != nil {
		log.Panic(err)
	}

	return int64(result)
}
