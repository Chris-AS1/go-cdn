package redis

import (
	"context"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/discovery/controller"
	"go-cdn/internal/tracing"
	"go-cdn/pkg/model"
	"time"

	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type RedisRepository struct {
	ctx    context.Context
	client *redis.Client
}

func New(ctx context.Context, dc *discovery.Controller, cfg *config.Config) (*RedisRepository, error) {
	_, span := tracing.Tracer.Start(ctx, "rd/New")
	defer span.End()

	rc := &RedisRepository{
		ctx: context.Background(),
	}
	err := rc.connect(dc, cfg)
	return rc, err
}

func (rc *RedisRepository) connect(dc *discovery.Controller, cfg *config.Config) error {
	address, err := rc.GetConnectionString(dc, cfg)
	if err != nil {
		return err
	}

	rc.client = redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     cfg.Cache.RedisPassword,
		DB:           cfg.Cache.RedisDB,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	_, err = rc.client.Ping(rc.ctx).Result()
	return err
}

func (rc *RedisRepository) CloseConnection() error {
	return rc.client.Close()
}

func (rc *RedisRepository) GetConnectionString(dc *discovery.Controller, cfg *config.Config) (string, error) {
	address, err := dc.DiscoverService(cfg.Cache.RedisAddress)
	if err != nil {
		return "", err
	}

	return address, nil
}

func (rc *RedisRepository) GetFile(ctx context.Context, id_hash string) (*model.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "rd/GetFile")
	span.SetAttributes(attribute.String("rd.hash", id_hash))
	defer span.End()

	bytes, err := rc.client.Get(rc.ctx, id_hash).Bytes()

	// Documentation at https://redis.uptrace.dev/guide/go-redis.html#redis-nil
	switch {
	case err == redis.Nil:
		return nil, repository.ErrKeyDoesNotExist
	case err != nil:
		return nil, err
	}

	return &model.StoredFile{IDHash: id_hash, Filename: "", Content: bytes}, nil
}
func (rc *RedisRepository) GetFileList(ctx context.Context) (*[]model.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "rd/GetFileList")
	defer span.End()
	return nil, fmt.Errorf("not implemented")
}

func (rc *RedisRepository) AddFile(ctx context.Context, file *model.StoredFile) error {
	_, span := tracing.Tracer.Start(ctx, "rd/AddFile")
	span.SetAttributes(attribute.String("rd.hash", file.IDHash))
	defer span.End()

	_, err := rc.client.Set(rc.ctx, file.IDHash, file.Content, 0).Result()
	return err
}

func (rc *RedisRepository) RemoveFile(ctx context.Context, id_hash string) error {
	_, span := tracing.Tracer.Start(ctx, "rd/RemoveFile")
	span.SetAttributes(attribute.String("rd.hash", id_hash))
	defer span.End()

	_, err := rc.client.Del(rc.ctx, id_hash).Result()
	return err
}
