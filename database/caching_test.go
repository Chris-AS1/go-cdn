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

func TestRedisGetFromCache(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	rd_client, err := NewRedisClient(csl_client, &cfg)
	assert.Nil(t, err)

	bytes, err := rd_client.GetFromCache("000")
	assert.Nil(t, err)
	assert.NotNil(t, bytes)
}

func TestRedisAddToCache(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	rd_client, err := NewRedisClient(csl_client, &cfg)
	assert.Nil(t, err)

	bytes, err := rd_client.AddToCache("000", []byte{00, 00, 00})
	assert.Nil(t, err)
	assert.NotNil(t, bytes)
}

func TestRedisRemoveFromCache(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	csl_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	rd_client, err := NewRedisClient(csl_client, &cfg)
	assert.Nil(t, err)

	_, err = rd_client.RemoveFromCache("000")
	assert.Nil(t, err)
}
