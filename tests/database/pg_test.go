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

func TestPostgres(t *testing.T) {
	var db *database.Controller
	ctx := context.Background()

	cfg, err := config.New()
	assert.Nil(t, err)

	dc, err := discovery.BuildControllerFromConfigs(cfg)
	assert.Nil(t, err)

	t.Run("TestPostgresConnection", func(t *testing.T) {
		pg_repo, err := database.NewPostgresRepository(dc, cfg)
		db = database.NewController(pg_repo)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a connection failed
	if err != nil {
		return
	}

	t.Run("TestPostgresAddFile", func(t *testing.T) {
		test_file := &model.StoredFile{
			IDHash:   "0001",
			Filename: "test",
			Content:  []byte{00, 10, 20},
		}
		err = db.AddFile(ctx, test_file)
		assert.Nil(t, err)
	})

	t.Run("TestPostgresGetFile", func(t *testing.T) {
		_, err = db.GetFile(ctx, "0001")
		assert.Nil(t, err)
	})

    // Fetch a nonexistent file
	t.Run("TestPostgresGetFile2", func(t *testing.T) {
		_, err = db.GetFile(ctx, "0002")
		assert.Nil(t, err)
	})

	t.Run("TestPostgresRemoveFile", func(t *testing.T) {
		err = db.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
