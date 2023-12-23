package database

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/database/controller"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/database/repository/postgres"
	"go-cdn/internal/discovery/controller"
	"go-cdn/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgres(t *testing.T) {
	var db *database.Controller
	ctx := context.Background()

	cfg, err := config.New()
	assert.Nil(t, err)

	dcb, err := discovery.NewControllerBuilder().FromConfigs(cfg)
	assert.Nil(t, err)
	dc := dcb.Build()

	t.Run("TestConnection", func(t *testing.T) {
		pg_repo, err := postgres.NewPostgresRepository(dc, cfg)
		db = database.NewController(pg_repo)
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
		err = db.AddFile(ctx, test_file)
		assert.Nil(t, err)
	})

	t.Run("TestGetFile", func(t *testing.T) {
		stored_test_file, err := db.GetFile(ctx, "0001")
		assert.Nil(t, err)
		assert.Equal(t, "0001", stored_test_file.IDHash)
		assert.Equal(t, "test", stored_test_file.Filename)
		assert.NotNil(t, stored_test_file.Content)
	})

	// Fetch a nonexistent file
	t.Run("TestGetFileNotFound", func(t *testing.T) {
		_, err = db.GetFile(ctx, "0002")
		assert.ErrorIs(t, err, repository.ErrKeyDoesNotExist)
	})

	t.Run("TestRemoveFile", func(t *testing.T) {
		err = db.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
