package database

import (
	"go-cdn/config"
	"go-cdn/consul"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresConnection(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	_, err = NewPostgresClient(csl_client, &cfg)
	assert.Nil(t, err)
}

func TestPostgresMigrations(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	pg_client, err := NewPostgresClient(csl_client, &cfg)
	assert.Nil(t, err)

	err = pg_client.MigrateDB()
	assert.Nil(t, err)
}

func TestAddFile(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	pg_client, err := NewPostgresClient(csl_client, &cfg)
	assert.Nil(t, err)

	err = pg_client.AddFile("0001", "test_file", []byte{00, 00, 00})
	assert.Nil(t, err)
}

func TestRemoveFile(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	pg_client, err := NewPostgresClient(csl_client, &cfg)
	assert.Nil(t, err)

	err = pg_client.RemoveFile("0001")
	assert.Nil(t, err)
}

func TestGetFileList(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	pg_client, err := NewPostgresClient(csl_client, &cfg)
	assert.Nil(t, err)

	available_files, err := pg_client.GetFileList()
	assert.Nil(t, err)
	assert.NotNil(t, available_files)
	log.Print(available_files)
}
