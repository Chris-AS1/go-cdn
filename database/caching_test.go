package database

import (
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	csl_client, err := consul.NewConsulClient(&cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = NewRedisClient(csl_client, &cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}
