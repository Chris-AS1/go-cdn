package database

import (
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresConnection(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	csl_client, err := consul.NewConsulClient(&cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = NewPostgresClient(csl_client, &cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestPostgresMigrations(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	csl_client, err := consul.NewConsulClient(&cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	pg_client, err := NewPostgresClient(csl_client, &cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = pg_client.MigrateDB()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}
