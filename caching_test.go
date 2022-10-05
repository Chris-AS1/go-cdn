package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	r := ConnectRedis()
	assert.Equal(t, r, "ping: PONG")
}
