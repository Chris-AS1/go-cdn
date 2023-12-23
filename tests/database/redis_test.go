package database

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/database/controller"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/database/repository/redis"
	"go-cdn/internal/discovery/controller"
	"go-cdn/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	var cache *database.Controller
	ctx := context.Background()

	cfg, err := config.New()
	assert.Nil(t, err)

	dcb, err := discovery.NewControllerBuilder().FromConfigs(cfg)
	assert.Nil(t, err)
	dc := dcb.Build()

	t.Run("TestConnection", func(t *testing.T) {
		redis_repo, err := redis.New(dc, cfg)
		cache = database.New(redis_repo)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a connection failed
	if err != nil {
		return
	}

	t.Run("TestAddFile", func(t *testing.T) {
		test_file := &model.StoredFile{
			IDHash:   "0001",
			Filename: "test",
			Content:  []byte{00, 10, 20},
		}
		err = cache.AddFile(ctx, test_file)
		assert.Nil(t, err)
	})

	t.Run("TestGetFile", func(t *testing.T) {
		stored_test_file, err := cache.GetFile(ctx, "0001")
		assert.Nil(t, err)
		assert.Equal(t, "0001", stored_test_file.IDHash)
		// filename is not stored
		assert.NotNil(t, stored_test_file.Content)
	})

	// Fetch a nonexistent file
	t.Run("TestGetFileNotFound", func(t *testing.T) {
		_, err := cache.GetFile(ctx, "0002")
		assert.ErrorIs(t, err, repository.ErrKeyDoesNotExist)
	})

	t.Run("TestRemoveFile", func(t *testing.T) {
		err = cache.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
