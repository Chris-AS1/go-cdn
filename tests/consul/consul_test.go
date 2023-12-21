package consul_test

import (
	"go-cdn/internal/config"
	"go-cdn/internal/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsul(t *testing.T) {
	var consul_client *consul.ConsulClient

	cfg, err := config.New()
	assert.Nil(t, err)

	t.Run("TestConsulConnection", func(t *testing.T) {
		consul_client, err = consul.NewConsulClient(&cfg)
		assert.Nil(t, err)
	})

	// Don't even attempt to run other tests if a client isn't returned
	if err != nil {
		return
	}

	t.Run("TestConsulRegistration", func(t *testing.T) {
		err := consul_client.RegisterService(&cfg)
		assert.Nil(t, err)
	})

	// Looks for itself after registering
	t.Run("TestConsulServiceDiscovery", func(t *testing.T) {
		address, port, err := consul_client.DiscoverService(cfg.Consul.ConsulServiceName)
		assert.Nil(t, err)
		assert.NotSame(t, address, "")
		assert.NotSame(t, port, 0)
	})

	t.Run("TestConsulDeregistration", func(t *testing.T) {
		err := consul_client.DeregisterService(&cfg)
		assert.Nil(t, err)
	})

}
