package database

import (
	"context"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/discovery"
	"go-cdn/internal/tracing"
	"go-cdn/pkg/model"
	"go-cdn/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type RedisRepository struct {
	ctx    context.Context
	client *redis.Client
}

func NewRedisRepository(csl *discovery.Controller, cfg *config.Config) (*RedisRepository, error) {
	rc := &RedisRepository{
		ctx: context.Background(),
	}
	err := rc.connect(csl, cfg)
	return rc, err
}

func (rc *RedisRepository) GetConnectionString(csl *discovery.Controller, cfg *config.Config) (string, error) {
	var err error
	var address string
	var port int
	if cfg.Consul.ConsulEnable {
		// Discovers Redis from Consul
		address, err = csl.DiscoverService(cfg.Redis.RedisAddress)
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

func (rc *RedisRepository) connect(csl *discovery.Controller, cfg *config.Config) error {
	address, err := rc.GetConnectionString(csl, cfg)
	if err != nil {
		return err
	}

	rc.client = redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     cfg.Redis.RedisPassword,
		DB:           cfg.Redis.RedisDB,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	_, err = rc.client.Ping(rc.ctx).Result()
	return err
}

func (rc *RedisRepository) GetFile(ctx context.Context, id_hash string) (*model.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "rdGetFromCache")
	span.SetAttributes(attribute.String("rd.hash", id_hash))
	defer span.End()

	bytes, err := rc.client.Get(rc.ctx, id_hash).Bytes()

	// Documentation at https://redis.uptrace.dev/guide/go-redis.html#redis-nil
	switch {
	case err == redis.Nil:
		return nil, utils.ErrorRedisKeyDoesNotExist
	case err != nil:
		return nil, err
	}

	return &model.StoredFile{IDHash: id_hash, Filename: "", Content: bytes}, nil
}
func (rc *RedisRepository) GetFileList(ctx context.Context) (*[]model.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "rdGetFromCache")
    defer span.End()
	return nil, fmt.Errorf("not implemented")
}

func (rc *RedisRepository) AddFile(ctx context.Context, file *model.StoredFile) error {
	_, span := tracing.Tracer.Start(ctx, "rdAddToCache")
	span.SetAttributes(attribute.String("rd.hash", file.IDHash))
	defer span.End()

	_, err := rc.client.Set(rc.ctx, file.IDHash, file.Content, 0).Result()
	return err
}

func (rc *RedisRepository) RemoveFile(ctx context.Context, id_hash string) error {
	_, span := tracing.Tracer.Start(ctx, "rdRemoveFromCache")
	span.SetAttributes(attribute.String("rd.hash", id_hash))
	defer span.End()

	_, err := rc.client.Del(rc.ctx, id_hash).Result()
	return err
}

func (rc *RedisRepository) CloseConnection() error {
	return rc.client.Close()
}
