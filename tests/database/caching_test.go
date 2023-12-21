package database_test

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/consul"
	"go-cdn/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	var redis_client *database.RedisRepository
	ctx := context.Background()

	cfg, err := config.New()
	assert.Nil(t, err)

	consul_client, err := consul.NewConsulClient(cfg)
	assert.Nil(t, err)

	t.Run("TestRedisConnection", func(t *testing.T) {
		redis_client, err = database.NewRedisRepository(consul_client, cfg)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a connection failed
	if err != nil {
		return
	}

	t.Run("TestRedisAddToCache", func(t *testing.T) {
		err = redis_client.AddFile(ctx, "0001", []byte{00, 00, 00})
		assert.Nil(t, err)
	})

	t.Run("TestRedisGetFromCache", func(t *testing.T) {
		bytes, err := redis_client.GetFile(ctx, "0001")
		assert.Nil(t, err)
		assert.NotNil(t, bytes)
	})

	t.Run("TestRedisRemoveFromCache", func(t *testing.T) {
		_, err = redis_client.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
