package database

import (
	"context"
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgres(t *testing.T) {
	var postgres_client *PostgresClient
	ctx := context.Background()

	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	consul_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	t.Run("TestPostgresConnection", func(t *testing.T) {
		postgres_client, err = NewPostgresClient(consul_client, &cfg)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a connection failed
	if err != nil {
		return
	}

	t.Run("TestPostgresMigrations", func(t *testing.T) {
		err = postgres_client.MigrateDB()
		assert.Nil(t, err)
	})

	t.Run("TestPostgresAddFile", func(t *testing.T) {
		err = postgres_client.AddFile(ctx, "0001", "test_file", []byte{00, 00, 00})
		assert.Nil(t, err)
	})

	t.Run("TestPostgresGetFile", func(t *testing.T) {
		_, err = postgres_client.GetFile(ctx, "0001")
		assert.Nil(t, err)
	})

	t.Run("TestPostgresRemoveFile", func(t *testing.T) {
		err = postgres_client.RemoveFile(ctx, "0001")
		assert.Nil(t, err)
	})
}
