package database

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/database"
	"go-cdn/internal/discovery"
	"go-cdn/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	var cache *database.Controller
	ctx := context.Background()

	cfg, err := config.New()
	assert.Nil(t, err)

	dc, err := discovery.BuildControllerFromConfigs(cfg)
	assert.Nil(t, err)

	t.Run("TestRedisConnection", func(t *testing.T) {
		redis_repo, err := database.NewRedisRepository(dc, cfg)
		cache = database.NewController(redis_repo)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a connection failed
	if err != nil {
		return
	}

	t.Run("TestRedisAddToCache", func(t *testing.T) {
		test_file := &model.StoredFile{
			IDHash:   "0001",
			Filename: "test",
			Content:  []byte{00, 10, 20},
		}
		err = cache.AddFile(ctx, test_file)
		assert.Nil(t, err)
	})

	t.Run("TestRedisGetFromCache", func(t *testing.T) {
		bytes, err := cache.GetFile(ctx, "0001")
		assert.Nil(t, err)
		assert.NotNil(t, bytes)
	})

	t.Run("TestRedisRemoveFromCache", func(t *testing.T) {
		err = cache.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
