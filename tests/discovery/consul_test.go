package discovery

import (
	"github.com/stretchr/testify/assert"
	"go-cdn/internal/config"
	"go-cdn/internal/discovery/controller"
	"go-cdn/internal/discovery/repository/consul"
	"go-cdn/internal/discovery/repository/dummy"
	"strings"
	"testing"
)

// TODO port to testcontainers
func TestConsul(t *testing.T) {
	cfg, err := config.New()
	assert.Nil(t, err)

	consul_repo, err := consul.NewConsulRepo(
		cfg.GetConsulConfig(),
		cfg.GetConsulServiceDefinition(),
	)
	assert.Nil(t, err)

	dc := discovery.NewController(consul_repo)

	// Don't attempt to run other tests if a client isn't returned
	if err != nil {
		return
	}

	t.Run("TestRegistration", func(t *testing.T) {
		err := dc.RegisterService()
		assert.Nil(t, err)
	})

	// Looks for itself after registering
	t.Run("TestServiceDiscovery", func(t *testing.T) {
		full_address, err := dc.DiscoverService(cfg.Consul.ConsulServiceName)
		assert.Nil(t, err)

		spl_full_address := strings.Split(full_address, ":")

		address, port := spl_full_address[0], spl_full_address[1]
		assert.Nil(t, err)
		assert.NotEqual(t, address, "")
		assert.NotEqual(t, port, 0)
	})

	t.Run("TestDeregistration", func(t *testing.T) {
		err := dc.DeregisterService()
		assert.Nil(t, err)
	})
}

func TestDummy(t *testing.T) {
	dc := discovery.NewController(dummy.NewDummyRepo())

	t.Run("TestServiceDiscovery", func(t *testing.T) {
		full_address, err := dc.DiscoverService("localhost:1234")
		assert.Nil(t, err)

		spl_full_address := strings.Split(full_address, ":")

		address, port := spl_full_address[0], spl_full_address[1]
		assert.Nil(t, err)
		assert.Equal(t, "localhost", address)
		assert.Equal(t, "1234", port)
	})
}
