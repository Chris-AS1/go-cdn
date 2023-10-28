package database

import (
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresConnectionViaConsul(t *testing.T) {
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
