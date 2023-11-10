package consul

import (
	"go-cdn/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsul(t *testing.T) {
	var consul_client *ConsulClient

	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	t.Run("TestConsulConnection", func(t *testing.T) {
		consul_client, err = NewConsulClient(&cfg)
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

	t.Run("TestConsulDeregistration", func(t *testing.T) {
		err := consul_client.DeregisterService(&cfg)
		assert.Nil(t, err)
	})

	t.Run("TestConsulServiceDiscovery", func(t *testing.T) {
		address, port, err := consul_client.DiscoverService(cfg.DatabaseProvider.DatabaseAddress)
		assert.Nil(t, err)
		assert.NotSame(t, address, "")
		assert.NotSame(t, port, 0)
	})
}
