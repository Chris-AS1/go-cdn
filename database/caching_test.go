package database

import (
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	_, err = NewRedisClient(csl_client, &cfg)
	assert.Nil(t, err)
}
