package database

import (
	"go-cdn/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = NewRedisClient(&cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}
